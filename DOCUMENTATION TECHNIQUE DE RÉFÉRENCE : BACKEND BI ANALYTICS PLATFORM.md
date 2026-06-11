# DOCUMENTATION TECHNIQUE DE RÉFÉRENCE : BACKEND BI ANALYTICS PLATFORM

*Version Mise à Jour – Intégration du Moteur de Templating Dynamique et Couche Sémantique Avancée*

---

## 1. VISION ARCHITECTURALE & PARADIGME

La plateforme est une architecture logicielle industrielle, agnostique et hautement scalable. Elle est conçue pour l'analyse de données de masse (Big Data) et supporte de fortes contraintes de concurrence sans compromis sur la sécurité ou la fraîcheur des données.

Elle repose sur la séparation stricte entre la gestion du trafic réseau, la modélisation sémantique et le calcul vectorisé lourd.

### Le Principe du Polyglottisme Ciblé

* **Go** fait office de chef d'orchestre réseau et de garant de l'intégrité des métadonnées. Son modèle de concurrence basé sur les *Goroutines* lui permet de gérer des milliers de connexions simultanées légères (HTTP, WebSockets) avec une empreinte mémoire minimale.
* **Rust** représente le muscle informatique du système. Sans Garbage Collector, il garantit une exécution à la vitesse du matériel, une sécurité mémoire absolue et une manipulation colonnaire ultra-rapide des flux de données.

---

## 2. CARTOGRAPHIE COMPLÈTE DES MICROSERVICES

L'écosystème backend est divisé en **4 microservices découplés**, autonomes, hautement spécialisés et encapsulés dans des conteneurs isolés.

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

* **Rôle :** Point d'entrée unique (Reverse Proxy & API Gateway) pour l'application.
* **Fonctions :** * Authentification, enregistrement et cycle de vie des utilisateurs (**Users**).
* Gestion et validation des **Rôles** (RBAC - Role-Based Access Control) pour l'accès macro aux ressources.
* Interception, décodage et validation cryptographique des jetons JWT à chaque requête.



### 🗂️ Service 2 : Metadata & Semantic Layer Service (Écrit en GO)

* **Rôle :** Gardien de la configuration, des palettes visuelles et de la modélisation sémantique.
* **Fonctions :**
* **Gestion des Datasets (Modèles Sémantiques) :** Enregistrement des chaînes de connexion, configuration des schémas, des jointures inter-tables complexes et stockage des requêtes SQL dynamiques de base.
* **Gestion des Charts & Dashboards :** Stockage des structures de graphiques et de la configuration dynamique des styles de colonnes (JSON `column_styles` contenant le mapping des couleurs personnalisées modifiées par l'utilisateur depuis l'UI, ex: bleu pour `ville_naissance`).
* **Sécurité RLS (Row-Level Security) :** Enregistrement des règles de segmentation logique des lignes de données par utilisateur ou par rôle.
* **Persistance :** Gère et interroge la base de données PostgreSQL interne dédiée aux métadonnées de l'application.



### 🚀 Service 3 : Query & Analytics Engine (Écrit en RUST)

* **Rôle :** Le cœur d'exécution et de transformation analytique. Ce service fonctionne exclusivement comme un worker d'arrière-plan sans API publique directe.
* **Fonctions :**
* Dépile en continu les requêtes de calcul poussées dans la file d'attente de messages Valkey.
* **Moteur d'Évaluation Dynamique (Jinja2-like) :** Analyse le code SQL sémantique à l'aide d'un moteur de template pour résoudre les structures conditionnelles (`{% if %}`), les boucles et injecter les variables contextuelles de l'UI et de l'utilisateur connecté.
* **Sécurité RLS :** Compile le SQL nettoyé, utilise son parseur pour analyser l'arbre syntaxique (AST) et y **injecter mathématiquement la clause `WHERE` du RLS** spécifique à l'utilisateur.
* Exécute la requête finale sécurisée en mode **Pushdown** sur le Data Warehouse (DWH) cible.
* Réceptionne la donnée, applique le dictionnaire de couleurs colonnaires dynamiques selon la configuration du Chart, et structure le payload de sortie.
* Gère la **pagination stricte** pour l'affichage UI et le **streaming vectorisé par Chunks** via Apache Arrow pour le SQL Lab et les exports massifs.



### 🔄 Service 4 : Cache & Sync Orchestrator (Écrit en RUST)

