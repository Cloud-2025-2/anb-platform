import { useState } from "react";
import api from "../lib/api";
import { setToken } from "../lib/auth";
import { useNavigate, Link } from "react-router-dom";

export default function Login() {
  const [email, setEmail] = useState(""); 
  const [password, setPassword] = useState("");
  const [msg, setMsg] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const nav = useNavigate();

  const submit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault(); setMsg(null);
    try {
      setLoading(true);
      const { data } = await api.post<{access_token:string}>("/auth/login", { email, password });
      setToken(data.access_token);
      nav("/upload");
    } catch {
      setMsg("Invalid credentials");
    } finally { setLoading(false); }
  };

  return (
    <div className="card" style={{padding:24}}>
      <h1>Log in to your account</h1>
      <p className="helper">Or <Link to="/signup">create a new account</Link></p>

      <form onSubmit={submit} className="form" style={{maxWidth:480}}>
        <div className="field">
          <label>Email</label>
          <input className="input" value={email}
            onChange={e=>setEmail(e.target.value)} required />
        </div>
        <div className="field">
          <label>Password</label>
          <input type="password" className="input" value={password}
            onChange={e=>setPassword(e.target.value)} required />
        </div>
        <button className="btn btn-primary" disabled={loading}>
          {loading ? "Logging in..." : "Log in"}
        </button>
        {msg && <div className="error" style={{marginTop:6}}>{msg}</div>}
      </form>
    </div>
  );
}
