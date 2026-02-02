// src/pages/Register.jsx
import '../styles/Register.css'
import bg from '../images/background1.svg';
import logo from '../images/logo.svg';
import { Link } from 'react-router-dom';
import { useNavigate } from 'react-router-dom';
import { useState } from "react";

export default function Register() {
  const navigate = useNavigate(); // ⭐ 关键
  const [phone, setPhone] = useState("");      // 手机号输入
  const [password, setPassword] = useState(""); // 密码输入

  const handleLogin = async () => {
  try {
    const res = await fetch("http://localhost:8000/api/login/phone_number", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        phone_number: phone,
        password: password,
      }),
    });

    const data = await res.json();
    console.log(data); // 调试用

    if (data.code === 200) {
      // 登录成功
      localStorage.setItem("token", data.data.token_class.token);
      localStorage.setItem("refresh_token", data.data.token_class.refresh_token);

      // 再跳转
      navigate("/chat"); // 原来你跳转的页面
    } else {
      alert(data.message); // 后端返回错误提示
    }
  } catch (err) {
    console.error(err);
    alert("网络错误，请稍后重试");
  }
};


  const handleUsernameLogin = () => {
  navigate("/Login1"); // 替换成你想跳转的页面路径

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
          <p>如果您还没有账户</p>
          <p>请点击</p>
          <Link to="/Register">立即注册</Link>
        </div>
      </div>

  {/* 右侧 */}
      <div className="register-right">
        <h1>登录</h1>
        <p className="subtitle">高效办公，文件询觅</p>
        <div className="form">
          {/* 在手机号输入框上方加提示 */}
          <label>手机号</label>
          <span
            className="input-tip"
            onClick={handleUsernameLogin}
            style={{ cursor: "pointer" }} // 保持样式不变，鼠标悬停显示可点击
          >
            UID登录
          </span>
          <input
              type="text"
              placeholder="输入你的手机号"
              value={phone}
              onChange={(e) => setPhone(e.target.value)}
            />
            <label>密码</label>
            <input
              type="password"
              placeholder="输入你的密码"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
            />
          <span className="input-tip">忘记密码？</span>
          <button className="submit-btn" onClick={handleLogin}>
             登录
        </button>
        </div>
      </div>
    </div>
  )
}
