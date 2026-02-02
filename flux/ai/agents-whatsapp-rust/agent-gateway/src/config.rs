use std::env;

#[derive(Clone)]
pub struct Config {
    pub mongodb_uri: String,
    pub mongodb_database: String,
    pub redis_uri: String,
    pub broker_url: String,
}

impl Config {
    pub fn from_env() -> anyhow::Result<Self> {
        Ok(Self {
            mongodb_uri: env::var("MONGODB_URI")
                .unwrap_or_else(|_| "mongodb://localhost:27017".to_string()),
            mongodb_database: env::var("MONGODB_DATABASE")
                .unwrap_or_else(|_| "messaging_app".to_string()),
            redis_uri: env::var("REDIS_URI")
                .unwrap_or_else(|_| "redis://localhost:6379".to_string()),
            broker_url: env::var("BROKER_URL").unwrap_or_else(|_| {
                "http://default-broker.homelab-services.svc.cluster.local".to_string()
            }),
        })
    }
}
