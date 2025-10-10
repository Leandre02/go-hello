/_ Fonctionnalités du service de monitoring _/

- Surveiller l'état des services (HTTP, TCP, ICMP)
- Enregistrer les résultats des vérifications dans une base de données
- Alerter en cas de panne ( code d'erreur 404, 500, etc. ) ( A venir ) via des Regex
- Fournir des statistiques de performance
- Interface web pour visualiser les données
- API REST pour intégration avec d'autres systèmes
- Authentification et autorisation des utilisateurs ( A venir)

/_ Technologies utilisées _/

- Langage de programmation : Go
- Base de données : PostgreSQL
- Framework web : net/http (standard library)
- Outils de développement : Air (rechargement automatique), Docker (conteneurisation)
- Frontend : HTML, CSS, JavaScript (A venir)
- Regex (A venir)

/_ Matrice d'Eisenhower _/

- Important et Urgent : Connexion à la base de données, Vérification des services, Enregistrement des résultats
- Important mais pas Urgent : Interface web, API REST, Authentification ( A venir )
- Pas Important mais Urgent : Configuration de l'environnement de développement, Tests unitaires
- Pas Important et pas Urgent : Alerte avancée, Statistiques détaillées

/_ Architecture du projet _/
-- Source d'inspiration : https://github.com/prometheus/prometheus
-- Modele de disposition : https://github.com/golang-standards/project-layout

-- Source note de cours : https://www.w3schools.com/go/index.php

-- Synthaxe de Go : https://www.w3schools.com/go/go_formatting_verbs.php

-- Les tableaux en Go :https://www.w3schools.com/go/go_arrays.php

-- Le context : https://pkg.go.dev/golang.org/x/net/context

/_ Definition des concepts techniques de Go _/

## 🔧 Concepts fondamentaux

### Le Context

Le context est un package qui permet de gérer l'annulation, les timeouts et la transmission de valeurs à travers les goroutines. Dans notre projet de monitoring, il est essentiel pour :

- Gérer les timeouts des requêtes HTTP vers les services surveillés
- Annuler les vérifications en cours si nécessaire
- Transmettre des métadonnées comme les identifiants de requête

**Exemple d'utilisation :**

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
resp, err := http.Get(moniteur.URL) // Utilise le context pour timeout
```

### Une Goroutine

C'est une fonction qui s'exécute de manière concurrente (en parallèle) avec d'autres goroutines. Les goroutines sont légères et permettent de surveiller plusieurs services simultanément sans bloquer l'application principale.

**Dans notre monitoring :**

- Chaque vérification de service peut s'exécuter dans sa propre goroutine
- Permet de surveiller des centaines de services en parallèle
- Utilise beaucoup moins de mémoire qu'un thread traditionnel

### Les Channels

Les channels sont des "tuyaux" qui permettent aux goroutines de communiquer entre elles de manière sûre. Ils permettent d'échanger des données sans risque de corruption.

**Utilisation dans le monitoring :**

```go
resultChan := make(chan StatutMoniteur, 100)
// Une goroutine envoie les résultats
go func() { resultChan <- statut }()
// Une autre goroutine reçoit et traite
statut := <-resultChan
```

### Les Interfaces

Une interface définit un contrat (ensemble de méthodes) qu'un type doit respecter. Elle permet une programmation flexible et modulaire.

**Exemple dans notre projet :**

```go
type Checker interface {
    Check(ctx context.Context, url string) (StatutMoniteur, error)
}
// HTTPChecker, TCPChecker peuvent implémenter cette interface
```

### Les Structs

Les structs sont des types personnalisés qui regroupent des données liées. Elles sont l'équivalent des classes dans d'autres langages.

**Nos structs principales :**

- `Moniteur` : Représente un service à surveiller
- `StatutMoniteur` : Contient le résultat d'une vérification
- `Alert` : Représente une alerte générée

### Les Pointeurs

Les pointeurs stockent l'adresse mémoire d'une variable plutôt que sa valeur. Ils permettent de modifier des données sans les copier.

**Usage typique :**

```go
func (m *MonitorService) Check(moniteur *Moniteur) error {
    // Le * permet de modifier directement l'objet
}
```

### Error Handling

Go utilise des valeurs d'erreur explicites plutôt que des exceptions. Chaque fonction pouvant échouer retourne une erreur.

**Pattern typique :**

```go
statut, err := checkService(url)
if err != nil {
    log.Printf("Erreur lors de la vérification: %v", err)
    return err
}
```

### Les Packages

Les packages organisent le code en modules réutilisables. Notre projet utilise :

- `net/http` : Pour les requêtes HTTP
- `database/sql` : Pour la base de données
- `time` : Pour la gestion du temps
- Nos packages internes : `models`, `services`, `repos`

### JSON Marshal/Unmarshal

Go peut automatiquement convertir des structs en JSON et vice-versa grâce aux tags.

**Exemple :**

```go
type StatutMoniteur struct {
    URL    string    `json:"url"`
    Statut bool      `json:"statut"`
    Date   time.Time `json:"date"`
}
```

### Les Slices

Les slices sont des tableaux dynamiques qui peuvent grandir ou rétrécir selon les besoins.

**Usage dans le monitoring :**

```go
var moniteurs []Moniteur
moniteurs = append(moniteurs, nouveauMoniteur)
```

## 🚀 Concepts avancés pour le monitoring

### Worker Pools

Pattern pour limiter le nombre de goroutines concurrentes et gérer la charge.

### Rate Limiting

Contrôler la fréquence des vérifications pour éviter de surcharger les services surveillés.

### Graceful Shutdown

Arrêter proprement l'application en terminant les vérifications en cours.

### Middleware Pattern

Chaîner des fonctions pour ajouter des fonctionnalités (logging, auth, metrics).

Ces concepts forment la base de notre architecture de monitoring robuste et performante !

    --- A la racine du projet ---

- main.go : point d'entrée de l'application
- .air.toml : configuration pour le rechargement automatique lors du développement !important : c'est ici que je dois configurer le chemin vers le fichier main.go
- go.mod : gestion des dépendances du projet
- .gitignore : fichiers et dossiers à ignorer par Git
- Readme.txt : documentation du projet
- .dockerignore : fichiers et dossiers à ignorer par Docker
- Dockerfile : instructions pour construire l'image Docker

  --- Dossier src : Dossier de rangement de mes sous-dossiers ---
  --- Dossier src/database ---

- schema.sql : script SQL pour créer la base de données et les tables nécessaires
- dbtrigger.sql : script SQL pour créer les triggers de la base de données

  --- Dossier src/models ---

- MoniteurModel.go : définit le modèle de données pour les moniteurs

  --- Dossier src/repos ---

- pg.go : gestion de la connexion à la base de données PostgreSQL
- MoniteurRepos.go : dépôt pour gérer les opérations sur les moniteurs
- demo.go : code de démarrage pour le dépôt initial dans l'application

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

Notes de cours pour le langage Go

- https://gowebexamples.com/hello-world/

/_ Problematique rencontrée _/

- Probleme d'organisation des fichiers (résolu en créant un dossier src pour ranger les sous-dossiers)
- Problème avec Air : ne trouvait pas le fichier main.go (résolu en configurant le chemin dans .air.toml)
- Problème avec la connexion à la base de données PostgreSQL (résolu en vérifiant les paramètres de connexion et en s'assurant que le service PostgreSQL est en cours d'exécution)





Voici des explications claires sur les notions que tu as rencontrées dans ton code, avec un extrait et référence officielle pour chacune :

1. pgx
Qu'est-ce que pgx ?

pgx est un pilote PostgreSQL écrit entièrement en Go. Il offre une interface native haute performance pour PostgreSQL, en exposant des fonctionnalités spécifiques à ce SGBD (comme LISTEN/NOTIFY, COPY), tout en pouvant également être utilisé comme driver compatible database/sql.

Pourquoi l'utiliser ? - Pour un accès efficace à la base avec support spécialisé PostgreSQL.

Extrait doc officielle :
"pgx driver is a low-level, high performance interface that exposes PostgreSQL-specific features such as LISTEN/NOTIFY and COPY. It also includes an adapter for the standard database/sql interface."[pgx github officiel]

2. context (abrégé ctx dans Go)
Qu'est-ce que le contexte ?

context.Context permet de transmettre autour d’une requête des informations comme un délai d’expiration (timeout), une annulation, et des métadonnées. Il est utilisé pour gérer proprement la durée de vie d’opérations asynchrones ou dépendantes de ressources.

Pourquoi c’est important ?

Cela permet d'éviter les fuites de goroutines, d'interrompre des requêtes longues, et de propager des signaux d’annulation dans toute la chaîne d’appels.

Doc officielle :
"The Context type carries deadlines, cancelation signals, and other request-scoped values across API boundaries and goroutines."[golang context pkg]

3. Handler / HandlerFunc
Définition :

En Go, un Handler est une interface HTTP centrale qui gère une requête HTTP et prépare une réponse. Son rôle est d’exécuter la logique métier correspondante.

Un HandlerFunc est une fonction avec la signature func(ResponseWriter, *Request) qui est convertible en Handler.

Pourquoi utiliser ces abstractions ?

Elles permettent de composer et d’enchaîner des traitements HTTP de façon propre et modulaire - comme un middleware ou un routeur.

Doc officielle :
"Handler is an interface that responds to an HTTP request. HandlerFunc is a type that allows using ordinary functions as HTTP handlers."[net/http package]

Pour commencer avec pgx, voici un extrait d’exemple officiel :

go
conn, err := pgx.Connect(context.Background(), "postgres://user:pass@localhost/db")
if err != nil {
   // gérer erreur
}
defer conn.Close(context.Background())

var name string
err = conn.QueryRow(context.Background(), "SELECT name FROM table WHERE id=$1", 42).Scan(&name)
Cela montre la liaison directe entre pgx, context, et les requêtes SQL.

Sources :

pgx GitHub - PostgreSQL Driver and Toolkit

Go context package

Go net/http package - Handler

