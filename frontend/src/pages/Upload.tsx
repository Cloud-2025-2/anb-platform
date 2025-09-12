import { useState } from "react";
import api from "../lib/api";

const MAX_MB = 100;

export default function Upload() {
  const [title, setTitle] = useState("");
  const [file, setFile] = useState<File | null>(null);
  const [msg, setMsg] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  const onPick = (f: File | null) => {
    if (!f) return setFile(null);
    if (f.type !== "video/mp4") return setMsg("Only MP4 files are allowed");
    if (f.size > MAX_MB * 1024 * 1024) return setMsg("Maximum 100 MB");
    setMsg(null); setFile(f);
  };

  const submit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault(); if (!file) return setMsg("Please select an MP4 file");
    const fd = new FormData();
    fd.append("title", title);
    fd.append("video_file", file);
    try {
      setLoading(true);
      const { data } = await api.post<{message?:string}>("/videos/upload", fd, {
        headers:{ "Content-Type":"multipart/form-data" }
      });
      setMsg(data?.message ?? "Uploaded. Processing in progress.");
      setTitle(""); setFile(null);
    } catch { setMsg("Upload error"); }
    finally { setLoading(false); }
  };

  return (
    <div className="card" style={{padding:24}}>
      <h1>Upload new video</h1>
      <form onSubmit={submit} className="form" style={{maxWidth:520}}>
        <div className="field">
          <label>Title</label>
          <input className="input" value={title}
            onChange={e=>setTitle(e.target.value)} required />
        </div>
        <div className="field">
          <label>MP4 file (≤ 100 MB)</label>
          <input type="file" className="input" accept="video/mp4"
            onChange={e=>onPick(e.target.files?.[0] ?? null)} required />
        </div>
        <button className="btn btn-primary" disabled={loading}>
          {loading ? "Uploading..." : "Upload"}
        </button>
        {msg && <div className="helper" style={{marginTop:6}}>{msg}</div>}
      </form>
    </div>
  );
}
