import { useState, useEffect, useRef } from 'react';
import '../styles/ChatApp.css';
import { useNavigate, useLocation, useParams } from "react-router-dom";

const ChatApp = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const params = useParams();

  const messagesEndRef = useRef(null);

  // ===== 获取 token =====
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

  // ===== 获取当前用户ID =====
  const getCurrentUserId = () => {
    try {
      const userStr = sessionStorage.getItem("user");
      if (!userStr) return null;
      const user = JSON.parse(userStr);
      return String(user.id);
    } catch (err) {
      console.error(err);
      return null;
    }
  };
  const currentUserId = getCurrentUserId();

  // ===== 页面状态 =====
  const [contacts, setContacts] = useState([]);
  const [activeContact, setActiveContact] = useState(null);
  const [messages, setMessages] = useState([]);
  const [input, setInput] = useState('');
  const [file, setFile] = useState(null);
  const [loadingContacts, setLoadingContacts] = useState(false);
  const [loadingMessages, setLoadingMessages] = useState(false);

  const friendName = location.state?.friendName;
  const conversationIdFromState = location.state?.conversationId;

  // ===== 自动滚到底部 =====
  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
  };

  // ===== 获取联系人 =====
  const fetchContacts = async () => {
    try {
      setLoadingContacts(true);
      const res = await fetch('http://localhost:8000/api/auth/conversations', {
        headers: { Authorization: `Bearer ${token}` }
      });
      const data = await res.json();
      if (data.code === 200) setContacts(data.data);
    } catch (err) {
      console.error(err);
    } finally {
      setLoadingContacts(false);
    }
  };

  // ===== 获取消息记录 =====
  const fetchMessages = async (conversation_id) => {
    if (!conversation_id) return;
    if (typeof conversation_id === "object") conversation_id = conversation_id.id;
    conversation_id = String(conversation_id);

    try {
      setLoadingMessages(true);
      const res = await fetch(`http://localhost:8000/api/auth/conversations/${conversation_id}`, {
        headers: { Authorization: `Bearer ${token}` }
      });
      const data = await res.json();
      if (data.code === 200) setMessages([...data.data].reverse());
      else setMessages([]);
    } catch (err) {
      console.error(err);
      setMessages([]);
    } finally {
      setLoadingMessages(false);
    }
  };

  // ===== 初始化联系人 =====
  useEffect(() => { fetchContacts(); }, []);

  // ===== 初始化会话 =====
  useEffect(() => {
    let id = params.conversationId || conversationIdFromState;
    if (typeof id === "object") id = id.id;
    if (!id || contacts.length === 0) return;

    const contact = contacts.find(c => String(c.conversation_id) === id) || {
      conversation_id: id,
      name: friendName || "聊天对象"
    };

    setActiveContact(contact);
    fetchMessages(id);
  }, [params.conversationId, conversationIdFromState, contacts]);

  // ===== 自动滚到底部 =====
  useEffect(() => { scrollToBottom(); }, [messages]);

  // ===== 点击联系人 =====
  const handleSelectContact = (contact) => {
    const id = String(contact.conversation_id);
    setActiveContact(contact);
    fetchMessages(id);
    navigate(`/chat/${id}`);
  };

  // ===== 发送文本消息 =====
  const sendMessage = async () => {
    if (!input.trim() || !activeContact) return;
    const conversationId = String(activeContact.conversation_id);

    try {
      const res = await fetch('http://localhost:8000/api/auth/messages/text', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
        body: JSON.stringify({ conversation_id: conversationId, content: input })
      });
      const data = await res.json();
      if (data.code === 201) {
        setMessages(prev => [
          ...prev,
          {
            message_id: data.data,
            sender_id: String(currentUserId),
            content: input,
            status: 0,
            updated_at: new Date().toISOString()
          }
        ]);
        setInput("");
      }
    } catch (err) { console.error(err); }
  };

  // ===== 发送文件消息 =====
  const sendFile = async () => {
    if (!file || !activeContact) return;
    const conversationId = String(activeContact.conversation_id);

    const formData = new FormData();
    formData.append('conversation_id', conversationId);
    formData.append('file', file);

    try {
      const res = await fetch('http://localhost:8000/api/auth/messages/file', {
        method: 'POST',
        headers: { Authorization: `Bearer ${token}` },
        body: formData
      });

      const data = await res.json();

      if (data.code === 201) {
        setMessages(prev => [
          ...prev,
          {
            message_id: data.data,
            sender_id: String(currentUserId),
            content: `[文件] ${file.name}`,
            status: 0,
            updated_at: new Date().toISOString()
          }
        ]);
        setFile(null);
      }
    } catch (err) { console.error(err); }
  };

  return (
    <div className="chat-app">
      {/* 左导航 */}
      <div className="chat-sidebar">
        <div className="nav-buttons">
          <button onClick={() => navigate("/chat")}>💬</button>
          <button onClick={() => navigate("/addfriend")}>👥</button>
          <button onClick={() => navigate("/chatpage")}>📝</button>
          <button onClick={() => navigate("/persional")}>⚙️</button>
        </div>
      </div>

      {/* 联系人 */}
      <div className="chat-middle">
        <div className="search">
          <input placeholder="🔍搜索联系人" />
        </div>

        {loadingContacts ? (
          <p>加载联系人中...</p>
        ) : (
          <div className="contacts">
            {contacts.map((c) => (
              <div
                key={String(c.conversation_id)}
                className={`contact ${String(activeContact?.conversation_id) === String(c.conversation_id) ? 'active' : ''}`}
                onClick={() => handleSelectContact(c)}
              >
                <div className="contact-avatar" />
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

      {/* 聊天区 */}
      <div className="chat-main">
        {!activeContact ? (
          <div className="main-content">
            <div className="icon">💬</div>
            <h2>高效办公，文件秒寻</h2>
            <p>选择一个会话开始聊天</p>
          </div>
        ) : (
          <div className="chat-panel">
            <div className="chat-top">
              <span className="chat-name">{activeContact.name}</span>
              <button className="chat-back" onClick={() => setActiveContact(null)}>返回</button>
            </div>

            <div className="chat-body">
              {loadingMessages ? (
                <p>加载消息中...</p>
              ) : messages.length === 0 ? (
                <p>暂无消息</p>
              ) : (
                messages.map((msg) => {
                  const isMe = String(msg.sender_id) === String(currentUserId);
                  return (
                    <div key={msg.message_id} className={`msg ${isMe ? 'self' : 'other'}`}>
                      {!isMe && <div className="avatar other" />}
                      <div className="bubble">{typeof msg.content === "string" ? msg.content : msg.content?.text || ""}</div>
                      {isMe && <div className="avatar self" />}
                    </div>
                  );
                })
              )}
              <div ref={messagesEndRef} />
            </div>

            <div className="chat-footer">
              <input
                type="text"
                value={input}
                onChange={(e) => setInput(e.target.value)}
                placeholder="输入消息..."
                onKeyDown={(e) => { if (e.key === 'Enter' && !e.shiftKey) { e.preventDefault(); sendMessage(); } }}
              />
              <input
                type="file"
                onChange={(e) => setFile(e.target.files[0])}
                style={{ marginLeft: 8 }}
              />
              <button onClick={sendMessage}>发送</button>
              <button onClick={sendFile} disabled={!file}>发送文件</button>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};

export default ChatApp;