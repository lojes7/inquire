from http import HTTPStatus

import dashscope

from config import settings


def _apply_dashscope_base_url() -> None:
    if settings.dashscope_base_http_api_url:
        dashscope.base_http_api_url = settings.dashscope_base_http_api_url


def call_dashscope_text_embedding(
    input_text: str,
    model: str | None = None,
) -> dict:
    _apply_dashscope_base_url()
    selected_model = (model or settings.dashscope_embedding_model).strip()
    if not selected_model:
        raise ValueError("Embedding model cannot be empty.")

    response = dashscope.TextEmbedding.call(
        model=selected_model,
        input=input_text,
        api_key=settings.dashscope_api_key,
    )

    if response.status_code != HTTPStatus.OK:
        raise Exception(
            f"DashScope TextEmbedding Error: code={response.status_code}, message={response.message}",
        )

    if not response.output:
        raise Exception("DashScope TextEmbedding returned empty output")

    return {
        "model": selected_model,
        "output": response.output,
        "usage": response.usage,
        "request_id": response.request_id,
    }
