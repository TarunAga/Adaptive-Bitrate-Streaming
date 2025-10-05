/**
 * Main Application Module
 * Coordinates all frontend functionality including navigation, video library, and notifications
 */

class StreamingApp {
    constructor() {
        this.apiBase = 'http://localhost:8081/api/v1';
        this.currentSection = 'home';
        this.videos = [];
        this.isLoading = false;
        
        this.init();
    }

    /**
     * Initialize the application
     */
    async init() {
        try {
            // Wait for DOM to be ready
            if (document.readyState === 'loading') {
                document.addEventListener('DOMContentLoaded', () => this.setup());
            } else {
                this.setup();
            }
        } catch (error) {
            console.error('App initialization error:', error);
        }
    }

    /**
     * Setup application after DOM is ready
     */
    setup() {
        this.bindEvents();
        this.updateAuthUI();
        this.loadVideos();
        this.testAPIConnection();
        
        // Show initial section
        this.showSection(this.currentSection);
    }

    /**
     * Bind application events
     */
    bindEvents() {
        // Navigation buttons
        const navButtons = document.querySelectorAll('.nav-btn[data-section]');
        navButtons.forEach(btn => {
            btn.addEventListener('click', (e) => {
                const section = e.target.getAttribute('data-section');
                this.showSection(section);
            });
        });

        // Auth state changes
        document.addEventListener('authStateChange', (e) => {
            this.handleAuthStateChange(e.detail);
        });

        // Upload success
        document.addEventListener('uploadSuccess', (e) => {
            this.handleUploadSuccess(e.detail);
        });

        // Video library filters
        document.getElementById('statusFilter')?.addEventListener('change', (e) => {
            this.filterVideos(e.target.value);
        });

        // Window events
        window.addEventListener('resize', this.handleResize.bind(this));
        window.addEventListener('beforeunload', this.handleBeforeUnload.bind(this));
    }

    /**
     * Show specific section
     */
    showSection(sectionName) {
        // Update current section
        this.currentSection = sectionName;
        
        // Hide all sections
        const sections = document.querySelectorAll('.content-section');
        sections.forEach(section => {
            section.classList.remove('active');
        });
        
        // Show target section
        const targetSection = document.getElementById(`${sectionName}Section`);
        if (targetSection) {
            targetSection.classList.add('active');
        }
        
        // Update navigation buttons
        const navButtons = document.querySelectorAll('.nav-btn[data-section]');
        navButtons.forEach(btn => {
            const btnSection = btn.getAttribute('data-section');
            btn.classList.toggle('active', btnSection === sectionName);
        });
        
        // Section-specific actions
        switch (sectionName) {
            case 'library':
                this.loadVideos();
                break;
            case 'upload':
                if (!window.authManager?.isAuthenticated()) {
                    window.authManager?.showAuthModal('login');
                    return;
                }
                break;
        }
    }

    /**
     * Handle authentication state changes
     */
    handleAuthStateChange(detail) {
        this.updateAuthUI();
        
        if (detail.type === 'login' || detail.type === 'signup') {
            // Redirect to upload if they were trying to access it
            if (this.currentSection === 'upload') {
                this.showSection('upload');
            }
            // Load user's videos
            this.loadVideos();
        } else if (detail.type === 'logout') {
            // Clear videos and redirect to home
            this.videos = [];
            this.renderVideoGrid();
            this.showSection('home');
        }
    }

    /**
     * Handle successful upload
     */
    handleUploadSuccess(detail) {
        // Refresh video library
        setTimeout(() => {
            this.loadVideos();
        }, 1000);
    }

    /**
     * Update UI based on authentication state
     */
    updateAuthUI() {
        const isAuthenticated = window.authManager?.isAuthenticated();
        
        // Update upload section availability
        if (!isAuthenticated && this.currentSection === 'upload') {
            this.showSection('home');
        }
    }

    /**
     * Load user videos from API
     */
    async loadVideos() {
        if (!window.authManager?.isAuthenticated()) {
            this.videos = [];
            this.renderVideoGrid();
            return;
        }

        try {
            this.setLoading(true);
            
            const response = await window.authManager.apiRequest('/videos');
            const result = await response.json();
            
            if (result.success && result.videos) {
                this.videos = result.videos;
            } else {
                this.videos = [];
                console.warn('Failed to load videos:', result.message);
            }
        } catch (error) {
            console.error('Error loading videos:', error);
            this.videos = [];
            
            if (error.message.includes('Authentication')) {
                // Auth error handled by authManager
                return;
            } else {
                this.showToast('Failed to load videos', 'error');
            }
        } finally {
            this.setLoading(false);
            this.renderVideoGrid();
        }
    }

    /**
     * Filter videos by status
     */
    filterVideos(status) {
        const filteredVideos = status === 'all' 
            ? this.videos 
            : this.videos.filter(video => video.status === status);
        
        this.renderVideoGrid(filteredVideos);
    }

    /**
     * Render video grid
     */
    renderVideoGrid(videos = this.videos) {
        const videoGrid = document.getElementById('videoGrid');
        const emptyState = document.getElementById('emptyState');
        
        if (!videoGrid || !emptyState) return;

        if (videos.length === 0) {
            videoGrid.innerHTML = '';
            emptyState.classList.remove('hidden');
        } else {
            emptyState.classList.add('hidden');
            videoGrid.innerHTML = videos.map(video => this.createVideoCard(video)).join('');
        }
    }

