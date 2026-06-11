

# DOCUMENTATION TECHNIQUE DE RÉFÉRENCE : BACKEND BI ANALYTICS PLATFORM

## 1. VISION ARCHITECTURALE & PARADIGME

La plateforme repose sur le découpage strict entre la gestion du réseau (I/O asynchrones), la modélisation sémantique et le calcul brut (Data Processing).

### Le Principe du Polyglottisme Ciblé

* **Go** est le gestionnaire de trafic et le garant de l'intégrité de la couche sémantique. Il brille par sa capacité à maintenir des milliers de connexions simultanées légères (HTTP, WebSockets) via les Goroutines.
* **Rust** est le moteur de traitement de données. Il garantit une exécution à la vitesse du matériel, une sécurité mémoire absolue sans Garbage Collector, et une manipulation vectorisée des flux Big Data.

---

## 2. CARTOGRAPHIE COMPLÈTE DES MICROSERVICES

L'écosystème backend est divisé en **4 microservices découplés**, autonomes et communicant de manière asynchrone ou via des protocoles légers.

```text
[ FRONTEND UI (React / Next.js) ]
               │
               ├── (HTTP REST / WebSockets)
               ▼
┌────────────────────────────────────────────────────────┐
│ 1. AUTH & IAM GATEWAY (GO)                             │
└──────────────────────┬─────────────────────────────────┘
                       │
                       ▼ (Routage Interne)
┌────────────────────────────────────────────────────────┐
│ 2. METADATA & SEMANTIC LAYER SERVICE (GO)              │
│    └── Persistance : PostgreSQL (Configurations)        │
└──────────────────────┬─────────────────────────────────┘
                       │
                       ▼ (Bus de Messages)
┌────────────────────────────────────────────────────────┐
│ ⚡ BUS DE MESSAGES & HOT CACHE : VALKEY                 │
└──────────────────────┬─────────────────────────────────┘
                       │
        ┌──────────────┴──────────────┐
        ▼ (Queue Consommation)        ▼ (Warm-up / Sync)
┌──────────────────────────────┐┌──────────────────────────────┐
│ 3. QUERY & ANALYTICS ENGINE  ││ 4. CACHE & SYNC ORCHESTRATOR │
│    (RUST / WORKER)           ││    (RUST / BACKGROUND)       │
└──────────────┬───────────────┘└──────────────┬───────────────┘
               │                               │
               ▼                               ▼
┌────────────────────────────────────────────────────────┐
│ 💾 DEEP CACHE PHYSIQUE : PARQUET / ZSTD (Disque Local)  │
└──────────────────────────────┬─────────────────────────┘
                               │ (Cache Miss)
                               ▼
              [ DATA WAREHOUSE CLIENT (DWH) ]

```

### 🛰️ Service 1 : Auth & IAM Gateway (Écrit en GO)

* **Rôle :** Point d'entrée unique (Reverse Proxy & API Gateway) pour le Frontend.
* **Fonctions :** * Authentification et enregistrement des utilisateurs (**Users**).
* Gestion et validation des **Rôles** (RBAC - Role-Based Access Control) pour l'accès macro aux dashboards.
* Interception, décodage et validation des jetons JWT.



### 🗂️ Service 2 : Metadata & Semantic Layer Service (Écrit en GO)

* **Rôle :** Cerveau de la configuration et de la modélisation sémantique.
* **Fonctions :**
* **Gestion des Datasets (Modèles Sémantiques) :** Enregistrement des connexions directes, configurations des jointures inter-tables complexes basées sur les dimensions, ou stockage des requêtes SQL personnalisées des Data Engineers.
* **Gestion des Charts & Dashboards :** Stockage des structures de graphiques, des rapports et de la configuration dynamique des styles de colonnes (JSON `column_styles` contenant le mapping des couleurs personnalisées appliquées par l'utilisateur depuis l'UI, ex: bleu pour `ville_de_naissance`).
* **Sécurité RLS (Row-Level Security) :** Stockage des règles de segmentation logique des lignes de données par rôle/utilisateur.
* **Persistance :** Gère et interroge la base de données PostgreSQL interne dédiée aux métadonnées.



### 🚀 Service 3 : Query & Analytics Engine (Écrit en RUST)