* **Rôle :** Le gestionnaire autonome du cycle de vie des données froides, lourdes et précalculées.
* **Fonctions :**
* Planifie et orchestre le rafraîchissement automatique et prédictif du Deep Cache Physique (par exemple, exécution nocturne programmée ou via Webhook après une ingestion réussie dans le DWH).
* Invalide les clés de cache de manière atomique dès qu'une modification intervient sur un modèle sémantique ou qu'un utilisateur change la palette de couleurs d'un graphique depuis l'UI.



---

## ⚡ 3. LA STRATÉGIE DE CACHE À DOUBLE ÉTAGE (HOT vs DEEP)

Le système implémente une matrice de cache hybride à deux niveaux pour concilier instantanéité de l'affichage et traitement de volumes de centaines de millions de lignes.

### Étage 1 : Le Hot Cache (In-Memory / Valkey)

* **Technologie :** Valkey (Moteur clé-valeur in-memory ultra-rapide).
* **Cible :** Les sessions, les métadonnées actives, les structures de dashboards, ainsi que les résultats de requêtes légères (KPIs unitaires, agrégations simples, graphiques de tendance de premier niveau).
* **Gouvernance RLS :** La clé de cache d'un graphique intègre systématiquement le contexte de sécurité de l'appelant :

$$\text{Cache Key} = \text{Hash}(\text{Chart ID} + \text{Filtres UI} + \text{Règle RLS Utilisateur})$$



Cela garantit l'étanchéité absolue : deux utilisateurs ayant des règles RLS différentes sur un même graphique ne partageront jamais la même zone de mémoire vive.
* **Latence :** $< 2\text{ ms}$.

### Étage 2 : Le Deep Cache Physique (Sur Disque / Parquet & ZSTD)

* **Technologie :** Fichiers colonnaires Apache Parquet locaux gérés nativement par Rust, compressés avec l'algorithme de pointe **ZSTD** (Zstandard).
* **Cible :** Les requêtes lourdes, les rapports volumineux contenant de grands tableaux, les historiques analytiques et les données brutes lues par le SQL Lab.
* **Le mécanisme de performance :** * **Compression Colonnaire :** Le format Parquet regroupe les données par colonne. Associé à la compression ZSTD (niveau 3), les données répétitives subissent un ratio de compression moyen de **1:8 à 1:10**. 5 Go de cache physique réel peuvent ainsi stocker l'équivalent de 40 à 50 Go de données brutes.
* **Memory-Mapping (mmap) & Lazy Evaluation :** Le moteur Rust n'ouvre pas le fichier Parquet global en RAM. Il utilise le *Memory-Mapping* pour créer un pointeur sur le fichier physique. Si l'utilisateur applique un filtre ou un tri sur la colonne `solde_compte` depuis l'UI, Rust lit **uniquement les octets correspondants à cette colonne** sur le disque. La RAM reste totalement propre.
* **Asynchronisme & Export sans perte :** Lors d'un export de masse (ex: 200M de lignes), le système bascule en asynchrone. Rust extrait la table du DWH par micro-blocs, formate le binaire Parquet/CSV compressé directement dans un flux continu de fichiers vers un stockage local ou un dossier partagé (NFS/MinIO), notifie Go via WebSockets une fois terminé, et livre un lien de téléchargement direct sans jamais saturer la mémoire du backend.



---

## 🔬 4. LE PIPELINE D'EXÉCUTION DES REQUÊTES SQL

Lorsqu'une demande de données ou de graphiques est envoyée par le Frontend, le traitement suit un pipeline linéaire strict et hautement sécurisé au sein du moteur Rust.

