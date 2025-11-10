/* Service de vérification HTTP simple
 * Projet de session A25
 * By : Leandre Kanmegne
 * 
 * Fait un GET sur une URL et retourne le statut
 * Gère les erreurs réseau et les codes HTTP
 * Limite la taille de la réponse luee pour éviter d'abuser de la mémoire
 */
package services

import (
	"context"
	"io"
	"net/http"
	"strings"
	"time"

	"example.com/go-hello/src/internal/models"
)

// VerifierURL fait une requête GET et retourne le statut
func VerifierURL(ctx context.Context, url string) models.StatutMoniteur {
	// ajoute http:// si manquant
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = "http://" + url
	}

	client := &http.Client{Timeout: 10 * time.Second}

	debut := time.Now()
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	req.Header.Set("User-Agent", "MoniteurMVP/1.0")

	statut := models.StatutMoniteur{
		URL:      url,
		VerifieA: time.Now(),
	}

	resp, err := client.Do(req)
	if err != nil {
		// erreur réseau
		statut.EstDisponible = false
		statut.MessageErreur = err.Error()
		statut.CodeStatutHTTP = 0
		statut.Latence = time.Since(debut)
		return statut
	}
	defer resp.Body.Close()

	// limite la lecture pour pas abuser
	io.Copy(io.Discard, io.LimitReader(resp.Body, 512*1024))

	statut.CodeStatutHTTP = resp.StatusCode
	statut.Latence = time.Since(debut)
	statut.EstDisponible = resp.StatusCode >= 200 && resp.StatusCode < 400
	if !statut.EstDisponible {
		statut.MessageErreur = http.StatusText(resp.StatusCode)
	}

	return statut
}