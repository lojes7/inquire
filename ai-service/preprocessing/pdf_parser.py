"""
PDF 文本提取器，使用 PyMuPDF (fitz) 按页逐页读取，对大文件内存友好。
"""

from .parser import BaseParser


class PDFParser(BaseParser):
    """使用 PyMuPDF 提取 PDF 中的文本。"""

    def extract_text(self, file_path: str) -> str:
        import fitz  # PyMuPDF

        doc = fitz.open(file_path)
        try:
            parts: list[str] = []
            for page in doc:  # type: ignore
                text = page.get_text()
                if text:
                    parts.append(text)
            return "\n\n".join(parts)
        finally:
            doc.close()
