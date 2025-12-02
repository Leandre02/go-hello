/* Point d'entrée du serveur HTTP
 * Projet de session A25
 * By : Leandre Kanmegne
 * 
 * Lance le serveur HTTP et gère l'arrêt avec les signaux système
 * Point d'entrée principal de l'application
 * Utilisation de WaitGroup en Go qui est un outil de synchronisation qui permet d’attendre que plusieurs goroutines aient fini leur travail avant de continuer.
 */

package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"example.com/go-hello/src/internal/routes"
	"example.com/go-hello/src/repos"
)

func main() {
	// récupère l'URL de connexion PostgreSQL
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Println("[ERREUR] DATABASE_URL non défini dans les variables d'environnement")
		log.Fatal("Impossible de démarrer sans connexion à la base de données")
	}

	// connexion à PostgreSQL
	depot, err := repos.NouvelleConnexion(dsn)
	if err != nil {
		log.Fatalf("Erreur connexion base de données : %v", err)
	}
	defer func() {
		if errFermeture := depot.Fermer(); errFermeture != nil {
			log.Printf("Erreur fermeture base : %v", errFermeture)
		}
	}()

	// setup de l'application avec les dépendances
	app := routes.ServicesApp{
		Depot: depot,
	}

	// création du router HTTP
	mux := routes.EnregistrerRoutes(app)

	// contexte pour gérer l'arrêt propre
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	serveur := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	// démarrage du serveur dans une goroutine
	go func() {
		log.Println("Serveur démarré sur :8080")
		if err := serveur.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Erreur démarrage serveur : %v", err)
		}
	}()

	// attente du signal d'interruption
	<-ctx.Done()
	log.Println("Arrêt du serveur en cours...")

	// timeout pour l'arrêt gracieux
	ctxShutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := serveur.Shutdown(ctxShutdown); err != nil {
		log.Fatalf("Erreur lors de l'arrêt du serveur : %v", err)
	}

	log.Println("Serveur arrêté")
}