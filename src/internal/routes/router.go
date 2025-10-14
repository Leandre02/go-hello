/*
 * Routes HTTP du serveur (MVP francophone)
 * By : Leandre Kanmegne
 * Version : 1.0
 * Commentaires : Organisation inspirée des meilleures pratiques Go
 * Source : dev.to/kengowada/go-routing-101-handling-and-grouping-routes-with-nethttp-4k0e [web:25]
 */

package routes

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"example.com/go-hello/src/internal/models"
	"example.com/go-hello/src/internal/services"
	"example.com/go-hello/src/repos"
)

// Struct centrale des dépendances applicatives pour les routes
type ServicesApp struct {
    Depot repos.Repo // Interface de persistance des moniteurs et statuts
}

// Représente le corps attendu pour les vérifications d'URL
type RequeteVerification struct {
    URL string `json:"url"`
}

// Structure utilisée pour exposer les statuts à l'extérieur (API REST)
type StatutVue struct {
    EstDisponible bool      `json:"est_disponible"`
    CodeHTTP      int       `json:"code_http"`
    LatenceMs     int64     `json:"latence_ms"`
    MessageErreur string    `json:"message_erreur"`
    VerifieA      time.Time `json:"verifie_a"`
    URL           string    `json:"url"`
}

// Convertit un modèle métier StatutMoniteur en vue API StatutVue, facilitant le découplage métier/exposition
func vueDepuisModele(s models.StatutMoniteur) StatutVue {
    return StatutVue{
        EstDisponible: s.EstDisponible,
        CodeHTTP:      s.CodeStatutHTTP,
        LatenceMs:     s.Latence.Milliseconds(),
        MessageErreur: s.MessageErreur,
        VerifieA:      s.VerifieA,
        URL:           s.URL,
    }
}

// Permet d'envoyer une réponse JSON au client, en garantissant le bon format et le code HTTP
func ecrireJSON(w http.ResponseWriter, code int, v any) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(code)
    _ = json.NewEncoder(w).Encode(v)
}

// Active le support du CORS pour permettre les appels API depuis n'importe quel navigateur/client frontend
func activerCORS(w http.ResponseWriter) {
    w.Header().Set("Access-Control-Allow-Origin", "*")
    w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
    w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

// Ajoute le moniteur pour l'URL spécifiée s'il n'existe pas encore, retourne l'identifiant du moniteur
func obtenirIDMoniteur(ctx context.Context, depot repos.Repo, url string) (int, error) {
    _ = depot.AjouterMoniteur(ctx, models.Moniteur{URL: url, Nom: url, Type: "http"})
    mons, err := depot.ListerMoniteurs(ctx)
    if err != nil {
        return 0, err
    }
    for _, m := range mons {
        if m.URL == url {
            return m.ID, nil
        }
    }
    return 0, errors.New("moniteur introuvable après ajout")
}

// Handler de vérification de disponibilité : reçoit une URL, effectue un check, stocke le résultat, et retourne le statut
func HandlerVerification(app ServicesApp) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        activerCORS(w)
        if r.Method == http.MethodOptions {
            w.WriteHeader(http.StatusNoContent)
            return
        }

        r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
        defer r.Body.Close()

        var reqBody RequeteVerification
        if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil || strings.TrimSpace(reqBody.URL) == "" {
            http.Error(w, "Corps invalide: attendu {\"url\":\"...\"}", http.StatusBadRequest)
            return
        }

        ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
        defer cancel()

        statut := services.VerifierURL(ctx, reqBody.URL)
        if id, err := obtenirIDMoniteur(ctx, app.Depot, statut.URL); err == nil {
            statut.MoniteurID = id
            _ = app.Depot.EnregistrerStatutMoniteur(ctx, statut)
        }

       ecrireJSON(w, http.StatusOK, map[string]any{
           "statut": vueDepuisModele(statut),
       })
    }
}

// Handler pour récupérer les derniers statuts des moniteurs (attend un paramètre facultatif 'limit')
func HandlerResultats(app ServicesApp) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        activerCORS(w)
        if r.Method == http.MethodOptions {
            w.WriteHeader(http.StatusNoContent)
            return
        }
        if r.Method == http.MethodDelete {
            if err := app.Depot.ViderTout(r.Context()); err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
            }
            ecrireJSON(w, http.StatusOK, map[string]any{"ok": true})
            return
        }

        mons, err := app.Depot.ListerMoniteurs(r.Context())
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        var tous []models.StatutMoniteur
        for _, m := range mons {
            s, err := app.Depot.DerniersStatutsMoniteur(r.Context(), m.ID)
            if err == nil {
                tous = append(tous, s...)
            }
        }
        sort.Slice(tous, func(i, j int) bool { return tous[i].VerifieA.After(tous[j].VerifieA) })

        limite := 50
        if v := r.URL.Query().Get("limit"); v != "" {
            if n, err := strconv.Atoi(v); err == nil && n > 0 {
                limite = n
            }
        }
        if limite > len(tous) {
            limite = len(tous)
        }
        selection := tous[:limite]

        vues := make([]StatutVue, 0, len(selection))
        for _, s := range selection {
            vues = append(vues, vueDepuisModele(s))
        }
        ecrireJSON(w, http.StatusOK, map[string]any{
            "resultats": vues,
        })
    }
}

// Handler de santé de l'application, utilisé pour les probes ou monitorings externes
func HandlerEtatApplication() http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        activerCORS(w)
        if r.Method == http.MethodOptions {
            w.WriteHeader(http.StatusNoContent)
            return
        }
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("OK"))
    }
}

// EnregistrerRoutes enregistre les handlers et RETOURNE un http.Handler.
func EnregistrerRoutes(app ServicesApp) http.Handler {

    routes := http.NewServeMux() // Utilisation d'un ServeMux pour un routage simple 

    routes.HandleFunc("/api/verifier", HandlerVerification(app))
    routes.HandleFunc("/api/resultats", HandlerResultats(app))
    routes.HandleFunc("/api/etat", HandlerEtatApplication())
    
    routes.Handle("/", http.FileServer(http.Dir("/web")))
    return routes
}
