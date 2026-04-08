import { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import "../styles/AddFriendPage.css";
import axios from "axios";

export default function AddFriendPage() {
  const navigate = useNavigate();

  // 当前 tab
  const [activeTab, setActiveTab] = useState("add");

  // ===== 好友申请模块 =====
  const [friendRequests, setFriendRequests] = useState([]);
  const [loadingRequests, setLoadingRequests] = useState(false);
  const [errorRequests, setErrorRequests] = useState("");

  // ===== 搜索模块 =====
  const [keyword, setKeyword] = useState("");          // 输入 ID/UID
  const [searchType, setSearchType] = useState("id");  // 搜索类型: "id" | "uid"
  const [stranger, setStranger] = useState(null);      // 搜索结果
  const [searching, setSearching] = useState(false);   // 搜索 loading
  const [friends, setFriends] = useState([]);
  const [loadingFriends, setLoadingFriends] = useState(false);

  // ===== 获取 token 的安全方法 =====
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

  // ===== 获取好友申请列表 =====
  const fetchFriendRequests = async () => {
    setLoadingRequests(true);
    setErrorRequests("");

    try {
      const token = getToken();
      if (!token) throw new Error("缺少登录 token");

      const res = await fetch("http://localhost:8000/api/auth/friendship_requests", {
        headers: { Authorization: `Bearer ${token}` },
      });
      const result = await res.json();

      if (result.code === 200) setFriendRequests(result.data);
      else setErrorRequests(result.message || "加载失败");
    } catch (err) {
      console.error(err);
      setErrorRequests("网络错误或 token 无效");
    } finally {
      setLoadingRequests(false);
    }
  };

  // ===== 同意 / 拒绝好友申请 =====
  const handleAccept = async (requestId) => {
    try {
      const token = getToken();
      if (!token) throw new Error("缺少登录 token");

      const res = await fetch(
        `http://localhost:8000/api/auth/friendship_requests/${requestId}`,
        { method: "POST", headers: { Authorization: `Bearer ${token}` } }
      );
      const result = await res.json();

      if (result.code === 200) fetchFriendRequests();
      else alert(result.message || "同意失败");
    } catch (err) {
      alert("网络错误或 token 无效");
    }
  };

  const handleReject = async (requestId) => {
    try {
      const token = getToken();
      if (!token) throw new Error("缺少登录 token");

      const res = await fetch(
        `http://localhost:8000/api/auth/friendship_requests/${requestId}`,
        { method: "DELETE", headers: { Authorization: `Bearer ${token}` } }
      );
      const result = await res.json();

      if (result.code === 200) fetchFriendRequests();
      else alert(result.message || "拒绝失败");
    } catch (err) {
      alert("网络错误或 token 无效");
    }
  };

  // ===== 搜索陌生人（支持 手机号 / UID） =====
  const handleSearchUser = async () => {
    if (!keyword.trim()) {
      alert("请输入 手机号 或 UID");
      return;
    }

    setSearching(true);
    setStranger(null);

    try {
      const token = getToken();
      if (!token) throw new Error("缺少登录 token");

      const url =
        searchType === "id"
          ? `http://localhost:8000/api/auth/info/strangers/id/${keyword}`
          : `http://localhost:8000/api/auth/info/strangers/uid/${keyword}`;

      const res = await fetch(url, {
        headers: { Authorization: `Bearer ${token}` },
      });
      const result = await res.json();

      if (result.code === 200) setStranger(result.data);
      else alert(result.message || "未找到用户");
    } catch (err) {
      console.error(err);
      alert("网络错误或 token 无效");
    } finally {
      setSearching(false);
    }
  };

  // ===== 发送好友申请 =====
  const handleSendFriendRequest = async (receiverId, receiverName) => {
  // 👉 弹出输入框
    const message = prompt("请输入好友申请备注：", "你好，我们加个好友吧");

    // 👉 用户点取消
    if (message === null) return;

    const user = JSON.parse(sessionStorage.getItem("user") || "{}");
    const senderName = user.name || "我的昵称";

    const body = {
      receiver_id: receiverId,
      sender_name: senderName,
      verification_message: message, // ✅ 用用户输入
    };

    console.log("发送请求体:", body);

    try {
      const token = getToken();
      if (!token) throw new Error("缺少登录 token");

      const res = await fetch("http://localhost:8000/api/auth/friendship_requests", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify(body),
      });

      const result = await res.json();
      console.log("后端返回:", result);

      if (result.code === 201) {
        alert("好友申请发送成功");
        if (activeTab === "request") fetchFriendRequests();
      } else {
        alert(`发送失败: ${result.message || "未知原因"}`);
      }
    } catch (err) {
      console.error(err);
      alert("网络错误或 token 无效");
    }
  };
  //点击聊天按钮自动切到聊天页面并加载与该好友的聊天记录//
  const handleChat = async (friendId, friendName) => {
    try {
      const token = getToken();
      if (!token) throw new Error("缺少 token");

      console.log("准备发送聊天请求");
      console.log("friendId:", friendId, "friendName:", friendName);
      console.log("token:", token);

      // 确保 id 是字符串
      const body = { id: String(friendId) };
      console.log("请求体 body:", body);

      const res = await fetch("http://localhost:8000/api/auth/conversations/private", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify(body),
      });

      console.log("fetch 完成，状态码:", res.status);

      const text = await res.text();
      console.log("返回原始文本:", text);

      let result;
      try {
        result = JSON.parse(text);
        console.log("解析后的 JSON:", result);
      } catch (err) {
        console.error("JSON 解析失败", err);
        alert("返回的 JSON 解析失败，请检查后端返回值");
        return;
      }

      if (result.code === 201 || result.code === 0) {
        const conversationId = result.data;
        console.log("会话创建成功，conversationId:", conversationId);

        // 跳转到 Chat 页面并传递 conversationId + friendName
        navigate("/chat", {
          state: { conversationId, friendId, friendName },
        });
      } else {
        console.error("创建会话失败，后端返回:", result);
        alert(`创建会话失败: ${result.message || "未知错误"}`);
      }
    } catch (err) {
      console.error("handleChat 出错:", err);
      alert("网络错误或 token 无效");
    }
};

  // ===== 切到「好友申请」自动加载 =====
  useEffect(() => {
  if (activeTab === "request") fetchFriendRequests();
  if (activeTab === "list") fetchFriends();
}, [activeTab]);
  //获取好友函数//
  const fetchFriends = async () => {
    setLoadingFriends(true);

    try {
      const token = getToken();
      if (!token) throw new Error("缺少 token");

      const res = await fetch("http://localhost:8000/api/auth/friendships", {
        headers: {
          Authorization: `Bearer ${token}`,
        },
      });

      const result = await res.json();

      if (result.code === 200) {
        setFriends(result.data);
      } else {
        alert(result.message || "获取好友失败");
      }
    } catch (err) {
      console.error(err);
      alert("网络错误或 token 无效");
    } finally {
      setLoadingFriends(false);
    }
  };

  console.log("当前tab:", activeTab);
  console.log("好友列表:", friends);
  return (
    <div className="chatpage-app">
      {/* 左侧导航栏 */}
      <div className="chat-sidebar">
        <div className="nav-buttons">
          <button onClick={() => navigate("/chat")}>💬</button>
          <button onClick={() => navigate("/addfriend")}>👥</button>
          <button onClick={() => navigate("/chatpage")}>📝</button>
          <button onClick={() => navigate("/persional")}>⚙️</button>
        </div>
      </div>

      {/* 中间菜单栏 */}
      <aside className="addfriend-middle">
        <h3>好友管理</h3>
        <div className="menu">
          <div
            className={`menu-item ${activeTab === "add" ? "active" : ""}`}
            onClick={() => setActiveTab("add")}
          >
            添加好友
          </div>
          <div
            className={`menu-item ${activeTab === "request" ? "active" : ""}`}
            onClick={() => setActiveTab("request")}
          >
            好友申请
          </div>
          <div
            className={`menu-item ${activeTab === "list" ? "active" : ""}`}
            onClick={() => setActiveTab("list")}
          >
            我的好友
          </div>
        </div>
      </aside>

      {/* 右侧主内容 */}
      <main className="addfriend-main">
        <header className="header">
          <div className="header-left">
            {activeTab === "add" && "添加好友"}
            {activeTab === "request" && "好友申请"}
            {activeTab === "list" && "我的好友"}
          </div>
        </header>

        <section className="addfriend-content">
          {/* ===== 添加好友 ===== */}
          {activeTab === "add" && (
            <>
              <div className="card">
                <h4>搜索用户</h4>

                <div className="search-box">
                  <select
                    value={searchType}
                    onChange={(e) => setSearchType(e.target.value)}
                  >
                    <option value="id">通过 手机号</option>
                    <option value="uid">通过 UID</option>
                  </select>

                  <input
                    placeholder="请输入用户 手机号 或 UID"
                    value={keyword}
                    onChange={(e) => setKeyword(e.target.value)}
                  />

                  <button
                    className="primary-btn"
                    onClick={handleSearchUser}
                    disabled={searching}
                  >
                    {searching ? "搜索中..." : "搜索"}
                  </button>
                </div>
              </div>

              <div className="card">
                <h4>搜索结果</h4>
                {!stranger && <div className="tip">暂无搜索结果</div>}
                {stranger && (
                  <div className="user-item">
                    <div className="avatar" />
                    <div className="user-info">
                      <div className="name">{stranger.name}</div>
                      <div className="desc">ID: {stranger.id}</div>
                    </div>
                    <button
                      className="outline-btn"
                      onClick={() => handleSendFriendRequest(stranger.id, stranger.name)}
                    >
                      添加好友
                    </button>
                  </div>
                )}
              </div>
            </>
          )}

          {/* ===== 好友申请 ===== */}
          {activeTab === "request" && (
            <div className="card">
              <h4>好友申请</h4>

              {loadingRequests && <div className="tip">加载中...</div>}
              {errorRequests && <div className="error">{errorRequests}</div>}
              {!loadingRequests && friendRequests.length === 0 && (
                <div className="tip">暂无好友申请</div>
              )}

              {friendRequests.map((item) => (
                <div className="user-item" key={item.request_id}>
                  <div className="avatar" />
                  <div className="user-info">
                    <div className="name">
                      {item.sender_name || item.sender_nickname || item.sender_id || "未知用户"}
                    </div>
                    <div className="desc">
                      {item.verification_message || "请求添加你为好友"}
                    </div>
                  </div>
                  <div className="actions">
                    <button
                      className="primary-btn small"
                      onClick={() => handleAccept(item.request_id)}
                    >
                      同意
                    </button>
                    <button
                      className="ghost-btn small"
                      onClick={() => handleReject(item.request_id)}
                    >
                      拒绝
                    </button>
                  </div>
                </div>
              ))}
            </div>
          )}

          {/* ===== 我的好友 ===== */}
          {activeTab === "list" && (
            <div className="card">
              <h4>好友列表</h4>

              {loadingFriends && <div className="tip">加载中...</div>}

              {!loadingFriends && friends.length === 0 && (
                <div className="tip">暂无好友</div>
              )}

              {friends.map((item) => (
                <div className="user-item" key={item.friendship_id}>
                  <div className="avatar" />

                  <div className="user-info">
                    <div className="name">
                      {item.friend_remark || item.friend_id}
                    </div>
                    <div className="desc offline">离线</div>
                  </div>

                  <button
                    className="ghost-btn"
                    onClick={() => handleChat(item.friend_id, item.friend_remark)}//待确定字样
                  >
                    聊天
                  </button>
                </div>
              ))}
            </div>
          )}
        </section>
      </main>
    </div>
  );
}