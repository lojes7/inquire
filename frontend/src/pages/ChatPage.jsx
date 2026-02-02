import "../styles//ChatPage.css";
import { useNavigate } from "react-router-dom";

export default function ChatPage() {

const navigate = useNavigate();
  return (
    <div className="chatpage-app">
      {/* 左侧导航栏 */}
      <div className="chat-sidebar">
        <div className="nav-buttons">
          <button onClick={() => navigate("/chat")}>💬</button>
          <button onClick={() => navigate("/addfriend")}>👥</button>
          <button onClick={() => navigate("/chatpage")}>📝</button>
          <button>➕</button>
        </div>
        {/* 底部设置按钮 */}
        <div className="sidebar-footer">
          <button>🔔</button>
           <button onClick={() => navigate("/persional")}>
            ⚙️</button>
        </div>
      </div>

      {/* 中间会话列表 */}
      <aside className="chat-middle">
        <button className="new-chat">＋ 新建对话</button>
        <div className="chat-list">
          <div className="chat-item active">
            产品需求文档查询
            <small>10分钟前</small>
          </div>
          <div className="chat-item">
            竞品分析相关资料
            <small>1小时前</small>
          </div>
          <div className="chat-item">
            设计规范文档
            <small>昨天</small>
          </div>
        </div>
      </aside>

      {/* 右侧聊天主区 */}
      <main className="chat-main">
        <header className="header">
          <div className="header-left">全部知识库</div>
          <div className="header-right">
            <button>清空对话</button>
            <button>导出结果</button>
          </div>
        </header>

        <section className="content">
          <div className="message user">
            帮我找一下产品设计相关的文档
          </div>
          <div className="message bot">
            好的，我为您找到了 3 份相关文档：
            <div className="doc-list">
              <div className="doc-item">
                <span>产品需求文档_v3.2.docx</span>
                <small>2.5MB</small>
              </div>
              <div className="doc-item">
                <span>界面设计规范_2024.pdf</span>
                <small>4.8MB</small>
              </div>
              <div className="doc-item">
                <span>竞品分析报告.pdf</span>
                <small>3.6MB</small>
              </div>
            </div>
          </div>
          <div className="message user">
            能详细说明一下产品需求文档的主要内容吗？
          </div>
          <div className="message bot">
            产品需求文档主要包含以下几个部分：<br />
            1. 产品概述和目标用户<br />
            2. 功能需求详细说明<br />
            3. 非功能性需求<br />
            4. 用户体验设计要求<br />
            5. 技术实现方案
          </div>
        </section>

        <footer className="footer">
          <div className="actions">
            <button>搜索文档</button>
            <button>数据分析</button>
            <button>内容总结</button>
            <button>智能推荐</button>
          </div>
          <div className="input-box">
            <textarea placeholder="输入您的问题，Enter发送，Shift+Enter换行" />
            <button className="send-btn">➤</button>
          </div>
        </footer>
      </main>
    </div>
  );
}
