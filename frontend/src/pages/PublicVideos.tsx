import { useEffect, useState } from "react";
import api from "../lib/api";
import { isLoggedIn } from "../lib/auth";
import type { AxiosError } from "axios";

type Pub = { video_id: string; title: string; processed_url?: string; votes?: number };

export default function PublicVideos() {
  const [items, setItems] = useState<Pub[]>([]);
  const [msg, setMsg] = useState<string | null>(null);

  useEffect(() => {
    api.get<Pub[]>("/public/videos")
      .then(r => setItems(r.data))
      .catch(() => setMsg("No se pudo cargar la lista pública"));
  }, []);

  const vote = async (id: string) => {
    try {
      await api.post(`/public/videos/${id}/vote`);
      setMsg("Voto registrado.");
    } catch (err: unknown) {
      const ax = err as AxiosError<{message?:string}>;
      setMsg(ax.response?.data?.message ?? "No se pudo votar"); // si ya votó, backend devuelve 400 con mensaje
    }
  };

  return (
    <div>
      <h1>Explore</h1>
      {msg && <div className="helper" style={{marginBottom:10}}>{msg}</div>}
      <div className="list">
        {items.map(v => (
          <div className="item" key={v.video_id}>
            <div className="thumb" />
            <div>
              <div className="title">{v.title}</div>
              <div className="meta">{(v.votes ?? 0)} votos</div>
            </div>
            <div style={{display:"flex", gap:8}}>
              {v.processed_url && <a className="btn" href={v.processed_url} target="_blank">Ver</a>}
              {isLoggedIn() && <button className="btn btn-primary" onClick={() => vote(v.video_id)}>Votar</button>}
            </div>
          </div>
        ))}
        {items.length===0 && <div className="helper">Aún no hay videos públicos.</div>}
      </div>
    </div>
  );
}