* **Rôle :** Le cœur d'exécution analytique. Ce service n'expose aucune API HTTP publique ; il fonctionne comme un worker d'arrière-plan.
* **Fonctions :**
* Consomme les tickets de calcul poussés dans la file d'attente Valkey.
* Compile le SQL sémantique et utilise un parseur pour **injecter mathématiquement la clause RLS** de l'utilisateur dans l'AST (Abstract Syntax Tree) SQL afin d'assurer l'étanchéité absolue des données.
* Exécute la requête finale en mode **Pushdown** sur le Data Warehouse (DWH) du client.
* Réceptionne la donnée, applique le dictionnaire de couleurs colonnaires dynamiques et structure le payload final.
* Gère la **pagination stricte** pour l'affichage UI et le **streaming vectorisé** par micro-paquets de lignes (via Apache Arrow) pour le SQL Lab ou les exports, sans jamais charger l'intégralité d'une table en RAM.



### 🔄 Service 4 : Cache & Sync Orchestrator (Écrit en RUST)

* **Rôle :** Le gestionnaire autonome du cycle de vie des données froides et lourdes.
* **Fonctions :**
* Planifie et orchestre le rafraîchissement automatique et prédictif du Deep Cache Physique (la nuit ou sur déclenchement par webhook après l'ingestion du DWH).
* Invalide les clés de cache de manière atomique dès qu'une modification s'opère sur un modèle sémantique ou qu'un utilisateur modifie la palette de couleurs d'un chart depuis l'UI.



---

## ⚡ 3. LA STRATÉGIE DE CACHE À DOUBLE ÉTAGE (HOT vs DEEP)

Pour concilier temps réel, volumes massifs (tables de 200M+ de lignes), requêtes lourdes et haute concurrence, le système implémente une matrice de cache révolutionnaire à deux niveaux.

### Étage 1 : Le Hot Cache (In-Memory / Valkey)

* **Technologie :** Valkey (Moteur de clés-valeurs in-memory ultra-rapide).
* **Cible :** Les métadonnées actives, les sessions, les structures de dashboards, ainsi que les résultats de requêtes légères (KPIs, agrégations simples, graphiques de tendance de premier niveau).
* **Gouvernance RLS :** La clé de cache d'un graphique est calculée dynamiquement sous la forme :

$$\text{Cache Key} = \text{Hash}(\text{Chart ID} + \text{Filtres UI} + \text{Règle RLS Utilisateur})$$



Cela garantit que deux utilisateurs ayant des règles RLS différentes sur le même graphique ne partageront jamais le même cache en mémoire.
* **Latence :** $< 2\text{ ms}$.

### Étage 2 : Le Deep Cache Physique (Sur Disque / Parquet & ZSTD)

* **Technologie :** Fichiers colonnaires Apache Parquet locaux gérés nativement par Rust, compressés avec l'algorithme **ZSTD** (Zstandard).
* **Cible :** Les requêtes ultra-lourdes, les rapports volumineux contenant de grands tableaux, les historiques analytiques et les données brutes lues par le SQL Lab.
* **Le mécanisme de performance :** * **Compression Colonnaire :** Le format Parquet regroupe les données par colonne. Associé à la compression ZSTD (niveau 3 par défaut), les données répétitives (comme les noms de pays ou de villes) subissent un ratio de compression moyen de **1:8 à 1:10**. 5 Go d'espace disque alloués peuvent ainsi contenir jusqu'à 50 Go de données réelles.
* **Memory-Mapping (mmap) & Lazy Evaluation :** Le moteur Rust n'ouvre pas le fichier Parquet en RAM. Il utilise le *Memory-Mapping* pour créer un index virtuel sur le disque. Si l'utilisateur applique un filtre ou un tri sur la colonne `solde_compte` depuis l'UI, Rust lit **uniquement les octets correspondants à cette colonne** sur le disque. La RAM reste totalement propre.
* **Asynchronisme & Export sans perte :** Lors d'une demande d'export massif (200M de lignes), le système bascule en asynchrone. Rust extrait la table du DWH par micro-blocs, formate le binaire Parquet/CSV compressé directement dans un flux continu de fichiers vers un stockage local ou partagé (NFS/MinIO), notifie Go via WebSockets une fois terminé, et livre un lien de téléchargement direct sans jamais saturer la mémoire du backend.



---

## 📦 4. MANIFESTE COMPLET DES PACKAGES & DEPENDANCES

Chaque microservice embarque exclusivement des librairies industrielles de haute performance, typées statiquement, éliminant les allocations de mémoire inutiles.

### 🛠️ Environnement Go (Services 1 & 2)

* **`github.com/gofiber/fiber/v3` (Framework Web) :** Construit au-dessus de `fasthttp`. C'est le moteur de routage HTTP et WebSocket le plus rapide en Go. Il utilise au mieux le modèle des Goroutines pour l'asynchronisme massif des utilisateurs.
* **`entgo.io/ent` (ORM Sémantique) :** ORM moderne basé sur les graphes et la génération de code. Contrairement aux ORMs classiques (comme GORM), il supprime toute utilisation d'interfaces vides (`interface{}`), garantissant un code 100 % typé statiquement à la compilation, une sécurité totale au runtime et des performances pures pour exécuter les jointures de modèles sémantiques dans PostgreSQL.
* **`github.com/jackc/pgx/v5` (Driver PostgreSQL) :** Driver natif PostgreSQL supportant le protocole binaire complet, le pipelining de requêtes et le pooling de connexions asynchrones.
* **`github.com/valkey-io/valkey-go` (Client Valkey) :** Client officiel optimisé pour Valkey. Il supporte le multiplexage automatique des connexions réseau (une seule connexion TCP partagée par toutes les Goroutines de manière non bloquante).
* **`github.com/golang-jwt/jwt/v5` (Sécurité) :** Package de gestion et de vérification cryptographique des jetons d'accès, supportant les signatures asymétriques (RS256).

### 🦀 Environnement Rust (Services 3 & 4)

* **`tokio` (Runtime Asynchrone) :** Le standard industriel pour l'exécution asynchrone orientée événements (*Event-driven*). Il gère le pool de threads de fond pour écouter Valkey et exécuter les calculs sans bloquer le système.
* **`polars` (Moteur de DataFrames) :** Librairie de manipulation colonnaire ultra-rapide écrite en Rust. Elle surpasse Pandas et gère de manière native l'évaluation paresseuse (*LazyFrame*), le requêtage de fichiers Parquet et le streaming vectorisé.
* **`sqlparser` (Parseur SQL) :** Crate de qualité industrielle permettant de compiler une chaîne SQL brute en un arbre syntaxique abstrait (AST). C'est l'outil de sécurité qui permet d'inspecter les requêtes des utilisateurs et d'y **injecter dynamiquement la clause de filtrage RLS** de manière mathématique, interdisant toute forme d'injection SQL.
* **`parquet` (Stockage Colonnaire) :** Crate native implémentant les spécifications d'Apache Parquet avec l'extension de compression native `zstd` pour la gestion de notre Deep Cache Physique.
* **`serde` & `serde_json` (Sérialisation) :** Framework de sérialisation/désérialisation ultra-rapide basé sur des macros de compilation. Il transforme les structures de données Rust en JSON (et inversement) à la vitesse du processeur, sans allocations mémoires superflues.
* **`clickhouse-rs` / `tokio-postgres` (Connecteurs DWH) :** Connecteurs asynchrones natifs communiquant avec les bases de données analytiques des clients via des protocoles binaires par blocs, optimisant le débit de transfert lors du streaming de données brutes.

---

## 🔄 5. INFRASTRUCTURE & FLUX DE TRAITEMENT (REQUEST/RESPONSE)

Pour résumer visuellement l'interaction de tout le backend, voici le cycle de vie d'une requête complexe initiée par un utilisateur sur un graphique personnalisé :

1. **Requête UI :** L'utilisateur ouvre son dashboard. Le navigateur envoie une requête HTTP à la Gateway **Go (Fiber)**. Le jeton est validé par **golang-jwt**.
2. **Vérification de la Configuration :** Le service **Go (Metadata)** extrait la structure du Chart depuis PostgreSQL via **Ent**, récupère le JSON `column_styles` (la couleur que l'utilisateur a configurée pour ce graphique précis) et calcule le hash unique de la requête incluant les filtres et le RLS de l'utilisateur.
3. **Routage vers la Matrice de Cache :**
* *Scénario A (Hit Hot Cache) :* Le JSON complet est présent dans **Valkey**. Go le récupère via **valkey-go** et le renvoie en **< 2ms**. Fin du flux.
* *Scénario B (Hit Deep Cache) :* La donnée n'est pas dans Valkey, mais un fichier Parquet correspondant existe sur le disque local. Le worker **Rust (Query Engine)** ouvre le fichier via **Polars (mmap)**, extrait instantanément les lignes requises par la pagination (ex: 100 lignes), y injecte la configuration de couleurs, convertit le tout via **Serde** et le passe à Go qui répond à l'utilisateur.
* *Scénario C (Cache Miss Total) :* La requête est poussée dans la file d'attente Valkey. Un worker **Rust (Tokio)** prend le job, parse le SQL sémantique avec **sqlparser**, injecte le filtre RLS, exécute la requête sur le Data Warehouse client, récupère le résultat agrégé ou streamé, écrit le fichier `.parquet` compressé en **ZSTD** via la brique **parquet** pour les prochaines fois, applique la couleur de la colonne et livre le flux final à l'API Go pour affichage immédiat.
