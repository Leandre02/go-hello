/*
 * Middleware pour la journalisation des requêtes HTTP
 * Projet de session A25
 * Auteur : Leandre Kanmegne
 * Version : 1.0
 *
 * Cette fonction middleware intercepte chaque requête HTTP,
 * logge la méthode, le chemin, l'adresse IP de l'appelant,
 * la durée de traitement, et le statut de la réponse.
 *
 * Utile pour monitorer l'activité serveur et diagnostiquer les problèmes.
 Explications :

Le middleware Journalisateur enveloppe le handler HTTP donné.

Un wrapper leveledResponseWriter capture le code HTTP écrit dans la réponse.

Le middleware logge la méthode HTTP, le chemin d’URL, l’adresse IP (prenant en compte le header X-Forwarded-For si présent), le statut HTTP, et le temps d’exécution.

La sortie est au format simple pour être lisible dans les logs classiques ou systèmes agrégateurs.

Tu peux enregistrer ce middleware chez ton mux ou serveur HTTP comme :

go
mux := http.NewServeMux()
// Enregistrer routes...

logMiddleware := middleware.Journalisateur(mux)
http.ListenAndServe(":8080", logMiddleware)
Cette approche est conforme aux bonnes pratiques actuelles de middleware Go. N’hésite pas à demander pour une version plus avancée avec plus de métadonnées ou sorties JSON structurées./*

Middleware pour la journalisation des requêtes HTTP

Projet de session A25

Auteur : Leandre Kanmegne

Version : 1.0

Cette fonction middleware intercepte chaque requête HTTP,

logge la méthode, le chemin, l'adresse IP de l'appelant,

la durée de traitement, et le statut de la réponse.

Utile pour monitorer l'activité serveur et diagnostiquer les problèmes.
*/
package middleware

import (
"log"
"net/http"
"time"
)

// Journalisateur est un middleware HTTP qui logge les requêtes et leur durée.
func Journalisateur(next http.Handler) http.Handler {
return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
debut := time.Now()
    // wrapper pour capturer le code HTTP retourné
    lrw := &leveledResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}

    // Appel du handler principal
    next.ServeHTTP(lrw, r)
    
    duree := time.Since(debut)

    ip := r.RemoteAddr
    if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
        ip = forwarded
    }

    log.Printf("%s %s %s %d %s\n",
        r.Method,
        r.URL.Path,
        ip,
        lrw.statusCode,
        duree)
})
}

// leveledResponseWriter permet de capturer le code HTTP écrit dans la réponse
type leveledResponseWriter struct {
http.ResponseWriter
statusCode int
}

func (lrw *leveledResponseWriter) WriteHeader(code int) {
lrw.statusCode = code
lrw.ResponseWriter.WriteHeader(code)
}