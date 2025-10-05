import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { videoAPI } from '../services/api';
import './VideoUpload.css';

const VideoUpload = () => {
  const [formData, setFormData] = useState({
    title: '',
    description: '',
    file: null
  });
  const [uploading, setUploading] = useState(false);
  const [uploadProgress, setUploadProgress] = useState(0);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');
  
  const navigate = useNavigate();

  const handleChange = (e) => {
    if (e.target.name === 'file') {
      setFormData({
        ...formData,
        file: e.target.files[0]
      });
    } else {
      setFormData({
        ...formData,
        [e.target.name]: e.target.value
      });
    }
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError('');
    setSuccess('');

    if (!formData.file) {
      setError('Please select a video file');
      return;
    }

    const allowedTypes = ['video/mp4', 'video/avi', 'video/mov', 'video/wmv'];
    if (!allowedTypes.includes(formData.file.type)) {
      setError('Please select a valid video file (MP4, AVI, MOV, WMV)');
      return;
    }

    setUploading(true);

    try {
      const uploadData = new FormData();
      uploadData.append('title', formData.title);
      uploadData.append('description', formData.description);
      uploadData.append('video', formData.file);

      await videoAPI.upload(uploadData);
      
      setSuccess('Video uploaded successfully! It will be processed for streaming.');
      setFormData({ title: '', description: '', file: null });
      
      // Reset file input
      document.getElementById('file').value = '';
      
      // Redirect to videos page after 2 seconds
      setTimeout(() => {
        navigate('/videos');
      }, 2000);
      
    } catch (err) {
      setError(err.response?.data?.error || 'Upload failed');
    } finally {
      setUploading(false);
      setUploadProgress(0);
    }
  };

  return (
    <div className="upload-container">
      <div className="upload-form-container">
        <h2>Upload Video</h2>
        
        {error && <div className="error-message">{error}</div>}
        {success && <div className="success-message">{success}</div>}
        
        <form className="upload-form" onSubmit={handleSubmit}>
          <div className="form-group">
            <label htmlFor="title">Video Title</label>
            <input
              type="text"
              id="title"
              name="title"
              value={formData.title}
              onChange={handleChange}
              required
              disabled={uploading}
            />
          </div>
          
          <div className="form-group">
            <label htmlFor="description">Description</label>
            <textarea
              id="description"
              name="description"
              value={formData.description}
              onChange={handleChange}
              rows="4"
              disabled={uploading}
            />
          </div>
          
          <div className="form-group">
            <label htmlFor="file">Video File</label>
            <input
              type="file"
              id="file"
              name="file"
              onChange={handleChange}
              accept="video/*"
              required
              disabled={uploading}
            />
            <small className="file-info">
              Supported formats: MP4, AVI, MOV, WMV (Max size: 500MB)
            </small>
          </div>
          
          {uploading && uploadProgress > 0 && (
            <div className="progress-container">
              <div className="progress-bar">
                <div 
                  className="progress-fill" 
                  style={{ width: `${uploadProgress}%` }}
                ></div>
              </div>
              <span className="progress-text">{uploadProgress}%</span>
            </div>
          )}
          
          <button type="submit" disabled={uploading} className="upload-button">
            {uploading ? 'Uploading...' : 'Upload Video'}
          </button>
        </form>
      </div>
    </div>
  );
};

export default VideoUpload;