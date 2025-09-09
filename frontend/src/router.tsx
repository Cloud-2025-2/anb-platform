import { createBrowserRouter } from "react-router-dom";
import App from "./App";
import Login from "./pages/Login";
import Signup from "./pages/Signup";
import Upload from "./pages/Upload";
import MyVideos from "./pages/MyVideos";
import VideoDetail from "./pages/VideoDetail";
import PublicVideos from "./pages/PublicVideos";
import Rankings from "./pages/Rankings";
import PrivateRoute from "./components/PrivateRoute";

export const router = createBrowserRouter([
  {
    path: "/",
    element: <App />,
    children: [
      { index: true, element: <PublicVideos /> },
      { path: "rankings", element: <Rankings /> },
      { path: "login", element: <Login /> },
      { path: "signup", element: <Signup /> },
      { path: "upload", element: <PrivateRoute element={<Upload />} /> },
      { path: "my-videos", element: <PrivateRoute element={<MyVideos />} /> },
      { path: "videos/:id", element: <PrivateRoute element={<VideoDetail />} /> },
    ],
  },
]);
