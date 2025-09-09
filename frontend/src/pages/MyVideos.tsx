import { useEffect, useState } from "react";
import api from "../lib/api";
import { Link } from "react-router-dom";

type Item = {
  video_id: string; title: string; status: "uploaded"|"processed";
  processed_url?: string;
};

function MyVideos() {
  const [items, setItems] = useState<Item[]>([]);
  const [err, setErr] = useState<string | null>(null);

  useEffect(() => {
    api.get("/videos")
      .then(r => setItems(r.data))
      .catch(e => setErr(e?.response?.data?.message || "Error al cargar"));
  }, []);

  return (
    <div>
      <h2>Mis videos</h2>
      {err && <small>{err}</small>}
      <ul>
        {items.map(v => (
          <li key={v.video_id}>
            <strong>{v.title}</strong> — {v.status}
            {" "}<Link to={`/videos/${v.video_id}`}>detalle</Link>
            {" "}{v.processed_url && <a href={v.processed_url} target="_blank">descargar</a>}
          </li>
        ))}
      </ul>
    </div>
  );
}

export default MyVideos;
