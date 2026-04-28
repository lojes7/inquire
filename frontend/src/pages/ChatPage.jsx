import { useState, useEffect, useRef } from "react";
import "../styles/ChatPage.css";
import Sidebar from "../components/Sidebar";

/*
=========================
获取 token
=========================
*/
const getToken = () => {
  const stored = sessionStorage.getItem("token");
  if (!stored) return null;

  try {
    const parsed = JSON.parse(stored);
    return parsed?.token || stored;
  } catch {
    return stored;
  }
};

export default function ChatPage() {
  /* =========================
     状态
  ========================= */
  const [files, setFiles] = useState([]);
  const [selectedFile, setSelectedFile] = useState(null);
  const [activeFileId, setActiveFileId] = useState(null);
  const [messages, setMessages] = useState([]);

  const [uploading, setUploading] = useState(false);
  const [uploadSuccess, setUploadSuccess] = useState(false);

  const [token, setToken] = useState(null);

  // ✅ 新增聊天状态
  const [input, setInput] = useState("");
  const [sending, setSending] = useState(false);

  const messagesEndRef = useRef(null);

  const conversationId = "1";

  /* =========================
     初始化 token
  ========================= */
  useEffect(() => {
    const t = getToken();
    setToken(t);
  }, []);

  /* =========================
     自动滚动
  ========================= */
  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [messages]);

  /* =========================
     获取文件列表
  ========================= */
  const fetchFiles = async (tk) => {
    if (!tk) return;

    try {
      const res = await fetch("/api/auth/files", {
        headers: {
          Authorization: `Bearer ${tk}`,
        },
      });

      let data = null;

      try {
        const text = await res.text();

        console.log("原始返回：", text); // 🔥 用来排查后端问题

        data = text ? JSON.parse(text) : {};
      } catch (err) {
        console.error("JSON解析失败：", err);
        data = {};
      }

      const fileList =
        data.files ||
        data.data?.files ||
        data.data?.list ||
        data.list ||
        [];

      setFiles(fileList);
    } catch (err) {
      console.error("获取文件失败", err);
    }
  };

  useEffect(() => {
    if (token) fetchFiles(token);
  }, [token]);

  /* =========================
     上传文件
  ========================= */
  const handleFileChange = async (e) => {
    const file = e.target.files[0];
    if (!file) return;

    if (!token) {
      alert("未登录");
      return;
    }

    setSelectedFile(file);
    setUploading(true);
    setUploadSuccess(false);

    const formData = new FormData();
    formData.append("file", file);

    try {
      const res = await fetch("/api/auth/files/upload", {
        method: "POST",
        headers: {
          Authorization: `Bearer ${token}`,
        },
        body: formData,
      });

      const data = await res.json();

      if (res.ok) {
        setUploadSuccess(true);
        await fetchFiles(token);

        if (data.file_info?.file_id) {
          setActiveFileId(String(data.file_info.file_id));
        }
      } else {
        alert(data.message || "上传失败");
      }
    } catch (err) {
      alert(err.message || "上传失败");
    } finally {
      setUploading(false);
    }
  };

  /* =========================
     点击文件（保留）
  ========================= */
  const sendWorkspaceFile = async (file) => {
    if (!token) {
      alert("未登录");
      return;
    }

    setActiveFileId(String(file.file_id));

    try {
      const res = await fetch("/api/auth/messages/workspace-file", {
        method: "POST",
        headers: {
          Authorization: `Bearer ${token}`,
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          conversation_id: conversationId,
          file_id: file.file_id,
        }),
      });

      const data = await res.json();

      if (res.ok) {
        setMessages((prev) => [
          ...prev,
          { type: "user", content: `发送文件：${file.file_name}` },
          { type: "bot", content: `已接收文件：${file.file_name}` },
        ]);
      }
    } catch (err) {
      console.error(err);
    }
  };

  /* =========================
     ✅ AI对话（新增）
  ========================= */
  const sendMessage = async () => {
  if (!input.trim()) return;

  if (!token) {
    alert("未登录");
    return;
  }

  const userText = input;

  setMessages((prev) => [
    ...prev,
    { type: "user", content: userText },
  ]);

  setInput("");
  setSending(true);

  try {
    const res = await fetch("/api/auth/files/search", {
      method: "POST",
      headers: {
        Authorization: `Bearer ${token}`,
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        query: userText,
      }),
    });

    // ✅ 安全解析
    let data = null;

    try {
      const text = await res.text();
      console.log("原始返回：", text);
      data = text ? JSON.parse(text) : {};
    } catch (err) {
      console.error("JSON解析失败：", err);
      data = {};
    }

    if (res.ok) {
      const fileList = data.files || data.data?.files || [];

      if (fileList.length === 0) {
        setMessages((prev) => [
          ...prev,
          { type: "bot", content: "未找到相关内容" },
        ]);
      } else {
        const text = fileList
          .map(
            (f, i) =>
              `${i + 1}. ${f.file_name}（score: ${f.score}）`
          )
          .join("\n");

        setMessages((prev) => [
          ...prev,
          { type: "bot", content: text },
        ]);
      }
    } else {
      alert(data.message || `请求失败（${res.status}）`);
    }
  } catch (err) {
    console.error("请求异常：", err);
    alert("请求失败");
  } finally {
    setSending(false);
  }
};

  return (
    <div className="chatpage-app">
      <Sidebar />

      <main className="workspace">
        {/* 左侧文件区（保留） */}
        <aside className="file-panel">
          <div className="file-header">
            <h2>知识资产</h2>

            <div
              className="upload-box"
              onClick={() =>
                document.getElementById("fileUpload").click()
              }
            >
              📄 上传文件
            </div>

            <input
              id="fileUpload"
              type="file"
              hidden
              onChange={handleFileChange}
            />

            {selectedFile && (
              <p className="file-name">
                已选择: {selectedFile.name}
                {uploading && "（上传中...）"}
                {uploadSuccess && "（成功 ✅）"}
              </p>
            )}
          </div>

          <div className="file-list">
            {!files || files.length === 0 ? (
              <p>暂无文件</p>
            ) : (
              files.map((file) => (
                <div
                  key={file.file_id}
                  className={`file-item ${
                    activeFileId === String(file.file_id)
                      ? "active"
                      : ""
                  }`}
                  onClick={() => sendWorkspaceFile(file)}
                >
                  📄 {file.file_name}
                </div>
              ))
            )}
          </div>
        </aside>

        {/* 右侧聊天区（增强） */}
        <section className="chat-area">
          <header className="chat-header">
            🤖 AI 对话
          </header>

          <div className="chat-content">
            {messages.length === 0 && (
              <p className="empty">暂无对话</p>
            )}

            {messages.map((msg, i) => (
              <div
                key={i}
                className={`chat-row ${msg.type}`}
              >
                <div className={`message ${msg.type}`}>
                  {msg.content}
                </div>
              </div>
            ))}

            <div ref={messagesEndRef} />
          </div>

          {/* 输入框 */}
          <div className="chat-input">
            <input
              type="text"
              placeholder="输入你的问题..."
              value={input}
              onChange={(e) => setInput(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === "Enter") sendMessage();
              }}
            />

            <button onClick={sendMessage} disabled={sending}>
              {sending ? "发送中..." : "发送"}
            </button>
          </div>
        </section>
      </main>
    </div>
  );
}