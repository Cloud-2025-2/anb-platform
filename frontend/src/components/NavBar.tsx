import { useState, useEffect } from "react";
import { Link, NavLink } from "react-router-dom";
import { isLoggedIn, clearToken } from "../lib/auth";

export default function NavBar() {
  const [loggedIn, setLoggedIn] = useState(isLoggedIn());

  useEffect(() => {
    // Listen for storage changes to update navbar when auth state changes
    const handleStorageChange = () => {
      setLoggedIn(isLoggedIn());
    };

    window.addEventListener('storage', handleStorageChange);
    
    // Also check periodically in case of same-tab changes
    const interval = setInterval(() => {
      setLoggedIn(isLoggedIn());
    }, 1000);

    return () => {
      window.removeEventListener('storage', handleStorageChange);
      clearInterval(interval);
    };
  }, []);

  return (
    <nav className="nav">
      <div className="brand">
        <span className="logo" /> VideoContest
      </div>

      <div className="navlinks">
        <NavLink to="/" className={({isActive}) => isActive ? "active" : ""}>Home</NavLink>
        <NavLink to="/rankings" className={({isActive}) => isActive ? "active" : ""}>Leaderboard</NavLink>
        {loggedIn && (
          <>
            <NavLink to="/upload" className={({isActive}) => isActive ? "active" : ""}>Upload</NavLink>
            <NavLink to="/my-videos" className={({isActive}) => isActive ? "active" : ""}>My Videos</NavLink>
          </>
        )}
      </div>

      <div className="right">
        {!loggedIn ? (
          <>
            <Link to="/login" className="btn">Log in</Link>
            <Link to="/signup" className="btn btn-primary">Sign up</Link>
          </>
        ) : (
          <button className="btn" onClick={() => { clearToken(); setLoggedIn(false); location.href="/login"; }}>
            Log out
          </button>
        )}
      </div>
    </nav>
  );
}
