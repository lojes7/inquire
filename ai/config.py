import os
from pydantic_settings import BaseSettings, SettingsConfigDict


class Settings(BaseSettings):
    model_config = SettingsConfigDict(
        env_file=".env",
        env_file_encoding="utf-8",
        extra="ignore",
    )

    dashscope_api_key: str = os.getenv("DASHSCOPE_API_KEY", "")
    dashscope_multimodal_embedding_model: str = os.getenv("DASHSCOPE_MULTIMODAL_EMBEDDING_MODEL", "qwen3-vl-embedding")
    dashscope_base_http_api_url: str = os.getenv("DASHSCOPE_BASE_HTTP_API_URL", "")
    port: int = int(os.getenv("PORT", "8001"))


settings = Settings()
if not settings.dashscope_api_key:
    raise ValueError("DASHSCOPE_API_KEY environment variable is missing or empty.")
