# Inquire AI Service 

## 至少需要
- Python 3.11+
- Aliyun DashScope API Key

## How to run locally
1. `cd ai`
2. 在 `.env` 文件中配置：
   - `DASHSCOPE_API_KEY`
3. `pip install -r requirements.txt`
4. `python main.py`

## 请求示例

### Text Embedding
```bash
curl -X POST http://localhost:8001/ask \
     -H "Content-Type: application/json" \
     -d '{"input_text": "衣服的质量杠杠的"}'
```

返回的 `answer` 字段包含 `model`、`output`、`usage` 和 `request_id`。