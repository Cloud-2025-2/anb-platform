import { useEffect, useState } from "react";
import api from "../lib/api";

type Row = { position?: number; username?: string; city?: string; votes: number };
type RankResp = Row[] | { data: Row[]; total_pages?: number };

export default function Rankings() {
  const [rows, setRows] = useState<Row[]>([]);
  const [city, setCity] = useState<string>("");
  const [page, setPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);

  const load = async () => {
    const { data } = await api.get<RankResp>("/public/rankings", {
      params: { city: city || undefined, page, limit: 10 }
    });

    if (Array.isArray(data)) {
      setRows(data);
      setTotalPages(1);
    } else {
      setRows(data.data ?? []);
      setTotalPages(data.total_pages ?? 1);
    }
  };

  useEffect(() => { load(); }, [city, page]);

  useEffect(() => { load(); }, [city, page]);

  return (
    <div>
      <h1>Leaderboard</h1>

      <div style={{display:"flex", gap:12, marginBottom:10}}>
        <select className="select" value={city} onChange={e=>{setPage(1); setCity(e.target.value);}}>
          <option value="">All cities</option>
          {/* si quieres una lista dinámica de ciudades, podrías cargarla de /public/rankings?cities=1 */}
          <option>Bogotá</option><option>Medellín</option><option>Cali</option>
        </select>
      </div>

      <table className="table">
        <thead>
          <tr><th style={{width:80}}>Rank</th><th>User</th><th style={{width:160}}>Votes</th></tr>
        </thead>
        <tbody>
          {rows.map((r, i) => (
            <tr key={i}>
              <td>#{r.position ?? (i + 1 + (page-1)*10)}</td>
              <td>{r.username ?? "Usuario"}{r.city ? ` (${r.city})` : ""}</td>
              <td>{r.votes}</td>
            </tr>
          ))}
          {rows.length===0 && (
            <tr><td colSpan={3} className="helper" style={{padding:20}}>Sin resultados.</td></tr>
          )}
        </tbody>
      </table>

      <div className="pager">
        <button className="btn" onClick={()=>setPage(p=>Math.max(1,p-1))} disabled={page<=1}>‹</button>
        <span className="helper">Page {page} / {totalPages}</span>
        <button className="btn" onClick={()=>setPage(p=>Math.min(totalPages,p+1))} disabled={page>=totalPages}>›</button>
      </div>
    </div>
  );
}
