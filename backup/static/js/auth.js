/**
 * Authentication Module
 * Handles user registration, login, logout, and JWT token management
 */

class AuthManager {
    constructor() {
        this.apiBase = 'http://localhost:8081/api/v1';
        this.token = null;
        this.user = null;
        this.refreshTimer = null;
        
        // Initialize authentication state
        this.loadStoredAuth();
        this.bindEvents();
    }

    /**
     * Load authentication data from localStorage
     */
    loadStoredAuth() {
        try {
            this.token = localStorage.getItem('authToken');
            const userData = localStorage.getItem('userData');
            
            if (this.token && userData) {
                this.user = JSON.parse(userData);
                this.validateToken();
            }
        } catch (error) {
            console.error('Error loading stored auth:', error);
            this.clearAuth();
        }
    }

    /**
     * Bind authentication events
     */
    bindEvents() {
        // Login button
        document.getElementById('loginBtn')?.addEventListener('click', () => {
            this.showAuthModal('login');
        });

        // Signup button
        document.getElementById('signupBtn')?.addEventListener('click', () => {
            this.showAuthModal('signup');
        });

        // Logout button
        document.getElementById('logoutBtn')?.addEventListener('click', () => {
            this.logout();
        });

        // Close auth modal
        document.getElementById('closeAuth')?.addEventListener('click', () => {
            this.hideAuthModal();
        });

        // Switch between login and signup
        document.getElementById('showSignup')?.addEventListener('click', (e) => {
            e.preventDefault();
            this.switchAuthMode('signup');
        });

        document.getElementById('showLogin')?.addEventListener('click', (e) => {
            e.preventDefault();
            this.switchAuthMode('login');
        });

        // Form submissions
        document.getElementById('loginSubmit')?.addEventListener('click', (e) => {
            e.preventDefault();
            this.handleLogin();
        });

        document.getElementById('signupSubmit')?.addEventListener('click', (e) => {
            e.preventDefault();
            this.handleSignup();
        });

        // Close modal on background click
        document.getElementById('authModal')?.addEventListener('click', (e) => {
            if (e.target.id === 'authModal') {
                this.hideAuthModal();
            }
        });

        // Enter key submission
        document.getElementById('loginForm')?.addEventListener('keypress', (e) => {
            if (e.key === 'Enter') {
                e.preventDefault();
                this.handleLogin();
            }
        });

        document.getElementById('signupForm')?.addEventListener('keypress', (e) => {
            if (e.key === 'Enter') {
                e.preventDefault();
                this.handleSignup();
            }
        });
    }

    /**
     * Show authentication modal
     */
    showAuthModal(mode = 'login') {
        const modal = document.getElementById('authModal');
        if (!modal) return;

        this.switchAuthMode(mode);
        modal.classList.add('active');
        
        // Focus first input
        setTimeout(() => {
            const firstInput = modal.querySelector('input:not([type="hidden"])');
            firstInput?.focus();
        }, 100);
    }

    /**
     * Hide authentication modal
     */
    hideAuthModal() {
        const modal = document.getElementById('authModal');
        if (!modal) return;

        modal.classList.remove('active');
        this.clearAuthForms();
    }

    /**
     * Switch between login and signup modes
     */
    switchAuthMode(mode) {
        const authTitle = document.getElementById('authTitle');
        const loginForm = document.getElementById('loginForm');
        const signupForm = document.getElementById('signupForm');

        if (mode === 'signup') {
            authTitle.textContent = 'Sign Up';
            loginForm.classList.add('hidden');
            signupForm.classList.remove('hidden');
        } else {
            authTitle.textContent = 'Login';
            loginForm.classList.remove('hidden');
            signupForm.classList.add('hidden');
        }
    }

    /**
     * Clear authentication forms
     */
    clearAuthForms() {
        // Clear login form
        document.getElementById('loginEmail').value = '';
        document.getElementById('loginPassword').value = '';

        // Clear signup form
        document.getElementById('signupName').value = '';
        document.getElementById('signupEmail').value = '';
        document.getElementById('signupPassword').value = '';
        document.getElementById('confirmPassword').value = '';
    }

