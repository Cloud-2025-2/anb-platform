import type { ReactElement } from "react";
import { Navigate } from "react-router-dom";
import { isLoggedIn } from "../lib/auth";

type Props = { element: ReactElement };

export default function PrivateRoute({ element }: Props) {
  return isLoggedIn() ? element : <Navigate to="/login" replace />;
}
