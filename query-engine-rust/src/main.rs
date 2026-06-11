mod parser;
mod executor;

use redis::AsyncCommands;
use std::env;
use dotenvy::dotenv;
use futures_util::StreamExt;

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    dotenv().ok();
    let redis_url = env::var("VALKEY_URL").unwrap_or_else(|_| "redis://valkey:6379".to_string());

    let client = redis::Client::open(redis_url)?;

    println!("Query Engine started, listening for configuration updates...");

    let mut pubsub = client.get_async_connection().await?.into_pubsub();
    pubsub.subscribe("config:updates").await?;

    let mut stream = pubsub.on_message();

    tokio::spawn(async move {
        while let Some(msg) = stream.next().await {
            let payload: String = msg.get_payload().unwrap();
            println!("Configuration update received: {}", payload);
        }
    });

    loop {
        tokio::time::sleep(tokio::time::Duration::from_secs(60)).await;
    }
}
