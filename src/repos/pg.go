/*
 * Gestion de la persistance PostgreSQL pour le monitoring
 * Projet de session A25
 * Auteur : Leandre Kanmegne
 * Version : 1.0
 *
 * Ce module implémente l'interface repos.Repo via une base PostgreSQL,
 * utilisant le driver pgx/stdlib pour de meilleures performances.
 *
 * Il gère la connexion avec pool, le CRUD des moniteurs et l'enregistrement
 * des statuts, ainsi que la gestion des valeurs NULL.
 *
 * Basé sur les meilleures pratiques Go + PGX 2025 :
 * - driver pgx pour performance et support natif Postgres
 * - gestion explicite du pool de connexions
 * - utilisation précise de sql.Null types pour NULL en base
 * - context.Context dans toutes les méthodes pour timeout et annulation
 *
 * Sources:
 * https://dev.to/mx_tech/go-with-postgresql-best-practices-for-performance-and-safety-47d7 [web:73]
 * https://betterstack.com/community/guides/scaling-go/postgresql-pgx-golang/ [web:74]
 */

package repos

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"example.com/go-hello/src/internal/models"
	_ "github.com/jackc/pgx/v5/stdlib" // Import du driver pgx pour database/sql
)

// Postgres implémente repos.Repo avec une connexion PostgreSQL.
type Postgres struct {
    db *sql.DB
}

// NouvelleConnexion créé une connexion PostgreSQL basée sur DSN (ex: postgres://user:pass@localhost:5432/ma_db?sslmode=disable).
func NouvelleConnexion(dsn string) (*Postgres, error) {
    db, err := sql.Open("pgx", dsn)
    if err != nil {
        return nil, err
    }

    // Configuration du pool de connexions
    db.SetMaxOpenConns(15)
    db.SetMaxIdleConns(15)
    db.SetConnMaxLifetime(30 * time.Minute)

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    // Vérification de la connexion
    if err := db.PingContext(ctx); err != nil {
        _ = db.Close()
        return nil, err
    }

    return &Postgres{db: db}, nil
}

// Fermer ferme proprement la connexion à la base.
func (p *Postgres) Fermer() error {
    return p.db.Close()
}

// AjouterMoniteur ajoute ou met à jour un moniteur grâce à un upsert sur l'URL.
// Retourne uniquement une erreur pour rester compatible avec l'interface repos.Repo.
func (p *Postgres) AjouterMoniteur(ctx context.Context, m models.Moniteur) error {
    if m.URL == "" {
        return errors.New("l'URL du moniteur est obligatoire")
    }
    if m.Type == "" {
        m.Type = "http"
    }
    const requete = `
        INSERT INTO monitoring.moniteurs (nom, url, type)
        VALUES ($1, $2, $3)
        ON CONFLICT (url)
        DO UPDATE SET
            nom = COALESCE(NULLIF(EXCLUDED.nom, ''), monitoring.moniteurs.nom),
            type = COALESCE(NULLIF(EXCLUDED.type, ''), monitoring.moniteurs.type);
    `
    _, err := p.db.ExecContext(ctx, requete, m.Nom, m.URL, m.Type)
    return err
}

// SupprimerMoniteur supprime un moniteur via son identifiant.
func (p *Postgres) SupprimerMoniteur(ctx context.Context, url string) error {
    res, err := p.db.ExecContext(ctx, `DELETE FROM monitoring.moniteurs WHERE url=$1`, url)
    if err != nil {
        return err
    }
    if n, _ := res.RowsAffected(); n == 0 {
        return errors.New("moniteur introuvable")
    }
    return nil
}

// ListerMoniteurs retourne tous les moniteurs ordonnés par ID.
func (p *Postgres) ListerMoniteurs(ctx context.Context) ([]models.Moniteur, error) {
    rows, err := p.db.QueryContext(ctx, `SELECT id, nom, url, type FROM monitoring.moniteurs ORDER BY id ASC`)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var moniteurs []models.Moniteur
    for rows.Next() {
        var m models.Moniteur
        if err := rows.Scan(&m.ID, &m.Nom, &m.URL, &m.Type); err != nil {
            return nil, err
        }
        moniteurs = append(moniteurs, m)
    }

    return moniteurs, rows.Err()
}

// EnregistrerStatutMoniteur insère un statut de moniteur en base.
func (p *Postgres) EnregistrerStatutMoniteur(ctx context.Context, s models.StatutMoniteur) error {
    if s.URL == "" {
        return errors.New("l'URL est obligatoire pour un statut")
    }
    if s.VerifieA.IsZero() {
        s.VerifieA = time.Now()
    }
    const requete = `
        INSERT INTO monitoring.statuts (moniteur_id, url, est_disponible, code_http, message_erreur, latence_ms, verifie_a)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
    `
    // Convertir MoniteurID (int) en int64 pour la DB, ou NULL si 0
    var monID any
    if s.MoniteurID == 0 {
        monID = nil
    } else {
        monID = int64(s.MoniteurID)
    }
    _, err := p.db.ExecContext(ctx, requete,
        monID, s.URL, s.EstDisponible, s.CodeStatutHTTP,
        valeurNullString(s.MessageErreur), s.Latence.Milliseconds(), s.VerifieA,
    )
    return err
}

// DerniersStatutsMoniteur récupère les derniers statuts (mis à jour en limitation côté appelant).
func (p *Postgres) DerniersStatutsMoniteur(ctx context.Context, moniteurID int) ([]models.StatutMoniteur, error) {
    const requete = `
        SELECT moniteur_id, url, est_disponible, code_http, message_erreur, latence_ms, verifie_a
        FROM monitoring.statuts
        WHERE moniteur_id = $1
        ORDER BY verifie_a DESC
    `
    rows, err := p.db.QueryContext(ctx, requete, moniteurID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var statuts []models.StatutMoniteur
    for rows.Next() {
        var s models.StatutMoniteur
        var monID sql.NullInt64
        var msg sql.NullString
        var latMs sql.NullInt64
        if err := rows.Scan(&monID, &s.URL, &s.EstDisponible, &s.CodeStatutHTTP, &msg, &latMs, &s.VerifieA); err != nil {
            return nil, err
        }
        if monID.Valid {
            s.MoniteurID = int(monID.Int64)
        }
        if msg.Valid {
            s.MessageErreur = msg.String
        }
        if latMs.Valid {
            s.Latence = time.Duration(latMs.Int64) * time.Millisecond
        }
        statuts = append(statuts, s)
    }
    return statuts, rows.Err()
}

// Helpers pour gérer valeurs NULL PostgreSQL
func valeurNullString(v string) any {
    if v == "" {
        return nil
    }
    return v
}

// ViderTout supprime toutes les données (statuts puis moniteurs) et réinitialise les séquences.
func (p *Postgres) ViderTout(ctx context.Context) error {
    // Utiliser TRUNCATE sur les tables du schéma monitoring
    _, err := p.db.ExecContext(ctx, `
        TRUNCATE TABLE monitoring.statuts, monitoring.moniteurs RESTART IDENTITY CASCADE;
    `)
    return err
}

