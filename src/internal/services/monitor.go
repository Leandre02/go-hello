/* Service de monitoring avec gestion avancée
 * Projet de session A25
 * By : Leandre Kanmegne
 * 
 * Encapsule la logique de vérification HTTP avec client configuré,
 * limitation du parallélisme et enregistrement en BD
 * Permet de vérifier des moniteurs et URLs directement
 * Gère les erreurs et les délais d'attente
 * Définit le service de monitoring utilisant le pattern repository
 * Limite le nombre de requêtes HTTP simultanées via un sémaphore
 * 
 * Sources:
 * https://pkg.go.dev/context - gestion des timeouts et annulations
 * https://golang.org/pkg/net/http - client HTTP
 * https://threedots.tech/post/repository-pattern-in-go/ - pattern repository
 */

package services

import (
	"context"
	"crypto/tls"
	"io"
	"net"
	"net/http"
	"time"

	"example.com/go-hello/src/internal/models"
	"example.com/go-hello/src/repos"
)

// Gère la vérification des moniteurs
type ServiceMoniteur struct {
	repo         repos.Repo
	clientHTTP   *http.Client
	semaphore    chan struct{}
	seuilLenteMs int64
}

// Crée un service de monitoring
func NewServiceMoniteur(repo repos.Repo, maxParallele int, delaiRequete time.Duration, seuilLenteMs int64) *ServiceMoniteur {
	if maxParallele <= 0 {
		maxParallele = 5
	}
	if delaiRequete <= 0 {
		delaiRequete = 10 * time.Second
	}
	if seuilLenteMs <= 0 {
		seuilLenteMs = 800
	}

	// configure le transport HTTP
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   3 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout:   4 * time.Second,
		ResponseHeaderTimeout: delaiRequete,
		ExpectContinueTimeout: 1 * time.Second,
		IdleConnTimeout:       60 * time.Second,
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   10,
		ForceAttemptHTTP2:     true,
		TLSClientConfig:       &tls.Config{MinVersion: tls.VersionTLS12},
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   delaiRequete,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 5 {
				return http.ErrUseLastResponse
			}
			return nil
		},
	}

	return &ServiceMoniteur{
		repo:         repo,
		clientHTTP:   client,
		semaphore:    make(chan struct{}, maxParallele),
		seuilLenteMs: seuilLenteMs,
	}
}

// check un moniteur et enregistre le résultat
func (s *ServiceMoniteur) VerifierMoniteur(ctx context.Context, moniteur models.Moniteur) (models.StatutMoniteur, error) {
	// limite le nombre de requêtes HTTP simultanées
	s.semaphore <- struct{}{}
	defer func() { <-s.semaphore }()

	// ajoute un timeout si pas déjà défini
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, s.clientHTTP.Timeout)
		defer cancel()
	}

	debut := time.Now()
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, moniteur.URL, nil)
	req.Header.Set("User-Agent", "ServiceMoniteur/1.0")

	statut := models.StatutMoniteur{
		MoniteurID: moniteur.ID,
		URL:        moniteur.URL,
	}

	resp, err := s.clientHTTP.Do(req)
	if err != nil {
		// erreur réseau
		statut.EstDisponible = false
		statut.MessageErreur = err.Error()
		statut.Latence = time.Since(debut)
		statut.VerifieA = time.Now()
		s.repo.EnregistrerStatutMoniteur(ctx, statut)
		return statut, nil
	}
	defer resp.Body.Close()

	// limite la lecture
	io.Copy(io.Discard, io.LimitReader(resp.Body, 1<<20))

	statut.CodeStatutHTTP = resp.StatusCode
	statut.Latence = time.Since(debut)
	statut.EstDisponible = resp.StatusCode >= 200 && resp.StatusCode < 400
	if !statut.EstDisponible && statut.MessageErreur == "" {
		statut.MessageErreur = http.StatusText(resp.StatusCode)
	}
	statut.VerifieA = time.Now()

	s.repo.EnregistrerStatutMoniteur(ctx, statut)

	return statut, nil
}

// Vérifie une URL directement
func (s *ServiceMoniteur) VerifierURL(ctx context.Context, url string) (models.StatutMoniteur, error) {
	return s.VerifierMoniteur(ctx, models.Moniteur{URL: url})
}

// Récupère les derniers statuts
func (s *ServiceMoniteur) DerniersResultatsPourMoniteur(ctx context.Context, moniteurID int, n int) ([]models.StatutMoniteur, error) {
	statuts, err := s.repo.DerniersStatutsMoniteur(ctx, moniteurID)
	if err != nil {
		return nil, err
	}
	if n <= 0 || n >= len(statuts) {
		return statuts, nil
	}
	return statuts[len(statuts)-n:], nil
}

// Retourne tous les moniteurs
func (s *ServiceMoniteur) ListerMoniteurs(ctx context.Context) ([]models.Moniteur, error) {
	return s.repo.ListerMoniteurs(ctx)
}