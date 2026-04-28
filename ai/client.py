from http import HTTPStatus

import dashscope

from config import settings


def _apply_dashscope_base_url() -> None:
    if settings.dashscope_base_http_api_url:
        dashscope.base_http_api_url = settings.dashscope_base_http_api_url


def call_dashscope_multimodal_fusion_embedding(
    input_data: list[dict],
    model: str | None = None,
) -> dict:
    _apply_dashscope_base_url()
    selected_model = (model or settings.dashscope_multimodal_embedding_model).strip()
    fusion_dimension = 1024
    if not selected_model:
        raise ValueError("Multimodal embedding model cannot be empty.")

    response = dashscope.MultiModalEmbedding.call(
        api_key=settings.dashscope_api_key,
        model=selected_model,
        input=input_data,
        enable_fusion=True,
        dimension=fusion_dimension,
    )

    if response.status_code != HTTPStatus.OK:
        raise Exception(
            f"DashScope MultiModalEmbedding Error: code={response.status_code}, message={response.message}",
        )

    if not response.output:
        raise Exception("DashScope MultiModalEmbedding returned empty output")

    return {
        "model": selected_model,
        "enable_fusion": True,
        "dimension": fusion_dimension,
        "output": response.output,
        "usage": response.usage,
        "request_id": response.request_id,
    }
