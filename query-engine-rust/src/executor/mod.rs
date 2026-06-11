use async_trait::async_trait;
use polars::prelude::*;

#[async_trait]
pub trait DataSource {
    async fn fetch_data(&self, query: &str) -> PolarsResult<DataFrame>;
}

// Implémentations spécifiques (Postgres, Clickhouse, API, etc.) seront ajoutées ici
