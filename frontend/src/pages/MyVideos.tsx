import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import api from "../lib/api";

type Item = {
  ID: string;
  Title: string;
  Status: "uploaded" | "processed" | "processing" | "failed" | "published";
  ProcessedURL?: string;
};

export default function MyVideos() {
  const [items, setItems] = useState<Item[]>([]);
  useEffect(() => { api.get<Item[]>("/videos").then(r=>setItems(r.data)); }, []);

  return (
    <div>
      <h1>My Videos</h1>
      <div className="list">
        {items.map(v => (
          <div className="item" key={v.ID}>
            <div className="thumb" />
            <div>
              <div className="title">{v.Title}</div>
              <div className="meta">{v.ProcessedURL ? "Ready to watch" : "Waiting / processing"}</div>
            </div>
            <div style={{display:"grid", gap:8, justifyItems:"end"}}>
              <span className={`badge ${v.Status}`}>{v.Status}</span>
              <div style={{display:"flex", gap:8}}>
                <Link to={`/videos/${v.ID}`} className="btn">Details</Link>
              </div>
            </div>
          </div>
        ))}
        {items.length===0 && <div className="helper">Aún no tienes videos.</div>}
      </div>
    </div>
  );
}
