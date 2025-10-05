/**
 * Upload Module
 * Handles video file upload with drag-and-drop, progress tracking, and validation
 */

class UploadManager {
    constructor() {
        this.apiBase = 'http://localhost:8081/api/v1';
        this.currentUpload = null;
        this.supportedFormats = ['mp4', 'avi', 'mov', 'mkv', 'webm', 'flv', 'wmv'];
        this.maxFileSize = 2 * 1024 * 1024 * 1024; // 2GB
        
        this.bindEvents();
    }

    /**
     * Bind upload-related events
     */
    bindEvents() {
        // File input change
        const fileInput = document.getElementById('videoFileInput');
        fileInput?.addEventListener('change', (e) => {
            this.handleFileSelect(e.target.files);
        });

        // Drag and drop events
        const uploadArea = document.getElementById('uploadArea');
        if (uploadArea) {
            uploadArea.addEventListener('dragover', this.handleDragOver.bind(this));
            uploadArea.addEventListener('dragleave', this.handleDragLeave.bind(this));
            uploadArea.addEventListener('drop', this.handleDrop.bind(this));
            uploadArea.addEventListener('click', () => {
                if (window.authManager?.isAuthenticated()) {
                    fileInput?.click();
                } else {
                    window.authManager?.showAuthModal('login');
                }
            });
        }

        // Upload button
        document.getElementById('uploadBtn')?.addEventListener('click', (e) => {
            e.preventDefault();
            this.handleUpload();
        });

        // Listen for auth state changes
        document.addEventListener('authStateChange', (e) => {
            this.updateUploadUI(e.detail.isAuthenticated);
        });
    }

    /**
     * Handle drag over event
     */
    handleDragOver(e) {
        e.preventDefault();
        e.stopPropagation();
        
        const uploadArea = document.getElementById('uploadArea');
        uploadArea?.classList.add('dragover');
        
        e.dataTransfer.dropEffect = 'copy';
    }

    /**
     * Handle drag leave event
     */
    handleDragLeave(e) {
        e.preventDefault();
        e.stopPropagation();
        
        const uploadArea = document.getElementById('uploadArea');
        uploadArea?.classList.remove('dragover');
    }

    /**
     * Handle file drop event
     */
    handleDrop(e) {
        e.preventDefault();
        e.stopPropagation();
        
        const uploadArea = document.getElementById('uploadArea');
        uploadArea?.classList.remove('dragover');
        
        if (!window.authManager?.isAuthenticated()) {
            this.showToast('Please login to upload videos', 'error');
            window.authManager?.showAuthModal('login');
            return;
        }
        
        const files = Array.from(e.dataTransfer.files);
        this.handleFileSelect(files);
    }

    /**
     * Handle file selection
     */
    handleFileSelect(files) {
        if (!files || files.length === 0) return;
        
        const file = files[0];
        
        // Validate file
        const validation = this.validateFile(file);
        if (!validation.valid) {
            this.showToast(validation.error, 'error');
            return;
        }
        
        // Show file info and upload form
        this.displaySelectedFile(file);
        this.showUploadForm();
    }

    /**
     * Validate uploaded file
     */
    validateFile(file) {
        // Check if file exists
        if (!file) {
            return { valid: false, error: 'No file selected' };
        }
        
        // Check file type
        if (!file.type.startsWith('video/')) {
            return { valid: false, error: 'Please select a video file' };
        }
        
        // Check file extension
        const extension = file.name.split('.').pop()?.toLowerCase();
        if (!extension || !this.supportedFormats.includes(extension)) {
            return { 
                valid: false, 
                error: `Unsupported format. Supported formats: ${this.supportedFormats.join(', ')}` 
            };
        }
        
        // Check file size
        if (file.size > this.maxFileSize) {
            return { 
                valid: false, 
                error: `File size too large. Maximum size: ${this.formatFileSize(this.maxFileSize)}` 
            };
        }
        
        return { valid: true };
    }

