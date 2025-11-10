/* Middleware pour logger les requêtes HTTP
 * Projet de session A25
 * By : Leandre Kanmegne
 * 
 * Intercepte chaque requête et log la méthode, URL, IP, durée et statut HTTP
 * Utilise un wrapper (http.ResponseWriter) pour capturer le code de statut HTTP
 * http.ResponseWriter est un objet pour écrire les réponses HTTP en Go
 * Retourne un handler HTTP (fonction) qui peut être utilisé dans la chaîne de middleware pour permettre le logging
 * récupère l'IP réelle du client si derrière un proxy (serveur intermédiaire) via l'en-tête X-Forwarded-For
 */
package middleware

import (
	"log"
	"net/http"
	"time"
)

// Journalisateur des requêtes HTTP
func Journalisateur(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		debut := time.Now()

		// wrapper pour capturer le code HTTP
		wrapper := &wrapperReponse{ResponseWriter: w, statusCode: http.StatusOK}

		// appel du handler
		next.ServeHTTP(wrapper, req)

		duree := time.Since(debut)

		// récupère l'IP réelle si derrière un proxy
		ip := req.RemoteAddr
		if forwarded := req.Header.Get("X-Forwarded-For"); forwarded != "" {
			ip = forwarded
		}

		log.Printf("%s %s %s %d %s\n",
			req.Method,
			req.URL.Path,
			ip,
			wrapper.statusCode,
			duree)
	})
}

// Capture le code HTTP écrit
type wrapperReponse struct {
	http.ResponseWriter
	statusCode int
}

// Capture le code de statut HTTP
func (w *wrapperReponse) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}