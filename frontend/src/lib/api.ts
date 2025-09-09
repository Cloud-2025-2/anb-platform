import axios from "axios";

const api = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || "http://localhost:8000/api",
});

api.interceptors.request.use((config) => {
  const tok = localStorage.getItem("anb_access_token");
  if (tok) config.headers.Authorization = `Bearer ${tok}`;
  return config;
});

export default api;
