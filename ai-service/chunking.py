"""
Token 感知的文本分块模块。

设计要点：
1. Token 估算：中文 ~1 字符/token，英文 ~4 字符/token，采用加权平均 ~2.5 字符/token。
2. 优先在自然边界（段落、句子）处分割，保留上下文语义连贯性。
3. 生成器模式：逐块产出，不在内存中同时持有所有块。
4. 配置化的 chunk_size 与 overlap，方便后续调优。
"""

import re
from collections.abc import Generator

# 默认配置：最大 512 tokens / 块，50 tokens 重叠
DEFAULT_CHUNK_TOKENS = 512
DEFAULT_OVERLAP_TOKENS = 50
# 粗略估算：默认约 2.5 字符 ≈ 1 token（中英文混合的保守估计）
_CHARS_PER_TOKEN = 2.5

# 段落分隔（连续空行）
_PARAGRAPH_PATTERN = re.compile(r"\n\s*\n")
# 句子分隔（中英文句末标点后跟空格/换行）
_SENTENCE_PATTERN = re.compile(r"(?<=[。！？.!?\n])\s*")


def estimate_tokens(text: str) -> int:
    """粗略估算文本的 token 数量。"""
    if not text:
        return 0
    return max(1, int(len(text) / _CHARS_PER_TOKEN))


def chunk_text(
    text: str,
    chunk_tokens: int = DEFAULT_CHUNK_TOKENS,
    overlap_tokens: int = DEFAULT_OVERLAP_TOKENS,
) -> Generator[str, None, None]:
    """
    将文本按 token 感知的方式切分为多个块。

    Args:
        text: 待分块的原始文本。
        chunk_tokens: 每块最大 token 数（默认 512）。
        overlap_tokens: 相邻块之间的重叠 token 数（默认 50）。

    Yields:
        每个文本块。
    """
    if not text or not text.strip():
        return

    max_chars = int(chunk_tokens * _CHARS_PER_TOKEN)
    overlap_chars = int(overlap_tokens * _CHARS_PER_TOKEN)

    if len(text) <= max_chars:
        yield text
        return

    # 第一步：按段落切分
    paragraphs = _PARAGRAPH_PATTERN.split(text)
    paragraphs = [p.strip() for p in paragraphs if p.strip()]

    # 对每个段落，如果仍然超出 max_chars，按句子切分
    segments: list[str] = []
    for para in paragraphs:
        if len(para) <= max_chars:
            segments.append(para)
        else:
            sentences = _SENTENCE_PATTERN.split(para)
            sentences = [s.strip() for s in sentences if s.strip()]
            for sent in sentences:
                if len(sent) <= max_chars:
                    segments.append(sent)
                else:
                    # 句子仍然过长（极少见的长句/无标点），强制按字符数切分
                    for i in range(0, len(sent), max_chars):
                        segments.append(sent[i : i + max_chars])

    # 第二步：将 segments 合并成 chunk，保留 overlap
    current_chunk: list[str] = []
    current_len = 0
    overlap_buffer: list[str] = []

    for seg in segments:
        seg_len = len(seg)

        if current_len + seg_len <= max_chars:
            current_chunk.append(seg)
            current_len += seg_len
        else:
            # 当前块已满，产出
            if current_chunk:
                yield "\n\n".join(current_chunk)

            # 根据 overlap 保留尾部上下文，作为下一块的起始
            if overlap_chars > 0 and current_chunk:
                overlap_text = "\n\n".join(current_chunk)
                tail = _extract_tail(overlap_text, overlap_chars)
                overlap_buffer = [tail] if tail else []
                current_chunk = list(overlap_buffer)
                current_len = sum(len(c) for c in current_chunk)
            else:
                current_chunk = []
                current_len = 0

            # 将当前 segment 放入下一块
            if seg_len <= max_chars:
                current_chunk.append(seg)
                current_len += seg_len
            else:
                # segment 本身超出 max_chars（已经是强制切分后的），直接逐段产出
                for i in range(0, len(seg), max_chars):
                    sub = seg[i : i + max_chars]
                    yield sub

    # 产出最后剩余的块
    if current_chunk:
        yield "\n\n".join(current_chunk)


def _extract_tail(text: str, max_chars: int) -> str:
    """从文本末尾提取最多 max_chars 个字符作为重叠上下文。"""
    if len(text) <= max_chars:
        return text
    return text[-max_chars:]
