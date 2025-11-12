/* Schéma PostgreSQL pour le service de monitoring
 * Projet de session A25
 * By : Leandre Kanmegne
 * 
 * Crée les tables, index et vues nécessaires au monitoring
 * À exécuter AVANT dbtrigger.sql
 */
SET
    client_min_messages TO WARNING;

-- création du schéma
CREATE SCHEMA IF NOT EXISTS monitoring;

-- table des moniteurs (services à surveiller)
CREATE TABLE IF NOT EXISTS monitoring.moniteurs (
    id BIGSERIAL PRIMARY KEY,
    nom TEXT NOT NULL,
    url TEXT NOT NULL UNIQUE,
    type TEXT NOT NULL DEFAULT 'http',
    actif BOOLEAN NOT NULL DEFAULT TRUE,
    cree_a TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- table des statuts (historique des vérifications)
CREATE TABLE IF NOT EXISTS monitoring.statuts (
    id BIGSERIAL PRIMARY KEY,
    moniteur_id BIGINT REFERENCES monitoring.moniteurs(id) ON DELETE CASCADE,
    url TEXT NOT NULL,
    code_http INTEGER,
    est_disponible BOOLEAN NOT NULL,
    message_erreur TEXT,
    latence_ms INTEGER,
    verifie_a TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- index pour les requêtes fréquentes
CREATE INDEX IF NOT EXISTS idx_moniteurs_url ON monitoring.moniteurs (url);

CREATE INDEX IF NOT EXISTS idx_statuts_moniteur_ts ON monitoring.statuts (moniteur_id, verifie_a DESC);

-- vue pour récupérer le dernier statut de chaque moniteur
CREATE
OR REPLACE VIEW monitoring.v_dernier_statut AS
SELECT
    DISTINCT ON (s.moniteur_id) s.moniteur_id,
    s.code_http,
    s.est_disponible,
    s.message_erreur,
    s.latence_ms,
    s.verifie_a,
    m.url,
    m.nom
FROM
    monitoring.statuts AS s
    JOIN monitoring.moniteurs AS m ON m.id = s.moniteur_id
ORDER BY
    s.moniteur_id,
    s.verifie_a DESC;