##  Description

Service de monitoring simple qui vérifie automatiquement l'état de sites web et APIs. L'outil check régulièrement l'accessibilité et la rapidité de réponse des différents services, puis expose les résultats via une API REST et une interface web.

##  Fonctionnalités

-  Surveillance des services (HTTP, TCP, ICMP)
-  Enregistrement des résultats dans PostgreSQL
-  Alertes automatiques en cas de panne (codes 404, 500, etc.)
-  Statistiques de performance (latence, disponibilité)
-  Interface web pour visualiser les données
-  API REST pour intégration avec d'autres systèmes
-  Authentification et autorisation (à venir)
-  Alertes avancées avec regex (à venir)

## Technologies utilisées

- **Langage** : Go 1.23
- **Base de données** : PostgreSQL 16
- **Framework web** : net/http (standard library)
- **Driver BD** : pgx v5
- **Développement** : Air (rechargement auto), Docker
- **Frontend** : HTML, CSS, JavaScript vanilla

## Matrice d'Eisenhower

- Important et Urgent : Connexion à la base de données, Vérification des services, Enregistrement des résultats


- Important mais pas Urgent : Interface web, API REST, Authentification ( A venir )


- Pas Important mais Urgent : Configuration de l'environnement de développement, Tests unitaires


- Pas Important et pas Urgent : Alerte avancée, Statistiques détaillées


## Architecture du projet


- main.go : point d'entrée de l'application


- .air.toml : configuration pour le rechargement automatique lors du développement !important : c'est ici que je dois configurer le chemin vers le fichier main.go


- go.mod : gestion des dépendances du projet


- .gitignore : fichiers et dossiers à ignorer par Git


- Readme.txt : documentation du projet


- .dockerignore : fichiers et dossiers à ignorer par Docker


- Dockerfile : instructions pour construire l'image Docker

  --- Dossier src : Dossier de rangement de mes sous-dossiers ---


  --- Dossier src/database ---

- init.sql : script SQL pour créer la base de données et les tables nécessaires


- dbtrigger.sql : script SQL pour créer les triggers de la base de données





  --- Dossier src/models ---

- MoniteurModel.go : définit le modèle de données pour les moniteurs



  --- Dossier src/repos ---

- pg.go : gestion de la connexion à la base de données PostgreSQL

- Repos.go : dépôt pour gérer les opérations sur les moniteurs





  --- Dossier src/services ---

- MoniteurService.go : service pour la logique métier liée aux moniteurs

- Planificateur.go : service pour la planification automatique des tâches


  --- Dossier src/controllers ---
- MoniteurController.go : contrôleur pour gérer les requêtes HTTP liées aux moniteurs


  --- Dossier src/routes ---

- MoniteurRoutes.go : définit les routes HTTP pour les moniteurs


  --- Dossier src/middleware ---

- AuthMiddleware.go : middleware pour l'authentification des utilisateurs ( A venir )





  --- Dossier src/view ---

- index.html : page HTML principale pour l'interface web


- styles.css : styles CSS pour l'interface web


- script.js : scripts JavaScript pour l'interface web



/_ Source _/


Notes de cours pour la BD PostgreSQL

- https://www.w3schools.com/postgresql/postgresql_create_table.php


- https://bd1.profinfo.ca/notes_de_cours/section_1.4/#afficher-les-tables


- https://bd2.profinfo.ca/mysql/creation_table/#syntaxe-de-base


- https://gowebexamples.com/hello-world/


- https://www.postgresql.org/docs/9.1/datatype-numeric.html



  --- Remarque importante sur les types de données Serial et Bigserial ---

* Bigserial est spécifique à PostgreSQL et est utilisé pour les colonnes qui nécessitent des valeurs uniques et auto-incrémentées, souvent utilisées pour les clés primaires.

* Bigserial permet de stocker des entiers auto-incrémentés de grande taille, allant de 1 à 9223372036854775807 vs Serial qui va de 1 à 2147483647.




## Installation et démarrage

### Prérequis

