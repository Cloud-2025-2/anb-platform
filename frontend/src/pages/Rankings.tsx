import { useEffect, useState } from "react";
import api from "../lib/api";

type Row = { position?: number; username?: string; city?: string; votes: number };

export default function Rankings() {
  const [rows, setRows] = useState<Row[]>([]);
  const [cities, setCities] = useState<string[]>([]);
  const [city, setCity] = useState<string>("");
  const [page, setPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);

  const load = async () => {
    const { data } = await api.get<Row[]>("/public/rankings", {
      params: { city: city || undefined, limit: 50 }
    });
    setRows(data);
    setTotalPages(1); // No pagination needed since we get all results
  };

  useEffect(() => { 
    load(); 
    // Load cities list
    api.get<string[]>("/public/cities")
      .then(r => setCities(r.data))
      .catch(() => {}); // Ignore errors, fallback to hardcoded cities
  }, [city, page]);

  return (
    <div>
      <h1>Leaderboard</h1>

      <div style={{display:"flex", gap:12, marginBottom:10}}>
        <select className="select" value={city} onChange={e=>{setPage(1); setCity(e.target.value);}}>
          <option value="">All cities</option>
          {cities.map(c => <option key={c} value={c}>{c}</option>)}
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
              <td>{r.username ?? "User"}{r.city ? ` (${r.city})` : ""}</td>
              <td>{r.votes}</td>
            </tr>
          ))}
          {rows.length===0 && (
            <tr><td colSpan={3} className="helper" style={{padding:20}}>No results.</td></tr>
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
