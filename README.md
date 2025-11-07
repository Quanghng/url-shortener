# URL Shortener (Go)

Un service web performant de raccourcissement et de gestion d’URLs, entièrement développé en Go.
Ce projet met en œuvre des concepts avancés de Go pour la conception d’un système concurrent, combinant API REST, traitements asynchrones et interface en ligne de commande (CLI), le tout dans une architecture claire et modulaire.

---

## 🚀 Vue d’ensemble

L’application **URL Shortener** permet de transformer une URL longue en un lien court et unique.
Lorsqu’un utilisateur accède à ce lien, le service le redirige instantanément vers l’URL originale tout en enregistrant le clic en tâche de fond.

Un **moniteur d’URLs** intégré vérifie également périodiquement la disponibilité des liens originaux et notifie tout changement d’état dans les logs du serveur.

L’interaction se fait à la fois via une **API RESTful** et une **interface CLI** complète.

---

## 🧠 Fonctionnalités principales

### 1. Raccourcissement d’URLs

* Génération de codes courts uniques (6 caractères alphanumériques).
* Gestion automatique des collisions grâce à une logique de retry.

### 2. Redirection instantanée

* Redirection immédiate (HTTP 302) vers l’URL originale.
* Enregistrement asynchrone des clics via **Goroutines** et **Channels bufferisés**, garantissant que la redirection n’est jamais bloquée.

### 3. Surveillance des URLs

* Vérification périodique (intervalle configurable) de la disponibilité des URLs longues.
* Notifications dans les logs lors d’un changement d’état (accessible ↔ inaccessible).

### 4. API REST (framework Gin)

* `GET /health` → Vérifie l’état du service.
* `POST /api/v1/links` → Crée une nouvelle URL courte (`{"long_url": "..."}`).
* `GET /{shortCode}` → Redirige vers l’URL originale et déclenche l’enregistrement du clic.
* `GET /api/v1/links/{shortCode}/stats` → Affiche les statistiques d’un lien (nombre total de clics).

### 5. Interface CLI (Cobra)

* `./url-shortener run-server` → Lance le serveur, les workers et le moniteur d’URLs.
* `./url-shortener create --url="https://..."` → Crée une URL courte depuis la ligne de commande.
* `./url-shortener stats --code="xyz123"` → Affiche les statistiques d’un lien donné.
* `./url-shortener migrate` → Exécute les migrations pour la base de données.

### 6. Fonctionnalités avancées (optionnelles)

* Alias personnalisés pour les URLs.
* Expiration des liens après une durée définie.
* Limitation de débit (rate limiting) par adresse IP pour la création de liens.

---

## 🏗️ Architecture du projet

Une structure modulaire et claire, séparant les responsabilités entre les différentes couches :

```
url-shortener/
├── cmd/
│   ├── root.go                 # Commande racine CLI (Cobra)
│   ├── server/server.go        # Logique de la commande 'run-server'
│   └── cli/
│       ├── create.go           # Commande 'create'
│       ├── stats.go            # Commande 'stats'
│       └── migrate.go          # Commande 'migrate'
├── internal/
│   ├── api/handlers.go         # Handlers HTTP (Gin)
│   ├── models/
│   │   ├── link.go             # Modèle GORM 'Link'
│   │   └── click.go            # Modèle GORM 'Click'
│   ├── services/
│   │   ├── link_service.go     # Logique métier pour les liens
│   │   └── click_service.go    # Logique métier pour les clics (optionnelle)
│   ├── workers/click_worker.go # Worker asynchrone pour les clics
│   ├── monitor/url_monitor.go  # Moniteur d’état des URLs
│   ├── config/config.go        # Chargement de configuration (Viper)
│   └── repository/
│       ├── link_repository.go  # Accès aux données 'Link'
│       └── click_repository.go # Accès aux données 'Click'
├── configs/config.yaml         # Fichier de configuration par défaut
├── go.mod                      # Dépendances du module Go
├── go.sum                      # Sommes de contrôle
└── README.md                   # Documentation du projet
```

---

## ⚙️ Installation et utilisation

### 1. Cloner et préparer le projet

```bash
git clone https://github.com/axellelanca/urlshortener.git
cd urlshortener
go mod tidy
```

### 2. Compiler le binaire

```bash
go build -o url-shortener
```

### 3. Initialiser la base de données

```bash
./url-shortener migrate
```

Cette commande crée la base de données SQLite `url_shortener.db` et ses tables.

### 4. Lancer le serveur

```bash
./url-shortener run-server
```

Lance :

* le serveur API REST,
* les workers asynchrones d’enregistrement des clics,
* et le moniteur d’URLs.

---

## 🧩 Exemples d’utilisation

### Créer une URL courte

```bash
./url-shortener create --url="https://www.example.com/ma-longue-url"
```

Sortie :

```
URL courte créée avec succès:
Code: XYZ123
URL complète: http://localhost:8080/XYZ123
```

### Accéder à l’URL courte

Ouvrez votre navigateur et allez sur `http://localhost:8080/XYZ123`.
Vous serez redirigé immédiatement vers l’URL originale, et un clic sera enregistré en arrière-plan.

### Consulter les statistiques d’un lien

```bash
./url-shortener stats --code="XYZ123"
```

Sortie :

```
Statistiques pour le code court: XYZ123
URL longue: https://www.example.com/ma-longue-url
Total de clics: 1
```

### Vérifier l’état du service (API Health Check)

```bash
curl http://localhost:8080/health
```

Réponse :

```json
{"status":"ok"}
```

---

## 🔍 Moniteur d’URLs

Le moniteur intégré vérifie régulièrement la disponibilité de chaque URL longue.
Si un lien devient inaccessible, une notification est générée dans les logs :

```
[NOTIFICATION] Le lien XYZ123 (https://url-hors-ligne.com) est passé de ACCESSIBLE à INACCESSIBLE !
```

---

## 🛑 Arrêter le serveur

Pour stopper le service, appuyez sur `Ctrl + C` dans le terminal où il est en cours d’exécution.
L’arrêt propre du serveur sera confirmé dans les logs.

---

## 🧰 Technologies utilisées

* **Go** (Goroutines, Channels, Interfaces)
* **Gin** – Framework web RESTful
* **GORM** – ORM pour SQLite
* **Cobra** – Framework CLI
* **Viper** – Gestion de la configuration

---

## 📚 Concepts clés

* Traitements asynchrones non bloquants.
* Architecture propre basée sur les patterns Repository et Service.
* Intégration cohérente entre API et CLI.
* Surveillance et notifications concurrentes.
* Bonne gestion des erreurs et des configurations.

---

## 📄 Licence

Projet distribué sous licence **MIT**.
