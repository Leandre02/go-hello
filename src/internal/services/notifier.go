/* Service d'alertes pour le monitoring
 * Projet de session A25
 * By : Leandre Kanmegne
 * 
 * Gère la création et l'envoi des alertes basées sur les statuts
 * Définit les structures d'alertes et d'étiquettes
 * Permet de créer des alertes à partir des statuts de moniteurs
 * 
 * Sources:
 * https://prometheus.io/docs/alerting/latest/overview/ - concepts d'alerting
 */

package services

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"example.com/go-hello/src/internal/models"
)

// Représente les labels d'une alerte
type Etiquettes map[string]string

// Représente un événement d'alerte
type Alerte struct {
	Etiquettes  Etiquettes `json:"labels"`
	Annotations Etiquettes `json:"annotations"`
	Debut       time.Time  `json:"startsAt,omitempty"`
	Fin         time.Time  `json:"endsAt,omitempty"`
	SourceURL   string     `json:"generatorURL,omitempty"`
}

// Crée une alerte depuis un statut
func NouvelleAlerte(statut models.StatutMoniteur, baseURL string) *Alerte {
	if statut.EstDisponible {
		return nil
	}

	niveau := determinerGravite(statut.CodeStatutHTTP)

	return &Alerte{
		Etiquettes: Etiquettes{
			"alertname": "ServiceIndisponible",
			"service":   statut.URL,
			"niveau":    niveau,
			"code_http": fmt.Sprintf("%d", statut.CodeStatutHTTP),
		},
		Annotations: Etiquettes{
			"description": creerDescription(statut),
			"resume":      fmt.Sprintf("Service %s en panne", statut.URL),
		},
		Debut:     statut.VerifieA,
		SourceURL: fmt.Sprintf("%s/moniteurs/%s", baseURL, statut.URL),
	}
}

// Retourne le niveau de gravité
func determinerGravite(codeHTTP int) string {
	switch {
	case codeHTTP >= 500:
		return "critique"
	case codeHTTP >= 400:
		return "avertissement"
	case codeHTTP == 0:
		return "critique"
	default:
		return "info"
	}
}

// Crée une description 
func creerDescription(statut models.StatutMoniteur) string {
	if statut.MessageErreur != "" {
		return fmt.Sprintf("Le service %s a échoué : %s (Code %d, Latence %v)",
			statut.URL, statut.MessageErreur, statut.CodeStatutHTTP, statut.Latence)
	}
	return fmt.Sprintf("Le service %s a retourné le code %d (Latence %v)",
		statut.URL, statut.CodeStatutHTTP, statut.Latence)
}

// Get récupère une valeur d'étiquette
func (e Etiquettes) Get(cle string) string {
	return e[cle]
}

// Set définit une étiquette
func (e Etiquettes) Set(cle, valeur string) {
	e[cle] = valeur
}

// Hash calcule un hash unique des étiquettes
func (e Etiquettes) Hash() uint64 {
	if len(e) == 0 {
		return 0
	}
	
	cles := make([]string, 0, len(e))
	for k := range e {
		cles = append(cles, k)
	}
	sort.Strings(cles)
	
	parties := make([]string, 0, len(cles))
	for _, k := range cles {
		parties = append(parties, k+"="+e[k])
	}
	
	donnees := []byte(strings.Join(parties, ","))
	somme := sha256.Sum256(donnees)
	return binary.BigEndian.Uint64(somme[:8])
}

// Retourne le nom de l'alerte
func (a *Alerte) Nom() string {
	return a.Etiquettes.Get("alertname")
}

// Retourne le hash de l'alerte
func (a *Alerte) Hash() uint64 {
	return a.Etiquettes.Hash()
}

// Check si l'alerte est résolue
func (a *Alerte) EstResolue() bool {
	return a.EstResolueA(time.Now())
}

// Check si l'alerte est résolue à une date donnée
func (a *Alerte) EstResolueA(date time.Time) bool {
	if a.Fin.IsZero() {
		return false
	}
	return !a.Fin.After(date)
}

// Retourne une représentation texte
func (a *Alerte) String() string {
	base := fmt.Sprintf("%s[%s]", a.Nom(), fmt.Sprintf("%016x", a.Hash())[:7])
	if a.EstResolue() {
		return base + "[résolue]"
	}
	return base + "[active]"
}

// Gère l'envoi des notifications
type ServiceNotifications struct {
	urlBase string
}

// Crée le service de notifications
func NouvelleServiceNotifications(urlBase string) *ServiceNotifications {
	return &ServiceNotifications{urlBase: urlBase}
}

// Envoie une notification
func (sn *ServiceNotifications) Notifier(ctx context.Context, statut models.StatutMoniteur, gravite string, motif string) error {
	if statut.EstDisponible {
		log.Printf("[ALERTE][%s] %s — %d ms (URL=%s, code=%d)\n",
			gravite, motif, statut.Latence.Milliseconds(), statut.URL, statut.CodeStatutHTTP)
	} else {
		log.Printf("[ALERTE][%s] %s — ERREUR=\"%s\" (URL=%s)\n",
			gravite, motif, statut.MessageErreur, statut.URL)
	}
	return nil
}