import React, { useState, useEffect, useRef, useCallback } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import Hls from 'hls.js';
import { videoAPI } from '../services/api';
import './VideoPlayer.css';

const VideoPlayer = () => {
  const { videoId } = useParams();
  const navigate = useNavigate();
  const videoRef = useRef(null);
  const hlsRef = useRef(null);
  
  const [video, setVideo] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [playerError, setPlayerError] = useState('');

  useEffect(() => {
    fetchVideo();
  }, [videoId]);

  useEffect(() => {
    return () => {
      // Cleanup HLS instance
      if (hlsRef.current) {
        hlsRef.current.destroy();
      }
    };
  }, []);

  const fetchVideo = async () => {
    try {
      const response = await videoAPI.getVideo(videoId);
      const videoData = response.data;
      
      if (videoData.status !== 'ready') {
        setError('Video is not ready for streaming yet. Please try again later.');
        setLoading(false);
        return;
      }
      
      setVideo(videoData);
      setLoading(false);
      
      // Initialize HLS player after video data is loaded
      setTimeout(() => initializePlayer(videoData), 100);
      
    } catch (err) {
      setError(err.response?.data?.error || 'Failed to load video');
      setLoading(false);
    }
  };

  const initializePlayer = (videoData) => {
    const video = videoRef.current;
    if (!video || !videoData.hls_url) return;

    setPlayerError('');

    if (Hls.isSupported()) {
      const hls = new Hls({
        enableWorker: true,
        lowLatencyMode: true,
        backBufferLength: 90
      });

      hls.loadSource(videoData.hls_url);
      hls.attachMedia(video);

      hls.on(Hls.Events.MANIFEST_PARSED, () => {
        console.log('HLS manifest parsed, starting playback');
      });

      hls.on(Hls.Events.ERROR, (event, data) => {
        console.error('HLS error:', data);
        if (data.fatal) {
          switch (data.type) {
            case Hls.ErrorTypes.NETWORK_ERROR:
              setPlayerError('Network error occurred. Please check your connection.');
              hls.startLoad();
              break;
            case Hls.ErrorTypes.MEDIA_ERROR:
              setPlayerError('Media error occurred. Trying to recover...');
              hls.recoverMediaError();
              break;
            default:
              setPlayerError('Fatal error occurred. Please refresh the page.');
              hls.destroy();
              break;
          }
        }
      });

      hlsRef.current = hls;

    } else if (video.canPlayType('application/vnd.apple.mpegurl')) {
      // Native HLS support (Safari)
      video.src = videoData.hls_url;
    } else {
      setPlayerError('HLS is not supported in this browser.');
    }
  };

  const formatDate = (dateString) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    });
  };

  const formatFileSize = (bytes) => {
    if (!bytes) return 'Unknown';
    const mb = bytes / (1024 * 1024);
    return `${mb.toFixed(1)} MB`;
  };

  if (loading) {
    return (
      <div className="player-container">
        <div className="loading">Loading video...</div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="player-container">
        <div className="error-message">
          {error}
          <button onClick={() => navigate('/videos')} className="back-button">
            Back to Videos
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="player-container">
      <div className="player-wrapper">
        <div className="video-container">
          {playerError && (
            <div className="player-error">
              {playerError}
            </div>
          )}
          
          <video
            ref={videoRef}
            controls
            className="video-player"
            poster={video.thumbnail_url}
          >
            Your browser does not support the video tag.
          </video>
        </div>

        <div className="video-details">
          <h1 className="video-title">{video.title}</h1>
          
          {video.description && (
            <div className="video-description">
              <p>{video.description}</p>
            </div>
          )}

          <div className="video-metadata">
            <div className="metadata-item">
              <strong>Uploaded:</strong> {formatDate(video.created_at)}
            </div>
            <div className="metadata-item">
              <strong>File Size:</strong> {formatFileSize(video.file_size)}
            </div>
            <div className="metadata-item">
              <strong>Status:</strong> 
              <span className="status-ready">Ready for Streaming</span>
            </div>
          </div>

          <div className="player-actions">
            <button onClick={() => navigate('/videos')} className="back-button">
              ← Back to Videos
            </button>
          </div>
        </div>
      </div>
    </div>
  );
};

export default VideoPlayer;