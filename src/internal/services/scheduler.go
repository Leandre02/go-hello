/* Planificateur de vérifications automatiques
 * Projet de session A25
 * By : Leandre Kanmegne
 * 
 * Lance des vérifications périodiques sur tous les moniteurs
 * Utilise des goroutines pour les vérifications en parallèle
 * Gère les erreurs et les délais d'attente
 * 
 * Sources:
 * https://dev.to/jones_charles_ad50858dbc0/building-a-go-concurrency-task-scheduler-efficient-task-processing-unleashed-4fhg
 * https://nghiant3223.github.io/2025/04/15/go-scheduler.html
 */
package services

import (
	"context"
	"sync"
	"time"

	"example.com/go-hello/src/internal/models"
	"example.com/go-hello/src/repos"
)

// Planificateur orchestre les vérifications périodiques
type Planificateur struct {
	repo      repos.Repo
	service   *ServiceMoniteur
	interval  time.Duration
	stopFn    context.CancelFunc
	wg        sync.WaitGroup
	enCoursMu sync.Mutex
	enCours   bool
}

// Crée un planificateur
func NewPlanificateur(repo repos.Repo, service *ServiceMoniteur, interval time.Duration) *Planificateur {
	if interval <= 0 {
		interval = 60 * time.Second
	}
	return &Planificateur{
		repo:     repo,
		service:  service,
		interval: interval,
	}
}

// Lance la boucle de vérifications
func (p *Planificateur) Demarrer(parent context.Context) {
	p.enCoursMu.Lock()
	if p.enCours {
		p.enCoursMu.Unlock()
		return
	}
	p.enCours = true
	p.enCoursMu.Unlock()

	ctx, cancel := context.WithCancel(parent)
	p.stopFn = cancel
	p.wg.Add(1)

	go func() {
		defer p.wg.Done()
		ticker := time.NewTicker(p.interval)
		defer ticker.Stop()

		// première vérification immédiate
		p.executerCycle(ctx)

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				p.executerCycle(ctx)
			}
		}
	}()
}

// Stop le planificateur
func (p *Planificateur) Arreter() {
	p.enCoursMu.Lock()
	if !p.enCours {
		p.enCoursMu.Unlock()
		return
	}
	p.enCours = false
	p.enCoursMu.Unlock()

	if p.stopFn != nil {
		p.stopFn()
	}
	p.wg.Wait()
}

// Vérifie tous les moniteurs en parallèle
func (p *Planificateur) executerCycle(ctx context.Context) {
	moniteurs, err := p.repo.ListerMoniteurs(ctx)
	if err != nil || len(moniteurs) == 0 {
		return
	}

	var wg sync.WaitGroup
	wg.Add(len(moniteurs))
	
	for i := range moniteurs {
		moniteur := moniteurs[i]
		go func(m models.Moniteur) {
			defer wg.Done()
			p.service.VerifierMoniteur(ctx, m)
		}(moniteur)
	}
	
	wg.Wait()
}