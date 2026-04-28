from abc import ABC, abstractmethod


class BaseParser(ABC):
    """文件文本提取抽象基类，所有 parser 需实现 extract_text 方法。"""

    @abstractmethod
    def extract_text(self, file_path: str) -> str:
        """从文件路径提取纯文本，返回提取的文本字符串。"""
        ...
