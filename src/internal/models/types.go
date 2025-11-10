/* Structures de données pour le monitoring
 * Projet de session A25
 * By : Leandre Kanmegne
 * Date : 01-10-2025
 * 
 * Définit les modèles Moniteur et StatutMoniteur utilisés partout dans l'app
 */
package models

import "time"

// Moniteur représente un service à surveiller
type Moniteur struct {
	ID   int    `json:"id"`
	Nom  string `json:"nom"`
	URL  string `json:"url"`
	Type string `json:"type"` // http, https, tcp
}

// StatutMoniteur représente le résultat d'une vérification
type StatutMoniteur struct {
	MoniteurID     int           `json:"moniteur_id"`
	EstDisponible  bool          `json:"est_disponible"`
	MessageErreur  string        `json:"message_erreur"`
	CodeStatutHTTP int           `json:"code_statut_http"`
	VerifieA       time.Time     `json:"verifie_a"`
	URL            string        `json:"url"`
	Latence        time.Duration `json:"latence"`
}

// NouveauStatutMoniteur crée un nouveau statut
func NouveauStatutMoniteur(moniteurID int, url string, estDisponible bool, messageErreur string, codeStatutHTTP int, latence time.Duration) StatutMoniteur {
	return StatutMoniteur{
		MoniteurID:     moniteurID,
		EstDisponible:  estDisponible,
		MessageErreur:  messageErreur,
		CodeStatutHTTP: codeStatutHTTP,
		VerifieA:       time.Now(),
		URL:            url,
		Latence:        latence,
	}
}