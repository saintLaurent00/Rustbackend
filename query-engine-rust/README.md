# ⚡ Query Engine (Rust)

Le moteur de calcul haute performance de la plateforme, responsable de la transformation, sécurisation et exécution des requêtes SQL.

## 🚀 Responsabilités
- **Templating SQL** : Résolution des requêtes dynamiques via le moteur Tera (Jinja2).
- **Injection RLS** : Modification de l'AST (Abstract Syntax Tree) SQL pour injecter mathématiquement les conditions de sécurité sans risque d'injection SQL.
- **Abstraction Multi-sources** : Capacité à interroger des bases de données, APIs et fichiers via Polars.
- **Réactivité** : Mise à jour de la configuration à chaud via Valkey Pub/Sub.

## 🏗️ Architecture
- **Runtime** : Tokio (Asynchrone).
- **Analyse SQL** : sqlparser-rs.
- **Traitement de données** : Polars (DataFrames colonnaires ultra-rapides).
- **Communication** : Redis/Valkey pour la file d'attente et la synchronisation.

## 📂 Structure du projet
- `src/parser/templating.rs` : Moteur de résolution des templates SQL.
- `src/parser/rls.rs` : Analyseur AST pour l'injection sécurisée des filtres de lignes.
- `src/executor/` : Abstraction pour les connecteurs de données.
