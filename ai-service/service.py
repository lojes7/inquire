import logging
from pathlib import Path
from urllib.parse import unquote, urlparse

from client import call_dashscope_multimodal_fusion_embedding
from chunking import chunk_text
from preprocessing import get_parser
from schemas import AskRequest, EmbedResponse, EmbedVectorItem, MultiModalInputItem

logger = logging.getLogger(__name__)



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


# qwen3-vl-embedding 原生支持的媒体 MIME 类型前缀
_NATIVE_MEDIA_PREFIXES = {
    "image/",
    "video/",
    "audio/",
}


def _is_native_media_type(file_type: str) -> bool:
    """判断 MIME 类型是否被 DashScope 原生支持（可直接作为 media input）。"""
    if not file_type:
        return False
    normalized = file_type.lower().split(";")[0].strip()
    return any(normalized.startswith(p) for p in _NATIVE_MEDIA_PREFIXES)


def _call_embed_for_text(text: str) -> list[float]:
    """对单个文本调用 DashScope 多模态融合嵌入，返回 1024 维向量。"""
    result = call_dashscope_multimodal_fusion_embedding(
        input_data=[{"text": text.strip()}],
    )
    # DashScope 融合嵌入输出结构：output["embedding"] 是一个 float 列表
    output = result.get("output", {})
    if isinstance(output, dict):
        embedding = output.get("embedding")
        if isinstance(embedding, list) and embedding:
            return embedding
    raise Exception("DashScope did not return a valid embedding for text input.")


def _call_embed_for_media(file_path: str, media_type: str) -> list[float]:
    """对媒体文件（image/video/audio）调用 DashScope 多模态融合嵌入。"""
    input_item: dict[str, str] = {}
    if media_type.startswith("image/"):
        input_item["image"] = _normalize_local_media_path(file_path, "image")
    elif media_type.startswith("video/"):
        input_item["video"] = _normalize_local_media_path(file_path, "video")
    elif media_type.startswith("audio/"):
        input_item["audio"] = _normalize_local_media_path(file_path, "audio")
    else:
        raise ValueError(f"Unsupported media type for direct embedding: {media_type}")

    result = call_dashscope_multimodal_fusion_embedding(input_data=[input_item])
    output = result.get("output", {})
    if isinstance(output, dict):
        embedding = output.get("embedding")
        if isinstance(embedding, list) and embedding:
            return embedding
    raise Exception(f"DashScope did not return a valid embedding for media input: {file_path}")


def process_embed_request(file_path: str, file_id: int, file_type: str) -> EmbedResponse:
    """
    文件嵌入管线：
    1. 判断文件类型是否被 DashScope 原生支持
    2. 原生媒体 → 直接调用 DashScope，返回 1 个向量
    3. 不支持 → 提取文本 → 分块 → 逐块调用 DashScope → 返回 N 个向量
    """
    # 校验文件存在
    resolved = Path(file_path).resolve()
    if not resolved.is_file():
        raise ValueError(f"File not found: {resolved}")

    normalized_type = file_type.lower().split(";")[0].strip()

    if _is_native_media_type(normalized_type):
        # 直接作为媒体文件发送给 DashScope
        logger.info("File %s (%s) → native media, direct embedding", file_id, normalized_type)
        vector = _call_embed_for_media(str(resolved), normalized_type)
        return EmbedResponse(
            file_id=file_id,
            vectors=[EmbedVectorItem(number=0, vector=vector)],
            chunk_count=1,
        )

    # 不支持直接嵌入 → 提取文本 → 分块 → 逐块嵌入
    parser = get_parser(str(resolved), normalized_type)
    if parser is None:
        raise ValueError(
            f"No parser available for file type '{normalized_type}' "
            f"(file_id={file_id}, path={resolved})"
        )

    logger.info("File %s (%s) → extract text via %s", file_id, normalized_type, type(parser).__name__)
    extracted_text = parser.extract_text(str(resolved))

    if not extracted_text or not extracted_text.strip():
        raise ValueError(f"File {file_id} produced empty text after extraction")

    logger.info("File %s → extracted %d chars, chunking...", file_id, len(extracted_text))
    chunks = list(chunk_text(extracted_text))

    if not chunks:
        raise ValueError(f"File {file_id} chunking produced no chunks")

    logger.info("File %s → %d chunks, generating embeddings...", file_id, len(chunks))

    vectors: list[EmbedVectorItem] = []
    for i, chunk in enumerate(chunks):
        try:
            vec = _call_embed_for_text(chunk)
            vectors.append(EmbedVectorItem(number=i, vector=vec))
        except Exception as e:
            logger.error("Embedding failed for file %s chunk %d: %s", file_id, i, e)
            raise Exception(f"Failed to embed chunk {i} of file {file_id}: {e}") from e

    return EmbedResponse(
        file_id=file_id,
        vectors=vectors,
        chunk_count=len(chunks),
    )

