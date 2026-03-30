// src/pages/ChatApp.jsx
import { useState, useEffect } from 'react';
import '../styles/ChatApp.css';
import { useNavigate, useLocation, useParams } from "react-router-dom";

const ChatApp = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const params = useParams();

  // ===== token 获取统一方法 =====
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
  const token = getToken();

  // ===== 页面状态 =====
  const [contacts, setContacts] = useState([]);
  const [activeContact, setActiveContact] = useState(null);
  const [messages, setMessages] = useState([]);
  const [input, setInput] = useState('');
  const [loadingContacts, setLoadingContacts] = useState(false);
  const [loadingMessages, setLoadingMessages] = useState(false);

  const friendName = location.state?.friendName;
  const conversationIdFromState = location.state?.conversationId;

  // ===== 获取联系人列表 =====
  const fetchContacts = async () => {
    try {
      setLoadingContacts(true);
      const res = await fetch('http://localhost:8000/api/auth/conversations', {
        headers: { 'Authorization': `Bearer ${token}` },
      });
      const data = await res.json();
      if (data.code === 200) {
        setContacts(data.data);
      } else {
        console.error('加载联系人失败:', data.message);
      }
    } catch (err) {
      console.error(err);
    } finally {
      setLoadingContacts(false);
    }
  };

  // ===== 获取聊天记录 =====
  const fetchMessages = async (conversation_id) => {
    if (!conversation_id) return;
    try {
      setLoadingMessages(true);
      const res = await fetch(
        `http://localhost:8000/api/auth/conversations/${conversation_id}`,
        { headers: { Authorization: `Bearer ${token}` } }
      );
      const data = await res.json();
      if (data.code === 200) {
        setMessages(data.data);
      } else {
        console.error('加载消息失败:', data.message);
        setMessages([]);
      }
    } catch (err) {
      console.error(err);
      setMessages([]);
    } finally {
      setLoadingMessages(false);
    }
  };

  // ===== 初始化和点击联系人 =====
  useEffect(() => {
    fetchContacts();
  }, []);

  useEffect(() => {
    // 选择要激活的会话
    const id = params.conversationId || conversationIdFromState;
    if (!id || contacts.length === 0) return;

    const contact = contacts.find(c => c.conversation_id === id) || {
      conversation_id: id,
      name: friendName || "聊天对象",
    };

    setActiveContact(contact);
    fetchMessages(id);
  }, [params.conversationId, conversationIdFromState, contacts]);

  // ===== 点击联系人 =====
  const handleSelectContact = (contact) => {
    setActiveContact(contact);
    fetchMessages(contact.conversation_id);
    navigate(`/chat/${contact.conversation_id}`); // 更新 URL
  };

  // ===== 发送消息 =====
  const sendMessage = async () => {
    if (!input.trim() || !activeContact || !activeContact.conversation_id) return;

    try {
      const res = await fetch('http://localhost:8000/api/auth/messages/text', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`,
        },
        body: JSON.stringify({
          conversation_id: String(activeContact.conversation_id),
          content: input,
        }),
      });

      const data = await res.json();
      console.log("发送消息返回:", data);

      if (data.code === 201) {
        setMessages([
          ...messages,
          {
            message_id: data.data,
            sender_id: 'me',
            sender_name: '我',
            status: 0,
            updated_at: new Date().toISOString(),
            content: input,
          },
        ]);
        setInput('');
      } else {
        console.error('发送失败:', data.message);
      }
    } catch (err) {
      console.error('请求异常:', err);
    }
  };

  return (
    <div className="chat-app">
      {/* 左侧导航栏 */}
      <div className="chat-sidebar">
        <div className="nav-buttons">
          <button onClick={() => navigate("/chat")}>💬</button>
          <button onClick={() => navigate("/addfriend")}>👥</button>
          <button onClick={() => navigate("/chatpage")}>📝</button>
          <button>➕</button>
        </div>
        <div className="sidebar-footer">
          <button>🔔</button>
          <button onClick={() => navigate("/persional")}>⚙️</button>
        </div>
      </div>

      {/* 中间联系人列表 */}
      <div className="chat-middle">
        <div className="search">
          <input type="text" placeholder="🔍搜索联系人" />
        </div>
        {loadingContacts ? (
          <p>加载联系人中...</p>
        ) : (
          <div className="contacts">
            {contacts.map((c) => (
              <div
                key={c.conversation_id}
                className={`contact ${activeContact?.conversation_id === c.conversation_id ? 'active' : ''}`}
                onClick={() => handleSelectContact(c)}
              >
                <div className="contact-avatar"></div>
                <div className="contact-info">
                  <div className="contact-name">
                    <span>{c.name}</span>
                    <span>{c.time}</span>
                  </div>
                  <div className="contact-message">{c.last_message}</div>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      {/* 右侧聊天区 */}
      <div className="chat-main">
        {!activeContact ? (
          <div className="main-content">
            <div className="icon">💬</div>
            <h2>高效办公，文件秒寻</h2>
            <p>选择一个会话开始聊天，或使用AI检索快速找到文件</p>
            <button>+ 创建会话</button>
          </div>
        ) : (
          <div className="chat-panel">
            <div className="chat-top">
              <span className="chat-name">{activeContact.name}</span>
              <button className="chat-back" onClick={() => setActiveContact(null)}>
                返回
              </button>
            </div>

            <div className="chat-body">
              {loadingMessages ? (
                <p>加载消息中...</p>
              ) : messages.length === 0 ? (
                <p>暂无消息</p>
              ) : (
                messages.map((msg) => (
                  <div key={msg.message_id} className={`msg ${msg.sender_id === 'me' ? 'self' : 'other'}`}>
                    {msg.sender_id !== 'me' && <div className="avatar other" />}
                    <div className="bubble">
                      {msg.status === 0
                        ? msg.content
                        : msg.status === 3
                        ? <a href={msg.content.file_url} target="_blank" rel="noreferrer">{msg.content.file_name}</a>
                        : '[系统消息]'}
                    </div>
                    {msg.sender_id === 'me' && <div className="avatar self" />}
                  </div>
                ))
              )}
            </div>

            <div className="chat-footer">
              <div className="chat-tools">
                <button title="文件检索">📎</button>
                <button title="表情">😊</button>
                <button title="链接">🔗</button>
              </div>
              <input
                value={input}
                onChange={(e) => setInput(e.target.value)}
                placeholder="输入消息…"
                onKeyDown={(e) => {
                  if (e.key === 'Enter' && !e.shiftKey) {
                    e.preventDefault();
                    sendMessage();
                  }
                }}
              />
              <button className="send-btn" onClick={sendMessage}>发送</button>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};

export default ChatApp;