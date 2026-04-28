import { useState } from "react";
import "../styles/ChatPage.css";
import Sidebar from "../components/Sidebar";

export default function ChatPage() {
  const [selectedFile, setSelectedFile] = useState(null);

  const handleFileChange = (e) => {
    const file = e.target.files[0];
    if (file) setSelectedFile(file);
  };

  return (
    <div className="chatpage-app">
      <Sidebar />

      {/* ✅ 主体（两栏） */}
      <main className="workspace">
        
        {/* ================= 左栏：文件区 ================= */}
        <aside className="file-panel">
          <div className="file-header">
            <h2>知识资产</h2>

            {/* 上传 */}
            <div
              className="upload-box"
              onClick={() =>
                document.getElementById("fileUpload").click()
              }
            >
              <span>📄</span>
              <p>上传文件</p>
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
              </p>
            )}
          </div>

          {/* 文件列表 */}
          <div className="file-list">
            <div className="file-item active">
              Q3_财务报告.pdf
            </div>
            <div className="file-item">
              项目策划书.docx
            </div>
            <div className="file-item">
              竞品分析.pdf
            </div>
          </div>
        </aside>

        {/* ================= 右栏：聊天区 ================= */}
       <section className="chat-area">

        {/* 顶部 */}
        <header className="chat-header">
          <div className="current-file">
            📎 当前上下文：Q3_财务报告.pdf
          </div>

          <div className="header-actions">
            <button className="action-btn clear">
              <span>🗑</span>
              清空
            </button>
            <button className="action-btn export">
              <span>⬇️</span>
              导出
            </button>
          </div>
        </header>

        {/* 聊天内容（滚动区） */}
        <div className="chat-content">

          {/* AI */}
          <div className="chat-row bot">
            <div className="avatar">🤖</div>
            <div className="message bot">
              AI 正在分析财务数据...
            </div>
          </div>

          {/* 用户 */}
          <div className="chat-row user">
            <div className="message user">
              请分析财务趋势
            </div>
          </div>

          {/* AI */}
          <div className="chat-row bot">
            <div className="avatar">🤖</div>
            <div className="message bot">
              增长 14.5%，风险在供应链波动
            </div>
          </div>

        </div>

        {/* 输入区（固定底部） */}
        <footer className="chat-footer">
          <div className="input-box">

            <textarea placeholder="输入你的问题..." />

            <button className="send-btn">➤</button>

          </div>
        </footer>

      </section>
      </main>
    </div>
  );
}