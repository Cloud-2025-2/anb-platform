import type { ReactElement } from "react";
import { Navigate } from "react-router-dom";
import { isLoggedIn } from "../lib/auth";

export default function PrivateRoute({ element }: { element: ReactElement }) {
  return isLoggedIn() ? element : <Navigate to="/login" replace />;
}