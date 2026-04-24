from typing import Any, Literal

from pydantic import BaseModel, Field, field_validator


class AskRequest(BaseModel):
    input_text: str = Field(..., description="Input text for embedding")
    model: str | None = Field(
        default=None,
        description="Optional embedding model override, defaults to server config",
    )

    @field_validator("input_text")
    @classmethod
    def validate_input_text(cls, value: str) -> str:
        if not value or not value.strip():
            raise ValueError("input_text cannot be empty.")
        return value


class AskResponse(BaseModel):
    answer: Any
    status: Literal["success", "error"]
