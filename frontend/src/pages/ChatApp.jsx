import { useState, useEffect, useRef } from 'react';
import '../styles/ChatApp.css';
import { useNavigate, useLocation, useParams } from "react-router-dom";
import Sidebar from "../components/Sidebar";

const ChatApp = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const params = useParams();

  const messagesEndRef = useRef(null);

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

  const token = getToken();

  /*
  =========================
  获取当前用户
  =========================
  */
  const getCurrentUser = () => {
    try {
      const userStr = sessionStorage.getItem("user");
      if (!userStr) return null;
      return JSON.parse(userStr);
    } catch {
      return null;
    }
  };

  const currentUser = getCurrentUser();
  const currentUserId = String(currentUser?.id || "");

  /*
  =========================
  状态 (严格保持原样)
  =========================
  */
  const [contacts, setContacts] = useState([]);
  const [activeContact, setActiveContact] = useState(null);
  const [messages, setMessages] = useState([]);
  const [input, setInput] = useState('');
  const [file, setFile] = useState(null);

  const [loadingContacts, setLoadingContacts] = useState(false);
  const [loadingMessages, setLoadingMessages] = useState(false);

  const [avatarMap, setAvatarMap] = useState({});
  const [conversationAvatarMap, setConversationAvatarMap] = useState({});
  const [conversationNameMap, setConversationNameMap] = useState({});

  const friendName = location.state?.friendName;
  const conversationIdFromState = location.state?.conversationId;

  /*
  =========================
  群聊功能状态 (严格保持原样)
  =========================
  */
  const [showGroupModal, setShowGroupModal] = useState(false);
  const [newGroupName, setNewGroupName] = useState('');
  const [friendRequests, setFriendRequests] = useState([]);
  const [selectedUserIds, setSelectedUserIds] = useState([]);

  /*
  =========================
  自动滚动
  =========================
  */
  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({
      behavior: "smooth"
    });
  };

  /*
  =========================
  获取头像 (仅修复判断逻辑，不动功能)
  =========================
  */
  const fetchAvatar = async (userId) => {
    if (!userId || userId === "undefined") return;
    const sId = String(userId);
    
    // 如果已经有 URL 了才跳过，否则继续请求
    if (avatarMap[sId] && avatarMap[sId].startsWith("blob:")) return;

    try {
      const res = await fetch(
        `http://localhost:8000/api/auth/head/${sId}`,
        {
          headers: {
            Authorization: `Bearer ${token}`
          }
        }
      );

      if (!res.ok) return;

      const blob = await res.blob();
      const url = URL.createObjectURL(blob);

      setAvatarMap(prev => ({
        ...prev,
        [sId]: url
      }));

    } catch (err) {
      console.error("头像加载异常", sId);
    }
  };

  /*
  =========================
  获取会话好友信息
  =========================
  */
  const fetchConversationFriendInfo = async (conversationId) => {
    const safeId = (typeof conversationId === 'object') ? conversationId.conversation_id : conversationId;
    if (!safeId) return;

    try {
      const res = await fetch(
        `http://localhost:8000/api/auth/conversations/${safeId}`,
        {
          headers: {
            Authorization: `Bearer ${token}`
          }
        }
      );

      const data = await res.json();

      if (data.code === 200 && Array.isArray(data.data)) {
        const otherMsg = data.data.find(
          msg => String(msg.sender_id) !== currentUserId
        );

        if (otherMsg) {
          const uId = String(otherMsg.sender_id);
          const uName = otherMsg.sender_name || otherMsg.name || otherMsg.username || friendName || "未知好友";

          setConversationAvatarMap(prev => ({ ...prev, [safeId]: uId }));
          setConversationNameMap(prev => ({ ...prev, [safeId]: uName }));
          
          // 获取 ID 后立即拉取头像
          fetchAvatar(uId);
        } else if (friendName) {
          setConversationNameMap(prev => ({ ...prev, [safeId]: friendName }));
        }
      }
    } catch (err) {
      console.error("获取会话好友失败", err);
    }
  };

  /*
  =========================
  获取文件blob
  =========================
  */
  const getFileBlobUrl = async (messageId) => {
    try {
      const res = await fetch(
        `http://localhost:8000/api/auth/files/${messageId}`,
        {
          headers: {
            Authorization: `Bearer ${token}`
          }
        }
      );
      if (!res.ok) throw new Error("下载失败");
      const blob = await res.blob();
      return URL.createObjectURL(blob);
    } catch (err) {
      return null;
    }
  };

  /*
  =========================
  获取联系人
  =========================
  */
  const fetchContacts = async () => {
    try {
      setLoadingContacts(true);
      const res = await fetch(
        'http://localhost:8000/api/auth/conversations',
        {
          headers: {
            Authorization: `Bearer ${token}`
          }
        }
      );

      const data = await res.json();

      if (data.code === 200) {
        setContacts(data.data);
        data.data.forEach(contact => {
          fetchConversationFriendInfo(contact.conversation_id);
        });
      }
    } catch (err) {
      console.error(err);
    } finally {
      setLoadingContacts(false);
    }
  };

  /*
  =========================
  获取消息
  =========================
  */
  const fetchMessages = async (conversation_id) => {
    const safeId = (typeof conversation_id === 'object') ? conversation_id.conversation_id : conversation_id;
    if (!safeId) return;

    try {
      setLoadingMessages(true);
      const res = await fetch(
        `http://localhost:8000/api/auth/conversations/${safeId}`,
        {
          headers: {
            Authorization: `Bearer ${token}`
          }
        }
      );

      const data = await res.json();

      if (data.code === 200 && Array.isArray(data.data)) {
        const parsedMessages = await Promise.all(
          data.data.map(async (msg) => {
            let content = {};
            try {
              content = typeof msg.content === "string" ? JSON.parse(msg.content) : msg.content;
            } catch {}

            if (content.file_name) {
              content.file_url = await getFileBlobUrl(msg.message_id);
            }
            fetchAvatar(String(msg.sender_id));
            return {
              ...msg,
              sender_id: String(msg.sender_id),
              content
            };
          })
        );
        setMessages(parsedMessages.reverse());
      }
    } catch (err) {
      console.error(err);
    } finally {
      setLoadingMessages(false);
    }
  };

  /*
  =========================
  群聊逻辑 (严格保持原样)
  =========================
  */
  const fetchFriendsForGroup = async () => {
    try {
      const res = await fetch('http://localhost:8000/api/auth/friendship_requests', {
        headers: { Authorization: `Bearer ${token}` }
      });
      const data = await res.json();
      if (Array.isArray(data)) {
        setFriendRequests(data);
      } else if (data.data) {
        setFriendRequests(data.data);
      }
    } catch (err) {
      console.error(err);
    }
  };

  const handleToggleFriend = (senderId) => {
    const id = Number(senderId);
    setSelectedUserIds(prev => 
      prev.includes(id) ? prev.filter(i => i !== id) : [...prev, id]
    );
  };

  const handleCreateGroupSubmit = async () => {
    if (!newGroupName.trim() || selectedUserIds.length === 0) return alert("请填写群名并选择成员");
    try {
      const res = await fetch('http://localhost:8000/api/auth/conversations/group', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
        body: JSON.stringify({
          group_name: newGroupName,
          member_ids: selectedUserIds
        })
      });
      const data = await res.json();
      if (data.code === 201 || data.code === 200) {
        setShowGroupModal(false);
        setNewGroupName('');
        setSelectedUserIds([]);
        fetchContacts();
        alert("发起群聊成功");
      }
    } catch (err) {
      console.error(err);
    }
  };

  /*
  =========================
  初始化
  =========================
  */
  useEffect(() => {
    fetchContacts();
    if (currentUserId) {
      fetchAvatar(currentUserId);
    }
  }, []);

  useEffect(() => {
    let id = params.conversationId || conversationIdFromState;
    if (!id) return;

    const contact = contacts.find(c => String(c.conversation_id) === String(id));
    setActiveContact(contact || { conversation_id: id, name: friendName || "聊天中" });
    
    fetchConversationFriendInfo(id);
    fetchMessages(id);

  }, [params.conversationId, conversationIdFromState, contacts.length]); 

  useEffect(() => {
    scrollToBottom();
  }, [messages]);

  /*
  =========================
  交互逻辑 (严格保持原样)
  =========================
  */
  const handleSelectContact = (contact) => {
    setActiveContact(contact);
    fetchMessages(contact.conversation_id);
    navigate(`/chat/${contact.conversation_id}`);
  };

  const sendMessage = async () => {
    if (!input.trim() || !activeContact) return;
    const res = await fetch(
      'http://localhost:8000/api/auth/messages/text',
      {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${token}`
        },
        body: JSON.stringify({
          conversation_id: activeContact.conversation_id,
          content: input
        })
      }
    );
    const data = await res.json();
    if (data.code === 201) {
      setMessages(prev => [
        ...prev,
        {
          message_id: data.data,
          sender_id: currentUserId,
          content: { text: input }
        }
      ]);
      setInput('');
    }
  };

  const sendFile = async () => {
    if (!file || !activeContact) return;
    const formData = new FormData();
    formData.append("conversation_id", activeContact.conversation_id);
    formData.append("file", file);
    const res = await fetch(
      'http://localhost:8000/api/auth/messages/file',
      {
        method: "POST",
        headers: { Authorization: `Bearer ${token}` },
        body: formData
      }
    );
    const data = await res.json();
    if (data.code === 201) {
      setMessages(prev => [
        ...prev,
        {
          message_id: data.data,
          sender_id: currentUserId,
          content: {
            file_name: file.name,
            file_type: file.type,
            file_url: URL.createObjectURL(file)
          }
        }
      ]);
      setFile(null);
    }
  };

  const getAvatar = (uid) => {
    const sId = String(uid);
    return avatarMap[sId] || "https://img.icons8.com/color/96/user.png";
  };

  return (
    <div className="chat-app">
      <Sidebar />
      <div className="chat-middle">
        <div className="search" style={{ display: 'flex', alignItems: 'center' }}>
          <input placeholder="🔍搜索联系人" style={{ flex: 1 }} />
          <button onClick={() => { fetchFriendsForGroup(); setShowGroupModal(true); }}
            style={{ marginLeft: '10px', background: 'none', border: 'none', cursor: 'pointer', fontSize: '20px' }}>
            ➕
          </button>
        </div>

        <div className="contacts">
          {contacts.map((c) => {
            const friendUid = conversationAvatarMap[c.conversation_id];
            return (
              <div key={c.conversation_id} className={`contact ${activeContact?.conversation_id === c.conversation_id ? 'active' : ''}`}
                onClick={() => handleSelectContact(c)}>
                <img className="contact-avatar" 
                  src={c.group_name ? "https://img.icons8.com/color/96/group.png" : getAvatar(friendUid)}
                  onError={(e) => e.currentTarget.src = "https://img.icons8.com/color/96/user.png"} />
                <div className="contact-info">
                  <div className="contact-name">
                    {c.group_name || conversationNameMap[c.conversation_id] || c.name || "未知会话"}
                  </div>
                  <div className="contact-message">{c.last_message || "暂无消息"}</div>
                </div>
              </div>
            );
          })}
        </div>
      </div>

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
              {activeContact.group_name || conversationNameMap[activeContact.conversation_id] || activeContact.name}
            </div>
            <div className="chat-body">
              {messages.map((msg) => (
                <div key={msg.message_id} className={`msg ${msg.sender_id === currentUserId ? "self" : "other"}`}>
                  {msg.sender_id !== currentUserId && (
                    <img className="avatar other" src={getAvatar(msg.sender_id)} />
                  )}
                  <div className="bubble">
                    {activeContact.group_name && msg.sender_id !== currentUserId && (
                      <div style={{ fontSize: '11px', color: '#999', marginBottom: '2px' }}>{msg.sender_name}</div>
                    )}
                    {msg.content.text && <span>{msg.content.text}</span>}
                    {msg.content.file_url && msg.content.file_type?.startsWith("image/") && (
                        <img className="chat-image" src={msg.content.file_url} />
                    )}
                    {msg.content.file_url && !msg.content.file_type?.startsWith("image/") && (
                        <a href={msg.content.file_url} download>📎 {msg.content.file_name}</a>
                    )}
                  </div>
                  {msg.sender_id === currentUserId && (
                    <img className="avatar self" src={getAvatar(currentUserId)} />
                  )}
                </div>
              ))}
              <div ref={messagesEndRef} />
            </div>
            <div className="chat-footer">
              <input value={input} onChange={(e) => setInput(e.target.value)}
                onKeyDown={(e) => e.key === "Enter" && sendMessage()} placeholder="输入消息..." />
              <input type="file" onChange={(e) => setFile(e.target.files[0])} />
              <button onClick={sendMessage}>发送</button>
              <button onClick={sendFile}>文件</button>
            </div>
          </div>
        )}
      </div>

      {showGroupModal && (
        <div style={{ position: 'fixed', top: 0, left: 0, width: '100vw', height: '100vh', backgroundColor: 'rgba(0,0,0,0.5)', display: 'flex', justifyContent: 'center', alignItems: 'center', zIndex: 2000 }}>
          <div style={{ backgroundColor: 'white', padding: '20px', borderRadius: '8px', width: '320px', maxHeight: '80vh', display: 'flex', flexDirection: 'column' }}>
            <h3 style={{ marginTop: 0 }}>发起群聊</h3>
            <input placeholder="输入群聊名称" value={newGroupName} onChange={(e) => setNewGroupName(e.target.value)}
              style={{ width: '100%', padding: '8px', marginBottom: '15px', boxSizing: 'border-box' }} />
            <div style={{ flex: 1, overflowY: 'auto', border: '1px solid #eee', padding: '10px' }}>
              {friendRequests.map((req) => (
                <label key={req.request_id} style={{ display: 'flex', alignItems: 'center', padding: '8px 0', cursor: 'pointer' }}>
                  <input type="checkbox" checked={selectedUserIds.includes(Number(req.sender_id))}
                    onChange={() => handleToggleFriend(req.sender_id)} />
                  <span style={{ marginLeft: '10px' }}>{req.sender_name}</span>
                </label>
              ))}
            </div>
            <div style={{ marginTop: '20px', display: 'flex', justifyContent: 'flex-end', gap: '10px' }}>
              <button onClick={() => setShowGroupModal(false)}>取消</button>
              <button onClick={handleCreateGroupSubmit} style={{ backgroundColor: '#007bff', color: 'white', border: 'none', padding: '6px 15px', borderRadius: '4px' }}>创建</button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default ChatApp;