import { useEffect, useState } from "react";
import { useParams, useNavigate } from "react-router-dom";
import api from "../lib/api";
import type { AxiosError } from "axios";

type Detail = {
  ID: string; Title: string; Status: "uploaded"|"processed"|"processing"|"failed"|"published";
  OriginalURL?: string; ProcessedURL?: string;
  is_public?: boolean;
};

export default function VideoDetail() {
  const { id } = useParams();
  const nav = useNavigate();
  const [d, setD] = useState<Detail | null>(null);
  const [msg, setMsg] = useState<string | null>(null);

  useEffect(() => {
    api.get<Detail>(`/videos/${id}`)
      .then(r => setD(r.data))
      .catch(() => setMsg("Could not load video details"));
  }, [id]);

  const remove = async () => {
  if (!confirm("Delete video? This action cannot be undone.")) return;
  try {
    await api.delete(`/videos/${id}`);
    nav("/my-videos");
  } catch (err: unknown) { 
    const ax = err as AxiosError<{ message?: string }>;
    setMsg(ax.response?.data?.message ?? "Could not delete video");
  }
};

  if (!d) return <div className="helper">Loading...</div>;

  return (
    <div>
      <h1>{d.Title}</h1>
      <div className="card" style={{padding:16}}>
        {d.ProcessedURL ? (
          <video src={d.ProcessedURL} controls style={{width:"100%", borderRadius:"12px"}} />
        ) : (
          <div className="thumb" style={{width:"100%", height:320, borderRadius:"12px", background:"#f0f0f0", display:"flex", alignItems:"center", justifyContent:"center", color:"#666", fontSize:"48px"}}>
            {d.Status === "processing" ? "⏳" : "🎥"}
          </div>
        )}
        <div style={{display:"flex", gap:10, marginTop:12, flexWrap:"wrap"}}>
          <span className={`badge ${d.Status}`}>{d.Status}</span>
          {/* Borra solo si no está publicado para votación */}
          {d.is_public !== true && (
            <button
              className="btn btn-danger"
              onClick={remove}
            >
              Delete
            </button>
          )}
        </div>
        {msg && <div className="error" style={{marginTop:8}}>{msg}</div>}
      </div>
    </div>
  );
}
