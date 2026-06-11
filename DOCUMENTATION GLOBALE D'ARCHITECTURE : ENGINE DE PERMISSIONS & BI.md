Pour une plateforme de Business Intelligence de niveau entreprise, la gestion des permissions ne peut pas se limiter à un simple contrôle "Tout ou Rien". Nous devons implémenter une **Matrice de Permissions Granulaire et Sémantique (ACL / RBAC Graph)**.

Un rôle (ex: `Finance_Manager`) doit pouvoir posséder des permissions sémantiques distinctes sur chaque objet :

* `CAN_READ` (Voir le graphique/dashboard).
* `CAN_EDIT` (Modifier la structure ou les couleurs `column_styles`).
* `CAN_SHARE` (Partager ou planifier l'envoi par mail).
* `CAN_GRANT` (Donner des droits à d'autres rôles).

Puisque nous utilisons **Ent**, ce framework prend tout son sens ici : il permet de modéliser ces relations complexes sous forme de **Graphes**. Les permissions, les rôles, les groupes et les objets (Charts, Dashboards, Datasets) deviennent des nœuds connectés par des arêtes (*Edges*) dans notre base PostgreSQL.

Voici la documentation d’architecture logicielle complète, définitive, restructurée pour intégrer la couche DAO au cœur du système de permissions sémantiques.

---

# 📑 DOCUMENTATION GLOBALE D'ARCHITECTURE : ENGINE DE PERMISSIONS & BI

## 1. Plan Complet de la Structure du Projet

```text
bi-analytics-platform/
├── platform_config.go          # FICHIER MAÎTRE : Configuration statique et compilation de la stack
├── .env.example                # Modèle de variables d'environnement système
├── docker-compose.yml          # Orchestration des conteneurs isolés de production
├── README.md                   # Guide d'onboarding et d'extension pour les développeurs
│
├── gateway-auth-go/            # MICROSERVICE 1 : API Gateway, Hybride Auth & IAM
│   ├── cmd/
│   │   └── main.go             # Point d'entrée, initialisation du serveur Fiber
│   ├── ent/                    # DAO IAM : Généré automatiquement par Entgo
│   │   ├── schema/             # Définition Code-First des entités de sécurité
│   │   │   ├── user.go         # Propriétés utilisateurs (Email, Password hash, Status)
│   │   │   ├── group.go        # Regroupement logique d'utilisateurs (ex: "Département Finance")
│   │   │   └── role.go         # Rôles de sécurité (ex: "Data_Analyst", "Executive")
│   │   └── client.go           # Client DAO binaire pour l'authentification et les groupes
│   └── internal/
│       ├── config/
│       │   └── loader.go       # Chargeur des structures de platform_config.go
│       └── auth/
│           ├── handler.go      # Réceptionne et valide les requêtes de login de l'UI
│           ├── service.go      # Route vers le bon provider (Local, LDAP, SSO)
│           └── provider/       
│               ├── local.go  
│               ├── ldap.go   
│               └── sso.go    
│
├── metadata-semantic-go/       # MICROSERVICE 2 : Couche Sémantique, Objets & Permissions ACL
│   ├── cmd/
│   │   └── main.go             # Point d'entrée du service
│   ├── ent/                    # DAO SÉMANTIQUE : Relations d'accès en Graphe
│   │   ├── schema/             
│   │   │   ├── dataset.go      # Abstractions de tables, jointures et scripts SQL
│   │   │   ├── chart.go        # Graphiques, configurations et JSON "column_styles"
│   │   │   ├── dashboard.go    # Collection de charts mis en page
│   │   │   ├── rls_rule.go     # Prédicats SQL de restriction de lignes (Row-Level Security)
│   │   │   └── permission.go   # NŒUD ACL : Fait le pont sémantique entre un Rôle/Groupe et un Objet
│   │   └── client.go           # Moteur d'interrogation de la matrice de droits
│   └── internal/
│       ├── dataset/            # Contrôleurs de gestion des modèles de données
│       ├── chart/              # Contrôleurs d'édition des graphiques et de leurs styles
│       ├── rls/                # Injection et validation des payloads RLS
│       └── permission/         # Évaluation sémantique des droits d'accès avant requête
│
├── query-engine-rust/          # MICROSERVICE 3 : Moteur de Calcul Analytique (Worker)
│   ├── Cargo.toml              
│   └── src/
│       ├── main.rs             # Écoute et dépile la file d'attente de requêtes Valkey
│       ├── config.rs           # Lit le fichier de configuration partagé
│       ├── parser/             
│       │   ├── mod.rs          
│       │   ├── templating.rs   # Moteur Tera (Jinja2) : Résolution des requêtes dynamiques
│       │   └── rls.rs          # Analyseur AST : Injection du filtre RLS validé par le DAO
│       └── executor/
│           └── pushdown.rs     # Exécution physique sur le Data Warehouse client
│
└── orchestrator-sync-rust/    # MICROSERVICE 4 : Planificateur, Deep Cache & Dispatcher Mail
    ├── Cargo.toml              
    └── src/
        ├── main.rs             # Daemon d'arrière-plan gérant les tâches chronométrées
        ├── cache/
        │   └── parquet_zstd.rs # Requêtage et écriture du Deep Cache via mmap
        ├── renderer/
        │   └── snapshot.rs     # Headless Chromium : Rendu visuel automatique des Dashboards
        └── mailing/
            └── dispatcher.rs   # Dispatcher de rapports visuels par e-mail (SMTP / API Gmail)

```

---

## 🔑 2. Conception du DAO Sémantique des Permissions (`/ent/schema/`)

Grâce à `Ent`, nous pouvons déclarer de manière limpide les privilèges fins qu'un **Rôle** ou qu'un **Groupe** possède sur un composant de la plateforme.

### Définition du Nœud de Permission : `permission.go`

Ce fichier définit précisément le niveau d'accès sémantique accordé sur un objet BI.

```go
package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Permission définit le schéma de la table des privilèges.
type Permission struct {
	ent.Schema
}

func (Permission) Fields() []ent.Field {
	return []ent.Field{
		// Définition sémantique fine des actions autorisées
		field.Enum("access_level").
			Values("CAN_READ", "CAN_EDIT", "CAN_SHARE", "CAN_GRANT").
			Default("CAN_READ"),
	}
}

// Edges définit les liaisons du Graphe. 
// Rôle : Relier cette permission à un Rôle (IAM) et aux objets sémantiques.
func (Permission) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("dataset", Dataset.Type).Unique(),
		edge.To("dashboard", Dashboard.Type).Unique(),
		edge.To("chart", Chart.Type).Unique(),
	}
}

```

### Définition de l'entité de restriction de données : `rls_rule.go`

Ce fichier est le DAO qui stocke la restriction SQL d'un utilisateur ou d'un rôle sur un dataset particulier.

```go
package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type RlsRule struct {
	ent.Schema
}

func (RlsRule) Fields() []ent.Field {
	return []ent.Field{
		field.String("name"),
		field.String("target_table"),
		// Exemple : "agence_id = {{ user.agence_id }}" ou "pays = 'CI'"
		field.String("sql_predicate"), 
	}
}

func (RlsRule) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("dataset", Dataset.Type).Ref("rls_rules").Unique(),
	}
}

```

---

## 🔄 3. Pipeline d'Interrogation du DAO Sémantique (`internal/permission/`)

Lorsqu’un utilisateur tente d'ouvrir ou de modifier un composant (ex: modifier la couleur d'un graphique), le service de métadonnées exécute une validation de graphe ultra-rapide.

Voici le code du validateur sémantique dans `/metadata-semantic-go/internal/permission/validator.go` :

```go
package permission

import (
	"context"
	"bi-metadata/ent"
	"bi-metadata/ent/permission"
	"bi-metadata/ent/chart"
)

type AccessValidator struct {
	entClient *ent.Client
}

func NewAccessValidator(client *ent.Client) *AccessValidator {
	return &AccessValidator{entClient: client}
}

// VerifyChartAccess vérifie si le rôle de l'utilisateur l'autorise à modifier ou lire un graphique.
func (v *AccessValidator) VerifyChartAccess(ctx context.Context, userRole string, chartID int, requiredLevel string) (bool, error) {
	// Requête de Graphe : On cherche s'il existe une permission associée au rôle de l'utilisateur,
	// pointant vers le graphique demandé, et ayant le niveau d'accès requis (ou supérieur).
	exists, err := v.entClient.Permission.
		Query().
		Where(
			permission.RoleName(userRole), // Filtre sur le rôle de la session
			permission.HasChartWith(chart.ID(chartID)), // Vérification du lien vers l'objet graphique
		).
		Where(
			// Vérifie la granularité sémantique (ex: CAN_EDIT)
			permission.AccessLevelEQ(permission.AccessLevel(requiredLevel)),
		).
		Exist(ctx)

	return exists, err
}

```

---

## 📜 4. Guide de Maintenance mis à jour (`README.md`)

```markdown
# 🚀 BI Analytics Platform Backend

Ce projet est une plateforme de Business Intelligence de classe entreprise. Il utilise **Go** et **Rust** combinés à un moteur DAO orienté graphes via **Ent** pour orchestrer la sécurité sémantique des données.

## 🔑 Fonctionnement de la Matrice Sémantique des Droits

Toutes nos entités (Utilisateurs, Groupes, Rôles, Datasets, Charts, Dashboards, Règles RLS) sont gérées sous forme de nœuds typés.

### 1. Comment ajouter un nouvel objet métier sécurisé (Ex: Un dossier de rapports) ?
1. Allez dans `metadata-semantic-go/ent/schema/`.
2. Créez le fichier `folder.go` en étendant `ent.Schema`.
3. Ajoutez une arête (`edge`) dans `permission.go` pointant vers votre nouvelle entité `Folder`.
4. Exécutez la commande de génération à la racine du microservice :
   ```bash
   go generate ./ent

```

Le framework va mettre à jour l'intégralité du client DAO. Les méthodes de vérification d'accès sémantiques (`CAN_READ`, `CAN_EDIT`) seront automatiquement disponibles pour vos contrôleurs.

### 2. Propagation vers le moteur de calcul Rust

Dès qu'une requête de données est demandée, le service Go valide le droit d'accès sémantique et extrait la clause du DAO `RlsRule`. Cette clause est envoyée dans la file d'attente Valkey avec la requête SQL. Le moteur Rust (`query-engine-rust`) intercepte le message, applique son parseur AST pour fusionner physiquement la règle RLS, et exécute la requête sécurisée sur le DWH.

```

---

## 🎯 Synthèse Métier de cette Architecture

En incluant l'ensemble des objets (Users, Groupes, Rôles, RLS, Charts, Dashboards) au sein du même écosystème DAO généré :
1.  **Zéro Fuite de Données :** Le contrôle sémantique des permissions s’effectue en amont du calcul. Si un rôle n'a pas la permission `CAN_READ` sur un graphique, le pipeline s'arrête immédiatement dans Go, avant même d'éveiller le worker Rust ou de solliciter le Deep Cache.
2.  **Auditabilité Totale :** La structure en graphes permet aux administrateurs de savoir instantanément, via une simple requête d'API, quel groupe ou quel utilisateur a accès à quel niveau d'information sur la plateforme.

```
