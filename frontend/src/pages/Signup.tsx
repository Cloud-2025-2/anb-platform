import { useState } from "react";
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

function Signup() {
  const [form, setForm] = useState<SignupForm>({
    first_name: "",
    last_name: "",
    email: "",
    password1: "",
    password2: "",
    city: "",
    country: "",
  });

  const [msg, setMsg] = useState<string | null>(null);

  const submit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    setMsg(null);
    try {
      await api.post("/auth/signup", form);
      setMsg("Usuario creado. Ahora inicia sesión.");
    } catch (err: unknown) {
      const ax = err as AxiosError<{ message?: string }>;
      setMsg(ax.response?.data?.message ?? "Error al registrar");
    }
  };

  return (
    <form onSubmit={submit} style={{ display: "grid", gap: 8 }}>
      <h2>Signup</h2>
      <input
        placeholder="Nombre"
        value={form.first_name}
        onChange={(e) => setForm({ ...form, first_name: e.target.value })}
      />
      <input
        placeholder="Apellido"
        value={form.last_name}
        onChange={(e) => setForm({ ...form, last_name: e.target.value })}
      />
      <input
        placeholder="Email"
        type="email"
        value={form.email}
        onChange={(e) => setForm({ ...form, email: e.target.value })}
      />
      <input
        type="password"
        placeholder="Contraseña"
        value={form.password1}
        onChange={(e) => setForm({ ...form, password1: e.target.value })}
      />
      <input
        type="password"
        placeholder="Confirmar contraseña"
        value={form.password2}
        onChange={(e) => setForm({ ...form, password2: e.target.value })}
      />
      <input
        placeholder="Ciudad"
        value={form.city}
        onChange={(e) => setForm({ ...form, city: e.target.value })}
      />
      <input
        placeholder="País"
        value={form.country}
        onChange={(e) => setForm({ ...form, country: e.target.value })}
      />
      <button>Crear cuenta</button>
      {msg && <small>{msg}</small>}
    </form>
  );
}

export default Signup;
