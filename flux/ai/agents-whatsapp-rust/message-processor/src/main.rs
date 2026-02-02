use axum::{routing::post, Router};
use mongodb::{Client as MongoClient, Database};
use redis::Client as RedisClient;
use std::sync::Arc;
use tracing::info;

mod config;
mod handlers;

use config::Config;

#[derive(Clone)]
pub struct AppState {
    pub mongo: MongoClient,
    pub db: Database,
    pub redis: RedisClient,
    pub config: Config,
}

#[tokio::main]
async fn main() -> anyhow::Result<()> {
    // Initialize tracing
    tracing_subscriber::fmt()
        .with_env_filter("message_processor=info")
        .init();

    info!("Starting message-processor...");

    // Load configuration
    let config = Config::from_env()?;

    // Connect to MongoDB
    let mongo = MongoClient::with_uri_str(&config.mongodb_uri).await?;
    let db = mongo.database(&config.mongodb_database);
    info!("Connected to MongoDB");

    // Connect to Redis
    let redis = RedisClient::open(config.redis_uri.as_str())?;
    info!("Connected to Redis");

    // Initialize DLQ collection
    shared::dlq::init_dlq_collection(&db).await?;
    info!("Initialized DLQ collection");

    // Create app state
    let state = Arc::new(AppState {
        mongo,
        db,
        redis,
        config,
    });

    // Build router
    let app = Router::new()
        .route("/", post(handlers::handle_event))
        .route("/health/live", axum::routing::get(|| async { "OK" }))
        .route("/health/ready", axum::routing::get(|| async { "OK" }))
        .with_state(state);

    // Start server
    let addr = "0.0.0.0:8080";
    info!("Listening on {}", addr);

    let listener = tokio::net::TcpListener::bind(addr).await?;
    axum::serve(listener, app).await?;

    Ok(())
}
