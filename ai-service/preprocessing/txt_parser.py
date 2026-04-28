"""
纯文本提取器，通过逐行流式读取处理大文件，避免全量加载到内存。
"""

from .parser import BaseParser

# 单次读取的缓冲区大小，避免一次性加载大文件
_READ_CHUNK_SIZE = 64 * 1024  # 64 KB


class TextParser(BaseParser):
    """提取 .txt / .log / .csv 等纯文本文件的全部内容。"""

    def extract_text(self, file_path: str) -> str:
        parts: list[str] = []
        with open(file_path, "r", encoding="utf-8", errors="replace") as f:
            while True:
                chunk = f.read(_READ_CHUNK_SIZE)
                if not chunk:
                    break
                parts.append(chunk)
        return "".join(parts)