    /**
     * Display selected file information
     */
    displaySelectedFile(file) {
        const uploadArea = document.getElementById('uploadArea');
        const uploadContent = uploadArea?.querySelector('.upload-content');
        
        if (uploadContent) {
            uploadContent.innerHTML = `
                <div class="upload-icon">✅</div>
                <h3>File Selected</h3>
                <div class="file-info">
                    <p><strong>Name:</strong> ${file.name}</p>
                    <p><strong>Size:</strong> ${this.formatFileSize(file.size)}</p>
                    <p><strong>Type:</strong> ${file.type}</p>
                </div>
                <button class="select-btn" onclick="document.getElementById('videoFileInput').click()">
                    Change File
                </button>
            `;
        }
    }

    /**
     * Show upload form
     */
    showUploadForm() {
        const uploadForm = document.getElementById('uploadForm');
        uploadForm?.classList.remove('hidden');
        
        // Focus on title input
        setTimeout(() => {
            document.getElementById('videoTitle')?.focus();
        }, 100);
    }

    /**
     * Hide upload form
     */
    hideUploadForm() {
        const uploadForm = document.getElementById('uploadForm');
        uploadForm?.classList.add('hidden');
        
        // Reset form
        this.resetUploadForm();
    }

    /**
     * Reset upload form
     */
    resetUploadForm() {
        const form = document.getElementById('uploadForm');
        if (form) {
            form.querySelector('#videoTitle').value = '';
            form.querySelector('#videoDescription').value = '';
        }
        
        const fileInput = document.getElementById('videoFileInput');
        if (fileInput) {
            fileInput.value = '';
        }
        
        // Reset upload area
        const uploadArea = document.getElementById('uploadArea');
        const uploadContent = uploadArea?.querySelector('.upload-content');
        if (uploadContent) {
            uploadContent.innerHTML = `
                <div class="upload-icon">📹</div>
                <h3>Drag & Drop Your Video</h3>
                <p>Or click to select a video file</p>
                <input type="file" id="videoFileInput" accept="video/*" hidden>
                <button class="select-btn" onclick="document.getElementById('videoFileInput').click()">
                    Select File
                </button>
            `;
        }
    }

    /**
     * Handle video upload
     */
    async handleUpload() {
        if (!window.authManager?.isAuthenticated()) {
            this.showToast('Please login to upload videos', 'error');
            window.authManager?.showAuthModal('login');
            return;
        }

        const fileInput = document.getElementById('videoFileInput');
        const title = document.getElementById('videoTitle')?.value.trim();
        const description = document.getElementById('videoDescription')?.value.trim();

        // Validation
        if (!fileInput?.files[0]) {
            this.showToast('Please select a video file', 'error');
            return;
        }

        if (!title) {
            this.showToast('Please enter a video title', 'error');
            return;
        }

        const file = fileInput.files[0];
        const validation = this.validateFile(file);
        if (!validation.valid) {
            this.showToast(validation.error, 'error');
            return;
        }

        // Prepare upload
        const uploadBtn = document.getElementById('uploadBtn');
        const originalText = uploadBtn?.textContent;

        try {
            uploadBtn.disabled = true;
            uploadBtn.textContent = 'Preparing...';

            // Show progress
            this.showProgress();

            // Create form data
            const formData = new FormData();
            formData.append('video', file);
            formData.append('title', title);
            if (description) {
                formData.append('description', description);
            }

            // Start upload
            await this.uploadFile(formData);

        } catch (error) {
            console.error('Upload error:', error);
            this.showToast(error.message || 'Upload failed', 'error');
            this.hideProgress();
        } finally {
            uploadBtn.disabled = false;
            uploadBtn.textContent = originalText;
        }
    }

    /**
     * Upload file with progress tracking
     */
    async uploadFile(formData) {
        return new Promise((resolve, reject) => {
            const xhr = new XMLHttpRequest();
            
            // Track upload progress
            xhr.upload.addEventListener('progress', (e) => {
                if (e.lengthComputable) {
                    const percentComplete = Math.round((e.loaded / e.total) * 100);
                    this.updateProgress(percentComplete, 'Uploading...');
                }
            });

            // Handle completion
            xhr.addEventListener('load', () => {
                try {
                    if (xhr.status >= 200 && xhr.status < 300) {
                        const response = JSON.parse(xhr.responseText);
                        if (response.success) {
                            this.handleUploadSuccess(response);
                            resolve(response);
                        } else {
                            reject(new Error(response.message || 'Upload failed'));
                        }
                    } else {
                        const response = JSON.parse(xhr.responseText);
                        reject(new Error(response.message || `HTTP ${xhr.status}: Upload failed`));
                    }
                } catch (error) {
                    reject(new Error('Invalid server response'));
                }
            });

            // Handle errors
            xhr.addEventListener('error', () => {
                reject(new Error('Network error during upload'));
            });

            xhr.addEventListener('abort', () => {
                reject(new Error('Upload cancelled'));
            });

            // Setup request
            xhr.open('POST', `${this.apiBase}/upload`);
            
            // Add authorization header
            const token = window.authManager?.getCurrentToken();
            if (token) {
                xhr.setRequestHeader('Authorization', `Bearer ${token}`);
            }

            // Start upload
            this.currentUpload = xhr;
            xhr.send(formData);
        });
    }

