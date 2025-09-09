import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import api from "../lib/api";

type Item = {
  video_id: string; title: string;
  status: "uploaded"|"processed"|"processing"|"failed";
  processed_url?: string;
};

export default function MyVideos() {
  const [items, setItems] = useState<Item[]>([]);
  useEffect(() => { api.get<Item[]>("/videos").then(r=>setItems(r.data)); }, []);

  return (
    <div>
      <h1>My Videos</h1>
      <div className="list">
        {items.map(v => (
          <div className="item" key={v.video_id}>
            <div className="thumb" />
            <div>
              <div className="title">{v.title}</div>
              <div className="meta">{v.processed_url ? "Ready to watch" : "Waiting / processing"}</div>
            </div>
            <div style={{display:"grid", gap:8, justifyItems:"end"}}>
              <span className={`badge ${v.status}`}>{v.status}</span>
              <div style={{display:"flex", gap:8}}>
                <Link to={`/videos/${v.video_id}`} className="btn">Details</Link>
                {v.processed_url && <a className="btn" href={v.processed_url} target="_blank">Download</a>}
              </div>
            </div>
          </div>
        ))}
        {items.length===0 && <div className="helper">Aún no tienes videos.</div>}
      </div>
    </div>
  );
}
