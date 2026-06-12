# 📊 Metadata & Semantic Layer Service (Go)

Ce service gère la couche sémantique de la plateforme, incluant la définition des datasets, la configuration des graphiques et la sécurité des données au niveau des lignes (RLS).

## 🚀 Responsabilités
- **Couche Sémantique** : Modélisation des sources de données (DB, API, Fichiers).
- **Gestion des Assets** : CRUD des graphiques (Charts) et des tableaux de bord (Dashboards).
- **Sécurité des données** : Stockage et validation des règles Row-Level Security (RLS).
- **Chiffrement** : Protection des identifiants de connexion aux sources de données via AES-GCM.

## 🏗️ Architecture
- **Framework Web** : Fiber v2.
- **DAO / ORM** : Ent.
- **Sécurité** : Module de chiffrement interne pour les données sensibles.

## 📂 Structure du projet
- `ent/schema/` : Définition des objets BI (Dataset, Chart, Dashboard, RlsRule).
- `internal/security/` : Logique de chiffrement/déchiffrement des secrets.
- `cmd/main.go` : Serveur API gérant les métadonnées.
