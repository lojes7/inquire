import { useState, useCallback, useRef, useEffect } from "react";
import Cropper from "react-easy-crop";
import Sidebar from "../components/Sidebar";
import "../styles/Persional.css";

export default function EditProfilePage() {
  const fileInputRef = useRef(null);

  /*
  =========================
  用户信息（不动）
  =========================
  */
  const [uid, setUid] = useState("");
  const [nickname, setNickname] = useState("");
  const [avatarUrl, setAvatarUrl] = useState("/default-avatar.png");

  /*
  =========================
  密码（不动）
  =========================
  */
  const [password, setPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");

  const [editingField, setEditingField] = useState("");

  /*
  =========================
  外观 / 通知（第二版UI需要）
  =========================
  */
  const [theme, setTheme] = useState("light");

  const [notifications, setNotifications] = useState({
    email: true,
    push: false,
    weekly: true,
  });

  const [avatarPreview, setAvatarPreview] = useState(null);

  const [crop, setCrop] = useState({ x: 0, y: 0 });
  const [zoom, setZoom] = useState(1);
  const [croppedAreaPixels, setCroppedAreaPixels] = useState(null);

  const onCropComplete = useCallback((_, croppedPixels) => {
    setCroppedAreaPixels(croppedPixels);
  }, []);

  /*
  =========================
  token（不动）
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
  初始化
  =========================
  */
  useEffect(() => {
  const user = JSON.parse(sessionStorage.getItem("user") || "{}");

  console.log("初始化 user:", user);

  if (user) {
    setUid(user.uid || "");
    setNickname(user.name || "");

    if (user.id) {
      console.log("开始请求头像 userId =", user.id);
      fetchAvatar(user.id);
    } else {
      console.log("❌ 没有 user.id");
    }
  }
}, []);

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
          headers: { Authorization: `Bearer ${getToken()}` },
        }
      );

      if (!res.ok) return;

      const blob = await res.blob();
      setAvatarUrl(URL.createObjectURL(blob));
    } catch {}
  };

  /*
  =========================
  保存 UID / 昵称（功能不动）
  =========================
  */
  const saveField = async (type) => {
    const token = getToken();
    const user = JSON.parse(sessionStorage.getItem("user") || "{}");

    let url = "";
    let body = {};

    if (type === "uid") {
      url = "http://localhost:8000/api/auth/me/uid";
      body = { uid };
    }

    if (type === "nickname") {
      url = "http://localhost:8000/api/auth/me/name";
      body = { name: nickname };
    }

    const res = await fetch(url, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${token}`,
      },
      body: JSON.stringify(body),
    });

    const data = await res.json();
    if (!res.ok) return alert(data.message || "修改失败");

    if (type === "uid") user.uid = uid;
    if (type === "nickname") user.name = nickname;

    sessionStorage.setItem("user", JSON.stringify(user));
    window.dispatchEvent(new Event("userUpdated"));

    setEditingField("");
    alert("保存成功");
  };

  /*
  =========================
  保存密码（不动）
  =========================
  */
  const savePassword = async () => {
    if (!password || !confirmPassword)
      return alert("密码不能为空");

    if (password !== confirmPassword)
      return alert("两次密码不一致");

    const res = await fetch(
      "http://localhost:8000/api/auth/me/password",
      {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${getToken()}`,
        },
        body: JSON.stringify({ password }),
      }
    );

    const data = await res.json();
    if (!res.ok) return alert(data.message || "修改失败");

    setPassword("");
    setConfirmPassword("");

    alert("密码修改成功");
  };

  /*
  =========================
  头像裁剪上传（不动）
  =========================
  */
  const getCroppedBlob = async (imageSrc, crop) => {
    const image = new Image();
    image.src = imageSrc;
    await new Promise((r) => (image.onload = r));

    const canvas = document.createElement("canvas");
    canvas.width = crop.width;
    canvas.height = crop.height;

    const ctx = canvas.getContext("2d");

    ctx.drawImage(
      image,
      crop.x,
      crop.y,
      crop.width,
      crop.height,
      0,
      0,
      crop.width,
      crop.height
    );

    return new Promise((resolve) => {
      canvas.toBlob((blob) => resolve(blob), "image/jpeg");
    });
  };

  const saveAvatar = async () => {
    if (!avatarPreview || !croppedAreaPixels)
      return alert("请选择头像");

    const blob = await getCroppedBlob(avatarPreview, croppedAreaPixels);

    const formData = new FormData();
    formData.append("file", blob, "avatar.jpg");

    const res = await fetch(
      "http://localhost:8000/api/auth/me/head",
      {
        method: "POST",
        headers: {
          Authorization: `Bearer ${getToken()}`,
        },
        body: formData,
      }
    );

    const data = await res.json();
    if (!res.ok) return alert(data.message || "上传失败");

    window.dispatchEvent(new Event("userUpdated"));

    setAvatarPreview(null);
    alert("头像更新成功");
  };

  /*
  =========================
  通知 / 主题
  =========================
  */
  const toggleNotification = (key) => {
    setNotifications((prev) => ({
      ...prev,
      [key]: !prev[key],
    }));
  };

  return (
    <div className="settings-layout">
      <Sidebar />

      <div className="settings-main">
        <div className="settings-container">

          {/* ================= 顶部 ================= */}
          <div className="settings-header">
            <h1>设置</h1>
            <p>管理你的账号与偏好设置</p>
          </div>

          {/* ================= 个人资料 ================= */}
          <section className="settings-card">
            <h2>👤 个人资料</h2>

            {/* 头像 */}
            <div className="setting-item avatar-item">
              <span>头像</span>

              <div className="avatar-right">
                <img
                  src={avatarPreview || avatarUrl}
                  className="real-avatar"
                />

                <button
                  className="edit-btn"
                  onClick={() => fileInputRef.current.click()}
                >
                  修改头像
                </button>

                <input
                  type="file"
                  hidden
                  ref={fileInputRef}
                  accept="image/*"
                  onChange={(e) =>
                    setAvatarPreview(
                      URL.createObjectURL(e.target.files[0])
                    )
                  }
                />
              </div>
            </div>

            {avatarPreview && (
              <>
                <div className="crop-box">
                  <Cropper
                    image={avatarPreview}
                    crop={crop}
                    zoom={zoom}
                    aspect={1}
                    cropShape="round"
                    showGrid={false}
                    onCropChange={setCrop}
                    onZoomChange={setZoom}
                    onCropComplete={onCropComplete}
                  />
                </div>

                <input
                  type="range"
                  min={1}
                  max={3}
                  step={0.1}
                  value={zoom}
                  onChange={(e) => setZoom(e.target.value)}
                />

                <button className="save-btn" onClick={saveAvatar}>
                  保存头像
                </button>
              </>
            )}

            {/* 昵称 */}
            <div
              className="setting-item clickable"
              onClick={() =>
                setEditingField(editingField === "nickname" ? "" : "nickname")
              }
            >
              <span>昵称</span>
              <span>{nickname} ＞</span>
            </div>

            {editingField === "nickname" && (
              <>
                <input
                  className="settings-input"
                  value={nickname}
                  onChange={(e) => setNickname(e.target.value)}
                />
                <button className="save-btn" onClick={() => saveField("nickname")}>
                  保存昵称
                </button>
              </>
            )}

            {/* UID */}
            <div
              className="setting-item clickable"
              onClick={() =>
                setEditingField(editingField === "uid" ? "" : "uid")
              }
            >
              <span>UID</span>
              <span>{uid} ＞</span>
            </div>

            {editingField === "uid" && (
              <>
                <input
                  className="settings-input"
                  value={uid}
                  onChange={(e) => setUid(e.target.value)}
                />
                <button className="save-btn" onClick={() => saveField("uid")}>
                  保存UID
                </button>
              </>
            )}
          </section>

          {/* ================= 通知 ================= */}
          <section className="settings-card">
            <h2>🔔 通知设置</h2>

            {[
              { key: "email", title: "邮件通知", desc: "重要更新和提醒" },
              { key: "push", title: "推送通知", desc: "实时消息提醒" },
              { key: "weekly", title: "每周摘要", desc: "获取周报" },
            ].map((item) => (
              <div
                key={item.key}
                className="toggle-row"
                onClick={() => toggleNotification(item.key)}
              >
                <div>
                  <p>{item.title}</p>
                  <span>{item.desc}</span>
                </div>

                <div
                  className={`toggle-switch ${
                    notifications[item.key] ? "active" : ""
                  }`}
                >
                  <span />
                </div>
              </div>
            ))}
          </section>

          {/* ================= 外观 ================= */}
          <section className="settings-card">
            <h2>🎨 外观设置</h2>

            <div className="theme-grid">
              {[
                { value: "light", label: "浅色模式", icon: "☀️" },
                { value: "dark", label: "深色模式", icon: "🌙" },
                { value: "system", label: "跟随系统", icon: "💻" },
              ].map((item) => (
                <button
                  key={item.value}
                  onClick={() => setTheme(item.value)}
                  className={`theme-card ${
                    theme === item.value ? "selected" : ""
                  }`}
                >
                  <div>{item.icon}</div>
                  <span>{item.label}</span>
                </button>
              ))}
            </div>
          </section>

          {/* ================= 安全 ================= */}
          <section className="settings-card">
            <h2>🔒 账号安全</h2>

            <input
              type="password"
              className="settings-input"
              placeholder="新密码"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
            />

            <input
              type="password"
              className="settings-input"
              placeholder="确认密码"
              value={confirmPassword}
              onChange={(e) => setConfirmPassword(e.target.value)}
            />

            <button className="save-btn" onClick={savePassword}>
              保存密码
            </button>
          </section>

          {/* ================= 保存按钮（UI对齐第二版） ================= */}
          <div className="save-row">
            <button className="save-btn">保存更改</button>
          </div>

        </div>
      </div>
    </div>
  );
}