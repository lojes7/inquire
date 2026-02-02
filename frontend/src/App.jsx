import { Routes, Route, Navigate } from 'react-router-dom';
import Register from './pages/Register';
import Login from './pages/Login';
import ChatApp from './pages/ChatApp';
import Login1 from './pages/Login1';
import ChatPage from './pages/ChatPage';
import AddFriendPage from './pages/AddFriendPage';
import Persional from './pages/Persional';

function App() {
  return (
    <Routes>
      <Route path="/" element={<Navigate to="/login" />} />
      <Route path="/register" element={<Register />} />
      <Route path="/login" element={<Login />} />
      <Route path="/chat" element={<ChatApp />} />
      <Route path="/login1" element={<Login1 />} />
      <Route path="/chatpage" element={<ChatPage />} />
      <Route path="/addfriend" element={<AddFriendPage />} />
      <Route path="/persional" element={<Persional />} />
    </Routes>
  );
}

export default App;
