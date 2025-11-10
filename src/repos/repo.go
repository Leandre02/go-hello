/* Interface pour les opérations de persistance
 * Projet de session A25
 * By : Leandre Kanmegne
 * Date : 01-10-2025
 *
 * Utilise le contexte d'exécution pour les opérations de la base de données
 * Définit les opérations pour accéder aux données (moniteurs et statuts)
 */
package repos

import (
	"context"

	"example.com/go-hello/src/internal/models"
)

// Définit les opérations de base pour la persistance
type Repo interface {
	// gestion des moniteurs
	AjouterMoniteur(ctx context.Context, moniteur models.Moniteur) error
	ListerMoniteurs(ctx context.Context) ([]models.Moniteur, error)
	SupprimerMoniteur(ctx context.Context, url string) error

	// gestion des statuts
	EnregistrerStatutMoniteur(ctx context.Context, statut models.StatutMoniteur) error
	DerniersStatutsMoniteur(ctx context.Context, moniteurID int) ([]models.StatutMoniteur, error)

	// utilitaire admin
	ViderTout(ctx context.Context) error // supprime tous les moniteurs et statuts
}