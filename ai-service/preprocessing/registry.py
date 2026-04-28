"""
解析器注册表，通过文件扩展名与 MIME 类型路由到对应的文本提取 parser。

新增格式时只需在此处注册即可，无需修改核心管线代码。
"""

from pathlib import Path

from .parser import BaseParser
from .txt_parser import TextParser
from .pdf_parser import PDFParser
from .docx_parser import DocxParser
from .pptx_parser import PptxParser

# 扩展名 → Parser 映射
_EXTENSION_PARSERS: dict[str, BaseParser] = {
    ".txt": TextParser(),
    ".log": TextParser(),
    ".csv": TextParser(),
    ".md": TextParser(),
    ".html": TextParser(),
    ".htm": TextParser(),
    ".xml": TextParser(),
    ".json": TextParser(),
    ".yaml": TextParser(),
    ".yml": TextParser(),
    ".py": TextParser(),
    ".js": TextParser(),
    ".ts": TextParser(),
    ".go": TextParser(),
    ".java": TextParser(),
    ".c": TextParser(),
    ".cpp": TextParser(),
    ".h": TextParser(),
    ".sh": TextParser(),
    ".pdf": PDFParser(),
    ".docx": DocxParser(),
    ".pptx": PptxParser(),
}

# MIME 类型 → Parser 映射（优先级低于扩展名）
_MIME_PARSERS: dict[str, BaseParser] = {
    "text/plain": TextParser(),
    "text/csv": TextParser(),
    "text/markdown": TextParser(),
    "text/html": TextParser(),
    "text/xml": TextParser(),
    "application/json": TextParser(),
    "application/xml": TextParser(),
    "application/pdf": PDFParser(),
    "application/vnd.openxmlformats-officedocument.wordprocessingml.document": DocxParser(),
    "application/vnd.openxmlformats-officedocument.presentationml.presentation": PptxParser(),
    # Go 的 http.DetectContentType 将 ZIP 系 Office 文档识别为 application/zip
    "application/zip": TextParser(),  # 兜底: 优先走扩展名匹配，无匹配时当纯文本
    "application/octet-stream": TextParser(),  # 兜底: 当纯文本尝试
}


def get_parser(file_path: str, mime_type: str | None = None) -> BaseParser | None:
    """根据文件路径和 MIME 类型查找对应的 parser，未匹配则返回 None。"""

    ext = Path(file_path).suffix.lower()
    if ext in _EXTENSION_PARSERS:
        return _EXTENSION_PARSERS[ext]

    if mime_type:
        mime = mime_type.lower().split(";")[0].strip()
        if mime in _MIME_PARSERS:
            return _MIME_PARSERS[mime]

    return None


def get_parser_for_file(file_path: str, mime_type: str | None = None) -> BaseParser:
    """
    与 get_parser 相同，但在未匹配时抛出 ValueError，
    上层无需额外判空。
    """
    parser = get_parser(file_path, mime_type)
    if parser is None:
        ext = Path(file_path).suffix.lower()
        raise ValueError(f"No parser registered for file extension '{ext}' or MIME type '{mime_type}'")
    return parser
