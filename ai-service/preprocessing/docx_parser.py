"""
DOCX 文本提取器，按段落逐段读取，支持大文档。
"""

from .parser import BaseParser


class DocxParser(BaseParser):
    """使用 python-docx 提取 .docx 文件中的文本。"""

    def extract_text(self, file_path: str) -> str:
        from docx import Document

        doc = Document(file_path)
        parts: list[str] = []
        for para in doc.paragraphs:
            text = para.text
            if text:
                parts.append(text)
        return "\n\n".join(parts)
