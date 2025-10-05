import React from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';
import './Navbar.css';

const Navbar = () => {
  const { user, logout } = useAuth();
  const navigate = useNavigate();

  const handleLogout = () => {
    logout();
    navigate('/login');
  };

  return (
    <nav className="navbar">
      <div className="navbar-container">
        <Link to="/" className="navbar-brand">
          Adaptive Streaming
        </Link>
        
        <div className="navbar-menu">
          {user ? (
            <>
              <Link to="/videos" className="navbar-item">
                Videos
              </Link>
              <Link to="/upload" className="navbar-item">
                Upload
              </Link>
              <span className="navbar-user">
                Welcome, {user.username}
              </span>
              <button onClick={handleLogout} className="navbar-logout">
                Logout
              </button>
            </>
          ) : (
            <>
              <Link to="/login" className="navbar-item">
                Login
              </Link>
              <Link to="/signup" className="navbar-item">
                Sign Up
              </Link>
            </>
          )}
        </div>
      </div>
    </nav>
  );
};

export default Navbar;