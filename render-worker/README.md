# 📸 Render Worker (Rust)

Service dédié à la génération de captures visuelles des dashboards pour les exports et les rapports mails. Découplé pour permettre un scaling indépendant.

## 🚀 Responsabilités
- **Rendu Headless** : Utilisation de Chromium pour effectuer le rendu exact des graphiques côté serveur.
- **Génération d'Images** : Capture de snapshots haute résolution.

## 🏗️ Architecture
- **Navigateur** : Chromiumoxide (contrôle de Headless Chromium en Rust).
- **Isolation** : S'exécute dans un conteneur spécifique optimisé pour Chromium.
