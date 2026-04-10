import { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import "../styles/AddFriendPage.css";
import Sidebar from "../components/Sidebar";

export default function AddFriendPage() {
  const navigate = useNavigate();

  const [activeTab, setActiveTab] = useState("add");

  const [friendRequests, setFriendRequests] = useState([]);
  const [loadingRequests, setLoadingRequests] = useState(false);
  const [errorRequests, setErrorRequests] = useState("");

  const [keyword, setKeyword] = useState("");
  const [searchType, setSearchType] = useState("id");
  const [stranger, setStranger] = useState(null);
  const [searching, setSearching] = useState(false);

  const [friends, setFriends] = useState([]);
  const [loadingFriends, setLoadingFriends] = useState(false);

  const [avatarMap, setAvatarMap] = useState({});

  // 控制是否显示搜索结果
  const [hasSearched, setHasSearched] = useState(false);

  // 记录已处理申请状态
  const [handledRequests, setHandledRequests] = useState({});

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

  /*
  =========================
  获取头像
  =========================
  */
  const fetchAvatar = async (userId) => {
    try {
      const token = getToken();

      const res = await fetch(
        `http://localhost:8000/api/auth/info/head/${userId}`,
        {
          headers: {
            Authorization: `Bearer ${token}`,
          },
        }
      );

      if (!res.ok) throw new Error("头像获取失败");

      const blob = await res.blob();
      const imageUrl = URL.createObjectURL(blob);

      setAvatarMap((prev) => ({
        ...prev,
        [String(userId)]: imageUrl,
      }));
    } catch (err) {
      console.error("头像加载失败:", userId);
    }
  };

  /*
  =========================
  获取好友申请
  =========================
  */
  const fetchFriendRequests = async () => {
    setLoadingRequests(true);
    setErrorRequests("");

    try {
      const token = getToken();

      const res = await fetch(
        "http://localhost:8000/api/auth/friendship_requests",
        {
          headers: {
            Authorization: `Bearer ${token}`,
          },
        }
      );

      const result = await res.json();

      if (result.code === 200) {
        setFriendRequests(result.data);

        // 初始化 handledRequests 状态
        const initialHandled = {};
        result.data.forEach((item) => {
          fetchAvatar(item.sender_id);

          // 假设接口返回 status 字段：accepted / rejected / pending
          if (item.status === "accepted" || item.status === "rejected") {
            initialHandled[item.request_id] = item.status;
          }
        });

        setHandledRequests(initialHandled);

      } else {
        setErrorRequests(result.message);
      }
    } catch (err) {
      setErrorRequests("加载失败");
    } finally {
      setLoadingRequests(false);
    }
  };

  /*
  =========================
  同意好友申请
  =========================
  */
  const handleAccept = async (requestId) => {
    try {
      const token = getToken();

      const res = await fetch(
        `http://localhost:8000/api/auth/friendship_requests/${requestId}`,
        {
          method: "POST",
          headers: {
            Authorization: `Bearer ${token}`,
          },
        }
      );

      const result = await res.json();

      if (result.code === 200) {
        alert("已同意好友申请");

        setHandledRequests((prev) => ({
          ...prev,
          [requestId]: "accepted",
        }));

        fetchFriends();
      }
    } catch (err) {
      alert("同意失败");
    }
  };

  /*
  =========================
  拒绝好友申请
  =========================
  */
  const handleReject = async (requestId) => {
    try {
      const token = getToken();

      const res = await fetch(
        `http://localhost:8000/api/auth/friendship_requests/${requestId}`,
        {
          method: "DELETE",
          headers: {
            Authorization: `Bearer ${token}`,
          },
        }
      );

      const result = await res.json();

      if (result.code === 200) {
        alert("已拒绝好友申请");

        setHandledRequests((prev) => ({
          ...prev,
          [requestId]: "rejected",
        }));
      }
    } catch (err) {
      alert("拒绝失败");
    }
  };

  /*
  =========================
  搜索用户
  =========================
  */
  const handleSearchUser = async () => {
    if (!keyword.trim()) return;

    setSearching(true);
    setHasSearched(true);

    try {
      const token = getToken();

      const url =
        searchType === "id"
          ? `http://localhost:8000/api/auth/info/strangers/id/${keyword}`
          : `http://localhost:8000/api/auth/info/strangers/uid/${keyword}`;

      const res = await fetch(url, {
        headers: {
          Authorization: `Bearer ${token}`,
        },
      });

      const result = await res.json();

      if (result.code === 200) {
        setStranger(result.data);

        fetchAvatar(result.data.id);
      } else {
        setStranger(null);
        alert("未找到用户");
      }
    } catch (err) {
      alert("搜索失败");
    } finally {
      setSearching(false);
    }
  };

  /*
  =========================
  发送好友申请
  =========================
  */
  const handleSendFriendRequest = async (receiverId) => {
    const message = prompt("请输入好友申请备注");

    if (message === null) return;

    const token = getToken();

    const user = JSON.parse(sessionStorage.getItem("user") || "{}");

    const body = {
      receiver_id: receiverId,
      sender_name: user.name,
      verification_message: message,
    };

    try {
      const res = await fetch(
        "http://localhost:8000/api/auth/friendship_requests",
        {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
            Authorization: `Bearer ${token}`,
          },
          body: JSON.stringify(body),
        }
      );

      const result = await res.json();

      if (result.code === 201) {
        alert("发送成功");
      }
    } catch (err) {
      alert("发送失败");
    }
  };

  /*
  =========================
  聊天
  =========================
  */
  const handleChat = async (friendId, friendName) => {
    const token = getToken();

    const body = { id: String(friendId) };

    const res = await fetch(
      "http://localhost:8000/api/auth/conversations/private",
      {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify(body),
      }
    );

    const result = await res.json();

    if (result.code === 201 || result.code === 0) {
      const conversationId = result.data;

      navigate("/chat", {
        state: { conversationId, friendId, friendName },
      });
    } else {
      alert("创建会话失败");
    }
  };

  /*
  =========================
  获取好友列表
  =========================
  */
  const fetchFriends = async () => {
    setLoadingFriends(true);

    try {
      const token = getToken();

      const res = await fetch("http://localhost:8000/api/auth/friendships", {
        headers: {
          Authorization: `Bearer ${token}`,
        },
      });

      const result = await res.json();

      if (result.code === 200) {
        setFriends(result.data);

        result.data.forEach((friend) => {
          fetchAvatar(friend.friend_id);
        });
      }
    } catch (err) {
      alert("获取好友失败");
    } finally {
      setLoadingFriends(false);
    }
  };

  /*
  =========================
  切换 tab 自动加载
  =========================
  */
  useEffect(() => {
    if (activeTab === "request") fetchFriendRequests();
    if (activeTab === "list") fetchFriends();
  }, [activeTab]);

  return (
    <div className="chatpage-app">
      <Sidebar />

     <aside className="addfriend-middle">
        <div className="addfriend-title">
          <h3>好友管理</h3>
          <p>添加、处理申请与查看好友</p>
        </div>

        <div
          className={`menu-item ${activeTab === "add" ? "active" : ""}`}
          onClick={() => setActiveTab("add")}
        >
          <span className="menu-icon">➕</span>
          <span>添加好友</span>
        </div>

        <div
          className={`menu-item ${activeTab === "request" ? "active" : ""}`}
          onClick={() => setActiveTab("request")}
        >
          <span className="menu-icon">📩</span>
          <span>好友申请</span>
        </div>

        <div
          className={`menu-item ${activeTab === "list" ? "active" : ""}`}
          onClick={() => setActiveTab("list")}
        >
          <span className="menu-icon">👥</span>
          <span>我的好友</span>
        </div>

        <div className="menu-footer">© 2026 MyChat</div>
      </aside>

      <main className="addfriend-main">
        <header className="header">
          <div className="header-left">
            {activeTab === "add" && "添加好友"}
            {activeTab === "request" && "好友申请"}
            {activeTab === "list" && "我的好友"}
          </div>
        </header>

        <section className="addfriend-content">
          {/* 添加好友 */}
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

              {hasSearched && (
                <div className="card">
                  <h4>搜索结果</h4>

                  {!stranger && <div className="tip">暂无搜索结果</div>}

                  {stranger && (
                    <div className="user-item">
                      <img
                        className="avatar"
                        src={avatarMap[String(stranger.id)] || "/default-avatar.png"}
                        onError={(e) =>
                          (e.currentTarget.src = "/default-avatar.png")
                        }
                      />

                      <div className="user-info">
                        <div className="name">{stranger.name}</div>
                        <div className="desc">ID: {stranger.id}</div>
                      </div>

                      <button
                        className="outline-btn"
                        onClick={() => handleSendFriendRequest(stranger.id)}
                      >
                        添加好友
                      </button>
                    </div>
                  )}
                </div>
              )}
            </>
          )}

          {/* 好友申请 */}
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
                  <img
                    className="avatar"
                    src={avatarMap[String(item.sender_id)] || "/default-avatar.png"}
                    onError={(e) =>
                      (e.currentTarget.src = "/default-avatar.png")
                    }
                  />

                  <div className="user-info">
                    <div className="name">
                      {item.sender_name ||
                        item.sender_nickname ||
                        item.sender_id ||
                        "未知用户"}
                    </div>

                    <div className="desc">
                      {item.verification_message || "请求添加你为好友"}
                    </div>
                  </div>

                  <div className="actions">
                    {!handledRequests[item.request_id] ? (
                      <>
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
                      </>
                    ) : (
                      <span
                        style={{
                          color:
                            handledRequests[item.request_id] === "accepted"
                              ? "#be49b4"
                              : "#ef4444",
                          fontWeight: "bold",
                        }}
                      >
                        {handledRequests[item.request_id] === "accepted"
                          ? "已同意"
                          : "已拒绝"}
                      </span>
                    )}
                  </div>
                </div>
              ))}
            </div>
          )}

          {/* 我的好友 */}
          {activeTab === "list" && (
            <div className="card">
              <h4>好友列表</h4>

              {loadingFriends && <div className="tip">加载中...</div>}

              {!loadingFriends && friends.length === 0 && (
                <div className="tip">暂无好友</div>
              )}

              {friends.map((item) => (
                <div className="user-item" key={item.friend_id}>
                  <img
                    className="avatar"
                    src={avatarMap[String(item.friend_id)] || "/default-avatar.png"}
                    onError={(e) =>
                      (e.currentTarget.src = "/default-avatar.png")
                    }
                  />

                  <div className="user-info">
                    <div className="name">{item.friend_remark || item.friend_id}</div>
                    <div className="desc offline">离线</div>
                  </div>

                  <button
                    className="ghost-btn chat-highlight"
                    onClick={() =>
                      handleChat(item.friend_id, item.friend_remark)
                    }
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