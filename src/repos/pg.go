/* Implémentation PostgreSQL du repo
 * Projet de session A25
 * By : Leandre Kanmegne
 * 
 * Gère la connexion et les opérations CRUD avec PostgreSQL
 * Configure un pool de connexions pour la gestion des performances de la BD
 * Utilise pgx (qui est un pilote PostgreSQL pour Go) comme driver pour de meilleures performances
 * Utilise la fonction ExecContext de pgx pour exécuter automatiquement la requête preparée
 * Utilise la fonction QueryContext de pgx pour exécuter les requêtes de sélection
 * 
 * Sources:
 * https://dev.to/mx_tech/go-with-postgresql-best-practices-for-performance-and-safety-47d7
 * https://betterstack.com/community/guides/scaling-go/postgresql-pgx-golang/
 * https://pkg.go.dev/github.com/jackc/pgx/v5
 */

package repos

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"example.com/go-hello/src/internal/models"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// Postgres implémente Repo avec PostgreSQL
type Postgres struct {
	db *sql.DB
}

// Crée une connexion PostgreSQL
func NouvelleConnexion(dsn string) (*Postgres, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	// config du pool de connexions
	db.SetMaxOpenConns(15) // nombre max de connexions ouvertes
	db.SetMaxIdleConns(15)  // nombre max de connexions inactives
	db.SetConnMaxLifetime(30 * time.Minute) // durée max de vie d'une connexion

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// check si la connexion marche
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, err
	}

    // retourne l'instance du repo
	return &Postgres{db: db}, nil
}

// Ferme la connexion à la base
func (p *Postgres) Fermer() error {
	return p.db.Close()
}

// Ajoute un moniteur
func (p *Postgres) AjouterMoniteur(ctx context.Context, moniteur models.Moniteur) error {
	if moniteur.URL == "" {
		return errors.New("l'URL du moniteur est obligatoire")
	}
	if moniteur.Type == "" {
		moniteur.Type = "http"
	}

    // Requete preparée pour insertion des données et gestion des doublons sur l'URL
	requete := `
		INSERT INTO monitoring.moniteurs (nom, url, type)
		VALUES ($1, $2, $3)
		ON CONFLICT (url) 
		DO UPDATE SET
			nom = COALESCE(NULLIF(EXCLUDED.nom, ''), monitoring.moniteurs.nom),
			type = COALESCE(NULLIF(EXCLUDED.type, ''), monitoring.moniteurs.type)
	`
	_, err := p.db.ExecContext(ctx, requete, moniteur.Nom, moniteur.URL, moniteur.Type) 
	return err
}

// Supprime un moniteur par URL
func (p *Postgres) SupprimerMoniteur(ctx context.Context, url string) error {
	resultat, err := p.db.ExecContext(ctx, `DELETE FROM monitoring.moniteurs WHERE url=$1`, url)
	if err != nil {
		return err
	}
	
	lignesAffectees, _ := resultat.RowsAffected()
	if lignesAffectees == 0 {
		return errors.New("moniteur introuvable")
	}
	
	return nil
}

// Retourne tous les moniteurs
func (p *Postgres) ListerMoniteurs(ctx context.Context) ([]models.Moniteur, error) {
	rows, err := p.db.QueryContext(ctx, `SELECT id, nom, url, type FROM monitoring.moniteurs ORDER BY id ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var moniteurs []models.Moniteur
	for rows.Next() {
		var moniteur models.Moniteur
		if err := rows.Scan(&moniteur.ID, &moniteur.Nom, &moniteur.URL, &moniteur.Type); err != nil {
			return nil, err
		}
		moniteurs = append(moniteurs, moniteur)
	}

	return moniteurs, rows.Err()
}

// Enregistre un statut dans la BD
func (p *Postgres) EnregistrerStatutMoniteur(ctx context.Context, statut models.StatutMoniteur) error {
	if statut.URL == "" {
		return errors.New("l'URL est obligatoire pour un statut")
	}
	if statut.VerifieA.IsZero() {
		statut.VerifieA = time.Now()
	}

	requete := `
		INSERT INTO monitoring.statuts (moniteur_id, url, est_disponible, code_http, message_erreur, latence_ms, verifie_a)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	// convertit MoniteurID en int64 ou NULL si 0
	var moniteurID any
	if statut.MoniteurID == 0 {
		moniteurID = nil
	} else {
		moniteurID = int64(statut.MoniteurID)
	}

	_, err := p.db.ExecContext(ctx, requete,
		moniteurID, statut.URL, statut.EstDisponible, statut.CodeStatutHTTP,
		valeurNullString(statut.MessageErreur), statut.Latence.Milliseconds(), statut.VerifieA,
	)
	return err
}

// DerniersStatutsMoniteur récupère les derniers statuts d'un moniteur
func (p *Postgres) DerniersStatutsMoniteur(ctx context.Context, moniteurID int) ([]models.StatutMoniteur, error) {
	requete := `
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
		var statut models.StatutMoniteur
		var moniteurIDNull sql.NullInt64
		var messageNull sql.NullString
		var latenceMs sql.NullInt64

        // Scan des valeurs de la ligne courante, avec gestion des valeurs NULL
		if err := rows.Scan(&moniteurIDNull, &statut.URL, &statut.EstDisponible, &statut.CodeStatutHTTP, &messageNull, &latenceMs, &statut.VerifieA); err != nil {
			return nil, err
		}

		if moniteurIDNull.Valid {
			statut.MoniteurID = int(moniteurIDNull.Int64)
		}
		if messageNull.Valid {
			statut.MessageErreur = messageNull.String
		}
		if latenceMs.Valid {
			statut.Latence = time.Duration(latenceMs.Int64) * time.Millisecond
		}

		statuts = append(statuts, statut)
	}

	return statuts, rows.Err()
}

// Retourne nil si la string est vide
func valeurNullString(valeur string) any {
	if valeur == "" {
		return nil
	}
	return valeur
}

// Supprime toutes les données et reset les séquences
func (p *Postgres) ViderTout(ctx context.Context) error {
	_, err := p.db.ExecContext(ctx, `
		TRUNCATE TABLE monitoring.statuts, monitoring.moniteurs RESTART IDENTITY CASCADE
	`)
	return err
}