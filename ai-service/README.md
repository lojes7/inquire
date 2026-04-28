# Inquire AI Service

## 功能说明

本服务支持多模态融合向量（`qwen3-vl-embedding`）。

多模态融合向量固定使用以下参数：

- `enable_fusion=true`
- `dimension=1024`

## 环境要求

- Python 3.11+
- DashScope API Key

## 配置说明

在 .env 中配置（未配置则使用默认值）：

- `DASHSCOPE_API_KEY`：必填
- `PORT`：服务端口，默认 `8001`

## 本地运行

1. `cd ai`
2. `pip install -r requirements.txt`
3. `python main.py`

## 接口说明

### POST `/ask`

接口固定执行多模态融合向量

### 多模态融合向量请求（本地绝对路径）

```bash
curl -X POST http://localhost:8001/ask \
   -H "Content-Type: application/json" \
   -d '{
      "input_data": [
         {"text": "这是一段测试文本，用于生成多模态融合向量"},
         {"image": "/data/inquire/media/demo.png"},
         {"video": "/data/inquire/media/demo.mp4"}
      ]
   }'
```

## 返回结构

`status=success` 时，`answer` 中包含：

- `model`
- `output`
- `usage`
- `request_id`

多模态请求还会返回：

- `enable_fusion`（固定为 `true`）
- `dimension`（固定为 `1024`）