mod mailing;

use std::env;
use dotenvy::dotenv;

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    dotenv().ok();

    println!("Orchestrator Sync started, handling background tasks...");

    // Boucle infinie pour les tâches planifiées
    loop {
        tokio::time::sleep(tokio::time::Duration::from_secs(3600)).await;
    }
}
