import React, { useState, useEffect } from 'react';
import { useAuth } from '../contexts/AuthContext';
import { videoAPI } from '../services/api'; // ✅ FIX: Import videoAPI specifically
import './VideoList.css';

const VideoList = () => {
    const [videos, setVideos] = useState([]); // ✅ FIX: Initialize as empty array
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState(null);
    const [filter, setFilter] = useState('all');
    const { user } = useAuth();

    useEffect(() => {
        if (user) {
            fetchVideos();
        } else {
            setLoading(false);
        }
    }, [user]);

    const fetchVideos = async () => {
        try {
            setLoading(true);
            setError(null);
            
            // ✅ FIX: Use the proper videoAPI method
            const result = await videoAPI.getVideos();
            
            if (result.success && Array.isArray(result.videos)) {
                setVideos(result.videos);
            } else {
                setVideos([]);
                if (result.error) {
                    setError(result.error);
                }
            }
        } catch (err) {
            console.error('Error fetching videos:', err);
            setError('Failed to load videos. Please try again.');
            setVideos([]); // ✅ Always ensure it's an array
        } finally {
            setLoading(false);
        }
    };

    // ✅ FIX: Safe filtering with array check
    const filteredVideos = videos.filter(video => {
        if (filter === 'all') return true;
        return video.status === filter;
    });

    const handleVideoClick = (videoId) => {
        window.location.href = `/player/${videoId}`;
    };

    // Rest of the component remains the same...
    const getStatusBadge = (status) => {
        const statusClasses = {
            'processing': 'status-processing',
            'completed': 'status-completed',
            'ready': 'status-ready',
            'failed': 'status-failed'
        };
        return statusClasses[status] || 'status-unknown';
    };

    if (!user) {
        return (
            <div className="video-list-container">
                <div className="auth-required">
                    <h2>Authentication Required</h2>
                    <p>Please log in to view your videos.</p>
                </div>
            </div>
        );
    }

    if (loading) {
        return (
            <div className="video-list-container">
                <div className="loading">
                    <div className="loading-spinner"></div>
                    <p>Loading your videos...</p>
                </div>
            </div>
        );
    }

    return (
        <div className="video-list-container">
            <div className="video-list-header">
                <h1>My Video Library</h1>
                <div className="video-list-controls">
                    <select 
                        value={filter} 
                        onChange={(e) => setFilter(e.target.value)}
                        className="filter-select"
                    >
                        <option value="all">All Videos</option>
                        <option value="processing">Processing</option>
                        <option value="ready">Ready</option>
                        <option value="completed">Completed</option>
                        <option value="failed">Failed</option>
                    </select>
                    <button onClick={fetchVideos} className="refresh-btn">
                        🔄 Refresh
                    </button>
                </div>
            </div>

            {error && (
                <div className="error-message">
                    <p>{error}</p>
                    <button onClick={fetchVideos} className="retry-btn">
                        Try Again
                    </button>
                </div>
            )}

            {filteredVideos.length > 0 ? (
                <div className="video-grid">
                    {filteredVideos.map((video) => (
                        <div key={video.id || video.video_id || video.ID} className="video-card">
                            <div 
                                className="video-thumbnail"
                                onClick={() => handleVideoClick(video.id || video.video_id || video.ID)}
                            >
                                <div className="video-thumbnail-placeholder">
                                    🎬
                                </div>
                                {(video.status === 'ready' || video.status === 'completed') && (
                                    <div className="play-overlay">
                                        <div className="play-button">▶</div>
                                    </div>
                                )}
                            </div>
                            
                            <div className="video-info">
                                <h3 className="video-title">{video.title || 'Untitled Video'}</h3>
                                <p className="video-meta">
                                    <span className="upload-date">
                                        {video.created_at ? new Date(video.created_at).toLocaleDateString() : 'Unknown date'}
                                    </span>
                                    <span className={`status-badge ${getStatusBadge(video.status)}`}>
                                        {video.status || 'Unknown'}
                                    </span>
                                </p>
                                {video.file_size && (
                                    <p className="video-size">
                                        Size: {(video.file_size / (1024 * 1024)).toFixed(1)} MB
                                    </p>
                                )}
                            </div>
                        </div>
                    ))}
                </div>
            ) : (
                <div className="empty-state">
                    <div className="empty-icon">📹</div>
                    <h3>No videos found</h3>
                    <p>
                        {filter === 'all' 
                            ? "You haven't uploaded any videos yet." 
                            : `No videos with status "${filter}" found.`
                        }
                    </p>
                    <button 
                        onClick={() => window.location.href = '/upload'} 
                        className="upload-btn"
                    >
                        Upload Your First Video
                    </button>
                </div>
            )}
        </div>
    );
};

export default VideoList;