/**
 * HLS Video Player Module
 * Handles adaptive bitrate streaming with HLS.js integration
 */

class HLSPlayer {
    constructor() {
        this.hls = null;
        this.videoElement = null;
        this.currentVideo = null;
        this.qualityLevels = [];
        this.isPlaying = false;
        this.currentQuality = 'auto';
        
        this.bindEvents();
    }

    /**
     * Bind player events
     */
    bindEvents() {
        // Close player modal
        document.getElementById('closePlayer')?.addEventListener('click', () => {
            this.closePlayer();
        });

        // Quality selector
        document.getElementById('qualitySelect')?.addEventListener('change', (e) => {
            this.changeQuality(e.target.value);
        });

        // Modal close on background click
        document.getElementById('playerModal')?.addEventListener('click', (e) => {
            if (e.target.id === 'playerModal') {
                this.closePlayer();
            }
        });

        // Keyboard shortcuts
        document.addEventListener('keydown', (e) => {
            const playerModal = document.getElementById('playerModal');
            if (playerModal?.classList.contains('active')) {
                this.handleKeyboardShortcuts(e);
            }
        });

        // Listen for video play events
        document.addEventListener('playVideo', (e) => {
            this.playVideo(e.detail.video);
        });
    }

    /**
     * Play video with HLS streaming
     */
    async playVideo(video) {
        if (!video) {
            this.showError('Video not found');
            return;
        }

        if (video.status !== 'ready') {
            this.showError('Video is still processing. Please try again later.');
            return;
        }

        this.currentVideo = video;
        this.showPlayerModal();
        this.loadVideo(video);
    }

    /**
     * Show player modal
     */
    showPlayerModal() {
        const modal = document.getElementById('playerModal');
        const title = document.getElementById('playerTitle');
        
        if (modal) {
            modal.classList.add('active');
            document.body.style.overflow = 'hidden'; // Prevent background scroll
        }
        
        if (title && this.currentVideo) {
            title.textContent = this.currentVideo.title || 'Untitled Video';
        }
    }

    /**
     * Close player modal
     */
    closePlayer() {
        const modal = document.getElementById('playerModal');
        
        if (modal) {
            modal.classList.remove('active');
            document.body.style.overflow = ''; // Restore scroll
        }

        // Stop and cleanup video
        this.stopVideo();
        this.currentVideo = null;
    }

    /**
     * Load video into player
     */
    async loadVideo(video) {
        this.videoElement = document.getElementById('hlsPlayer');
        
        if (!this.videoElement) {
            this.showError('Video player not found');
            return;
        }

        try {
            // Show loading state
            this.showLoading();

            // Get HLS playlist URL
            const hlsUrl = await this.getHLSUrl(video);
            
            if (!hlsUrl) {
                throw new Error('HLS playlist not available');
            }

            // Initialize HLS.js or use native HLS
            if (this.isHLSSupported()) {
                await this.initializeHLS(hlsUrl);
            } else if (this.videoElement.canPlayType('application/vnd.apple.mpegurl')) {
                // Native HLS support (Safari)
                this.videoElement.src = hlsUrl;
            } else {
                throw new Error('HLS not supported in this browser');
            }

            // Setup video event listeners
            this.setupVideoEventListeners();

        } catch (error) {
            console.error('Error loading video:', error);
            this.showError(error.message || 'Failed to load video');
        }
    }

    /**
     * Get HLS playlist URL for video
     */
    async getHLSUrl(video) {
        try {
            // Construct HLS URL based on your S3 structure
            // Format: https://s3-bucket-url/{userId}/{videoId}/hls/playlist.m3u8
            const baseUrl = `https://adaptive-bitrate-streaming-videos.s3.amazonaws.com`;
            const hlsUrl = `${baseUrl}/${video.user_id}/${video.video_id}/hls/playlist.m3u8`;
            
            // Verify URL exists
            const response = await fetch(hlsUrl, { method: 'HEAD' });
            if (response.ok) {
                return hlsUrl;
            }
            
            // Fallback: try without user_id structure
            const fallbackUrl = `${baseUrl}/${video.video_id}/hls/playlist.m3u8`;
            const fallbackResponse = await fetch(fallbackUrl, { method: 'HEAD' });
            if (fallbackResponse.ok) {
                return fallbackUrl;
            }
            
            throw new Error('HLS playlist not found');
        } catch (error) {
            console.error('Error getting HLS URL:', error);
            throw error;
        }
    }

    /**
     * Check if HLS.js is supported
     */
    isHLSSupported() {
        return window.Hls && window.Hls.isSupported();
    }

