/* Service de vérification HTTP simple
* - Timeouts raisonnables
* - Lecture limitée du corps de réponse
* - Noms en français et commentaires clairs
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

// VerifierURL effectue une requête GET avec timeout et renvoie un StatutMoniteur.
func VerifierURL(ctx context.Context, url string) models.StatutMoniteur {
	// Normaliser l'URL (ajouter http:// si manquant)
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = "http://" + url
	}

	client := &http.Client{Timeout: 10 * time.Second}

	debut := time.Now()
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	req.Header.Set("User-Agent", "MoniteurMVP/1.0") // User-Agent personnalisé

	r := models.StatutMoniteur{
		URL:      url,
		VerifieA: time.Now(),
	}

	rep, err := client.Do(req)
	if err != nil {
		r.EstDisponible = false
		r.MessageErreur = err.Error()
		r.CodeStatutHTTP = 0
		r.Latence = time.Since(debut)
		return r
	}
	defer rep.Body.Close()

	// Limiter la lecture du corps pour éviter les abus (512 Ko)
	const limite = 512 * 1024
	_, _ = io.Copy(io.Discard, io.LimitReader(rep.Body, limite))

	r.CodeStatutHTTP = rep.StatusCode
	r.Latence = time.Since(debut)
	r.EstDisponible = rep.StatusCode >= 200 && rep.StatusCode < 400
	if !r.EstDisponible {
		r.MessageErreur = http.StatusText(rep.StatusCode)
	}
	return r
}