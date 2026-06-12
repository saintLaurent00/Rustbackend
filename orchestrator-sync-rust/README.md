# 🔄 Orchestrator Sync (Rust)

Le cerveau d'arrière-plan de la plateforme, gérant les tâches planifiées, la synchronisation du cache et la distribution des rapports.

## 🚀 Responsabilités
- **Tâches Cron** : Planification des rafraîchissements de cache et des envois de rapports.
- **Mailing** : Expédition de mails avec pièces jointes (rapports BI) via SMTP.
- **Coordination** : Communication avec le Render Worker pour obtenir les visuels nécessaires.

## 🏗️ Architecture
- **Mailing** : Lettre (Crate Rust pour SMTP).
- **Planification** : Boucle événementielle asynchrone avec Tokio.
