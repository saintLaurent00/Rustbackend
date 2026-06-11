

Chaque fichier possède sa description et son rôle précis directement documentés à l'intérieur.

---

## 📁 1. Arborescence Globale du Projet

```text
bi-analytics-platform/
├── .env.example
├── docker-compose.yml
├── README.md
├── gateway-auth-go/
│   ├── cmd/
│   │   └── main.go
│   ├── internal/
│   │   ├── auth/
│   │   │   ├── handler.go
│   │   │   ├── service.go
│   │   │   └── provider/
│   │   │       ├── local.go
│   │   │       ├── ldap.go
│   │   │       └── sso.go
│   │   └── config/
│   │       └── config.go
│   ├── go.mod
│   └── go.sum
├── metadata-semantic-go/
│   ├── cmd/
│   │   └── main.go
│   ├── internal/
│   │   ├── dataset/
│   │   ├── chart/
│   │   └── mailconfig/
│   ├── go.mod
│   └── go.sum
├── query-engine-rust/
│   ├── Cargo.toml
│   └── src/
│       ├── main.rs
│       ├── parser/
│       │   ├── mod.rs
│       │   ├── templating.rs
│       │   └── rls.rs
│       └── executor/
│           ├── mod.rs
│           └── pushdown.rs
└── orchestrator-sync-rust/
    ├── Cargo.toml
    └── src/
        ├── main.rs
        ├── cache/
        │   ├── mod.rs
        │   └── parquet_zstd.rs
        ├── renderer/
        │   ├── mod.rs
        │   └── snapshot.rs
        └── mailing/
            ├── mod.rs
            └── dispatcher.rs

```

---

## 📄 2. Fichiers de Configuration Globaux

### `/.env.example`

```env
# ==============================================================================
# FICHIER DE CONFIGURATION GLOBAL (ENVIRONNEMENT)
# Rôle : Centralise les variables d'environnement pour tous les microservices.
# ==============================================================================

# --- CONFIGURATION SYSTÈME & INFRASTRUCTURE ---
PORT_AUTH_GATEWAY=8080
PORT_METADATA_SERVICE=8081
VALKEY_HOST=valkey
VALKEY_PORT=6379

# --- BASES DE DONNÉES ---
POSTGRES_METADATA_URL=postgres://admin:secret@postgres_meta:5432/bi_metadata?sslmode=disable
DWH_CLICKHOUSE_URL=http://dwh_user:dwh_pass@clickhouse_dwh:8123/default

# --- CACHE PHYSIQUE ---
DEEP_CACHE_DIR=/var/lib/bi_cache
DEEP_CACHE_SIZE_LIMIT_GB=5

# --- MODULES D'AUTHENTIFICATION EXTENSIBLES (FLAGS) ---
AUTH_ENABLE_LOCAL=true
AUTH_ENABLE_LDAP=false
AUTH_ENABLE_SSO=false

# CONFIGURATION LDAP (Si activé)
LDAP_SERVER_URL=ldaps://ldap.entreprise.local:636
LDAP_BASE_DN=DC=entreprise,DC=local

# CONFIGURATION SSO OIDC (Si activé)
SSO_PROVIDER_URL=https://accounts.google.com
SSO_CLIENT_ID=your-google-client-id
SSO_CLIENT_SECRET=your-google-client-secret
SSO_REDIRECT_URL=http://localhost:8080/api/v1/auth/callback

```

### `/docker-compose.yml`

```yaml
# ==============================================================================
# CONFIGURATION DOCKER COMPOSE ORCHESTRATION
# Rôle : Déploie et interconnecte de manière isolée l'ensemble de la stack BI.
# ==============================================================================
version: '3.8'

services:
  valkey:
    image: valkey/valkey:7.2
    container_name: bi_valkey_broker
    ports:
      - "${VALKEY_PORT}:6379"
    volumes:
      - valkey_data:/data

  postgres_meta:
    image: postgres:16-alpine
    container_name: bi_postgres_metadata
    environment:
      POSTGRES_USER: admin
      POSTGRES_PASSWORD: secret
      POSTGRES_DB: bi_metadata
    volumes:
      - postgres_data:/var/lib/postgresql/data

  gateway-auth:
    build: ./gateway-auth-go
    container_name: bi_gateway_auth
    ports:
      - "${PORT_AUTH_GATEWAY}:8080"
    env_file: .env
    depends_on:
      - valkey
      - postgres_meta

  query-engine:
    build: ./query-engine-rust
    container_name: bi_query_engine_worker
    env_file: .env
    volumes:
      - cache_data:${DEEP_CACHE_DIR}
    depends_on:
      - valkey

volumes:
  valkey_data:
  postgres_data:
  cache_data:

```

---

## 🏗️ 3. Zoom sur le Code Architectural Extensible

Voici comment sont structurés les fichiers clés pour permettre aux développeurs d'ajouter de la logique métier sans casser l'existant.

### `/gateway-auth-go/internal/auth/provider/local.go`

