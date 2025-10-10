/*
	Planification des tâches et gestion des horaires

* Projet de session A25
* By : Leandre Kanmegne
Ton fichier scheduler.go (Planificateur) est non seulement utile, mais essentiel dans un système de monitoring automatisé. Il orchestre l’exécution périodique des vérifications sur tous les moniteurs, ce qui évite de devoir déclencher manuellement ou par requêtes chaque check.

Pourquoi le scheduler est important
Automatisation des tâches : Il exécute automatiquement des cycles de vérification à intervalle régulier (60 sec par défaut), assurant un suivi continu des services surveillés.

Gestion propre de la concurrence : Avec l’usage de sync.WaitGroup et du contexte, il garantit que toutes les vérifications se terminent proprement avant d’enchaîner ou d’arrêter le service.

Robustesse au changement : Le planificateur peut être démarré, arrêté, redémarré facilement, facilitant la maintenance et le contrôle dans un environnement en production.

Adapté à Go idiomatique : L’utilisation de goroutines, contextes, mutex, et waitgroups est conforme aux meilleures pratiques d’écriture concurrente Go.

En résumé
Sans ce scheduler, tu aurais besoin de lancer manuellement ou par une autre couche tes vérifications périodiques, ce qui rendrait ton monitoring moins fiable et moins maintainable. Il est donc recommandé de garder ce planificateur qui complète parfaitement la logique métier encapsulée dans le service
Pour approfondir l’implémentation de scheduler dans Go et comprendre la concurrence dans la gestion des tâches, plusieurs ressources de qualité existent, notamment :

Construire un scheduler concurrent efficace en Go, gestion des tâches et workers : https://dev.to/jones_charles_ad50858dbc0/building-a-go-concurrency-task-scheduler-efficient-task-processing-unleashed-4fhg - Go Concurrency Task Scheduler

Concurrence en Go et routines légères : https://nghiant3223.github.io/2025/04/15/go-scheduler.html

Meilleures pratiques pour scheduling et context cancelation : Documentation officielle du package context et net/http https://www.xenonstack.com/insights/go-application-monitoring

Ces sources confortent ta conception actuelle qui allie robustesse, clarté et contrôle dans un projet Go moderne.
*/
package services

import (
	"context"
	"sync"
	"time"

	"example.com/go-hello/src/internal/models"
	"example.com/go-hello/src/repos"
)

// Planificateur lance des vérifications périodiques sur l'ensemble des moniteurs.
type Planificateur struct {
	repo      repos.Repo
	service   *ServiceMoniteur
	interval  time.Duration
	stopFn    context.CancelFunc
	wg        sync.WaitGroup
	enCoursMu sync.Mutex
	enCours   bool
}

// NewPlanificateur crée un scheduler simple (tick périodique).
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

// Demarrer lance la boucle de planification dans une goroutine.
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

		// Tick immédiat au démarrage
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

// Arreter termine proprement la boucle et attend la fin.
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

// executerCycle récupère la liste des moniteurs et les vérifie en parallèle bornée (via ServiceMoniteur).
func (p *Planificateur) executerCycle(ctx context.Context) {
	moniteurs, err := p.repo.ListerMoniteurs(ctx)
	if err != nil || len(moniteurs) == 0 {
		return
	}

	var wg sync.WaitGroup
	wg.Add(len(moniteurs))
	for i := range moniteurs {
		mon := moniteurs[i]
		go func(m models.Moniteur) {
			defer wg.Done()
			// chaque vérification hérite d'un timeout du service
			_, _ = p.service.VerifierMoniteur(ctx, m)
		}(mon)
	}
	wg.Wait()
}
