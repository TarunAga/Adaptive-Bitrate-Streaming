import React from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { AuthProvider, useAuth } from './contexts/AuthContext';
import Navbar from './components/Navbar';
import Login from './pages/Login';
import Signup from './pages/Signup';
import VideoUpload from './pages/VideoUpload';
import VideoList from './pages/VideoList';
import VideoPlayer from './pages/VideoPlayer';
import './App.css';

// Protected Route Component
const ProtectedRoute = ({ children }) => {
  const { token } = useAuth();
  return token ? children : <Navigate to="/login" />;
};

// Public Route Component (redirect to videos if already authenticated)
const PublicRoute = ({ children }) => {
  const { token } = useAuth();
  return !token ? children : <Navigate to="/videos" />;
};

function App() {
  return (
    <AuthProvider>
      <Router>
        <div className="App">
          <Navbar />
          <main className="main-content">
            <Routes>
              <Route path="/" element={<Navigate to="/videos" />} />
              <Route 
                path="/login" 
                element={
                  <PublicRoute>
                    <Login />
                  </PublicRoute>
                } 
              />
              <Route 
                path="/signup" 
                element={
                  <PublicRoute>
                    <Signup />
                  </PublicRoute>
                } 
              />
              <Route 
                path="/upload" 
                element={
                  <ProtectedRoute>
                    <VideoUpload />
                  </ProtectedRoute>
                } 
              />
              <Route 
                path="/videos" 
                element={
                  <ProtectedRoute>
                    <VideoList />
                  </ProtectedRoute>
                } 
              />
              <Route 
                path="/player/:videoId" 
                element={
                  <ProtectedRoute>
                    <VideoPlayer />
                  </ProtectedRoute>
                } 
              />
            </Routes>
          </main>
        </div>
      </Router>
    </AuthProvider>
  );
}

export default App;
