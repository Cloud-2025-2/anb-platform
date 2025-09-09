import { useEffect, useState } from "react";
import api from "../lib/api";
import { useParams, useNavigate } from "react-router-dom";

type Detail = {
  video_id: string; title: string; status: string;
  original_url?: string; processed_url?: string; votes?: number;
};

function VideoDetail() {
  const { id } = useParams();
  const nav = useNavigate();
  const [d, setD] = useState<Detail | null>(null);
  const [msg, setMsg] = useState<string | null>(null);

  useEffect(() => {
    api.get(`/videos/${id}`)
      .then(r => setD(r.data))
      .catch(() => setMsg("No se pudo cargar el detalle"));
  }, [id]);

  const remove = async () => {
    if (!confirm("¿Eliminar video?")) return;
    try {
      await api.delete(`/videos/${id}`);
      nav("/my-videos");
    } catch {
      setMsg("No se pudo eliminar");
    }
  };

  if (!d) return <div>Cargando...</div>;
  return (
    <div>
      <h2>{d.title}</h2>
      <p>Estado: {d.status}</p>
      {d.original_url && <p><a href={d.original_url} target="_blank">Original</a></p>}
      {d.processed_url && <p><a href={d.processed_url} target="_blank">Procesado</a></p>}
      <p>Votos: {d.votes ?? 0}</p>
      <button onClick={remove}>Eliminar</button>
      <small>{msg}</small>
    </div>
  );
}

export default VideoDetail;
