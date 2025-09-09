// src/pages/Upload.tsx
import { useState } from "react";
import api from "../lib/api";
import type { AxiosError } from "axios";

type UploadResp = { message?: string };

function Upload() {
  const [title, setTitle] = useState("");
  const [file, setFile] = useState<File | null>(null);
  const [msg, setMsg] = useState<string | null>(null);

  const submit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    if (!file) {
      setMsg("Selecciona un MP4");
      return;
    }
    const fd = new FormData();
    fd.append("title", title);
    fd.append("video_file", file);

    try {
      const { data } = await api.post<UploadResp>("/videos/upload", fd, {
        headers: { "Content-Type": "multipart/form-data" },
      });
      setMsg(data?.message ?? "Subido. Procesamiento en curso.");
      setTitle("");
      setFile(null);
    } catch (err: unknown) {
      const ax = err as AxiosError<{ message?: string }>;
      setMsg(ax.response?.data?.message ?? "Error al subir");
    }
  };

  return (
    <form onSubmit={submit} style={{ display: "grid", gap: 8 }}>
      <h2>Subir video</h2>
      <input
        placeholder="Título"
        value={title}
        onChange={(e) => setTitle(e.target.value)}
        required
      />
      <input
        type="file"
        accept="video/mp4"
        onChange={(e) => setFile(e.target.files?.[0] ?? null)}
        required
      />
      <button>Subir</button>
      {msg && <small>{msg}</small>}
    </form>
  );
}

export default Upload;
