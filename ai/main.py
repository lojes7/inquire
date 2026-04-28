from fastapi import FastAPI
from fastapi.responses import JSONResponse
import uvicorn
import traceback
from schemas import AskRequest, AskResponse
from service import process_ask_request
from config import settings
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
            content=AskResponse(answer=f"Parameter Error: {str(ve)}", status="error").model_dump()
        )
    except Exception as e:
        traceback.print_exc()
        return JSONResponse(
            status_code=500,
            content=AskResponse(answer=f"Server Error: {str(e)}", status="error").model_dump()
        )
@app.get("/health")
def health_check():
    return {"status": "ok", "service": "Inquire AI with DashScope"}
if __name__ == "__main__":
    uvicorn.run("main:app", host="0.0.0.0", port=settings.port, reload=True)
