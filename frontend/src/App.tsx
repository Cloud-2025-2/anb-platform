import { Outlet, Link, useNavigate } from "react-router-dom";
import { isLoggedIn, clearToken } from "./lib/auth";

export default function App() {
  const nav = useNavigate();

  const logout = () => {
    clearToken();
    nav("/login");
  };

  return (
    <div style={{ maxWidth: 900, margin: "0 auto", padding: 16 }}>
      <nav style={{ display: "flex", gap: 12, marginBottom: 16 }}>
        <Link to="/">Públicos</Link>
        <Link to="/rankings">Ranking</Link>
        {isLoggedIn() && (
          <>
            <Link to="/upload">Subir video</Link>
            <Link to="/my-videos">Mis videos</Link>
            <button onClick={logout}>Salir</button>
          </>
        )}
        {!isLoggedIn() && (
          <>
            <Link to="/login">Login</Link>
            <Link to="/signup">Signup</Link>
          </>
        )}
      </nav>
      <Outlet />
    </div>
  );
}
