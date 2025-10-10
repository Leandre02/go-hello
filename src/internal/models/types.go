/* Definit la structure des données utilisées dans l'application
* Projet de session A25
* By : Leandre Kanmegne
* Date : 01-10-2025
* Version : 1.0
* Commentaire : MVP initiale avec branchement sans postgreSQL et sans authentification. Stockage en local
 */
package models

import "time"

// Moniteur représente un service à surveiller
type Moniteur struct {
	ID  int    `json:"id"`  // Identifiant unique
	Nom string `json:"nom"` // Nom du service
	URL string `json:"url"` // Adresse du service 
	Type string `json:"type"` // Type de service (ex: "http", "https", "tcp")

}

// StatutMoniteur représente le statut d'un moniteur à un instant donné
type StatutMoniteur struct {
	MoniteurID       int           `json:"moniteur_id"`        // Identifiant du moniteur (si stocké en BD)
	EstDisponible    bool          `json:"est_disponible"`     // true si le service répond correctement
	MessageErreur    string        `json:"message_erreur"`     // Détail de l'erreur si indisponible
	CodeStatutHTTP   int           `json:"code_statut_http"`   // Code HTTP reçu (0 si échec réseau)
	VerifieA        time.Time     `json:"verifie_a"`          // Horodatage de la vérification
	URL              string        `json:"url"`                // Copie de l'URL pour traçabilité
	Latence          time.Duration `json:"latence"`            // Durée de la requête
}

// NouveauStatutMoniteur crée un nouveau statut de moniteur
func NouveauStatutMoniteur(moniteurID int, url string, estDisponible bool, messageErreur string, codeStatutHTTP int, latence time.Duration) StatutMoniteur {
	return StatutMoniteur{
		MoniteurID:     moniteurID,
		EstDisponible:  estDisponible,
		MessageErreur:  messageErreur,
		CodeStatutHTTP: codeStatutHTTP,
		VerifieA:      time.Now(),
		URL:            url,
		Latence:        latence,
	}	
}