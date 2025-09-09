import { Link, NavLink } from "react-router-dom";
import { isLoggedIn, clearToken } from "../lib/auth";

export default function NavBar() {
  return (
    <nav className="nav">
      <div className="brand">
        <span className="logo" /> VideoContest
      </div>

      <div className="navlinks">
        <NavLink to="/" className={({isActive}) => isActive ? "active" : ""}>Home</NavLink>
        <NavLink to="/rankings" className={({isActive}) => isActive ? "active" : ""}>Leaderboard</NavLink>
        {isLoggedIn() && (
          <>
            <NavLink to="/upload" className={({isActive}) => isActive ? "active" : ""}>Upload</NavLink>
            <NavLink to="/my-videos" className={({isActive}) => isActive ? "active" : ""}>My Videos</NavLink>
          </>
        )}
      </div>

      <div className="right">
        {!isLoggedIn() ? (
          <>
            <Link to="/login" className="btn">Log in</Link>
            <Link to="/signup" className="btn btn-primary">Sign up</Link>
          </>
        ) : (
          <button className="btn" onClick={() => { clearToken(); location.href="/login"; }}>
            Log out
          </button>
        )}
      </div>
    </nav>
  );
}
