import "./Button.css";

export default function Button({ children, onClick, type = "primary" }) {
  return (
    <button className={`btn btn-${type}`} onClick={onClick}>
      {children}
    </button>
  );
}
