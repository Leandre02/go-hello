# 📡 Service de Monitoring / Monitoring Service

![Go](https://img.shields.io/badge/Go-00ADD8?style=for-the-badge&logo=go&logoColor=white)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-4169E1?style=for-the-badge&logo=postgresql&logoColor=white)
![Docker](https://img.shields.io/badge/Docker-2496ED?style=for-the-badge&logo=docker&logoColor=white)
![HTML5](https://img.shields.io/badge/HTML5-E34F26?style=for-the-badge&logo=html5&logoColor=white)
![CSS3](https://img.shields.io/badge/CSS3-1572B6?style=for-the-badge&logo=css3&logoColor=white)
![JavaScript](https://img.shields.io/badge/JavaScript-F7DF1E?style=for-the-badge&logo=javascript&logoColor=black)

> 🇫🇷 Un service web de monitoring HTTP/HTTPS codé en Go, avec persistance PostgreSQL et interface web en temps réel.
>
> 🇬🇧 A Go-powered HTTP/HTTPS monitoring service with PostgreSQL persistence and a real-time web dashboard.

🌐 **Live demo:** [monitoring.memberbase.ca](https://monitoring.memberbase.ca/)

---

## 📖 Description

### 🇫🇷 Français

**Service de Monitoring** est une application fullstack minimaliste permettant de vérifier la disponibilité de sites web en temps réel. L'interface web communique avec une API REST Go, et les résultats sont persistés dans PostgreSQL. Un système de triggers SQL détecte automatiquement les transitions UP/DOWN et génère des alertes.

> 📚 Projet de session A25 — Technologies Émergentes  
> 👤 Par : Leandre Kanmegne

### 🇬🇧 English

**Monitoring Service** is a minimal fullstack application for checking website availability in real time. The web interface communicates with a Go REST API, and results are persisted in PostgreSQL. SQL triggers automatically detect UP/DOWN transitions and generate alerts.

> 📚 Session project A25 — Emerging Technologies  
> 👤 By: Leandre Kanmegne

---

## ✨ Features / Fonctionnalités

| 🇫🇷 | 🇬🇧 |
|---|---|
| ✅ Vérification HTTP/HTTPS en temps réel | ✅ Real-time HTTP/HTTPS status checks |
| 📊 Historique des statuts par site | 📊 Status history per monitored site |
| ⚡ Latence mesurée à chaque requête | ⚡ Latency measured on every request |
| 🔔 Alertes automatiques UP/DOWN (triggers SQL) | 🔔 Automatic UP/DOWN alerts (SQL triggers) |
| 🗑️ Réinitialisation complète de l'historique | 🗑️ Full history reset |
| 🔄 Auto-ping configurable (setInterval) | 🔄 Configurable auto-ping (setInterval) |
| 🐳 Environnement Docker complet (dev + prod) | 🐳 Full Docker environment (dev + prod) |
| 🧪 Tests unitaires avec race detector | 🧪 Unit tests with race detector |

---

## 🏗️ Architecture

```
📦 go-hello
├── 🖥️  src/
│   ├── cmd/server/main.go        → Entrypoint HTTP
│   ├── internal/
│   │   ├── middleware/logger.go  → Logging middleware
│   │   ├── models/types.go       → Structs (Moniteur, Statut)
│   │   ├── routes/router.go      → REST API endpoints
│   │   └── services/             → HTTP checker + tests
│   ├── repos/                    → Interface + implémentation PostgreSQL
│   └── database/
│       ├── init.sql              → Schéma & tables / Schema & tables
│       └── dbtrigger.sql         → Triggers alertes UP/DOWN
├── 🌐  web/                      → Front statique / Static frontend
├── 🐳  docker-compose.dev.yml    → Environnement dev
├── 🐳  dockerfile                → Build prod multi-stage
└── ⚙️  .env.example              → Variables d'environnement
```

---

## 🌐 API Endpoints

| Méthode | Route | Description 🇫🇷 | Description 🇬🇧 |
|---|---|---|---|
| `POST` | `/api/verifier` | Vérifier une URL | Check a URL |
| `GET` | `/api/resultats?limit=N` | Lister les résultats | List results |
| `DELETE` | `/api/resultats` | Vider l'historique | Clear history |
| `GET` | `/api/etat` | Santé de l'API | API health check |

---

## ⚙️ Prérequis / Prerequisites

- 🐳 [Docker Desktop](https://www.docker.com/products/docker-desktop/) installé / installed
- 🔧 Docker Compose disponible / available (inclus dans Docker Desktop / included)
- 🔌 Port `8080` libre sur la machine / free on the machine *(configurable dans / in `.env`)*
- 📄 Fichier `.env` configuré / configured *(voir / see `.env.example`)*

---

## 🚀 Démarrage / Getting Started

### 1. Cloner le projet / Clone the project
```bash
git clone https://github.com/Leandre02/go-hello.git
cd go-hello
```

### 2. Configurer l'environnement / Set up environment
```bash
cp .env.example .env
# Modifier les variables selon votre config / Edit variables as needed
```

### 3. Lancer / Start

```bash
# WSL
docker-compose -f docker-compose.dev.yml up

# Fedora / Linux
docker compose -f docker-compose.dev.yml up
```

---

## 🛠️ Commandes utiles / Useful Commands

```bash
# ▶️  Démarrer / Start
docker compose -f docker-compose.dev.yml up

# ⏹️  Arrêter / Stop
docker compose -f docker-compose.dev.yml down

# 📋  Voir les logs / View logs
docker compose -f docker-compose.dev.yml logs -f app

# 🧪  Tests avec race detector / Tests with race detector
go test ./... -race
```

---

## 🗄️ Base de données / Database

| Table | Description |
|---|---|
| `monitoring.moniteurs` | Sites surveillés / Monitored sites |
| `monitoring.statuts` | Historique des vérifications / Check history |
| `monitoring.alertes` | Alertes UP/DOWN générées / Generated UP/DOWN alerts |
| `monitoring.v_dernier_statut` | Vue : dernier statut par site / Last status per site |

---

## 🌐 Accès au projet / Project Access

> 🔗 **[monitoring.memberbase.ca](https://monitoring.memberbase.ca/)**

---

## 👤 Auteur / Author

**Leandre Kanmegne** — [GitHub](https://github.com/Leandre02)
