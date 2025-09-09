import { useEffect, useState } from "react";
import api from "../lib/api";
import { isLoggedIn } from "../lib/auth";
import type { AxiosError } from "axios";

type Pub = {
  video_id: string;
  title: string;
  processed_url?: string;
  votes?: number;
};

function PublicVideos() {
  const [items, setItems] = useState<Pub[]>([]);
  const [msg, setMsg] = useState<string | null>(null);

  useEffect(() => {
    api
      .get<Pub[]>("/public/videos")
      .then((r) => setItems(r.data))
      .catch((err: unknown) => {
        const ax = err as AxiosError<{ message?: string }>;
        setMsg(ax.response?.data?.message ?? "No se pudo cargar la lista pública");
      });
  }, []);

  const vote = async (id: string) => {
    try {
      await api.post(`/public/videos/${id}/vote`);
      setMsg("Voto registrado.");
    } catch (err: unknown) {
      const ax = err as AxiosError<{ message?: string }>;
      setMsg(ax.response?.data?.message ?? "No se pudo votar");
    }
  };

  return (
    <div>
      <h2>Videos públicos</h2>
      {msg && <small>{msg}</small>}
      <ul>
        {items.map((v) => (
          <li key={v.video_id}>
            <strong>{v.title}</strong>
            {v.processed_url && (
              <>
                {" "}
                —{" "}
                <a href={v.processed_url} target="_blank" rel="noreferrer">
                  ver
                </a>
              </>
            )}{" "}
            ({v.votes ?? 0} votos){" "}
            {isLoggedIn() && <button onClick={() => vote(v.video_id)}>Votar</button>}
          </li>
        ))}
      </ul>
    </div>
  );
}

export default PublicVideos;