    /**
     * Create video card HTML
     */
    createVideoCard(video) {
        const uploadDate = new Date(video.created_at).toLocaleDateString();
        const statusClass = `status-${video.status}`;
        const canPlay = video.status === 'ready';
        
        return `
            <div class="video-card" ${canPlay ? `onclick="window.app.playVideo('${video.video_id}')"` : ''}>
                <div class="video-thumbnail">
                    ${this.getVideoIcon(video.status)}
                </div>
                <div class="video-info">
                    <div class="video-title">${this.escapeHtml(video.title)}</div>
                    <div class="video-meta">
                        <span>Uploaded ${uploadDate}</span>
                        <span class="video-status ${statusClass}">
                            ${this.formatStatus(video.status)}
                        </span>
                    </div>
                    ${video.description ? `
                        <div class="video-description">
                            ${this.escapeHtml(video.description)}
                        </div>
                    ` : ''}
                </div>
            </div>
        `;
    }

    /**
     * Get video icon based on status
     */
    getVideoIcon(status) {
        switch (status) {
            case 'ready': return '🎬';
            case 'processing': return '⏳';
            case 'failed': return '❌';
            default: return '📹';
        }
    }

    /**
     * Format status for display
     */
    formatStatus(status) {
        switch (status) {
            case 'ready': return 'Ready';
            case 'processing': return 'Processing';
            case 'failed': return 'Failed';
            case 'uploaded': return 'Uploaded';
            default: return status.charAt(0).toUpperCase() + status.slice(1);
        }
    }

    /**
     * Play video
     */
    async playVideo(videoId) {
        const video = this.videos.find(v => v.video_id === videoId);
        
        if (!video) {
            this.showToast('Video not found', 'error');
            return;
        }
        
        if (video.status !== 'ready') {
            this.showToast('Video is still processing', 'warning');
            return;
        }
        
        // Dispatch play event to HLS player
        const event = new CustomEvent('playVideo', {
            detail: { video }
        });
        document.dispatchEvent(event);
    }

    /**
     * Set loading state
     */
    setLoading(isLoading) {
        this.isLoading = isLoading;
        
        const loadingOverlay = document.getElementById('loadingOverlay');
        if (loadingOverlay) {
            loadingOverlay.classList.toggle('hidden', !isLoading);
        }
    }

    /**
     * Show toast notification
     */
    showToast(message, type = 'info', duration = 3000) {
        const container = document.getElementById('toastContainer');
        if (!container) {
            console.log(`[${type.toUpperCase()}] ${message}`);
            return;
        }

        const toast = document.createElement('div');
        toast.className = `toast ${type}`;
        toast.innerHTML = `
            <div class="toast-content">
                <span class="toast-icon">${this.getToastIcon(type)}</span>
                <span class="toast-message">${this.escapeHtml(message)}</span>
            </div>
        `;

        container.appendChild(toast);

        // Auto remove
        setTimeout(() => {
            toast.style.animation = 'slideOut 0.3s ease forwards';
            setTimeout(() => {
                if (toast.parentNode) {
                    toast.parentNode.removeChild(toast);
                }
            }, 300);
        }, duration);

        // Click to dismiss
        toast.addEventListener('click', () => {
            if (toast.parentNode) {
                toast.parentNode.removeChild(toast);
            }
        });
    }

    /**
     * Get toast icon
     */
    getToastIcon(type) {
        switch (type) {
            case 'success': return '✅';
            case 'error': return '❌';
            case 'warning': return '⚠️';
            case 'info':
            default: return 'ℹ️';
        }
    }

    /**
     * Escape HTML to prevent XSS
     */
    escapeHtml(text) {
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }

    /**
     * Handle window resize
     */
    handleResize() {
        // Could implement responsive adjustments here
    }

    /**
     * Handle before unload
     */
    handleBeforeUnload(e) {
        // Check if upload is in progress
        if (window.uploadManager?.currentUpload) {
            e.preventDefault();
            e.returnValue = 'Upload in progress. Are you sure you want to leave?';
            return e.returnValue;
        }
    }

    /**
     * Test API connection
     */
    async testAPIConnection() {
        try {
            const response = await fetch(`${this.apiBase}/health`);
            const result = await response.json();
            
            if (result.success) {
                console.log('✅ API Connection Successful');
            } else {
                console.warn('⚠️ API Health Check Failed:', result);
            }
        } catch (error) {
            console.error('❌ API Connection Failed:', error);
            this.showToast('Unable to connect to server', 'error');
        }
    }

    /**
     * Get application statistics
     */
    getStats() {
        return {
            currentSection: this.currentSection,
            videosLoaded: this.videos.length,
            isAuthenticated: window.authManager?.isAuthenticated() || false,
            isLoading: this.isLoading,
            readyVideos: this.videos.filter(v => v.status === 'ready').length,
            processingVideos: this.videos.filter(v => v.status === 'processing').length
        };
    }

    /**
     * Refresh application data
     */
    async refresh() {
        await this.loadVideos();
        this.showToast('Data refreshed', 'success');
    }
}

// Global toast function for other modules
window.showToast = function(message, type = 'info', duration = 3000) {
    if (window.app) {
        window.app.showToast(message, type, duration);
    } else {
        console.log(`[${type.toUpperCase()}] ${message}`);
    }
};

// Create global app instance
window.app = new StreamingApp();

// Export for module systems
if (typeof module !== 'undefined' && module.exports) {
    module.exports = StreamingApp;
}