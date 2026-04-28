from typing import Any, Literal

from pydantic import BaseModel, Field, model_validator


class MultiModalInputItem(BaseModel):
    text: str | None = Field(default=None, description="Text content")
    image: str | None = Field(default=None, description="Local absolute image path")
    video: str | None = Field(default=None, description="Local absolute video path")
    audio: str | None = Field(default=None, description="Local absolute audio path")
    factor: float | None = Field(default=None, description="Optional modality weight")

    @model_validator(mode="after")
    def validate_single_modality(self) -> "MultiModalInputItem":
        modality_values = {
            "text": self.text,
            "image": self.image,
            "video": self.video,
            "audio": self.audio,
        }
        filled_modalities = [
            key for key, value in modality_values.items() if isinstance(value, str) and value.strip()
        ]
        if len(filled_modalities) != 1:
            raise ValueError(
                "Each input_data item must contain exactly one non-empty field among text/image/video/audio.",
            )
        return self


class AskRequest(BaseModel):
    input_data: list[MultiModalInputItem] = Field(
        ...,
        min_length=1,
        description="Multimodal input list for fusion embedding",
    )
    model: str | None = Field(
        default=None,
        description="Optional embedding model override, defaults to server config",
    )

    @model_validator(mode="after")
    def validate_mode_payload(self) -> "AskRequest":
        if not self.input_data:
            raise ValueError("input_data cannot be empty.")
        return self


class AskResponse(BaseModel):
    answer: Any
    status: Literal["success", "error"]


# -------------------- 文件嵌入相关 schema --------------------


class EmbedRequest(BaseModel):
    """文件嵌入请求体，Backend 调用 /embed 时传入。"""

    file_path: str = Field(..., description="文件在容器中的绝对路径")
    file_id: int = Field(..., description="files 表中的文件 ID", gt=0)
    file_type: str = Field(..., description="文件 MIME 类型，如 application/pdf")


class EmbedVectorItem(BaseModel):
    """单个向量块。"""

    number: int = Field(..., description="向量分片序号，从 0 递增")
    vector: list[float] = Field(..., description="1024 维浮点向量")


class EmbedResponse(BaseModel):
    """文件嵌入响应体。"""

    status: Literal["success", "error"] = "success"
    file_id: int = Field(..., description="与请求中相同的 file_id")
    vectors: list[EmbedVectorItem] = Field(default_factory=list, description="向量列表")
    chunk_count: int = Field(0, description="分块数量")
    answer: str | None = Field(default=None, description="错误时的描述信息")
