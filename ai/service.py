from client import call_dashscope_text_embedding
from schemas import AskRequest


def process_ask_request(request: AskRequest) -> dict:
    """Process embedding request with optional model override."""
    input_text = request.input_text.strip()
    model = request.model.strip() if request.model else None
    return call_dashscope_text_embedding(input_text=input_text, model=model)
