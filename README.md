# 📊 BI Analytics Platform Backend

Ce projet est une plateforme de Business Intelligence industrielle, modulaire et hautement sécurisée, développée en **Go** et **Rust**.

## 🏗️ Architecture

La plateforme est composée de 5 microservices :
* **`gateway-auth-go`** : Gestion des accès (Local, LDAP, SSO) et validation JWT.
* **`metadata-semantic-go`** : Gestion des datasets, charts et sécurité RLS. Secrets chiffrés en AES-GCM.
* **`query-engine-rust`** : Moteur de calcul, injection RLS via AST SQL et templating Tera.
* **`render-worker`** : Rendu visuel headless via Chromium.
* **`orchestrator-sync-rust`** : Tâches de fond et envoi de rapports par mail.

---

## 🚀 Démarrage Rapide

1. **Générer le fichier d'environnement :**
   ```bash
   cp .env.example .env
   ```

2. **Lancer l'infrastructure et les services :**
   ```bash
   docker-compose up --build
   ```

---

## 🧪 Comment tester l'application ?

Nous avons mis à disposition un script de vérification automatique pour s'assurer que votre environnement est prêt.

### 1. Script de test global
Ce script vérifie la compilation de chaque service, l'état de l'infrastructure Docker et la connectivité au broker Valkey.

```bash
./scripts/test_platform.sh
```

### 2. Accès aux interfaces de test
* **API Gateway** : `http://localhost:8080/api/v1/auth/login`
* **Service Métadonnées** : `http://localhost:8081/api/v1/health`
* **MailDev (Interface Mail)** : `http://localhost:1080` (Pour visualiser les rapports envoyés)
* **Valkey** : `redis://localhost:6379`

### 3. Test des Schémas (Go)
Pour chaque service Go, vous pouvez vérifier l'intégrité des schémas générés :
```bash
cd gateway-auth-go
go test ./ent/...
```

### 4. Test du Moteur de Calcul (Rust)
Vous pouvez lancer les vérifications de syntaxe et les tests unitaires :
```bash
cd query-engine-rust
cargo test
```

---

## 🛠️ Principes de Développement
Toute l'architecture repose sur la **Clean Architecture** et le découplage via des files d'attente de messages. Pour plus de détails sur chaque service, consultez les fichiers `README.md` dans leurs répertoires respectifs.
