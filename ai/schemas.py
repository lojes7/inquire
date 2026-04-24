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
    mode: Literal["multimodal_fusion"] = Field(
        default="multimodal_fusion",
        description="Request mode, only multimodal_fusion is supported",
    )
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
            raise ValueError("input_data cannot be empty when mode=multimodal_fusion.")
        return self


class AskResponse(BaseModel):
    answer: Any
    status: Literal["success", "error"]