    /**
     * Handle successful upload
     */
    handleUploadSuccess(response) {
        this.updateProgress(100, 'Upload completed!');
        
        setTimeout(() => {
            this.hideProgress();
            this.hideUploadForm();
            this.showToast('Video uploaded successfully! Processing will begin shortly.', 'success');
            
            // Dispatch upload success event
            const event = new CustomEvent('uploadSuccess', {
                detail: { videoId: response.video_id, response }
            });
            document.dispatchEvent(event);
            
            // Redirect to library
            setTimeout(() => {
                this.navigateToLibrary();
            }, 2000);
        }, 1000);
    }

    /**
     * Show upload progress
     */
    showProgress() {
        const progressContainer = document.getElementById('uploadProgress');
        progressContainer?.classList.remove('hidden');
        
        const uploadForm = document.getElementById('uploadForm');
        uploadForm?.classList.add('hidden');
    }

    /**
     * Update upload progress
     */
    updateProgress(percent, text = 'Uploading...') {
        const progressFill = document.getElementById('progressFill');
        const progressText = document.getElementById('progressText');
        const progressPercent = document.getElementById('progressPercent');

        if (progressFill) {
            progressFill.style.width = `${percent}%`;
        }
        if (progressText) {
            progressText.textContent = text;
        }
        if (progressPercent) {
            progressPercent.textContent = `${percent}%`;
        }
    }

    /**
     * Hide upload progress
     */
    hideProgress() {
        const progressContainer = document.getElementById('uploadProgress');
        progressContainer?.classList.add('hidden');
        
        // Reset progress
        this.updateProgress(0, 'Uploading...');
    }

    /**
     * Cancel current upload
     */
    cancelUpload() {
        if (this.currentUpload) {
            this.currentUpload.abort();
            this.currentUpload = null;
            this.hideProgress();
            this.showToast('Upload cancelled', 'warning');
        }
    }

    /**
     * Navigate to library section
     */
    navigateToLibrary() {
        // This will be handled by the main app
        if (window.app && window.app.showSection) {
            window.app.showSection('library');
        }
    }

    /**
     * Update upload UI based on auth state
     */
    updateUploadUI(isAuthenticated) {
        const uploadArea = document.getElementById('uploadArea');
        const uploadContent = uploadArea?.querySelector('.upload-content');
        
        if (!isAuthenticated && uploadContent) {
            uploadContent.innerHTML = `
                <div class="upload-icon">🔒</div>
                <h3>Login Required</h3>
                <p>Please login to upload videos</p>
                <button class="select-btn" onclick="window.authManager?.showAuthModal('login')">
                    Login
                </button>
            `;
        } else if (isAuthenticated && uploadContent && !uploadContent.querySelector('.file-info')) {
            // Reset to default state if logged in and no file selected
            this.resetUploadForm();
        }
    }

    /**
     * Format file size for display
     */
    formatFileSize(bytes) {
        if (bytes === 0) return '0 Bytes';
        
        const k = 1024;
        const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB'];
        const i = Math.floor(Math.log(bytes) / Math.log(k));
        
        return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
    }

    /**
     * Get upload statistics
     */
    getUploadStats() {
        return {
            supportedFormats: this.supportedFormats,
            maxFileSize: this.maxFileSize,
            maxFileSizeFormatted: this.formatFileSize(this.maxFileSize)
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

// Create global upload manager instance
window.uploadManager = new UploadManager();

// Export for module systems
if (typeof module !== 'undefined' && module.exports) {
    module.exports = UploadManager;
}