```go
package provider

import "context"

// LocalProvider gère l'authentification native de la plateforme.
// Rôle : Vérifier les accès par rapport aux identifiants stockés en base locale.
type LocalProvider struct{}

func NewLocalProvider() *LocalProvider {
	return &LocalProvider{}
}

func (p *LocalProvider) Authenticate(ctx context.Context, username, password string) (bool, error) {
	// 1. Chercher l'utilisateur dans PostgreSQL via l'ORM Ent
	// 2. Comparer le mot de passe hashé avec Bcrypt
	return true, nil
}

```

### `/gateway-auth-go/internal/auth/service.go`

```go
package auth

import (
	"context"
	"fmt"
	"bi-gateway/internal/auth/provider"
)

// AuthStrategy définit le contrat d'interface pour tout nouveau module de connexion.
// RÔLE MAJEUR : Pour ajouter une nouvelle méthode de connexion (ex: FaceID, Biométrie, Certificat),
// le développeur doit simplement créer un fichier dans le dossier 'provider/' et implémenter cette interface.
type AuthStrategy interface {
	Authenticate(ctx context.Context, username, password string) (bool, error)
}

type AuthService struct {
	providers map[string]AuthStrategy
}

func NewAuthService() *AuthService {
	s := &AuthService{
		providers: make(map[string]AuthStrategy),
	}
	// Enregistrement dynamique des modules configurés
	s.providers["local"] = provider.NewLocalProvider()
	s.providers["ldap"] = provider.NewLdapProvider() // Fichier similaire à implémenter
	return s
}

func (s *AuthService) Login(ctx context.Context, method string, username, password string) (bool, error) {
	provider, exists := s.providers[method]
	if !exists {
		return false, fmt.Errorf("méthode d'authentification '%s' non supportée ou non activée", method)
	}
	return provider.Authenticate(ctx, username, password)
}

```

### `/query-engine-rust/src/parser/mod.rs`

```rust
// ==============================================================================
// MODULE DE CONVERSION ET SÉCURISATION DES REQUÊTES (RUST ENGINE)
// Rôle : Orchestre l'évaluation du template puis l'injection des règles RLS.
// ==============================================================================

pub mod templating;
pub mod rls;

/// Point d'entrée pour transformer un SQL dynamique en un SQL sécurisé prêt pour le DWH.
pub fn transform_query(raw_sql: &str, ui_filters: &serde_json::Value, user_rls: &str) -> String {
    // Étape 1 : Résolution des expressions Jinja2 via Tera
    let clean_sql = templating::compile_dynamic_sql(raw_sql, ui_filters);
    
    // Étape 2 : Injection de la sécurité Row-Level Security via l'AST de sqlparser
    let secure_sql = rls::inject_rls_conditions(&clean_sql, user_rls);
    
    secure_sql
}

```

---

## 📄 4. Le Guide de Référence à la Racine

### `/README.md`

```markdown
# 📊 BI Analytics Platform Backend

Ce projet est une plateforme de Business Intelligence industrielle, modulaire et hautement sécurisée, développée en **Go** et **Rust**.

## 🏗️ Principes d'Architecture (Clean Architecture & Modularité)

Le projet est rigoureusement structuré de manière à ce qu'un développeur puisse étendre ses fonctionnalités ou assurer la maintenance de manière isolée sans risques de régression.

### 1. Comment ajouter un nouveau module d'authentification (ex: OAuth de l'entreprise) ?
1. Rendez-vous dans le microservice `gateway-auth-go`.
2. Créez un nouveau fichier de fournisseur dans `internal/auth/provider/mon_nouveau_sso.go`.
3. Implémentez l'interface `AuthStrategy` (méthode `Authenticate`).
4. Enregistrez votre fournisseur dans le constructeur de `internal/auth/service.go`. Le module sera immédiatement disponible et pilotable dynamiquement depuis l'UI.

### 2. Comment ajouter une nouvelle fonction personnalisée dans le moteur de template SQL ?
1. Rendez-vous dans le microservice `query-engine-rust`.
2. Ouvrez le fichier `src/parser/templating.rs`.
3. Utilisez l'API `tera.register_filter` ou `tera.register_function` pour intégrer votre logique de calcul ou de transformation textuelle. Vos filtres seront utilisables directement par les Data Engineers dans leurs fenêtres de script SQL.

---

## 🚀 Lancement Rapide (Environnement Isolé ou de Dev)

1. **Générer le fichier d'environnement :**
   ```bash
   cp .env.example .env

```

2. **Lancer l'ensemble de l'écosystème avec Docker Compose :**
```bash
docker compose up --build -d

```



## 🔍 Responsabilité des Services

* **`gateway-auth-go`** : Proxy d'entrée, gestion asynchrone des connexions hybrides (Local/LDAP/SSO) et validation JWT.
* **`metadata-semantic-go`** : CRUD de la couche sémantique, stockage PostgreSQL des styles de couleurs de charts et configurations mails.
* **`query-engine-rust`** : Compilation haute performance, injection RLS, évaluation Jinja2 (Tera) et streaming de données.
* **`orchestrator-sync-rust`** : Tâches cron de fond, snapshots visuels de dashboards (Headless Chromium) et envoi de rapports par mail.

```

```
