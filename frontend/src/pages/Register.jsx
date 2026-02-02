// src/pages/Register.jsx
import '../styles/Register.css';
import bg from '../images/background1.svg';
import logo from '../images/logo.svg';
import { Link, useNavigate } from 'react-router-dom';
import { useState } from 'react';

export default function Register() {
  const navigate = useNavigate();

  // 表单状态
  const [phone, setPhone] = useState('');
  const [wechat, setWechat] = useState('');
  const [password, setPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  // 提交注册
  const handleRegister = async () => {
    setError('');

    // 简单前端校验
    if (!phone || !wechat || !password || !confirmPassword) {
      setError('请填写完整信息');
      return;
    }
    if (password !== confirmPassword) {
      setError('两次输入的密码不一致');
      return;
    }

    setLoading(true);

    try {
      const res = await fetch('http://localhost:8000/api/register', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          name: wechat,
          password: password,
          phone_number: phone,
        }),
      });

      const data = await res.json();

      if (data.code === 201) {
        alert('注册成功！即将跳转到登录页');
        navigate('/login'); // 注册成功跳转登录
      } else {
        // 失败
        setError(data.message || '注册失败，请重试');
      }
    } catch (err) {
      console.error(err);
      setError('网络错误，请稍后重试');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="register-container">
      {/* 左侧 */}
      <div className="register-left">
        <div className="logo">
          <img src={logo} alt="logo" className="logo-icon" />
          <span className="logo-text">
            高效办公，文件<span className="logo-highlight">询觅</span>
          </span>
        </div>

        <div className="illustration">
          <img src={bg} alt="illustration" />
        </div>

        <div className="login-tip">
          <p>如果您已有账户</p>
          <p>请点击</p>
          <Link to="/login">立即登录</Link>
        </div>
      </div>

      {/* 右侧 */}
      <div className="register-right">
        <h1>注册</h1>
        <p className="subtitle">创建一个新用户吧！</p>

        <div className="form">
          {error && <div className="error-msg">{error}</div>}

          <label>手机号</label>
          <input
            type="text"
            placeholder="输入你的手机号"
            value={phone}
            onChange={(e) => setPhone(e.target.value)}
          />

          <label>微信号</label>
          <input
            type="text"
            placeholder="输入你的微信号"
            value={wechat}
            onChange={(e) => setWechat(e.target.value)}
          />

          <label>密码</label>
          <input
            type="password"
            placeholder="输入你的密码"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
          />

          <input
            type="password"
            placeholder="再次输入你的密码"
            className="margin-top"
            value={confirmPassword}
            onChange={(e) => setConfirmPassword(e.target.value)}
          />

          <button
            className="submit-btn"
            onClick={handleRegister}
            disabled={loading}
          >
            {loading ? '注册中...' : '创建'}
          </button>
        </div>
      </div>
    </div>
  );
}
