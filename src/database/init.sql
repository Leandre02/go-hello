/* Schéma PostgreSQL pour le service de monitoring */

-- Schéma PostgreSQL pour le service de monitoring (MVP)
-- À exécuter avant le lancement de l'app.

-- ============================================
--  init.sql  — Schéma, tables, index, vue
--  Exécuter AVANT dbtrigger.sql
-- ============================================

-- (optionnel mais pratique)
SET client_min_messages TO WARNING;

-- 1) Schéma
CREATE SCHEMA IF NOT EXISTS monitoring;

-- 2) Table des moniteurs
CREATE TABLE IF NOT EXISTS monitoring.moniteurs (
    id        BIGSERIAL PRIMARY KEY,
    url       TEXT        NOT NULL UNIQUE,
    actif     BOOLEAN     NOT NULL DEFAULT TRUE,
    cree_a    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 3) Table des statuts (historique des checks)
CREATE TABLE IF NOT EXISTS monitoring.statuts (
    id             BIGSERIAL PRIMARY KEY,
    moniteur_id    BIGINT     NOT NULL REFERENCES monitoring.moniteurs(id) ON DELETE CASCADE,
    code_http      INTEGER,
    est_disponible BOOLEAN    NOT NULL,
    message_erreur TEXT,
    latence_ms     INTEGER,       -- on stocke la latence en millisecondes
    verifie_a      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 4) Index utiles
CREATE INDEX IF NOT EXISTS idx_moniteurs_url           ON monitoring.moniteurs (url);
CREATE INDEX IF NOT EXISTS idx_statuts_moniteur_ts     ON monitoring.statuts (moniteur_id, verifie_a DESC);

-- 5) Vue: dernier statut par moniteur (pratique pour l’API)
--    Utilise DISTINCT ON pour ne garder que la ligne la plus récente par moniteur
CREATE OR REPLACE VIEW monitoring.v_dernier_statut AS
SELECT DISTINCT ON (s.moniteur_id)
       s.moniteur_id,
       s.code_http,
       s.est_disponible,
       s.message_erreur,
       s.latence_ms,
       s.verifie_a,
       m.url
FROM monitoring.statuts  AS s
JOIN monitoring.moniteurs AS m ON m.id = s.moniteur_id
ORDER BY s.moniteur_id, s.verifie_a DESC;

-- (optionnel) Définir le search_path pour cette session
-- SET search_path TO monitoring, public;
