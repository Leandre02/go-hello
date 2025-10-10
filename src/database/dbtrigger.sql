/* Trigger PostgreSQL pour mettre à jour automatiquement la date de modification */

-- Trigger pour mettre à jour automatiquement la date_modification des moniteurs

-- Table d'alertes
CREATE TABLE IF NOT EXISTS monitoring.alertes (
  id           BIGSERIAL PRIMARY KEY,
  moniteur_id  BIGINT NOT NULL REFERENCES monitoring.moniteurs(id) ON DELETE CASCADE,
  type         TEXT NOT NULL,                -- ex: 'DOWN'
  details      TEXT,
  cree_a       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Fonction de trigger (PL/pgSQL, pas de DELIMITER ici)
CREATE OR REPLACE FUNCTION monitoring.fn_log_alerte_down()
RETURNS TRIGGER AS $$
DECLARE
  prev_ok BOOLEAN;
BEGIN
  -- Dernier statut avant l'insert
  SELECT est_disponible
    INTO prev_ok
  FROM monitoring.statuts
  WHERE moniteur_id = NEW.moniteur_id
  ORDER BY verifie_a DESC
  LIMIT 1
  OFFSET 0;

  -- Si précédent était TRUE et nouveau est FALSE => alerte
  IF prev_ok IS TRUE AND NEW.est_disponible IS FALSE THEN
    INSERT INTO monitoring.alertes (moniteur_id, type, details)
    VALUES (NEW.moniteur_id, 'DOWN',
            COALESCE('code_http='||NEW.code_http||', latence_ms='||NEW.latence_ms, ''));
  END IF;

  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Déclencheur après chaque insert de statut
DROP TRIGGER IF EXISTS trg_log_alerte_down ON monitoring.statuts;
CREATE TRIGGER trg_log_alerte_down
AFTER INSERT ON monitoring.statuts
FOR EACH ROW
EXECUTE FUNCTION monitoring.fn_log_alerte_down();
