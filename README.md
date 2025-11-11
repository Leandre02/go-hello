# Service de Monitoring

Projet de session A25  
Par : Leandre Kanmegne

## 📋 Description

Service de monitoring simple qui vérifie automatiquement l'état de sites web et APIs. L'outil check régulièrement l'accessibilité et la rapidité de réponse des différents services, puis expose les résultats via une API REST et une interface web.

## ✨ Fonctionnalités

- ✅ Surveillance des services (HTTP, TCP, ICMP)
- ✅ Enregistrement des résultats dans PostgreSQL
- ✅ Alertes automatiques en cas de panne (codes 404, 500, etc.)
- ✅ Statistiques de performance (latence, disponibilité)
- ✅ Interface web pour visualiser les données
- ✅ API REST pour intégration avec d'autres systèmes
- 🔜 Authentification et autorisation (à venir)
- 🔜 Alertes avancées avec regex (à venir)

## 🛠️ Technologies utilisées

- **Langage** : Go 1.23
- **Base de données** : PostgreSQL 16
- **Framework web** : net/http (standard library)
- **Driver BD** : pgx v5
- **Développement** : Air (rechargement auto), Docker
- **Frontend** : HTML, CSS, JavaScript vanilla

## 📁 Architecture du projet

/
├── main.go                          # Point d'entrée
├── .env.example                     # Variables d'environnement
├── docker-compose*.yml              # Config Docker
├── src/
│   ├── internal/
│   │   ├── models/
│   │   │   └── types.go            # Structures de données
│   │   ├── routes/
│   │   │   └── router.go           # Routes HTTP
│   │   ├── services/
│   │   │   ├── http_checker.go    # Vérification HTTP simple
│   │   │   ├── monitor.go         # Service de monitoring avancé
│   │   │   ├── notifier.go        # Système d'alertes
│   │   │   └── scheduler.go       # Planificateur auto
│   │   └── middleware/
│   │       └── logger.go           # Logging des requêtes
│   ├── repos/
│   │   ├── repo.go                 # Interface repository
│   │   └── pg.go                   # Implémentation PostgreSQL
│   └── database/
│       ├── init.sql                # Schéma de base
│       └── dbtrigger.sql           # Triggers et alertes
└── web/
    ├── index.html                   # Interface utilisateur
    ├── script.js                    # Logique frontend
    └── styles.css                   # Styles


## 🚀 Installation et démarrage

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


## 📚 API Endpoints

### POST /api/verifier
Vérifie une URL et retourne son statut
json
Request:
{
  "url": "https://exemple.com"
}

Response:
{
  "statut": {
    "est_disponible": true,
    "code_http": 200,
    "latence_ms": 123,
    "verifie_a": "2025-01-10T14:30:00Z",
    "url": "https://exemple.com"
  }
}


### GET /api/resultats?limit=50
Récupère les derniers résultats
json
Response:
{
  "resultats": [
    {
      "est_disponible": true,
      "code_http": 200,
      "latence_ms": 123,
      "verifie_a": "2025-01-10T14:30:00Z",
      "url": "https://exemple.com"
    }
  ]
}


### DELETE /api/resultats
Vide toutes les données (moniteurs et statuts)

### GET /api/etat
Health check du serveur

## 🗄️ Base de données

### Tables principales

**monitoring.moniteurs**
- Stocke les services à surveiller

**monitoring.statuts**
- Historique de toutes les vérifications

**monitoring.alertes**
- Alertes générées automatiquement par triggers

### Triggers

Le système détecte automatiquement les transitions UP/DOWN et génère des alertes dans la table `monitoring.alertes`.

## 📖 Concepts techniques Go

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

## 🔧 Configuration avancée

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


## 📝 Commandes Docker utiles

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


## 🐛 Problèmes rencontrés et solutions

### Problème d'organisation
- **Solution** : Création du dossier `src/` pour mieux ranger les fichiers

### Air ne trouvait pas main.go
- **Solution** : Configuration du chemin dans `.air.toml`

### Connexion PostgreSQL échoue
- **Solution** : Vérifier que `DATABASE_URL` est bien défini et que PostgreSQL est démarré

### Erreur "address already in use"
- **Solution** : Arrêter le processus sur le port 8080 ou changer de port

## 📚 Sources et références

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

## 🎯 Prochaines étapes

- [ ] Ajouter l'authentification JWT
- [ ] Implémenter les alertes email/webhook
- [ ] Ajouter support pour TCP et ICMP
- [ ] Dashboard avec graphiques
- [ ] Export des données (CSV, JSON)
- [ ] API pour gérer les moniteurs (CRUD complet)
- [ ] Tests unitaires et d'intégration
- [ ] Déploiement sur Render ou autre service cloud