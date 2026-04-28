from fastapi import FastAPI
from fastapi.responses import JSONResponse
import uvicorn
import traceback
import logging

from schemas import AskRequest, AskResponse, EmbedRequest, EmbedResponse
from service import process_ask_request, process_embed_request
from config import settings

logging.basicConfig(level=logging.INFO, format="%(asctime)s [%(levelname)s] %(name)s: %(message)s")

app = FastAPI(
    title="Inquire AI Service",
    description="AI embedding layer using DashScope MultiModalEmbedding API",
)


@app.post("/ask", response_model=AskResponse)
async def ask_endpoint(request: AskRequest):
    try:
        answer = process_ask_request(request)
        return AskResponse(answer=answer, status="success")
    except ValueError as ve:
        return JSONResponse(
            status_code=400,
            content=AskResponse(answer=f"Parameter Error: {str(ve)}", status="error").model_dump(),
        )
    except Exception as e:
        traceback.print_exc()
        return JSONResponse(
            status_code=500,
            content=AskResponse(answer=f"Server Error: {str(e)}", status="error").model_dump(),
        )


@app.post("/embed", response_model=EmbedResponse)
async def embed_endpoint(request: EmbedRequest):
    """文件嵌入端点：接收文件路径，返回分块的向量列表。"""
    try:
        result = process_embed_request(
            file_path=request.file_path,
            file_id=request.file_id,
            file_type=request.file_type,
        )
        return result
    except ValueError as ve:
        return JSONResponse(
            status_code=400,
            content=EmbedResponse(
                status="error",
                file_id=request.file_id,
                answer=f"Parameter Error: {str(ve)}",
            ).model_dump(),
        )
    except Exception as e:
        traceback.print_exc()
        return JSONResponse(
            status_code=500,
            content=EmbedResponse(
                status="error",
                file_id=request.file_id,
                answer=f"Server Error: {str(e)}",
            ).model_dump(),
        )


@app.get("/health")
def health_check():
    return {"status": "ok", "service": "Inquire AI with DashScope"}


if __name__ == "__main__":
    uvicorn.run("main:app", host="0.0.0.0", port=settings.port, reload=True)