```text
[ SQL Dynamique (Modèle Sémantique avec Syntaxe Jinja2) ]
                           │
                           ▼
┌────────────────────────────────────────────────────────┐
│ 1. ÉVALUATION DU TEMPLATE (TERA ENGINE)                │
│    - Injection des filtres UI (dates, pays...)          │
│    - Résolution des boucles et conditions {% if %}     │
│    - Injection du contexte utilisateur connecté        │
└──────────────────────────┬─────────────────────────────┘
                           │
                           ▼ [ SQL Standard Nettoyé ]
┌────────────────────────────────────────────────────────┐
│ 2. ANALYSE SYNTAXIQUE DE SÉCURITÉ (SQLPARSER AST)       │
│    - Déconstruction de la requête en arbre syntaxique   │
│    - Blocage absolu de toute tentative d'injection SQL  │
└──────────────────────────┬─────────────────────────────┘
                           │
                           ▼
┌────────────────────────────────────────────────────────┐
│ 3. ENRICHISSEMENT RLS (ROW-LEVEL SECURITY)             │
│    - Injection mathématique des filtres de lignes       │
│    - Modification de l'AST avant compilation           │
└──────────────────────────┬─────────────────────────────┘
                           │
                           ▼ [ SQL Final Compilé Sécurisé ]
┌────────────────────────────────────────────────────────┐
│ 4. ROUTAGE & PUSHDOWN ANALYTIQUE                      │
│    - Vérification du Deep Cache Physique (Parquet)     │
│    - Si Cache Miss : Exécution directe sur le DWH      │
└────────────────────────────────────────────────────────┘

```

---

## 📦 5. MANIFESTE COMPLET DES PACKAGES & DÉPENDANCES

Chaque microservice embarque exclusivement des librairies industrielles de haute performance, typées statiquement, éliminant les allocations de mémoire inutiles.

### 🛠️ Environnement Go (Services 1 & 2)

* **`github.com/gofiber/fiber/v3` (Framework Web) :** Construit au-dessus de `fasthttp`. C'est le moteur de routage HTTP et WebSocket le plus rapide en Go. Il utilise au mieux le modèle des Goroutines pour l'asynchronisme massif des utilisateurs.
* **`entgo.io/ent` (ORM Sémantique) :** ORM moderne basé sur les graphes et la génération de code. Contrairement aux ORMs classiques (comme GORM), il supprime toute utilisation d'interfaces vides (`interface{}`), garantissant un code 100 % typé statiquement à la compilation, une sécurité totale au runtime et des performances pures pour exécuter les jointures de modèles sémantiques dans PostgreSQL.
* **`github.com/jackc/pgx/v5` (Driver PostgreSQL) :** Driver natif PostgreSQL supportant le protocole binaire complet, le pipelining de requêtes et le pooling de connexions asynchrones.
* **`github.com/valkey-io/valkey-go` (Client Valkey) :** Client officiel optimisé pour Valkey. Il supporte le multiplexage automatique des connexions réseau (une seule connexion TCP partagée par toutes les Goroutines de manière non bloquante).
* **`github.com/golang-jwt/jwt/v5` (Sécurité) :** Package de gestion et de vérification cryptographique des jetons d'accès, supportant les signatures asymétriques (RS256).

### 🦀 Environnement Rust (Services 3 & 4)

* **`tokio` (Runtime Asynchrone) :** Le standard industriel pour l'exécution asynchrone orientée événements (*Event-driven*). Il gère le pool de threads de fond pour écouter Valkey et exécuter les calculs sans bloquer le système.
* **`tera` (Moteur de Templating) :** Implémentation Rust native inspirée à 100 % de Jinja2. Elle compile les structures de requêtes, évalue les conditions et injecte les variables en quelques microsecondes sans le poids d'un runtime interprété.
* **`polars` (Moteur de DataFrames) :** Librairie de manipulation colonnaire ultra-rapide écrite en Rust. Elle surpasse Pandas et gère de manière native l'évaluation paresseuse (*LazyFrame*), le requêtage de fichiers Parquet et le streaming vectorisé.
* **`sqlparser` (Parseur SQL) :** Crate de qualité industrielle permettant de compiler une chaîne SQL brute en un arbre syntaxique abstrait (AST). C'est l'outil de sécurité qui permet d'inspecter les requêtes des utilisateurs et d'y injecter dynamiquement la clause de filtrage RLS de manière mathématique, interdisant toute forme d'injection SQL.
* **`parquet` (Stockage Colonnaire) :** Crate native implémentant les spécifications d'Apache Parquet avec l'extension de compression native `zstd` pour la gestion de notre Deep Cache Physique.
* **`serde` & `serde_json` (Sérialisation) :** Framework de sérialisation/désérialisation ultra-rapide basé sur des macros de compilation. Il transforme les structures de données Rust en JSON (et inversement) à la vitesse du processeur, sans allocations mémoires superflues.
* **`clickhouse-rs` / `tokio-postgres` (Connecteurs DWH) :** Connecteurs asynchrones natifs communiquant avec les bases de données analytiques des clients via des protocoles binaires par blocs, optimisant le débit de transfert lors du streaming de données brutes.
