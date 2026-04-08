import { useState, useCallback } from "react";
import { useNavigate } from "react-router-dom";
import Cropper from "react-easy-crop";
import "../styles/Persional.css";

export default function EditProfilePage() {
  const navigate = useNavigate();

  const [selectedType, setSelectedType] = useState("");

  // ===== 头像 =====
  const [avatar, setAvatar] = useState(null);
  const [avatarPreview, setAvatarPreview] = useState(null);

  const [avatarLoading, setAvatarLoading] = useState(false);
  const [avatarError, setAvatarError] = useState("");
  const [avatarSuccess, setAvatarSuccess] = useState(false);

  // ===== 裁剪 =====
  const [crop, setCrop] = useState({ x: 0, y: 0 });

  const [zoom, setZoom] = useState(1);

  const [croppedAreaPixels, setCroppedAreaPixels] = useState(null);

  const onCropComplete = useCallback((_, croppedPixels) => {
    setCroppedAreaPixels(croppedPixels);
  }, []);

  // 图片选择
  const handleAvatarChange = (e) => {
    const file = e.target.files[0];

    if (!file) return;

    setAvatar(file);

    setAvatarPreview(URL.createObjectURL(file));
  };

  // 裁剪图片函数
  const getCroppedImg = async (imageSrc, crop) => {
    const image = new Image();

    image.src = imageSrc;

    await new Promise((resolve) => (image.onload = resolve));

    const canvas = document.createElement("canvas");

    const ctx = canvas.getContext("2d");

    canvas.width = crop.width;
    canvas.height = crop.height;

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
      canvas.toBlob((blob) => {
        resolve(blob);
      }, "image/jpeg");
    });
  };

  // 上传头像
  const handleAvatarUpload = async (e) => {
    e.preventDefault();

    setAvatarError("");
    setAvatarSuccess(false);

    if (!avatarPreview || !croppedAreaPixels) {
      setAvatarError("请选择头像");
      return;
    }

    try {
      setAvatarLoading(true);

      const croppedBlob = await getCroppedImg(
        avatarPreview,
        croppedAreaPixels
      );

      const formData = new FormData();

      formData.append("file", croppedBlob, "avatar.jpg");

      const token = sessionStorage.getItem("token");

      const res = await fetch("http://localhost:8000/api/auth/me/head", {
        method: "POST",
        headers: {
          Authorization: `Bearer ${token}`,
        },
        body: formData,
      });

      const data = await res.json();

      if (!res.ok) {
        throw new Error(data.message || "上传失败");
      }

      setAvatarSuccess(true);

      setTimeout(() => navigate(-1), 1200);
    } catch (err) {
      setAvatarError(err.message);
    } finally {
      setAvatarLoading(false);
    }
  };

  return (
    <div className="profile-page">
      <div className="profile-card">
        <h2 className="profile-title">修改个人资料</h2>

        {!selectedType && (
          <div className="profile-select">
            <button
              className="profile-btn"
              onClick={() => setSelectedType("avatar")}
            >
              修改头像
            </button>
          </div>
        )}

        {selectedType === "avatar" && (
          <form onSubmit={handleAvatarUpload} className="profile-form">
            <label className="profile-label">上传头像</label>

            <input
              type="file"
              accept="image/*"
              className="profile-input"
              onChange={handleAvatarChange}
            />

            {avatarPreview && (
              <div className="crop-container">
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
            )}

            <input
              type="range"
              min={1}
              max={3}
              step={0.1}
              value={zoom}
              onChange={(e) => setZoom(e.target.value)}
            />

            {avatarError && <div className="profile-error">{avatarError}</div>}
            {avatarSuccess && (
              <div className="profile-success">头像上传成功</div>
            )}

            <div className="profile-btn-group">
              <button className="profile-btn" disabled={avatarLoading}>
                {avatarLoading ? "上传中..." : "保存头像"}
              </button>

              <button
                type="button"
                className="profile-btn secondary"
                onClick={() => setSelectedType("")}
              >
                返回
              </button>
            </div>
          </form>
        )}
      </div>
    </div>
  );
}