from pathlib import Path
from urllib.parse import unquote, urlparse

from client import call_dashscope_multimodal_fusion_embedding
from schemas import AskRequest, MultiModalInputItem


def _normalize_local_media_path(path_value: str, media_key: str) -> str:
    normalized = path_value.strip()
    if not normalized:
        raise ValueError(f"{media_key} path cannot be empty.")

    parsed = urlparse(normalized)
    if parsed.scheme in {"http", "https"}:
        raise ValueError(f"{media_key} must be a local absolute path, remote URL is not allowed.")

    if parsed.scheme == "file":
        if parsed.netloc not in {"", "localhost"}:
            raise ValueError(f"Unsupported file URI host for {media_key}: {parsed.netloc}")
        candidate = Path(unquote(parsed.path)).expanduser()
    elif parsed.scheme:
        # Keep compatibility with Windows absolute path formats like C:\\path\\to\\file.
        if len(parsed.scheme) == 1 and normalized[1:3] in {":\\", ":/"}:
            candidate = Path(normalized).expanduser()
        else:
            raise ValueError(f"Unsupported path scheme for {media_key}: {parsed.scheme}")
    else:
        candidate = Path(normalized).expanduser()

    if not candidate.is_absolute():
        raise ValueError(f"{media_key} must be an absolute local path.")

    resolved_path = candidate.resolve()
    if not resolved_path.is_file():
        raise ValueError(f"{media_key} local file not found: {resolved_path}")

    return str(resolved_path)


def _prepare_multimodal_item(item: MultiModalInputItem) -> dict:
    payload: dict = {}
    if item.factor is not None:
        payload["factor"] = item.factor

    if item.text and item.text.strip():
        payload["text"] = item.text.strip()
        return payload
    if item.image and item.image.strip():
        payload["image"] = _normalize_local_media_path(item.image, "image")
        return payload
    if item.video and item.video.strip():
        payload["video"] = _normalize_local_media_path(item.video, "video")
        return payload
    if item.audio and item.audio.strip():
        payload["audio"] = _normalize_local_media_path(item.audio, "audio")
        return payload

    raise ValueError("Invalid multimodal item: no available modality field.")


def process_ask_request(request: AskRequest) -> dict:
    """Process multimodal fusion embedding request."""
    model = request.model.strip() if request.model else None
    multimodal_input = [_prepare_multimodal_item(item) for item in request.input_data]
    return call_dashscope_multimodal_fusion_embedding(
        input_data=multimodal_input,
        model=model,
    )
