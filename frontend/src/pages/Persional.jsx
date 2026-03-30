import { useState } from "react";
import { useNavigate } from "react-router-dom";
import "../styles/Persional.css";

export default function EditProfilePage() {
  const navigate = useNavigate();

  // ===== 选择修改类型 =====
  const [selectedType, setSelectedType] = useState("");

  // ===== UID 表单状态 =====
  const [uid, setUid] = useState("");
  const [uidLoading, setUidLoading] = useState(false);
  const [uidError, setUidError] = useState("");
  const [uidSuccess, setUidSuccess] = useState(false);

  // ===== 昵称表单状态 =====
  const [nickname, setNickname] = useState("");
  const [nicknameLoading, setNicknameLoading] = useState(false);
  const [nicknameError, setNicknameError] = useState("");
  const [nicknameSuccess, setNicknameSuccess] = useState(false);

  // ===== 密码表单状态 =====
  const [password, setPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [passwordLoading, setPasswordLoading] = useState(false);
  const [passwordError, setPasswordError] = useState("");
  const [passwordSuccess, setPasswordSuccess] = useState(false);

  // 通用提交函数
  const handleSubmit = async (e, type) => {
    e.preventDefault();

    let url = "";
    let body = {};
    let setLoading;
    let setError;
    let setSuccess;

    switch (type) {
      case "uid":
        setLoading = setUidLoading;
        setError = setUidError;
        setSuccess = setUidSuccess;
        setUidError("");
        setUidSuccess(false);
        if (!uid.trim()) {
          setUidError("UID不能为空");
          return;
        }
        url = "http://localhost:8000/api/auth/me/uid";
        body = { uid };
        break;
      case "nickname":
        setLoading = setNicknameLoading;
        setError = setNicknameError;
        setSuccess = setNicknameSuccess;
        setNicknameError("");
        setNicknameSuccess(false);
        if (!nickname.trim()) {
          setNicknameError("昵称不能为空");
          return;
        }
        url = "http://localhost:8000/api/auth/me/name";
        body = { name: nickname };
        break;
      case "password":
        setLoading = setPasswordLoading;
        setError = setPasswordError;
        setSuccess = setPasswordSuccess;
        setPasswordError("");
        setPasswordSuccess(false);
        if (!password || !confirmPassword) {
          setPasswordError("密码不能为空");
          return;
        }
        if (password !== confirmPassword) {
          setPasswordError("两次输入的密码不一致");
          return;
        }
        url = "http://localhost:8000/api/auth/me/password";
        body = { password };
        break;
      default:
        return;
    }

    try {
      setLoading(true);
      const token = sessionStorage.getItem("token");

      const res = await fetch(url, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify(body),
      });

      const data = await res.json();
      if (!res.ok || (type === "uid" && data.code !== 201)) {
        throw new Error(data.message || "修改失败");
      }

      setSuccess(true);
      setTimeout(() => navigate(-1), 1200);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="profile-page">
      <div className="profile-card">
        <h2 className="profile-title">修改个人资料</h2>

        {/* ===== 选择修改类型 ===== */}
        {!selectedType && (
          <div className="profile-select">
            <button
              className="profile-btn"
              onClick={() => setSelectedType("uid")}
            >
              修改 UID
            </button>
            <button
              className="profile-btn"
              onClick={() => setSelectedType("nickname")}
            >
              修改昵称
            </button>
            <button
              className="profile-btn"
              onClick={() => setSelectedType("password")}
            >
              修改密码
            </button>
          </div>
        )}

        {/* ===== 修改 UID ===== */}
        {selectedType === "uid" && (
          <form onSubmit={(e) => handleSubmit(e, "uid")} className="profile-form">
            <label className="profile-label">UID</label>
            <input
              type="text"
              className="profile-input"
              placeholder="请输入新的UID"
              value={uid}
              onChange={(e) => setUid(e.target.value)}
              disabled={uidLoading}
            />
            {uidError && <div className="profile-error">{uidError}</div>}
            {uidSuccess && <div className="profile-success">UID修改成功</div>}
            <div className="profile-btn-group">
              <button type="submit" className="profile-btn" disabled={uidLoading}>
                {uidLoading ? "提交中..." : "保存UID修改"}
              </button>
              <button
                type="button"
                className="profile-btn secondary"
                onClick={() => setSelectedType("")}
              >
                返回选择
              </button>
            </div>
          </form>
        )}

        {/* ===== 修改昵称 ===== */}
        {selectedType === "nickname" && (
          <form
            onSubmit={(e) => handleSubmit(e, "nickname")}
            className="profile-form"
          >
            <label className="profile-label">昵称</label>
            <input
              type="text"
              className="profile-input"
              placeholder="请输入新的昵称"
              value={nickname}
              onChange={(e) => setNickname(e.target.value)}
              disabled={nicknameLoading}
            />
            {nicknameError && <div className="profile-error">{nicknameError}</div>}
            {nicknameSuccess && <div className="profile-success">昵称修改成功</div>}
           <div className="profile-btn-group">
            <button type="submit" className="profile-btn" disabled={nicknameLoading}>
              {nicknameLoading ? "提交中..." : "保存昵称修改"}
            </button>
            <button
              type="button"
              className="profile-btn secondary"
              onClick={() => setSelectedType("")}
            >
              返回选择
            </button>
          </div>
          </form>
        )}

        {/* ===== 修改密码 ===== */}
        {selectedType === "password" && (
          <form
            onSubmit={(e) => handleSubmit(e, "password")}
            className="profile-form"
          >
            <label className="profile-label">新密码</label>
            <input
              type="password"
              className="profile-input"
              placeholder="请输入新密码"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              disabled={passwordLoading}
            />
            <label className="profile-label">确认新密码</label>
            <input
              type="password"
              className="profile-input"
              placeholder="请再次输入新密码"
              value={confirmPassword}
              onChange={(e) => setConfirmPassword(e.target.value)}
              disabled={passwordLoading}
            />
            {passwordError && <div className="profile-error">{passwordError}</div>}
            {passwordSuccess && <div className="profile-success">密码修改成功</div>}
            <div className="profile-btn-group">
              <button type="submit" className="profile-btn" disabled={passwordLoading}>
                {passwordLoading ? "提交中..." : "保存密码修改"}
              </button>
              <button
                type="button"
                className="profile-btn secondary"
                onClick={() => setSelectedType("")}
              >
                返回选择
              </button>
            </div>
          </form>
        )}
      </div>
    </div>
  );
}