- Docker et Docker Compose
- Go 1.23+ (pour développement local)
- PostgreSQL 16 (si hors Docker)

### Démarrage rapide avec Docker

1. **Cloner le projet**

   git clone <url-du-projet>
   cd go-hello


2. **Configurer les variables d'environnement**

   cp .env.example .env
   # Éditer .env avec vos valeurs


3. **Lancer avec Docker Compose**


   # Mode développement (avec rechargement auto)
   docker-compose -f docker-compose.dev.yml up



4. **Accéder à l'application**
   - Interface web : http://localhost:8080
   - API : http://localhost:8080/api/
   - PostgreSQL : localhost:5432

### Développement local (sans Docker)

1. **Démarrer PostgreSQL**

   docker run --name monitoring_postgres \
     -e POSTGRES_PASSWORD=postgres \
     -e POSTGRES_DB=monitoring_database \
     -p 5432:5432 -d postgres:16


2. **Initialiser la base de données**

   psql -h localhost -U postgres -d monitoring_database -f src/database/init.sql
   psql -h localhost -U postgres -d monitoring_database -f src/database/dbtrigger.sql


3. **Configurer les variables**

   export DATABASE_URL="postgres://postgres:postgres@localhost:5432/monitoring_database?sslmode=disable"


4. **Lancer l'application**

   go run main.go


##  API Endpoints

### POST /api/verifier
Vérifie une URL et retourne son statut

### GET /api/resultats?limit=50
Récupère les derniers résultats


### DELETE /api/resultats
Vide toutes les données (moniteurs et statuts)

### GET /api/etat
Health check du serveur

##  Base de données

### Tables principales

**monitoring.moniteurs**
- Stocke les services à surveiller

**monitoring.statuts**
- Historique de toutes les vérifications

**monitoring.alertes**
- Alertes générées automatiquement par triggers

### Triggers

Le système détecte automatiquement les transitions UP/DOWN et génère des alertes dans la table `monitoring.alertes`.

##  Concepts techniques Go :

- pgx est un pilote PostgreSQL écrit entièrement en Go. Il offre une interface native haute performance pour PostgreSQL, en exposant des fonctionnalités spécifiques à ce SGBD (comme LISTEN/NOTIFY, COPY), tout en pouvant également être utilisé comme driver compatible database/sql.

Pourquoi l'utiliser ? - Pour un accès efficace à la base avec support spécialisé PostgreSQL.

- context.Context permet de transmettre autour d’une requête des informations comme un délai d’expiration (timeout), une annulation, et des métadonnées. Il est utilisé pour gérer proprement la durée de vie d’opérations asynchrones ou dépendantes de ressources.
Pourquoi c’est important ?
Cela permet d'éviter les fuites de goroutines, d'interrompre des requêtes longues, et de propager des signaux d’annulation dans toute la chaîne d’appels.

- En Go, un Handler est une interface HTTP centrale qui gère une requête HTTP et prépare une réponse. Son rôle est d’exécuter la logique métier correspondante.
Un HandlerFunc est une fonction avec la signature func(ResponseWriter, *Request) qui est convertible en Handler.
Pourquoi utiliser ces abstractions ?
Elles permettent de composer et d’enchaîner des traitements HTTP de façon propre et modulaire - comme un middleware ou un routeur.

### Context
Gestion des timeouts et annulations dans les requêtes
- Source : https://pkg.go.dev/context

### Goroutines et Channels
Permet de surveiller plusieurs services en parallèle
- Pattern de workers pour limiter la concurrence

### pgx Driver
Driver PostgreSQL performant avec support natif des features avancées
- Source : https://github.com/jackc/pgx

### Repository Pattern
Séparation claire entre logique métier et persistance
- Source : https://threedots.tech/post/repository-pattern-in-go/

- Source d'inspiration : https://github.com/prometheus/prometheus

-- Modele de disposition : https://github.com/golang-standards/project-layout

-- Source note de cours : https://www.w3schools.com/go/index.php

