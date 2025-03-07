# Quizzy (Backend)

## Configuration

Avant de démarrer le projet, il est nécessaire d'ajouter les variables d'environnement dans la configuration.

### 1. Créer le fichier `.env`

Ajoutez les variables suivantes dans un fichier `.env` à la racine du projet :

```
APP_ADDR=127.0.0.1:8000
APP_ENV=DEVELOPMENT
APP_FIREBASE_CONF_FILE=firebase.conf.json
APP_REDIS_URI=redis://localhost:6379
```

### 2. Description des variables d'environnement

| Variable | Description |
|----------|-------------|
| `APP_ADDR` | Adresse sur laquelle l'application démarre |
| `APP_FIREBASE_CONF_FILE` | Fichier de configuration Firestore |
| `APP_ENV` | Mode de démarrage (`DEVELOPMENT`, `TEST`, `PRODUCTION`) |
| `APP_REDIS_URI` | URL du service Redis |

### 3. Créer le fichier `firebase.conf.json`

Créez un fichier `firebase.conf.json` à la racine du projet et implémentez la partie Firestore selon votre configuration Firebase.

## Lancer le projet

Suivez ces étapes pour démarrer l'application :

### 1. Installer les dépendances
```sh
go mod tidy
```

### 2. Exporter les variables d'environnement (optionnel si le fichier `.env` est utilisé avec un loader)
Sur **Linux/macOS** :
```sh
export APP_ADDR=127.0.0.1:8000
export APP_ENV=DEVELOPMENT
export APP_FIREBASE_CONF_FILE=firebase.conf.json
export APP_REDIS_URI=redis://localhost:6379
```

Sur **Windows (PowerShell)** :
```powershell
$env:APP_ADDR="127.0.0.1:8000"
$env:APP_ENV="DEVELOPMENT"
$env:APP_FIREBASE_CONF_FILE="firebase.conf.json"
$env:APP_REDIS_URI="redis://localhost:6379"
```

### 3. Lancer les services avec Docker
Assurez-vous d'avoir Docker installé, puis exécutez la commande suivante :
```sh
docker-compose up -d
```

### 4. Démarrer le serveur
```sh
go run main.go
```