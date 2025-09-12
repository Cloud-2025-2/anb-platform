import { useEffect, useState } from "react";
import api from "../lib/api";
import { isLoggedIn } from "../lib/auth";
import type { AxiosError } from "axios";

type Pub = { 
  ID: string; 
  Title: string; 
  ProcessedURL?: string; 
  User: { FirstName: string; LastName: string; };
  Votes?: Array<any>; 
};

export default function PublicVideos() {
  const [items, setItems] = useState<Pub[]>([]);
  const [msg, setMsg] = useState<string | null>(null);

  useEffect(() => {
    api.get<Pub[]>("/public/videos")
      .then(r => setItems(r.data))
      .catch(() => setMsg("Could not load public videos"));
  }, []);

  const vote = async (id: string) => {
    try {
      await api.post(`/public/videos/${id}/vote`);
      setMsg("Vote registered successfully.");
      // Reload the videos to get updated vote counts
      const { data } = await api.get<Pub[]>("/public/videos");
      setItems(data);
    } catch (err: unknown) {
      const ax = err as AxiosError<{message?:string}>;
      setMsg(ax.response?.data?.message ?? "Could not vote"); // si ya votó, backend devuelve 400 con mensaje
    }
  };

  return (
    <div>
      <h1>Explore</h1>
      {msg && <div className="helper" style={{marginBottom:10}}>{msg}</div>}
      <div className="list">
        {items.map(v => (
          <div className="item" key={v.ID}>
            <div className="thumb">
              <div style={{background: '#f0f0f0', width: '100%', height: '100%', display: 'flex', alignItems: 'center', justifyContent: 'center', color: '#666', fontSize: '24px'}}>
                🎥
              </div>
            </div>
            <div>
              <div className="title">{v.Title}</div>
              <div className="meta">by {v.User.FirstName} {v.User.LastName} • {(v.Votes?.length ?? 0)} votes</div>
            </div>
            <div style={{display:"flex", gap:8}}>
              {v.ProcessedURL && <a className="btn" href={v.ProcessedURL} target="_blank">Watch</a>}
              {isLoggedIn() && <button className="btn btn-primary" onClick={() => vote(v.ID)}>Vote</button>}
            </div>
          </div>
        ))}
        {items.length===0 && <div className="helper">No public videos yet.</div>}
      </div>
    </div>
  );
}
