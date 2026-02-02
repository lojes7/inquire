import { useState } from "react";
import { useNavigate } from "react-router-dom";
import "../styles/Persional.css";

export default function EditProfilePage() {
  const navigate = useNavigate();

  const [uid, setUid] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState(false);

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError("");
    setSuccess(false);

    if (!uid.trim()) {
      setError("微信号不能为空");
      return;
    }

    try {
      setLoading(true);
      const token = localStorage.getItem("token");

      const res = await fetch("http://localhost:8000/api/auth/me/uid", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify({ uid }),
      });

      const data = await res.json();

      if (!res.ok || data.code !== 201) {
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

        <form onSubmit={handleSubmit} className="profile-form">
          <label className="profile-label">微信号</label>
          <input
            type="text"
            className="profile-input"
            placeholder="请输入新的微信号"
            value={uid}
            onChange={(e) => setUid(e.target.value)}
            disabled={loading}
          />

          {error && <div className="profile-error">{error}</div>}
          {success && <div className="profile-success">修改成功</div>}

          <button
            type="submit"
            className="profile-btn"
            disabled={loading}
          >
            {loading ? "提交中..." : "保存修改"}
          </button>
        </form>
      </div>
    </div>
  );
}

