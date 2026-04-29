import { NavLink, useNavigate } from "react-router-dom";
import { useEffect, useState } from "react";
import "../styles/Sidebar.css";
import logo from "../images/logo.svg";

const navItems = [
  { to: "/chat", icon: "chat" },
  { to: "/addfriend", icon: "group_add" },
  { to: "/chatpage", icon: "edit_note" },
  { to: "/calendar", icon: "calendar_month" },
  { to: "/persional", icon: "settings" },
];

export default function Sidebar() {
  const navigate = useNavigate();

  const [userInfo, setUserInfo] = useState({
    name: "",
    uid: "",
    id: "",
  });

  const [avatar, setAvatar] = useState("/default-avatar.png");

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

  const fetchAvatar = async (userId) => {
    try {
      const token = getToken();

      const res = await fetch(
        `http://localhost:8000/api/auth/info/head/${userId}?t=${Date.now()}`, // ✅ 防缓存
        {
          headers: {
            Authorization: `Bearer ${token}`,
          },
        }
      );

      if (!res.ok) throw new Error();

      const blob = await res.blob();
      if (!blob || blob.size === 0) throw new Error();

      const imageUrl = URL.createObjectURL(blob);

      // ✅ 防内存泄漏
      setAvatar((prev) => {
        if (prev && prev.startsWith("blob:")) {
          URL.revokeObjectURL(prev);
        }
        return imageUrl;
      });

    } catch {
      const purpleAvatar =
        "data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='100' height='100'%3E%3Crect width='100%25' height='100%25' fill='%23E6E6FA'/%3E%3C/svg%3E";

      setAvatar(purpleAvatar);
    }
  };

  const loadUserInfo = () => {
    const user = JSON.parse(sessionStorage.getItem("user") || "{}");

    setUserInfo({
      name: user.name || "未知用户",
      uid: user.uid || "暂无UID",
      id: user.id || "",
    });

    if (user.id) {
      setAvatar(""); // 👈 强制刷新
      setTimeout(() => fetchAvatar(user.id), 50);
    }
  };

  useEffect(() => {
    loadUserInfo();
    window.addEventListener("userUpdated", loadUserInfo);
    return () => {
      window.removeEventListener("userUpdated", loadUserInfo);
    };
  }, []);

  const handleLogout = () => {
    sessionStorage.clear();
    navigate("/login");
  };

  return (
    <aside className="sidebar mini">
      {/* Logo（只留图标） */}
      <div className="sidebar-logo">
        <img src={logo} alt="logo" className="logo-icon" />
      </div>

      {/* 用户头像 */}
      <div className="sidebar-user">
        <img
          className="user-avatar"
          src={avatar}
          alt="头像"
          onError={(e) => {
            e.target.src = "/default-avatar.png";
          }}
        />
      </div>

      {/* 导航 */}
      <nav className="sidebar-nav">
        {navItems.map((item) => (
          <NavLink
            key={item.to}
            to={item.to}
            className={({ isActive }) =>
              isActive ? "sidebar-item active" : "sidebar-item"
            }
          >
            <span className="material-symbols-outlined icon">
              {item.icon}
            </span>
          </NavLink>
        ))}
      </nav>

      {/* 退出 */}
      <div className="sidebar-bottom">
        <button
          onClick={handleLogout}
          className="sidebar-item logout-btn"
        >
          <span className="material-symbols-outlined">
            logout
          </span>
        </button>
      </div>
    </aside>
  );
}