    /**
     * Initialize HLS.js player
     */
    async initializeHLS(hlsUrl) {
        // Cleanup existing HLS instance
        if (this.hls) {
            this.hls.destroy();
        }

        this.hls = new window.Hls({
            debug: false,
            enableWorker: true,
            lowLatencyMode: false,
            backBufferLength: 90
        });

        // Error handling
        this.hls.on(window.Hls.Events.ERROR, (event, data) => {
            console.error('HLS.js error:', data);
            
            if (data.fatal) {
                switch (data.type) {
                    case window.Hls.ErrorTypes.NETWORK_ERROR:
                        this.showError('Network error loading video');
                        break;
                    case window.Hls.ErrorTypes.MEDIA_ERROR:
                        this.showError('Media error playing video');
                        break;
                    default:
                        this.showError('Fatal error playing video');
                        break;
                }
            }
        });

        // Manifest loaded
        this.hls.on(window.Hls.Events.MANIFEST_PARSED, (event, data) => {
            console.log('HLS manifest loaded with', data.levels.length, 'quality levels');
            this.qualityLevels = data.levels;
            this.updateQualitySelector();
            this.hideLoading();
        });

        // Level switched
        this.hls.on(window.Hls.Events.LEVEL_SWITCHED, (event, data) => {
            const level = this.hls.levels[data.level];
            this.updateQualityIndicator(level);
        });

        // Load and attach
        this.hls.loadSource(hlsUrl);
        this.hls.attachMedia(this.videoElement);
    }

    /**
     * Setup video element event listeners
     */
    setupVideoEventListeners() {
        if (!this.videoElement) return;

        // Playing state
        this.videoElement.addEventListener('playing', () => {
            this.isPlaying = true;
            this.hideLoading();
        });

        // Paused state
        this.videoElement.addEventListener('pause', () => {
            this.isPlaying = false;
        });

        // Waiting/buffering
        this.videoElement.addEventListener('waiting', () => {
            this.showBuffering();
        });

        // Can play
        this.videoElement.addEventListener('canplay', () => {
            this.hideBuffering();
        });

        // Error
        this.videoElement.addEventListener('error', (e) => {
            console.error('Video element error:', e);
            this.showError('Error playing video');
        });

        // Ended
        this.videoElement.addEventListener('ended', () => {
            this.isPlaying = false;
            // Could show related videos or replay options here
        });

        // Time update for progress
        this.videoElement.addEventListener('timeupdate', () => {
            // Could update custom progress bar here
        });
    }

    /**
     * Update quality selector dropdown
     */
    updateQualitySelector() {
        const qualitySelect = document.getElementById('qualitySelect');
        if (!qualitySelect || !this.qualityLevels.length) return;

        // Clear existing options except 'Auto'
        const autoOption = qualitySelect.querySelector('option[value="auto"]');
        qualitySelect.innerHTML = '';
        if (autoOption) {
            qualitySelect.appendChild(autoOption);
        }

        // Add quality levels
        this.qualityLevels.forEach((level, index) => {
            const option = document.createElement('option');
            option.value = index.toString();
            option.textContent = `${level.height}p (${Math.round(level.bitrate / 1000)}kbps)`;
            qualitySelect.appendChild(option);
        });
    }

    /**
     * Change video quality
     */
    changeQuality(quality) {
        if (!this.hls) return;

        this.currentQuality = quality;

        if (quality === 'auto') {
            this.hls.currentLevel = -1; // Auto selection
            this.showToast('Quality set to Auto', 'success');
        } else {
            const levelIndex = parseInt(quality);
            if (levelIndex >= 0 && levelIndex < this.qualityLevels.length) {
                this.hls.currentLevel = levelIndex;
                const level = this.qualityLevels[levelIndex];
                this.showToast(`Quality set to ${level.height}p`, 'success');
            }
        }
    }

    /**
     * Update quality indicator
     */
    updateQualityIndicator(level) {
        let indicator = document.querySelector('.quality-indicator');
        
        if (!indicator) {
            indicator = document.createElement('div');
            indicator.className = 'quality-indicator';
            
            const playerContainer = document.querySelector('.player-container');
            if (playerContainer) {
                playerContainer.style.position = 'relative';
                playerContainer.appendChild(indicator);
            }
        }

        if (level) {
            indicator.textContent = `${level.height}p`;
            indicator.classList.remove('hidden');
            
            // Auto-hide after 3 seconds
            setTimeout(() => {
                indicator.classList.add('hidden');
            }, 3000);
        }
    }

    /**
     * Show loading state
     */
    showLoading() {
        const playerContainer = document.querySelector('.player-container');
        if (!playerContainer) return;

        let loadingDiv = playerContainer.querySelector('.player-loading');
        if (!loadingDiv) {
            loadingDiv = document.createElement('div');
            loadingDiv.className = 'player-loading';
            loadingDiv.innerHTML = '<div class="loading-spinner"></div>';
            playerContainer.appendChild(loadingDiv);
        }
        
        loadingDiv.style.display = 'flex';
    }

    /**
     * Hide loading state
     */
    hideLoading() {
        const loadingDiv = document.querySelector('.player-loading');
        if (loadingDiv) {
            loadingDiv.style.display = 'none';
        }
    }

