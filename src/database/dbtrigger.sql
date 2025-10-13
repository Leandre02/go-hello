-- =======================================================
--  dbtrigger.sql — Table d'alertes + fonction + trigger
--  Exécuter APRES init.sql (tables déjà créées)
-- =======================================================

-- 1) Table des alertes (DOWN/UP)
CREATE TABLE IF NOT EXISTS monitoring.alertes (
    id          BIGSERIAL PRIMARY KEY,
    moniteur_id BIGINT     NOT NULL REFERENCES monitoring.moniteurs(id) ON DELETE CASCADE,
    type        TEXT       NOT NULL CHECK (type IN ('DOWN','UP')),
    details     TEXT,
    cree_a      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 2) Fonction de détection de transition d’état
--    Si dernier statut connu = UP (true) et nouveau = DOWN (false)  → alerte DOWN
--    Si dernier statut connu = DOWN (false) et nouveau = UP (true)  → alerte UP
CREATE OR REPLACE FUNCTION monitoring.fn_alert_on_transition()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
DECLARE
    prev_est_dispo BOOLEAN;
BEGIN
    -- Récupère l’état précédent (le plus récent avant NEW)
    SELECT s.est_disponible
      INTO prev_est_dispo
      FROM monitoring.statuts s
     WHERE s.moniteur_id = NEW.moniteur_id
       AND s.id <> NEW.id
     ORDER BY s.verifie_a DESC, s.id DESC
     LIMIT 1;

    -- Pas d’état précédent → pas d’alerte
    IF prev_est_dispo IS NULL THEN
        RETURN NEW;
    END IF;

    -- Transition UP → DOWN
    IF prev_est_dispo = TRUE AND NEW.est_disponible = FALSE THEN
        INSERT INTO monitoring.alertes (moniteur_id, type, details)
        VALUES (NEW.moniteur_id, 'DOWN',
                CONCAT('Indisponible; code_http=', COALESCE(NEW.code_http::text,'NULL'),
                       ', latence_ms=', COALESCE(NEW.latence_ms::text,'NULL')));
    END IF;

    -- Transition DOWN → UP
    IF prev_est_dispo = FALSE AND NEW.est_disponible = TRUE THEN
        INSERT INTO monitoring.alertes (moniteur_id, type, details)
        VALUES (NEW.moniteur_id, 'UP',
                CONCAT('De nouveau disponible; code_http=', COALESCE(NEW.code_http::text,'NULL'),
                       ', latence_ms=', COALESCE(NEW.latence_ms::text,'NULL')));
    END IF;

    RETURN NEW;
END;
$$;

-- 3) Trigger (idempotent: on drop puis on crée)
DROP TRIGGER IF EXISTS tr_statut_transition ON monitoring.statuts;

CREATE TRIGGER tr_statut_transition
AFTER INSERT ON monitoring.statuts
FOR EACH ROW
EXECUTE FUNCTION monitoring.fn_alert_on_transition();
