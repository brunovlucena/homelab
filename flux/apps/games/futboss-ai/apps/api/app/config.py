# FutBoss AI - Configuration
# Author: Bruno Lucena (bruno@lucena.cloud)

from pydantic_settings import BaseSettings
from functools import lru_cache


class Settings(BaseSettings):
    # API
    api_host: str = "0.0.0.0"
    api_port: int = 8000
    debug: bool = True

    # MongoDB
    mongodb_url: str = "mongodb://localhost:27017"
    mongodb_database: str = "futboss"

    # JWT
    jwt_secret: str = "super-secret-change-in-production"
    jwt_algorithm: str = "HS256"
    jwt_expiration_hours: int = 24

    # Ollama
    ollama_base_url: str = "http://localhost:11434"
    ollama_model: str = "llama3.2"

    # Game
    initial_tokens: int = 1000
    match_duration_minutes: int = 90
    token_to_brl_rate: float = 0.01

    class Config:
        env_file = ".env"


@lru_cache()
def get_settings() -> Settings:
    return Settings()

