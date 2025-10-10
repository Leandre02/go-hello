/*
 * Logique métier pour la surveillance des services
 * Projet de session A25
 * Auteur : Leandre Kanmegne
 * Version : 1.0
 *
 * Cette couche de service encapsule la logique métier de vérification des services HTTP.
 * 
 * Elle utilise un client HTTP configuré avec des timeouts rigoureux, gestion TLS minimale,
 * et contrôle du parallélisme via un canal sémaphore pour limiter la charge.
 * 
 * Chaque méthode applique le package context pour gérer l'annulation propre, les délais, et la propagation
 * d'informations entre goroutines.
 * 
 * Le module enregistre systématiquement les résultats des vérifications dans un dépôt abstrait
 * (interface repos.Repo), assurant la persistance séparée de la logique métier.
 * 
 * Le code s'inspire des bonnes pratiques idiomatiques Go modernes et des modèles reconnus dans l'écosystème :
 * - Utilisation de context.Context (https://pkg.go.dev/context) pour la gestion des deadlines et annulations [web:17]
 * - Mise en place d'un client HTTP robuste avec timeouts et gestion TLS (https://golang.org/pkg/net/http) [web:28]
 * - Limitation du parallélisme pour éviter la surcharge système, pattern de sémaphore canal (https://dev.to/kengowada/go-routing-101-handling-and-grouping-routes-with-nethttp-4k0e) [web:25]
 * - Séparation claire entre la couche métier (services) et la persistance (dépôt ou base de données), facilitant la testabilité et maintenance (https://threedots.tech/post/repository-pattern-in-go/) [web:13]
 * - Gestion des erreurs réseau et mesure précise de la latence avec reporting même en cas d'échec réseau (https://middleware.io/blog/golang-monitoring/) [web:36]
 *
 * Ce module est conçu pour s’intégrer parfaitement à l’infrastructure HTTP REST exposée par routes.go,
 * permettant une architecture claire, découplée et facile à tester.
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

// ServiceMoniteur encapsule la logique de vérification HTTP,
// gestion des timeouts, concurrence limitée et enregistrement des statuts.
type ServiceMoniteur struct {
    repo         repos.Repo     // Interface de persistance
    clientHTTP   *http.Client   // Client HTTP performant configurable
    sem          chan struct{}  // Semaphore pour limiter le parallélisme
    seuilLenteMs int64          // Seuil de latence pour générer des alertes de lenteur (ms)
}

// NewServiceMoniteur construit un service avec client HTTP sécurisé et paramètres configurables.
// maxParallele définit la limite de goroutines concurrentes,
// delaiRequete est le timeout pour chaque requête,
// seuilLenteMs définit la latence (en ms) au-delà de laquelle une alerte est générée.
func NewServiceMoniteur(repo repos.Repo, maxParallele int, delaiRequete time.Duration, seuilLenteMs int64) *ServiceMoniteur {
    if maxParallele <= 0 {
        maxParallele = 5
    }
    if delaiRequete <= 0 {
        delaiRequete = 10 * time.Second
    }
    if seuilLenteMs <= 0 {
        seuilLenteMs = 800 // 800 ms par défaut
    }

    // Transport HTTP personnalisé avec timeouts stricts et TLS moderne
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
        sem:          make(chan struct{}, maxParallele),
        seuilLenteMs: seuilLenteMs,
    }
}

// VerifierMoniteur effectue une requête HTTP GET sécurisée aux paramètres configurés.
// Elle limite le parallélisme, applique un contexte avec timeout, mesure la latence,
// et enregistre le statut dans le dépôt.
// Retourne le statut de disponibilité et une erreur si la requête n'a pas pu être effectuée.
func (s *ServiceMoniteur) VerifierMoniteur(ctx context.Context, mon models.Moniteur) (models.StatutMoniteur, error) {
    s.sem <- struct{}{}         // Prise de place dans la pool de workers
    defer func() { <-s.sem }()  // Libération

    if _, ok := ctx.Deadline(); !ok {
        var cancel context.CancelFunc
        ctx, cancel = context.WithTimeout(ctx, s.clientHTTP.Timeout)
        defer cancel()
    }

    debut := time.Now()
    req, _ := http.NewRequestWithContext(ctx, http.MethodGet, mon.URL, nil)
    req.Header.Set("User-Agent", "ServiceMoniteur/1.0 (+monitoring)")

    statut := models.StatutMoniteur{
        MoniteurID: mon.ID,
        URL:        mon.URL,
    }

    rsp, err := s.clientHTTP.Do(req)
    if err != nil {
        // En cas d'erreur réseau ou timeout
        statut.EstDisponible = false
        statut.MessageErreur = err.Error()
        statut.Latence = time.Since(debut)
        statut.VerifieA = time.Now()
        _ = s.repo.EnregistrerStatutMoniteur(ctx, statut)
        return statut, nil
    }
    defer rsp.Body.Close()

    // Limite de lecture pour éviter surcharge
    const limite = 1 << 20 // 1 MiB
    _, _ = io.Copy(io.Discard, io.LimitReader(rsp.Body, limite))

    statut.CodeStatutHTTP = rsp.StatusCode
    statut.Latence = time.Since(debut)
    statut.EstDisponible = rsp.StatusCode >= 200 && rsp.StatusCode < 400
    if !statut.EstDisponible && statut.MessageErreur == "" {
        statut.MessageErreur = http.StatusText(rsp.StatusCode)
    }
    statut.VerifieA = time.Now()

    _ = s.repo.EnregistrerStatutMoniteur(ctx, statut)

    return statut, nil
}

// VerifierURL est un raccourci simplifié pour checker une URL directement (sans config Moniteur)
func (s *ServiceMoniteur) VerifierURL(ctx context.Context, url string) (models.StatutMoniteur, error) {
    return s.VerifierMoniteur(ctx, models.Moniteur{URL: url})
}

// DerniersResultatsPourMoniteur récupère les derniers n statuts pour un moniteur donné.
// Si n <= 0, retourne tous les statuts disponibles.
func (s *ServiceMoniteur) DerniersResultatsPourMoniteur(ctx context.Context, moniteurID int, n int) ([]models.StatutMoniteur, error) {
    stats, err := s.repo.DerniersStatutsMoniteur(ctx, moniteurID)
    if err != nil {
        return nil, err
    }
    if n <= 0 || n >= len(stats) {
        return stats, nil
    }
    return stats[len(stats)-n:], nil
}

// ListerMoniteurs délègue la récupération des moniteurs au dépôt.
func (s *ServiceMoniteur) ListerMoniteurs(ctx context.Context) ([]models.Moniteur, error) {
    return s.repo.ListerMoniteurs(ctx)
}
