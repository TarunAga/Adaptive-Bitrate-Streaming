import React, { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';
import { authAPI } from '../services/api';
import './Auth.css';

const Login = () => {
  const [formData, setFormData] = useState({
    email: '',
    password: ''
  });
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  
  const { login } = useAuth();
  const navigate = useNavigate();

  const handleChange = (e) => {
    setFormData({
      ...formData,
      [e.target.name]: e.target.value
    });
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError('');
    setLoading(true);

    try {
            console.log('🔄 Attempting login with:', { email: formData.email });
            
            const response = await authAPI.login(formData);
            console.log('✅ Login response received:', response);

            // ✅ FIX: Handle the correct response format from your backend
            if (response.success && response.token && response.user) {
                // Store token and user data
                localStorage.setItem('authToken', response.token);
                localStorage.setItem('user', JSON.stringify(response.user));
                
                // Update auth context
                login(response.user, response.token);
                
                console.log('✅ Login successful, navigating to videos page');
                
                // Navigate to videos page
                navigate('/videos');
            } else {
                // Handle case where success is false or missing data
                console.error('❌ Invalid login response format:', response);
                setError(response.message || 'Login failed - invalid response format');
            }
        } catch (error) {
            console.error('❌ Login error:', error);
            
            // Handle different error formats
            if (error.message) {
                setError(error.message);
            } else if (error.error) {
                setError(error.error);
            } else if (typeof error === 'string') {
                setError(error);
            } else {
                setError('Login failed. Please check your credentials and try again.');
            }
        } finally {
            setLoading(false);
        }
  };

  return (
    <div className="auth-container">
      <form className="auth-form" onSubmit={handleSubmit}>
        <h2>Login</h2>
        
        {error && <div className="error-message">{error}</div>}
        
        <div className="form-group">
          <label htmlFor="email">Email</label>
          <input
            type="email"
            id="email"
            name="email"
            value={formData.email}
            onChange={handleChange}
            required
          />
        </div>
        
        <div className="form-group">
          <label htmlFor="password">Password</label>
          <input
            type="password"
            id="password"
            name="password"
            value={formData.password}
            onChange={handleChange}
            required
          />
        </div>
        
        <button type="submit" disabled={loading} className="auth-button">
          {loading ? 'Logging in...' : 'Login'}
        </button>
        
        <p className="auth-link">
          Don't have an account? <Link to="/signup">Sign up here</Link>
        </p>
      </form>
    </div>
  );
};

export default Login;