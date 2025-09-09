import { useState } from "react";
import api from "../lib/api";
import { setToken } from "../lib/auth";
import { useNavigate, Link } from "react-router-dom";

function Login() {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [msg, setMsg] = useState<string | null>(null);
  const nav = useNavigate();

  const submit = async (e: React.FormEvent) => {
    e.preventDefault();
    setMsg(null);
    try {
      const { data } = await api.post("/auth/login", { email, password });
      setToken(data?.access_token);
      nav("/upload");
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    } catch (e: any) {
      setMsg(e?.response?.data?.message || "Credenciales inválidas");
    }
  };

  return (
    <form onSubmit={submit} style={{ display: "grid", gap: 8 }}>
      <h2>Login</h2>
      <input placeholder="email" value={email} onChange={(e)=>setEmail(e.target.value)} required />
      <input type="password" placeholder="password" value={password} onChange={(e)=>setPassword(e.target.value)} required />
      <button>Entrar</button>
      <small>{msg}</small>
      <Link to="/signup">Crear cuenta</Link>
    </form>
  );
}

export default Login;
