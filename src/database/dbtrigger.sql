-- Script de triggers pour alertes automatiques
-- Projet de session A25
-- By : Leandre Kanmegne
--
-- Crée une table d'alertes et un trigger qui détecte quand
-- un moniteur passe de UP à DOWN (ou l'inverse)
-- À exécuter APRÈS init.sql
--
-- Source:
-- https: / / www.postgresql.org / docs / current / sql - createtrigger.html


-- Table pour stocker les alertes
CREATE TABLE IF NOT EXISTS monitoring.alertes (
    id BIGSERIAL PRIMARY KEY,
    moniteur_id BIGINT NOT NULL REFERENCES monitoring.moniteurs(id) ON DELETE CASCADE,
    type TEXT NOT NULL CHECK (type IN ('DOWN', 'UP')),
    details TEXT,
    cree_a TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Détecte les changements d'état et insère une alerte
CREATE
OR REPLACE FUNCTION monitoring.detecter_transition() RETURNS TRIGGER AS $ $ DECLARE ancien_etat BOOLEAN;

BEGIN -- récupère l'état précédent
SELECT
    est_disponible INTO ancien_etat
FROM
    monitoring.statuts
WHERE
    moniteur_id = NEW.moniteur_id
    AND id <> NEW.id
ORDER BY
    verifie_a DESC
LIMIT
    1;

-- Vérifie si un état précédent existe
IF ancien_etat IS NULL THEN RETURN NEW;

END IF;

-- si passage de disponible à indisponible
IF ancien_etat = TRUE
AND NEW.est_disponible = FALSE THEN
INSERT INTO
    monitoring.alertes (moniteur_id, type, details)
VALUES
    (
        NEW.moniteur_id,
        'DOWN',
        'Indisponible - HTTP ' || COALESCE(NEW.code_http :: TEXT, 'erreur')
    );

END IF;

-- si passage de indisponible à disponible
IF ancien_etat = FALSE
AND NEW.est_disponible = TRUE THEN
INSERT INTO
    monitoring.alertes (moniteur_id, type, details)
VALUES
    (
        NEW.moniteur_id,
        'UP',
        'Rétabli - HTTP ' || COALESCE(NEW.code_http :: TEXT, '200')
    );

END IF;

RETURN NEW;

END;

$ $ LANGUAGE plpgsql;

-- Trigger qui s'exécute après chaque insertion de statut
DROP TRIGGER IF EXISTS trigger_alerte_statut ON monitoring.statuts;

CREATE TRIGGER trigger_alerte_statut
AFTER
INSERT
    ON monitoring.statuts FOR EACH ROW EXECUTE FUNCTION monitoring.detecter_transition();