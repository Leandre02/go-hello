/* Routes HTTP du serveur
 * Projet de session A25
 * By : Leandre Kanmegne
 * 
 * Définit les endpoints de l'API REST pour le monitoring
 * - /api/verifier : vérifie une URL donnée
 * - /api/resultats : récupère les derniers statuts des moniteurs
 * - /api/etat : check de santé du serveur
 * Utilise le package net/http de Go pour gérer les routes et les handlers
 * Utilise le package context pour gérer les délais d'attente et annulations
 * Utilise le package encoding/json pour sérialiser/désérialiser les données JSON
 * Utilise le package sync pour gérer la concurrence
 * Se sert du mux de net/http pour router les requêtes
 * 
 * Source:
 * https://dev.to/kengowada/go-routing-101-handling-and-grouping-routes-with-nethttp-4k0e
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

// Regroupe les dépendances de l'app
type ServicesApp struct {
	Depot repos.Repo
}

// Représente le body pour vérifier une URL
type RequeteVerification struct {
	URL string `json:"url"`
}

// Représente un statut pour l'API
type StatutVue struct {
	EstDisponible bool      `json:"est_disponible"`
	CodeHTTP      int       `json:"code_http"`
	LatenceMs     int64     `json:"latence_ms"`
	MessageErreur string    `json:"message_erreur"`
	VerifieA      time.Time `json:"verifie_a"`
	URL           string    `json:"url"`
}

// Convertit un StatutMoniteur en StatutVue
func vueDepuisModele(statut models.StatutMoniteur) StatutVue {
	return StatutVue{
		EstDisponible: statut.EstDisponible,
		CodeHTTP:      statut.CodeStatutHTTP,
		LatenceMs:     statut.Latence.Milliseconds(),
		MessageErreur: statut.MessageErreur,
		VerifieA:      statut.VerifieA,
		URL:           statut.URL,
	}
}

// Envoie une réponse JSON
func ecrireJSON(w http.ResponseWriter, code int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(data)
}

// Active CORS pour permettre les appels depuis n'importe quel client
func activerCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

// Récupère ou crée l'ID d'un moniteur
func obtenirIDMoniteur(ctx context.Context, depot repos.Repo, url string) (int, error) {
	depot.AjouterMoniteur(ctx, models.Moniteur{URL: url, Nom: url, Type: "http"})
	
	moniteurs, err := depot.ListerMoniteurs(ctx)
	if err != nil {
		return 0, err
	}
	
	for _, moniteur := range moniteurs {
		if moniteur.URL == url {
			return moniteur.ID, nil
		}
	}
	
	return 0, errors.New("moniteur introuvable après ajout")
}

// Check une URL et retourne le résultat
func HandlerVerification(app ServicesApp) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		activerCORS(w)
		
		if req.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// limite la taille du body
		req.Body = http.MaxBytesReader(w, req.Body, 1<<20)
		defer req.Body.Close()

		var body RequeteVerification
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil || strings.TrimSpace(body.URL) == "" {
			http.Error(w, "Corps invalide: attendu {\"url\":\"...\"}", http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(req.Context(), 15*time.Second)
		defer cancel()

		statut := services.VerifierURL(ctx, body.URL)
		
		// enregistre dans la BD si possible
		if id, err := obtenirIDMoniteur(ctx, app.Depot, statut.URL); err == nil {
			statut.MoniteurID = id
			app.Depot.EnregistrerStatutMoniteur(ctx, statut)
		}

		ecrireJSON(w, http.StatusOK, map[string]any{
			"statut": vueDepuisModele(statut),
		})
	}
}

// Récupère les derniers statuts
func HandlerResultats(app ServicesApp) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		activerCORS(w)
		
		if req.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		
		if req.Method == http.MethodDelete {
			if err := app.Depot.ViderTout(req.Context()); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			ecrireJSON(w, http.StatusOK, map[string]any{"ok": true})
			return
		}

		moniteurs, err := app.Depot.ListerMoniteurs(req.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		
		var tousStatuts []models.StatutMoniteur
		for _, moniteur := range moniteurs {
			statuts, err := app.Depot.DerniersStatutsMoniteur(req.Context(), moniteur.ID)
			if err == nil {
				tousStatuts = append(tousStatuts, statuts...)
			}
		}
		
		// tri par date décroissante
		sort.Slice(tousStatuts, func(i, j int) bool {
			return tousStatuts[i].VerifieA.After(tousStatuts[j].VerifieA)
		})

		// limite le nombre de résultats
		limite := 50
		if valeur := req.URL.Query().Get("limit"); valeur != "" {
			if n, err := strconv.Atoi(valeur); err == nil && n > 0 {
				limite = n
			}
		}
		if limite > len(tousStatuts) {
			limite = len(tousStatuts)
		}
		
		selection := tousStatuts[:limite]

		vues := make([]StatutVue, 0, len(selection))
		for _, statut := range selection {
			vues = append(vues, vueDepuisModele(statut))
		}
		
		ecrireJSON(w, http.StatusOK, map[string]any{
			"resultats": vues,
		})
	}
}

// Check de santé du serveur
func HandlerEtatApplication() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		activerCORS(w)
		
		if req.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}
}

// Configure les routes HTTP
func EnregistrerRoutes(app ServicesApp) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/verifier", HandlerVerification(app))
	mux.HandleFunc("/api/resultats", HandlerResultats(app))
	mux.HandleFunc("/api/etat", HandlerEtatApplication())
	mux.Handle("/", http.FileServer(http.Dir("/web")))

	return mux
}