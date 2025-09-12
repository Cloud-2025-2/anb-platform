import { useState } from "react";
import { Link } from "react-router-dom";
import api from "../lib/api";
import type { AxiosError } from "axios";

type SignupForm = {
  first_name: string;
  last_name: string;
  email: string;
  password1: string;
  password2: string;
  city: string;
  country: string;
};

const INITIAL: SignupForm = {
  first_name: "",
  last_name: "",
  email: "",
  password1: "",
  password2: "",
  city: "",
  country: "",
};

export default function Signup() {
  const [form, setForm] = useState<SignupForm>(INITIAL);
  const [loading, setLoading] = useState(false);
  const [msg, setMsg] = useState<{ kind: "ok" | "err"; text: string } | null>(null);

  const on = (k: keyof SignupForm) =>
    (e: React.ChangeEvent<HTMLInputElement>) =>
      setForm({ ...form, [k]: e.target.value });

  const submit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    setMsg(null);

    if (!/.+@.+\..+/.test(form.email)) {
      setMsg({ kind: "err", text: "Invalid email" });
      return;
    }
    if (form.password1.length < 6) {
      setMsg({ kind: "err", text: "Password must be at least 6 characters" });
      return;
    }
    if (form.password1 !== form.password2) {
      setMsg({ kind: "err", text: "Passwords do not match" });
      return;
    }

    try {
      setLoading(true);
      await api.post("/auth/signup", form);
      setMsg({ kind: "ok", text: "Account created. You can now log in." });
      setForm(INITIAL);
    } catch (err: unknown) {
      const ax = err as AxiosError<{ message?: string }>;
      setMsg({
        kind: "err",
        text: ax.response?.data?.message ?? "Could not create account",
      });
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="card" style={{ padding: 24 }}>
      <h1>Create your account</h1>
      <p className="helper">Join our community and start competing!</p>

      <form onSubmit={submit} className="form" style={{ maxWidth: 560 }}>
        <div className="field">
          <label>First name</label>
          <input className="input" value={form.first_name} onChange={on("first_name")} required />
        </div>

        <div className="field">
          <label>Last name</label>
          <input className="input" value={form.last_name} onChange={on("last_name")} required />
        </div>

        <div className="field">
          <label>Email address</label>
          <input type="email" className="input" value={form.email} onChange={on("email")} required />
        </div>

        <div className="field">
          <label>City</label>
          <input className="input" value={form.city} onChange={on("city")} required />
        </div>

        <div className="field">
          <label>Country</label>
          <input className="input" value={form.country} onChange={on("country")} required />
        </div>

        <div className="field">
          <label>Password</label>
          <input
            type="password"
            className="input"
            value={form.password1}
            onChange={on("password1")}
            required
          />
        </div>

        <div className="field">
          <label>Confirm Password</label>
          <input
            type="password"
            className="input"
            value={form.password2}
            onChange={on("password2")}
            required
          />
        </div>

        <button className="btn btn-primary" disabled={loading}>
          {loading ? "Creating..." : "Sign up"}
        </button>

        {msg && (
          <div
            className={msg.kind === "ok" ? "success" : "error"}
            style={{ marginTop: 8 }}
          >
            {msg.text}
          </div>
        )}

        <p className="helper" style={{ marginTop: 12 }}>
          Already have an account? <Link to="/login">Log in</Link>
        </p>
      </form>
    </div>
  );
}