-- Synthaxe de Go : https://www.w3schools.com/go/go_formatting_verbs.php

-- Les tableaux en Go :https://www.w3schools.com/go/go_arrays.php

-- Le context : https://pkg.go.dev/golang.org/x/net/context


##  Configuration avancée

Toutes les configurations sont dans `.env` :


# Serveur
PORT=8080

# Monitoring
INTERVALLE_VERIFICATION_SECONDES=60
WORKERS_MAX_PARALLELES=5
SEUIL_LATENCE_LENTE_MS=800

# Timeouts
TIMEOUT_REQUETE_SECONDES=10
TIMEOUT_ARRET_SERVEUR_SECONDES=5


##  Commandes Docker utiles

# Build l'image
docker build -t monitoring:latest .

# Lancer manuellement
docker run -d --name monitoring \
  -e DATABASE_URL="..." \
  -p 8080:8080 monitoring:latest

# Voir les logs
docker logs -f monitoring

# Arrêter et supprimer
docker stop monitoring && docker rm monitoring

# Nettoyer tout
docker-compose down -v


##  Problèmes rencontrés et solutions

### Problème d'organisation
- **Solution** : Création du dossier `src/` pour mieux ranger les fichiers

### Air ne trouvait pas main.go
- **Solution** : Configuration du chemin dans `.air.toml`

### Confusion entre main.go et ./src/cmd/server/main.go
- **Solution** :  ls suivi de find . -name "main.go" -type  pour trouver le chemin de main et corriger dans docker-compose

### Connexion PostgreSQL échoue : Erreur : Failed to connect to database
- **Solution** : Vérifier que `DATABASE_URL` est bien défini dans le .env et que PostgreSQL est démarré sur le port 5432

### Erreur "address already in use"
- **Solution** : Arrêter le processus sur le port 8080 ou changer de port

### Probleme de configuration de la BD
- **Solution** : Utlisation du module Pgx de Go

### Erreur d'handling
- **Solution** : Go utilise des valeurs d'erreur explicites plutôt que des exceptions. Chaque fonction pouvant échouer retourne une erreur.

## Fichiers Frontend non servis (404 sur localhost:8080)
- **Solution** : Correction du chemin dans router.go

## Probleme de configuration de mes routes :
- **Solution** : curl http://localhost:8080/api/etat pour consulter l'etat de mes routes

## Variable DATABASE_URL harcodée dans docker-compose 
- **Solution** : Creation d'un fichier .env

##  Sources et références

### Documentation Go
- https://pkg.go.dev/context
- https://golang.org/pkg/net/http
- https://gowebexamples.com/

### PostgreSQL et pgx
- https://dev.to/mx_tech/go-with-postgresql-best-practices-for-performance-and-safety-47d7
- https://betterstack.com/community/guides/scaling-go/postgresql-pgx-golang/
- https://www.postgresql.org/docs/

### Architecture et patterns
- https://threedots.tech/post/repository-pattern-in-go/
- https://github.com/golang-standards/project-layout
- https://github.com/prometheus/prometheus (inspiration)

### Monitoring et alerting
- https://prometheus.io/docs/alerting/latest/overview/
- https://middleware.io/blog/golang-monitoring/

### Concurrence et scheduling
- https://dev.to/jones_charles_ad50858dbc0/building-a-go-concurrency-task-scheduler-efficient-task-processing-unleashed-4fhg
- https://nghiant3223.github.io/2025/04/15/go-scheduler.html

### Routing HTTP
- https://dev.to/kengowada/go-routing-101-handling-and-grouping-routes-with-nethttp-4k0e

##  Prochaines étapes

- [ ] Ajouter l'authentification JWT
- [ ] Implémenter les alertes email/webhook
- [ ] Ajouter support pour TCP et ICMP
- [ ] Dashboard avec graphiques
- [ ] Export des données (CSV, JSON)
- [ ] API pour gérer les moniteurs (CRUD complet)
- [ ] Tests unitaires et d'intégration
- [ ] Déploiement sur Render ou autre service cloud