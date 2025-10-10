/* Schéma PostgreSQL pour le service de monitoring */

-- Schéma PostgreSQL pour le service de monitoring (MVP)
-- À exécuter avant le lancement de l'app.

-- Schéma optionnel
CREATE SCHEMA IF NOT EXISTS monitoring;

-- 1) Cibles à surveiller
CREATE TABLE IF NOT EXISTS monitoring.moniteurs (
  id           BIGSERIAL PRIMARY KEY,
  url          TEXT NOT NULL UNIQUE,         -- ex: https://exemple.com
  actif        BOOLEAN NOT NULL DEFAULT TRUE,
  cree_a       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 2) Résultats des sondes (historique)
CREATE TABLE IF NOT EXISTS monitoring.statuts (
  id              BIGSERIAL PRIMARY KEY,
  moniteur_id     BIGINT NOT NULL REFERENCES monitoring.moniteurs(id) ON DELETE CASCADE,
  code_http       INTEGER,                   -- ex: 200, 301, 500 (NULL si échec avant HTTP)
  latence_ms      INTEGER,                   -- durée totale en ms (NULL si pas de réponse)
  est_disponible  BOOLEAN NOT NULL,          -- true/false (décidé par ton app)
  verifie_a       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  erreur          TEXT                       -- message d’erreur brut si échec
);

-- Index pour récupérer vite le dernier statut par moniteur
CREATE INDEX IF NOT EXISTS idx_statuts_moniteur_date
  ON monitoring.statuts (moniteur_id, verifie_a DESC);

-- 3) Vue pratique: dernier statut par moniteur
CREATE OR REPLACE VIEW monitoring.v_dernier_statut AS
SELECT DISTINCT ON (s.moniteur_id)
  s.moniteur_id,
  s.code_http,
  s.latence_ms,
  s.est_disponible,
  s.verifie_a,
  m.url
FROM monitoring.statuts s
JOIN monitoring.moniteurs m ON m.id = s.moniteur_id
ORDER BY s.moniteur_id, s.verifie_a DESC;
