import { useState, useEffect, useRef } from 'react';
import '../styles/ChatApp.css';
import { useNavigate, useLocation, useParams } from "react-router-dom";
import Sidebar from "../components/Sidebar";

const ChatApp = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const params = useParams();

  const messagesEndRef = useRef(null);
  const [previewFile, setPreviewFile] = useState(null);

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
  const currentUserId = String(currentUser?.id);

  /*
  =========================
  状态
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
  获取头像
  =========================
  */
  const fetchAvatar = async (userId) => {
    try {
      const res = await fetch(
        `http://localhost:8000/api/auth/info/head/${userId}`,
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
        [String(userId)]: url
      }));

    } catch (err) {
      console.error("头像加载失败", err);
    }
  };

  /*
  =========================
  获取会话好友信息
  =========================
  */
  const fetchConversationFriendInfo = async (conversationId) => {
    try {
      const res = await fetch(
        `http://localhost:8000/api/auth/conversations/${conversationId}`,
        {
          headers: {
            Authorization: `Bearer ${token}`
          }
        }
      );

      const data = await res.json();

      if (data.code === 200 && data.data.length > 0) {

        const otherMsg = data.data.find(
          msg => String(msg.sender_id) !== currentUserId
        );

        if (otherMsg) {

          fetchAvatar(otherMsg.sender_id);

          setConversationAvatarMap(prev => ({
            ...prev,
            [conversationId]: String(otherMsg.sender_id)
          }));

          setConversationNameMap(prev => ({
            ...prev,
            [conversationId]:
              otherMsg.sender_name ||
              otherMsg.name ||
              "未知好友"
          }));
        }
      }

    } catch (err) {
      console.error(err);
    }
  };

  /*
  =========================
  获取文件blob
  =========================
  */
  const getFileBlobUrl = async (messageId, fileType, fileName) => {
    try {
      const res = await fetch(
        `http://localhost:8000/api/auth/files/${messageId}`,
        {
          headers: {
            Authorization: `Bearer ${token}`
          }
        }
      );

      const blob = await res.blob();

      // ✅ 1. 优先用后端给的 fileType
      let type = fileType;

      // ✅ 2. 如果没有 type，就根据文件名推断
      if (!type && fileName) {
        const ext = fileName.split(".").pop().toLowerCase();

        const typeMap = {
          pdf: "application/pdf",
          txt: "text/plain",
          json: "application/json",
          csv: "text/csv",
          html: "text/html",
          mp4: "video/mp4",
          webm: "video/webm",
          mp3: "audio/mpeg",
          png: "image/png",
          jpg: "image/jpeg",
          jpeg: "image/jpeg",
          gif: "image/gif"
        };

        type = typeMap[ext] || "application/octet-stream";
      }

      // ✅ 3. 强制修复 blob 类型
      const fixedBlob = new Blob([blob], {
        type: type || blob.type
      });

      return URL.createObjectURL(fixedBlob);

    } catch (err) {
      console.error(err);
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
    if (!conversation_id) return;

    try {
      setLoadingMessages(true);

      const res = await fetch(
        `http://localhost:8000/api/auth/conversations/${conversation_id}`,
        {
          headers: {
            Authorization: `Bearer ${token}`
          }
        }
      );

      const data = await res.json();

      if (data.code === 200) {

        const parsedMessages = await Promise.all(
          data.data.map(async (msg) => {

            let content = {};

            try {
              content =
                typeof msg.content === "string"
                  ? JSON.parse(msg.content)
                  : msg.content;
            } catch {}

            if (content.file_name) {
              // ✅ 先修复 file_type
              if (!content.file_type && content.file_name) {
                const ext = content.file_name.split(".").pop().toLowerCase();
                const typeMap = {
                  pdf: "application/pdf",
                  txt: "text/plain",
                  json: "application/json",
                  csv: "text/csv",
                  html: "text/html",
                  mp4: "video/mp4",
                  webm: "video/webm",
                  mp3: "audio/mpeg",
                  png: "image/png",
                  jpg: "image/jpeg",
                  jpeg: "image/jpeg",
                  gif: "image/gif"
                };

                content.file_type = typeMap[ext] || "application/octet-stream";
              }

              // ✅ 再获取 URL
              content.file_url = await getFileBlobUrl(
                msg.message_id,
                content.file_type,
                content.file_name
              );
            }
            fetchAvatar(msg.sender_id);
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
    let id =
      params.conversationId ||
      conversationIdFromState;

    if (!id || contacts.length === 0) return;

    const contact =
      contacts.find(
        c => String(c.conversation_id) === String(id)
      ) || {
        conversation_id: id,
        name: friendName || "聊天对象"
      };

    setActiveContact(contact);

    fetchMessages(id);

  }, [
    params.conversationId,
    conversationIdFromState,
    contacts
  ]);

  useEffect(() => {
    scrollToBottom();
  }, [messages]);

  /*
  =========================
  点击联系人
  =========================
  */
  const handleSelectContact = (contact) => {
    setActiveContact(contact);

    fetchMessages(contact.conversation_id);

    navigate(`/chat/${contact.conversation_id}`);
  };

  /*
  =========================
  发文本
  =========================
  */
  const sendMessage = async () => {
    if (!input.trim()) return;

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
    const handleSend = () => {
      if (file) {
        sendFile();
      } else {
        sendMessage();
      }
    };

  /*
  =========================
  发文件
  =========================
  */
  const sendFile = async () => {
    if (!file) return;

    const formData = new FormData();

    formData.append("conversation_id", activeContact.conversation_id);
    formData.append("file", file);

    const res = await fetch(
      'http://localhost:8000/api/auth/messages/file',
      {
        method: "POST",
        headers: {
          Authorization: `Bearer ${token}`
        },
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

  return (
    <div className="chat-app">

      {/* 左导航栏（保持原样） */}

        <Sidebar />

      {/* 联系人 */}
      <div className="chat-middle">
        <div className="search">
          <input placeholder="🔍搜索联系人" />
        </div>

        <div className="contacts">
          {contacts.map((c) => (
            <div
              key={c.conversation_id}
              className="contact"
              onClick={() => handleSelectContact(c)}
            >
              <img
                className="contact-avatar"
                src={
                  avatarMap[
                    conversationAvatarMap[c.conversation_id]
                  ] || "/default-avatar.png"
                }
                onError={(e) =>
                  e.currentTarget.src = "/default-avatar.png"
                }
              />

              <div className="contact-info">
                <div className="contact-name">
                  {
                    conversationNameMap[c.conversation_id]
                    || c.name
                    || "未知好友"
                  }
                </div>

                <div className="contact-message">
                  {c.last_message}
                </div>
              </div>
            </div>
          ))}
        </div>
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
              {conversationNameMap[activeContact.conversation_id] || activeContact.name}
            </div>

            <div className="chat-body">

              {messages.map((msg) => (
                <div
                  key={msg.message_id}
                  className={`msg ${
                    msg.sender_id === currentUserId
                      ? "self"
                      : "other"
                  }`}
                >
                  {msg.sender_id !== currentUserId && (
                    <img
                      className="avatar other"
                      src={
                        avatarMap[msg.sender_id] ||
                        "/default-avatar.png"
                      }
                    />
                  )}
                  {/* 文本 */}
                  {msg.content.text && (
                    <div className="bubble">
                      {msg.content.text}
                    </div>
                  )}

                  {/* 图片 */}
                  {msg.content.file_url &&
                    msg.content.file_type?.startsWith("image/") && (
                      <img
                        className="chat-image"
                        src={msg.content.file_url}
                      />
                  )}

                  {/* 文件 */}
                  {msg.content.file_url &&
                    !msg.content.file_type?.startsWith("image/") && (
                      <div
                        className="file-card"
                        onClick={() => {
                          const file = msg.content;
                          let type = file.file_type;
                          if (!type && file.file_name) {
                            const ext = file.file_name.split(".").pop().toLowerCase();
                            const typeMap = {
                              pdf: "application/pdf",
                              txt: "text/plain",
                              json: "application/json",
                              csv: "text/csv",
                              html: "text/html",
                              mp4: "video/mp4",
                              webm: "video/webm",
                              mp3: "audio/mpeg",
                              png: "image/png",
                              jpg: "image/jpeg",
                              jpeg: "image/jpeg",
                              gif: "image/gif"
                            };

                            type = typeMap[ext];
                          }

                          setPreviewFile({
                            ...file,
                            file_type: type
                          });
                        }}
                      >
                        <div className="file-icon">📄</div>
                        <div className="file-info">
                          <div className="file-name">
                            {msg.content.file_name}
                          </div>

                          <div className="file-type">
                            {msg.content.file_type || "文件"}
                          </div>
                        </div>
                      </div>
                  )}
                    {msg.sender_id === currentUserId && (
                    <img
                      className="avatar self"
                      src={
                        avatarMap[currentUserId] ||
                        "/default-avatar.png"
                      }
                    />
                  )}
                </div>
              ))}

              <div ref={messagesEndRef} />

            </div>

            <div className="chat-footer">
              {/* 输入框 */}
              <input
                type="text"
                placeholder="输入消息..."
                value={input}
                onChange={(e) => setInput(e.target.value)}
                onKeyDown={(e) =>
                  e.key === "Enter" && handleSend()
                }
              />
              {/* 自定义文件按钮 */}
              <label className="file-btn">
                📎
                <input
                  type="file"
                  onChange={(e) => setFile(e.target.files[0])}
                />
              </label>
              {file && (
                <div className="selected-file">
                  {file.name}
                </div>
              )}
              {/* 发送按钮 */}
              <button onClick={handleSend}>
                发送
              </button>

            </div>

          </div>
        )}
      </div>
      {previewFile && (
      <div className="preview-modal">
        <div className="preview-content">

          <div className="preview-header">
            <span>{previewFile.file_name}</span>
            <button onClick={() => setPreviewFile(null)}>关闭</button>
          </div>
          <div className="preview-body">
            {/* 图片 */}
            {previewFile.file_type?.startsWith("image/") && (
              <img src={previewFile.file_url} className="preview-img" />
            )}
            {/* PDF */}
            {previewFile.file_type?.includes("pdf") && (
              <iframe
                src={previewFile.file_url}
                className="preview-frame"
                title="pdf"
                style={{ width: "100%", height: "100%" }}
              />
            )}
            {/* 视频 */}
            {previewFile.file_type?.startsWith("video/") && (
              <video
                src={previewFile.file_url}
                controls
                className="preview-video"
              />
            )}
            {/* 文本 */}
            {previewFile.file_type?.includes("text") && (
              <iframe
                src={previewFile.file_url}
                className="preview-frame"
                title="text"
                style={{ width: "100%", height: "100%" }}
              />
            )}
            {/* 兜底 */}
            {previewFile.file_type &&
              !previewFile.file_type.startsWith("image/") &&
              !previewFile.file_type.includes("pdf") &&
              !previewFile.file_type.startsWith("video/") &&
              !previewFile.file_type.includes("text") && (
                <div>暂不支持预览</div>
            )}
          </div>
        </div>
      </div>
    )}
        </div>
  );
};

export default ChatApp;