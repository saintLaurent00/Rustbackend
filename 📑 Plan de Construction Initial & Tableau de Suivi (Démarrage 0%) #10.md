### Clé de Lecture du Pipeline Fondamental
Avant de coder la première ligne, voici le chemin précis que va emprunter chaque octet de donnée. Ce flux dicte l'ordre logique dans lequel les fichiers doivent être créés.

```
[Mire UI Connexion] ──► gateway-auth-go (Vérifie platform_config.go)
                                │ (Si Valide : Génère JWT)
                                ▼
[Configuration BI]  ──► metadata-semantic-go (Génère le Graphe via le DAO Ent)
                                │ (Pousse la requête + la règle RLS dans Valkey)
                                ▼
[Moteur de Calcul]  ──► query-engine-rust (Prend le ticket dans Valkey)
                                │── 1. Évalue le template (Tera / Jinja2)
                                │── 2. Parse l'AST (sqlparser) et injecte le RLS
                                └── 3. Exécute le Pushdown sur le DWH
                                if Cache-Miss ──► Stocke dans Parquet / ZSTD
```

---

### Tableau de Bord de Suivi Rigoureux (Toutes Tâches à 0%)
| ID  | Module / Composant                  | Fichier Cible Précis                                      | Statut      | Avancement | Dépendance Bloquante |
|-----|-------------------------------------|-----------------------------------------------------------|-------------|------------|----------------------|
| 01  | Configuration Centrale              | /platform_config.go                                       | 🟢 VALIDÉE  | 100%       | Aucune               |
| 02  | Initialisation Docker               | /docker-compose.yml                                       | 🟢 VALIDÉE  | 100%       | ID 01                |
| 03  | Génération Schémas DAO IAM          | /gateway-auth-go/ent/schema/                              | 🔴 À lancer | 0%         | ID 02                |
| 04  | Module Auth Locale                  | /gateway-auth-go/internal/auth/provider/local.go           | 🔴 À lancer | 0%         | ID 03                |
| 05  | Module Auth LDAP                    | /gateway-auth-go/internal/auth/provider/ldap.go            | 🔴 À lancer | 0%         | ID 03                |
| 06  | Module Auth SSO                     | /gateway-auth-go/internal/auth/provider/sso.go             | 🔴 À lancer | 0%         | ID 03                |
| 07  | Routeur / Validateur JWT            | /gateway-auth-go/internal/auth/handler.go                  | 🔴 À lancer | 0%         | ID 04, 05, 06         |
| 08  | Génération Schémas DAO BI           | /metadata-semantic-go/ent/schema/                          | 🔴 À lancer | 0%         | ID 02                |
| 09  | Validateur Droits Sémantiques       | /metadata-semantic-go/internal/permission/                 | 🔴 À lancer | 0%         | ID 08                |
| 10  | Moteur de Template SQL              | /query-engine-rust/src/parser/templating.rs                | 🔴 À lancer | 0%         | ID 02 (Valkey)        |
| 11  | Modificateur AST (RLS)              | /query-engine-rust/src/parser/rls.rs                       | 🔴 À lancer | 0%         | ID 10                |
| 12  | Pooler de Connexions DWH            | /query-engine-rust/src/executor/pushdown.rs                | 🔴 À lancer | 0%         | ID 11                |
| 13  | Capture Visuelle Dashboard          | /orchestrator-sync-rust/src/renderer/snapshot.rs           | 🔴 À lancer | 0%         | ID 02                |
| 14  | Usine d'Expédition Mails            | /orchestrator-sync-rust/src/mailing/dispatcher.rs          | 🔴 À lancer | 0%         | ID 13                |

---

### Fichiers Initialisés ✅

#### platform_config.go
- Configuration centralisée avec variables d'env
- Stocke URLs DB, Redis, JWT Secret, Mail Provider

#### docker-compose.yml
- **PostgreSQL** 15-Alpine (IAM + Métadonnées)
- **Valkey** (Cache + Message Queue)
- **MailDev** (SMTP local pour tests)
- Healthchecks + volumes persistants

#### .gitignore
- Binaires Go/Rust, logs, .env, cache

#### README.md
- Architecture de haut niveau
- Démarrage rapide
- Plan de construction

#### CONTRIBUTING.md
- Workflow git (branches, commits)
- Structure des répertoires
- Checklist de validation

---

### Prochaine Étape Immédiate (Étape B)
**Créer l'infrastructure IAM :**
1. Schémas Ent (ID 03) → User, Role, Permission
2. Authentification Locale (ID 04)
3. Authentification LDAP (ID 05)
4. Authentification SSO (ID 06)
5. Handler JWT (ID 07)

➡️ Vérifie que `docker compose up -d` démarre sans erreur, puis lance l'ID 03 !
