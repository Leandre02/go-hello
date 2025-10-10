/*
 * Service d'alertes pour le système de monitoring
 * Projet de session A25
 * Auteur : Leandre Kanmegne
 * Version : 1.0
 *
 * Ce fichier implémente la gestion et la création des alertes issues des statuts de
 * moniteur. Les alertes contiennent des labels, des annotations, des informations temporelles,
 * et des fonctions pour leur identification et comparaison.
 *
 * Le service d'alerte permet aussi de générer des notifications basiques (logging)
 * en fonction du niveau de sévérité (critique, avertissement, info).
 *
 * Ce module est conforme aux meilleures pratiques d'architecture de systèmes d'alerting
 * dans le monitoring IT et applique une abstraction propre et évolutive.[web:67][web:71]
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

// Etiquettes représente une collection clé-valeur attachée aux alertes
type Etiquettes map[string]string

// Alerte modélise un événement d'alerte dans le système
type Alerte struct {
    // Labels définissant l'identité et le regroupement de l'alerte,
    // dont "alertname" est obligatoire.
    Etiquettes Etiquettes `json:"labels"`

    // Annotations apportant des infos complémentaires (description, résumé, etc.)
    Annotations Etiquettes `json:"annotations"`

    // Plage de validité temporelle de l'alerte
    Debut     time.Time `json:"startsAt,omitempty"`
    Fin       time.Time `json:"endsAt,omitempty"`
    SourceURL string    `json:"generatorURL,omitempty"`
}

// NouvelleAlerte crée une instance d'alerte à partir d'un statut de moniteur
func NouvelleAlerte(statut models.StatutMoniteur, baseURL string) *Alerte {
    if statut.EstDisponible {
        return nil // Pas d'alerte si service OK
    }

    niveau := determinerGravite(statut.CodeStatutHTTP)

    return &Alerte{
        Etiquettes: Etiquettes{
            "alertname":  "ServiceIndisponible",
            "service":    statut.URL,
            "niveau":     niveau,
            "code_http":  fmt.Sprintf("%d", statut.CodeStatutHTTP),
        },
        Annotations: Etiquettes{
            "description": creerDescription(statut),
            "resume":      fmt.Sprintf("Service %s en panne", statut.URL),
        },
        Debut:     statut.VerifieA,
        SourceURL: fmt.Sprintf("%s/moniteurs/%s", baseURL, statut.URL),
    }
}

// determinerGravite retourne la gravité d'une alerte selon le code HTTP
func determinerGravite(codeHTTP int) string {
    switch {
    case codeHTTP >= 500:
        return "critique"
    case codeHTTP >= 400:
        return "avertissement"
    case codeHTTP == 0:
        return "critique" // Erreur de connexion réseau
    default:
        return "info"
    }
}

// creerDescription génère un texte descriptif de l'alerte à partir du statut
func creerDescription(statut models.StatutMoniteur) string {
    if statut.MessageErreur != "" {
        return fmt.Sprintf("Le service %s a échoué : %s (Code %d, Latence %v)",
            statut.URL, statut.MessageErreur, statut.CodeStatutHTTP, statut.Latence)
    }
    return fmt.Sprintf("Le service %s a retourné le code %d (Latence %v)",
        statut.URL, statut.CodeStatutHTTP, statut.Latence)
}

// Get récupère la valeur d'un label
func (e Etiquettes) Get(cle string) string {
    return e[cle]
}

// Set fixe la valeur d'un label
func (e Etiquettes) Set(cle, valeur string) {
    e[cle] = valeur
}

// Hash calcule un hash cohérent basé sur les paires clé-valeur des étiquettes
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

// Nom retourne le nom de l'alerte (label "alertname")
func (a *Alerte) Nom() string {
    return a.Etiquettes.Get("alertname")
}

// Hash retourne un hash unique basé sur les étiquettes de l'alerte
func (a *Alerte) Hash() uint64 {
    return a.Etiquettes.Hash()
}

// EstResolue indique si l'alerte est résolue selon la date actuelle
func (a *Alerte) EstResolue() bool {
    return a.EstResolueA(time.Now())
}

// EstResolueA indique si l'alerte est résolue à une date donnée
func (a *Alerte) EstResolueA(date time.Time) bool {
    if a.Fin.IsZero() {
        return false
    }
    return !a.Fin.After(date)
}

// String retourne une représentation textuelle et courte de l'alerte
func (a *Alerte) String() string {
    base := fmt.Sprintf("%s[%s]", a.Nom(), fmt.Sprintf("%016x", a.Hash())[:7])
    if a.EstResolue() {
        return base + "[résolue]"
    }
    return base + "[active]"
}

// ServiceNotifications gère l'envoi ou la consignation des alertes sous forme de notifications.
type ServiceNotifications struct {
    urlBase string
}

// NouvelleServiceNotifications crée un gestionnaire basique de notifications.
func NouvelleServiceNotifications(urlBase string) *ServiceNotifications {
    return &ServiceNotifications{urlBase: urlBase}
}

// Notifier consigne une alerte dans les logs. 
// Cette implémentation est un MVP basique, qui peut être remplacé par un notifier plus avancé (email, webhook).
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
