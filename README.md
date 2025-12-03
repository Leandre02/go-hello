# Service de Monitoring

Projet de session A25  - Technologies Émergentes
Par : Leandre Kanmegne

# Prérequis :
- Docker Desktop installé : https://www.docker.com/products/docker-desktop/
- Docker compose disponible (Préconfiguré dans Docker Desktop récent)
- Port 8080 libre sur la machine (Valeur modifiable dans le .env au moment du build)
- Fichier .env configuré (voir .env.example) à la racine du projet

# Commandes:
- Démarrer: docker-compose -f docker-compose.dev.yml up
- Arrêter: docker-compose -f docker-compose.dev.yml down
- Voir logs: docker-compose -f docker-compose.dev.yml logs -f app
- Test avec race detector : Pour lancer: go test ./... -race dans la console wsl

Une fois démarré:

API et interface web accessibles sur: http://localhost:8080 (ou le port défini dans PORT).