    /**
     * Handle user login
     */
    async handleLogin() {
        const email = document.getElementById('loginEmail').value.trim();
        const password = document.getElementById('loginPassword').value;

        if (!email || !password) {
            this.showToast('Please fill in all fields', 'error');
            return;
        }

        if (!this.validateEmail(email)) {
            this.showToast('Please enter a valid email address', 'error');
            return;
        }

        const submitBtn = document.getElementById('loginSubmit');
        const originalText = submitBtn.textContent;

        try {
            submitBtn.disabled = true;
            submitBtn.textContent = 'Signing In...';

            const response = await fetch(`${this.apiBase}/auth/login`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({ email, password })
            });

            const result = await response.json();

            if (response.ok && result.success) {
                this.setAuth(result.token, result.user);
                this.showToast(`Welcome back, ${result.user.username}!`, 'success');
                this.hideAuthModal();
                
                // Dispatch login event
                this.dispatchAuthEvent('login', result.user);
            } else {
                this.showToast(result.message || 'Login failed', 'error');
            }
        } catch (error) {
            console.error('Login error:', error);
            this.showToast('Network error. Please try again.', 'error');
        } finally {
            submitBtn.disabled = false;
            submitBtn.textContent = originalText;
        }
    }

    /**
     * Handle user signup
     */
    async handleSignup() {
        const name = document.getElementById('signupName').value.trim();
        const email = document.getElementById('signupEmail').value.trim();
        const password = document.getElementById('signupPassword').value;
        const confirmPassword = document.getElementById('confirmPassword').value;

        // Validation
        if (!name || !email || !password || !confirmPassword) {
            this.showToast('Please fill in all fields', 'error');
            return;
        }

        if (!this.validateEmail(email)) {
            this.showToast('Please enter a valid email address', 'error');
            return;
        }

        if (password.length < 6) {
            this.showToast('Password must be at least 6 characters long', 'error');
            return;
        }

        if (password !== confirmPassword) {
            this.showToast('Passwords do not match', 'error');
            return;
        }

        const submitBtn = document.getElementById('signupSubmit');
        const originalText = submitBtn.textContent;

        try {
            submitBtn.disabled = true;
            submitBtn.textContent = 'Creating Account...';

            const response = await fetch(`${this.apiBase}/auth/register`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({
                    username: name,
                    email: email,
                    password: password,
                    first_name: name.split(' ')[0] || name,
                    last_name: name.split(' ').slice(1).join(' ') || ''
                })
            });

            const result = await response.json();

            if (response.ok && result.success) {
                this.setAuth(result.token, result.user);
                this.showToast(`Welcome, ${result.user.username}!`, 'success');
                this.hideAuthModal();
                
                // Dispatch signup event
                this.dispatchAuthEvent('signup', result.user);
            } else {
                this.showToast(result.message || 'Registration failed', 'error');
            }
        } catch (error) {
            console.error('Signup error:', error);
            this.showToast('Network error. Please try again.', 'error');
        } finally {
            submitBtn.disabled = false;
            submitBtn.textContent = originalText;
        }
    }

    /**
     * Handle user logout
     */
    logout() {
        this.clearAuth();
        this.showToast('You have been logged out', 'success');
        
        // Dispatch logout event
        this.dispatchAuthEvent('logout');
    }

    /**
     * Set authentication state
     */
    setAuth(token, user) {
        this.token = token;
        this.user = user;
        
        // Store in localStorage
        localStorage.setItem('authToken', token);
        localStorage.setItem('userData', JSON.stringify(user));
        
        // Update UI
        this.updateAuthUI();
        
        // Set up token refresh
        this.setupTokenRefresh();
    }

    /**
     * Clear authentication state
     */
    clearAuth() {
        this.token = null;
        this.user = null;
        
        // Clear localStorage
        localStorage.removeItem('authToken');
        localStorage.removeItem('userData');
        
        // Clear refresh timer
        if (this.refreshTimer) {
            clearTimeout(this.refreshTimer);
            this.refreshTimer = null;
        }
        
        // Update UI
        this.updateAuthUI();
    }

    /**
     * Update authentication UI
     */
    updateAuthUI() {
        const loginBtn = document.getElementById('loginBtn');
        const signupBtn = document.getElementById('signupBtn');
        const userMenu = document.getElementById('userMenu');
        const userNameDisplay = document.getElementById('userNameDisplay');

        if (this.isAuthenticated()) {
            // Hide auth buttons
            loginBtn?.classList.add('hidden');
            signupBtn?.classList.add('hidden');
            
            // Show user menu
            userMenu?.classList.remove('hidden');
            if (userNameDisplay) {
                userNameDisplay.textContent = this.user.username || this.user.email;
            }
        } else {
            // Show auth buttons
            loginBtn?.classList.remove('hidden');
            signupBtn?.classList.remove('hidden');
            
            // Hide user menu
            userMenu?.classList.add('hidden');
        }
    }

    /**
     * Validate token and refresh if needed
     */
    async validateToken() {
        if (!this.token) return false;

        try {
            const response = await fetch(`${this.apiBase}/auth/validate`, {
                method: 'GET',
                headers: {
                    'Authorization': `Bearer ${this.token}`
                }
            });

            if (response.ok) {
                const result = await response.json();
                if (result.success) {
                    this.updateAuthUI();
                    this.setupTokenRefresh();
                    return true;
                }
            }
            
            // Token is invalid
            this.clearAuth();
            return false;
        } catch (error) {
            console.error('Token validation error:', error);
            return false;
        }
    }

    /**
     * Setup automatic token refresh
     */
    setupTokenRefresh() {
        if (this.refreshTimer) {
            clearTimeout(this.refreshTimer);
        }

        // Refresh token 5 minutes before expiration (assuming 1 hour expiration)
        const refreshInterval = 55 * 60 * 1000; // 55 minutes
        
        this.refreshTimer = setTimeout(() => {
            this.refreshToken();
        }, refreshInterval);
    }

    /**
     * Refresh authentication token
     */
    async refreshToken() {
        if (!this.token) return;

        try {
            const response = await fetch(`${this.apiBase}/auth/refresh`, {
                method: 'POST',
                headers: {
                    'Authorization': `Bearer ${this.token}`
                }
            });

            if (response.ok) {
                const result = await response.json();
                if (result.success && result.token) {
                    this.token = result.token;
                    localStorage.setItem('authToken', this.token);
                    this.setupTokenRefresh();
                    return;
                }
            }
            
            // Refresh failed, logout user
            console.warn('Token refresh failed, logging out user');
            this.logout();
        } catch (error) {
            console.error('Token refresh error:', error);
            this.logout();
        }
    }

    /**
     * Get authentication header for API requests
     */
    getAuthHeader() {
        return this.token ? `Bearer ${this.token}` : null;
    }

    /**
     * Check if user is authenticated
     */
    isAuthenticated() {
        return !!(this.token && this.user);
    }

    /**
     * Get current user
     */
    getCurrentUser() {
        return this.user;
    }

    /**
     * Get current token
     */
    getCurrentToken() {
        return this.token;
    }

    /**
     * Validate email format
     */
    validateEmail(email) {
        const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
        return emailRegex.test(email);
    }

    /**
     * Show toast notification
     */
    showToast(message, type = 'info') {
        // This will be implemented by the main app
        if (window.showToast) {
            window.showToast(message, type);
        } else {
            console.log(`[${type.toUpperCase()}] ${message}`);
        }
    }

    /**
     * Dispatch authentication events
     */
    dispatchAuthEvent(type, user = null) {
        const event = new CustomEvent('authStateChange', {
            detail: { type, user, isAuthenticated: this.isAuthenticated() }
        });
        document.dispatchEvent(event);
    }

    /**
     * Make authenticated API request
     */
    async apiRequest(endpoint, options = {}) {
        const headers = {
            'Content-Type': 'application/json',
            ...options.headers
        };

        if (this.token) {
            headers.Authorization = `Bearer ${this.token}`;
        }

        const response = await fetch(`${this.apiBase}${endpoint}`, {
            ...options,
            headers
        });

        // Handle token expiration
        if (response.status === 401) {
            this.logout();
            throw new Error('Authentication expired. Please login again.');
        }

        return response;
    }
}

// Create global auth manager instance
window.authManager = new AuthManager();

// Export for module systems
if (typeof module !== 'undefined' && module.exports) {
    module.exports = AuthManager;
}