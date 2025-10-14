/* Definit l'interface pour les opérations de gestion des dépôts
* Projet de session A25
* By : Leandre Kanmegne
* Date : 01-10-2025
* Version : 1.0
* Commentaire : MVP initiale avec branchement sans postgreSQL et sans authentification. Stockage en local
 */
package repos

import (
	"context"

	"example.com/go-hello/src/internal/models"
)

// Repo définit le contrat des opérations de persistance.
type Repo interface {
	// Moniteurs
	AjouterMoniteur(ctx context.Context, moniteur models.Moniteur) error                          // Ajouter un moniteur
	ListerMoniteurs(ctx context.Context) ([]models.Moniteur, error)                               // Lister tous les moniteurs
	SupprimerMoniteur(ctx context.Context, url string) error                                      // Supprimer un moniteur par URL

	// StatutMoniteur
	EnregistrerStatutMoniteur(ctx context.Context, statut models.StatutMoniteur) error               // Ajouter un statut de moniteur
	DerniersStatutsMoniteur(ctx context.Context, moniteurID int) ([]models.StatutMoniteur, error)  // Lister tous les statuts d'un moniteur

	// Administration
	ViderTout(ctx context.Context) error // Supprime toutes les données (moniteurs et statuts)
}