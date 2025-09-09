import { useEffect, useState } from "react";
import type { AxiosError } from "axios";
import api from "../lib/api";

type RankRow = {
  position?: number;     // si el backend ya lo manda
  username?: string;     // o nombre del usuario
  city?: string;
  votes: number;
};

function Rankings() {
  const [rows, setRows] = useState<RankRow[]>([]);
  const [err, setErr] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  async function load() {
    setLoading(true);
    setErr(null);
    try {
      const { data } = await api.get<RankRow[]>("/public/rankings");
      setRows(Array.isArray(data) ? data : []);
    } catch (e: unknown) {
      const ax = e as AxiosError<{ message?: string }>;
      setErr(ax.response?.data?.message ?? "No se pudo cargar el ranking");
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => { load(); }, []);

  return (
    <div>
      <h2>Ranking</h2>
      {loading && <small>Cargando…</small>}
      {err && <small>{err}</small>}
      {!loading && !err && (
        <ol>
          {rows.map((r, i) => (
            <li key={i}>
              #{r.position ?? i + 1}{" "}
              {r.username ?? "Usuario"}{r.city ? ` (${r.city})` : ""} — {r.votes} votos
            </li>
          ))}
          {rows.length === 0 && <li>Sin datos aún.</li>}
        </ol>
      )}
      <button onClick={load} style={{ marginTop: 8 }}>Actualizar</button>
    </div>
  );
}

export default Rankings;
