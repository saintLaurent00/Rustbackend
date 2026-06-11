use chromiumoxide::browser::{Browser, BrowserConfig};
use chromiumoxide::handler::viewport::Viewport;
use futures::StreamExt;
use std::env;

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    let (browser, mut handler) =
        Browser::launch(BrowserConfig::builder()
            .viewport(Viewport::builder().width(1280).height(720).build())
            .build()?).await?;

    tokio::spawn(async move {
        while let Some(h) = handler.next().await {
            if h.is_err() { break; }
        }
    });

    println!("Render Worker started, waiting for render tasks...");

    // Ici nous écouterions une queue Valkey pour des tâches de rendu
    // Pour l'instant, on simule une attente
    loop {
        tokio::time::sleep(tokio::time::Duration::from_secs(60)).await;
    }
}
