/* Point d'entrée du serveur HTTP
* Objectif MVP: un service de monitoring simple et clair
* Résumé concret de ton projet de monitoring en Go
Ton projet contient plusieurs fichiers clés, chacun avec un rôle précis, ensemble formant une architecture claire, maintenable et évolutive selon les meilleures pratiques Go récentes.

1. Types et modèles (models/)
Définissent les structures de données métier : Moniteur, StatutMoniteur, Alertes.

Fournissent la base de tout le traitement en garantissant une typage clair.

2. Repository PostgreSQL (pg.go dans repos/)
Interface avec la base PostgreSQL via pgx, gérant le CRUD des moniteurs et le stockage des statuts.

Assure une abstraction de la persistance avec contexte, pool et gestion des NULL explicites.

3. Service métier (monitor.go, notifier.go, scheduler.go dans services/)
ServiceMoniteur contient la logique de vérification HTTP robuste, gestion des timeout, client personnalisé et parallélisme contrôlé.

ServiceNotifications génère les alertes à partir des statuts, avec sévérité et messages détaillés; elle peut notifier via logs ou se connecter à d’autres systèmes.

Planificateur orchestre le lancement périodique des vérifications, garantissant un monitoring continu et fiable.

4. Routes HTTP (routes.go)
Sont les points d’entrée exposés aux clients, transformant requêtes en appels métier, puis formatant les réponses JSON.

Utilise une structure ServicesApp injectant les dépendances comme Repo et ServiceMoniteur.

Applique la gestion CORS, la validation et traite les contextes HTTP.

5. Middleware (middleware.go)
Intercepte les requêtes HTTP pour journaliser méthode, URL, temps de traitement, et code HTTP.

Utile pour surveillance, analyse de performance et diagnostics.

Pourquoi ces fichiers ?
Séparation claire des responsabilités, appliquant l’architecture propre (Clean/Hexagonal) pour la clarté, testabilité et maintenabilité.

Couche modèle gère la définition des données.

Repository isole la persistance des règles métier.

Services métier contiennent la logique centrale, indépendamment des détails techniques externes.

Routes adaptent cette logique aux requêtes HTTP.

Middleware améliore la qualité opérationnelle.

Cette organisation facilite l’évolution future, l’extension fonctionnelle ou technique, et le travail collaboratif grace à des conventions claires et des interfaces stables.
 */

package main

import (
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "context"
    "time"

    "example.com/go-hello/src/internal/routes"
     "example.com/go-hello/src/repos"
)

// fonction main : initialisation et lancement du serveur HTTP
func main() {
    // DSN PostgreSQL - adapter selon ton environnement
    dsn := "postgresql://monitoring_database_zp7a_user:OBFLhG8AjhyhnLCjqWTXWlCz0FnRW8et@dpg-d3jchol6ubrc73co188g-a.oregon-postgres.render.com/monitoring_database_zp7a"

    // Initialisation du dépôt PostgreSQL
    depot, err := repos.NouvelleConnexion(dsn)
    if err != nil {
        log.Fatalf("Erreur connexion base de données : %v", err)
    }
    defer func() {
        if cerr := depot.Fermer(); cerr != nil {
            log.Printf("Erreur fermeture base : %v", cerr)
        }
    }()

    // Injection des dépendances au router (app agrégé)
    app := routes.ServicesApp{
        Depot: depot,
    }

    // Création du router avec injection du dépôt
    mux := routes.EnregistrerRoutes(app)

    // Configuration d'un contexte pour gérer l'arrêt propre
    ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
    defer stop()

    srv := &http.Server{
        Addr:    ":8080",
        Handler: mux,
    }

    if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
    log.Fatalf("serveur: %v", err)
}
    // Démarrage du serveur HTTP dans une goroutine pour permettre un arrêt propre
    go func() {
        log.Println("Serveur démarré sur :8080")
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("Erreur démarrage serveur : %v", err)
        }
    }()

    // Attente du signal d'interruption ou FIN
    <-ctx.Done()
    log.Println("Arrêt du serveur...")

    // Timeout pour arrêt gracieux
    ctxShutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    if err := srv.Shutdown(ctxShutdown); err != nil {
        log.Fatalf("Erreur lors de l'arrêt du serveur : %v", err)
    }

    log.Println("Serveur arrêté proprement")
}
