import { NavLink, useNavigate } from "react-router-dom";
import { useEffect, useState } from "react";
import "../styles/Sidebar.css";
import logo from "../images/logo.svg";

const navItems = [
  { to: "/chat", icon: "💬", label: "聊天" },
  { to: "/addfriend", icon: "👥", label: "好友" },
  { to: "/chatpage", icon: "📝", label: "文档" },
  { to: "/calendar", icon: "📅", label: "日程" },
  { to: "/persional", icon: "⚙️", label: "设置" },
  
];

export default function Sidebar() {
  const navigate = useNavigate();

  const [userInfo, setUserInfo] = useState({
    name: "",
    uid: "",
    id: "",
  });

  const [avatar, setAvatar] = useState("/default-avatar.png");

  /*
  ========================
  获取 token
  ========================
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
  ========================
  获取头像
  ========================
  */
  const fetchAvatar = async (userId) => {
    try {
      const token = getToken();

      const res = await fetch(
        `http://localhost:8000/api/auth/head/${userId}`,
        {
          headers: {
            Authorization: `Bearer ${token}`,
          },
        }
      );

      if (!res.ok) throw new Error();

      const blob = await res.blob();

      const imageUrl = URL.createObjectURL(blob);

      setAvatar(imageUrl);

    } catch (err) {
      console.error("头像加载失败");
    }
  };

  /*
  ========================
  加载用户信息
  ========================
  */
  const loadUserInfo = () => {
    const user = JSON.parse(
      sessionStorage.getItem("user") || "{}"
    );

    setUserInfo({
      name: user.name || "未知用户",
      uid: user.uid || "暂无UID",
      id: user.id || "",
    });

    if (user.id) {
      fetchAvatar(user.id);
    }
  };

  /*
  ========================
  初始化 + 监听用户更新
  ========================
  */
  useEffect(() => {
    loadUserInfo();

    window.addEventListener("userUpdated", loadUserInfo);

    return () => {
      window.removeEventListener("userUpdated", loadUserInfo);
    };
  }, []);

  /*
  ========================
  退出登录
  ========================
  */
  const handleLogout = () => {
    sessionStorage.clear();
    navigate("/login");
  };

  return (
    <aside className="sidebar">
      {/* Logo */}
      <div className="sidebar-logo">
        <img src={logo} alt="logo" className="logo-icon" />
        <span className="logo-text">询觅</span>
      </div>

      {/* 用户信息 */}
      <div className="sidebar-user">
        <img
          className="user-avatar"
          src={avatar}
          alt="头像"
          onError={(e) => {
            e.target.src = "/default-avatar.png";
          }}
        />

        <div className="user-info">
          <p className="user-name">{userInfo.name}</p>
          <p className="user-email">UID：{userInfo.uid}</p>
        </div>
      </div>

      {/* 主导航 */}
      <nav className="sidebar-nav">
        {navItems.map((item) => (
          <NavLink
            key={item.to}
            to={item.to}
            className={({ isActive }) =>
              isActive ? "sidebar-item active" : "sidebar-item"
            }
          >
            <span>{item.icon}</span>
            <span>{item.label}</span>
          </NavLink>
        ))}
      </nav>

      {/* 底部退出 */}
      <div className="sidebar-bottom">
        <button
          onClick={handleLogout}
          className="sidebar-item logout-btn"
        >
          <span>🚪</span>
          <span>退出登录</span>
        </button>
      </div>
    </aside>
  );
}