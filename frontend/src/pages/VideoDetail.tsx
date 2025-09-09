import { useEffect, useState } from "react";
import { useParams, useNavigate } from "react-router-dom";
import api from "../lib/api";
import type { AxiosError } from "axios";

type Detail = {
  video_id: string; title: string; status: "uploaded"|"processed"|"processing"|"failed";
  original_url?: string; processed_url?: string; votes?: number;
  is_public?: boolean; // si el backend lo expone; si no, omitimos la condición
};

export default function VideoDetail() {
  const { id } = useParams();
  const nav = useNavigate();
  const [d, setD] = useState<Detail | null>(null);
  const [msg, setMsg] = useState<string | null>(null);

  useEffect(() => {
    api.get<Detail>(`/videos/${id}`)
      .then(r => setD(r.data))
      .catch(() => setMsg("No se pudo cargar el detalle"));
  }, [id]);

  const remove = async () => {
  if (!confirm("¿Eliminar video? Esta acción no se puede deshacer.")) return;
  try {
    await api.delete(`/videos/${id}`);
    nav("/my-videos");
  } catch (err: unknown) { // ⬅️ sin any
    const ax = err as AxiosError<{ message?: string }>;
    setMsg(ax.response?.data?.message ?? "No se pudo eliminar");
  }
};

  if (!d) return <div className="helper">Cargando…</div>;

  return (
    <div>
      <h1>{d.title}</h1>
      <div className="card" style={{padding:16}}>
        {d.processed_url ? (
          <video src={d.processed_url} controls style={{width:"100%", borderRadius:"12px"}} />
        ) : (
          <div className="thumb" style={{width:"100%", height:320}} />
        )}
        <div style={{display:"flex", gap:10, marginTop:12, flexWrap:"wrap"}}>
          {d.original_url && <a className="btn" href={d.original_url} target="_blank">Original</a>}
          {d.processed_url && <a className="btn" href={d.processed_url} target="_blank">Download</a>}
          <span className={`badge ${d.status}`}>{d.status}</span>
          {/* Borra solo si no está publicado para votación (si no tienes el flag, deja visible y el backend valida) */}
          {!d.is_public && <button className="btn btn-danger" onClick={remove}>Eliminar</button>}
        </div>
        {msg && <div className="error" style={{marginTop:8}}>{msg}</div>}
      </div>
    </div>
  );
}