    /**
     * Show buffering indicator
     */
    showBuffering() {
        let bufferingDiv = document.querySelector('.buffering-indicator');
        
        if (!bufferingDiv) {
            bufferingDiv = document.createElement('div');
            bufferingDiv.className = 'buffering-indicator';
            bufferingDiv.innerHTML = `
                <div class="buffering-spinner"></div>
                <span>Buffering...</span>
            `;
            
            const playerContainer = document.querySelector('.player-container');
            if (playerContainer) {
                playerContainer.appendChild(bufferingDiv);
            }
        }
        
        bufferingDiv.classList.add('active');
    }

    /**
     * Hide buffering indicator
     */
    hideBuffering() {
        const bufferingDiv = document.querySelector('.buffering-indicator');
        if (bufferingDiv) {
            bufferingDiv.classList.remove('active');
        }
    }

    /**
     * Show error state
     */
    showError(message) {
        const playerContainer = document.querySelector('.player-container');
        if (!playerContainer) return;

        // Hide loading
        this.hideLoading();

        // Create error display
        let errorDiv = playerContainer.querySelector('.player-error');
        if (!errorDiv) {
            errorDiv = document.createElement('div');
            errorDiv.className = 'player-error';
            playerContainer.appendChild(errorDiv);
        }

        errorDiv.innerHTML = `
            <div class="player-error-icon">⚠️</div>
            <h3>Playback Error</h3>
            <p>${message}</p>
            <button class="retry-btn" onclick="window.hlsPlayer.retryLoad()">
                Retry
            </button>
        `;

        errorDiv.style.display = 'flex';
    }

    /**
     * Retry loading video
     */
    retryLoad() {
        const errorDiv = document.querySelector('.player-error');
        if (errorDiv) {
            errorDiv.style.display = 'none';
        }

        if (this.currentVideo) {
            this.loadVideo(this.currentVideo);
        }
    }

    /**
     * Stop video playback and cleanup
     */
    stopVideo() {
        if (this.videoElement) {
            this.videoElement.pause();
            this.videoElement.src = '';
            this.videoElement.load();
        }

        if (this.hls) {
            this.hls.destroy();
            this.hls = null;
        }

        this.isPlaying = false;
        this.qualityLevels = [];
        
        // Clear indicators
        const indicators = document.querySelectorAll('.quality-indicator, .buffering-indicator, .player-loading, .player-error');
        indicators.forEach(el => el.remove());
    }

    /**
     * Handle keyboard shortcuts
     */
    handleKeyboardShortcuts(e) {
        if (!this.videoElement) return;

        switch (e.code) {
            case 'Space':
                e.preventDefault();
                if (this.isPlaying) {
                    this.videoElement.pause();
                } else {
                    this.videoElement.play();
                }
                break;
                
            case 'Escape':
                e.preventDefault();
                this.closePlayer();
                break;
                
            case 'ArrowLeft':
                e.preventDefault();
                this.videoElement.currentTime = Math.max(0, this.videoElement.currentTime - 10);
                break;
                
            case 'ArrowRight':
                e.preventDefault();
                this.videoElement.currentTime = Math.min(
                    this.videoElement.duration, 
                    this.videoElement.currentTime + 10
                );
                break;
                
            case 'ArrowUp':
                e.preventDefault();
                this.videoElement.volume = Math.min(1, this.videoElement.volume + 0.1);
                break;
                
            case 'ArrowDown':
                e.preventDefault();
                this.videoElement.volume = Math.max(0, this.videoElement.volume - 0.1);
                break;
                
            case 'KeyM':
                e.preventDefault();
                this.videoElement.muted = !this.videoElement.muted;
                break;
                
            case 'KeyF':
                e.preventDefault();
                this.toggleFullscreen();
                break;
        }
    }

    /**
     * Toggle fullscreen mode
     */
    toggleFullscreen() {
        if (!this.videoElement) return;

        if (document.fullscreenElement) {
            document.exitFullscreen();
        } else {
            this.videoElement.requestFullscreen().catch(err => {
                console.error('Error attempting to enable fullscreen:', err);
            });
        }
    }

    /**
     * Get player statistics
     */
    getStats() {
        if (!this.hls || !this.videoElement) {
            return null;
        }

        return {
            currentLevel: this.hls.currentLevel,
            loadLevel: this.hls.loadLevel,
            buffered: this.videoElement.buffered.length > 0 ? 
                this.videoElement.buffered.end(0) - this.videoElement.currentTime : 0,
            currentTime: this.videoElement.currentTime,
            duration: this.videoElement.duration,
            volume: this.videoElement.volume,
            muted: this.videoElement.muted,
            qualityLevels: this.qualityLevels.length
        };
    }

    /**
     * Show toast notification
     */
    showToast(message, type = 'info') {
        if (window.showToast) {
            window.showToast(message, type);
        } else {
            console.log(`[${type.toUpperCase()}] ${message}`);
        }
    }
}

// Create global HLS player instance
window.hlsPlayer = new HLSPlayer();

// Export for module systems
if (typeof module !== 'undefined' && module.exports) {
    module.exports = HLSPlayer;
}