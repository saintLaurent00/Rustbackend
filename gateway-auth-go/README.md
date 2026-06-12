# 🔐 Gateway Auth Service (Go)

Ce microservice est le point d'entrée unique pour l'authentification et la gestion de l'identité (IAM) de la plateforme BI.

## 🚀 Responsabilités
- **Gestion des sessions** : Délivrance et validation de jetons JWT.
- **Authentification Hybride** : Support initial de l'authentification Locale (Email/Password), extensible vers LDAP et SSO.
- **RBAC (Role-Based Access Control)** : Gestion des utilisateurs, groupes, rôles et permissions via l'ORM Ent.

## 🏗️ Architecture
- **Framework Web** : Fiber v2 (Performance & Simplicité).
- **DAO / ORM** : Ent (Génération de code typé, approche orientée graphes).
- **Sécurité** : Hachage des mots de passe avec Bcrypt et signatures JWT asymétriques.

## 📂 Structure du projet
- `cmd/main.go` : Point d'entrée, initialisation du serveur et des connexions.
- `ent/schema/` : Définition des entités (User, Role, Group, Permission).
- `internal/auth/` : Logique métier de l'authentification et gestion des providers.

## 🛠️ Développement
Pour régénérer le code Ent après une modification de schéma :
```bash
go generate ./ent
```
