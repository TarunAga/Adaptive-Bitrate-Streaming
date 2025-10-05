import axios from 'axios';

const API_BASE_URL = 'http://localhost:8081/api/v1';

// Create axios instance
const api = axios.create({
    baseURL: API_BASE_URL,
    timeout: 30000, // 30 seconds timeout
    headers: {
        'Content-Type': 'application/json',
    },
});

// Request interceptor to add auth token
api.interceptors.request.use(
    (config) => {
        const token = localStorage.getItem('authToken');
        if (token) {
            config.headers.Authorization = `Bearer ${token}`;
        }
        return config;
    },
    (error) => {
        return Promise.reject(error);
    }
);

// Response interceptor to handle errors
api.interceptors.response.use(
    (response) => {
        return response;
    },
    (error) => {
        if (error.response) {
            console.error('API Error:', error.response.data);
            
            // Handle authentication errors
            if (error.response.status === 401) {
                localStorage.removeItem('authToken');
                localStorage.removeItem('user');
                window.location.href = '/login';
            }
        } else if (error.request) {
            console.error('Network Error:', error.message);
        }
        
        return Promise.reject(error);
    }
);

// ✅ AUTH ENDPOINTS
export const authAPI = {
    register: async (userData) => {
        try {
            console.log('📡 API: Sending register request:', { ...userData, password: '[HIDDEN]' });
            
            const response = await api.post('/auth/register', userData);
            
            console.log('✅ API: Register response:', response.data);
            return response.data;
        } catch (error) {
            console.error('❌ API: Register error:', error.response?.data || error.message);
            
            // ✅ FIX: Better error handling
            if (error.response?.data) {
                throw error.response.data;
            } else {
                throw { 
                    success: false, 
                    message: error.message || 'Registration failed' 
                };
            }
        }
    },

    login: async (credentials) => {
        try {
            console.log('📡 API: Sending login request:', { ...credentials, password: '[HIDDEN]' });
            
            const response = await api.post('/auth/login', credentials);
            
            console.log('✅ API: Login response:', {
                ...response.data,
                token: response.data.token ? `${response.data.token.substring(0, 20)}...` : 'none'
            });
            
            return response.data;
        } catch (error) {
            console.error('❌ API: Login error:', error.response?.data || error.message);
            
            // ✅ FIX: Better error handling for login
            if (error.response?.data) {
                // Backend returned an error response
                throw error.response.data;
            } else if (error.request) {
                // Network error
                throw { 
                    success: false, 
                    message: 'Network error - please check your connection' 
                };
            } else {
                // Other error
                throw { 
                    success: false, 
                    message: error.message || 'Login failed' 
                };
            }
        }
    },
};

// ✅ VIDEO ENDPOINTS
export const videoAPI = {
    getVideos: async () => {
        try {
            const response = await api.get('/videos');
            
            if (response.data) {
                if (response.data.success && Array.isArray(response.data.videos)) {
                    return {
                        success: true,
                        videos: response.data.videos
                    };
                } else if (Array.isArray(response.data.videos)) {
                    return {
                        success: true,
                        videos: response.data.videos
                    };
                } else if (Array.isArray(response.data)) {
                    return {
                        success: true,
                        videos: response.data
                    };
                } else {
                    return {
                        success: true,
                        videos: []
                    };
                }
            }
            
            return {
                success: true,
                videos: []
            };
        } catch (error) {
            console.error('Error fetching videos:', error);
            
            return {
                success: false,
                videos: [],
                error: error.response?.data?.message || 'Failed to fetch videos'
            };
        }
    },

    uploadVideo: async (videoData) => {
        try {
            const formData = new FormData();
            formData.append('video', videoData.file);
            formData.append('title', videoData.title);
            if (videoData.description) {
                formData.append('description', videoData.description);
            }

            const response = await api.post('/upload', formData, {
                headers: {
                    'Content-Type': 'multipart/form-data',
                },
                timeout: 300000, // 5 minutes for video upload
            });

            return response.data;
        } catch (error) {
            throw error.response?.data || { message: 'Upload failed' };
        }
    },

    getStreamingURL: async (videoId) => {
        try {
            const response = await api.get(`/video/${videoId}/stream`);
            return response.data;
        } catch (error) {
            throw error.response?.data || { message: 'Failed to get streaming URL' };
        }
    },
};

export const healthCheck = async () => {
    try {
        const response = await api.get('/health');
        return response.data;
    } catch (error) {
        throw error.response?.data || { message: 'Health check failed' };
    }
};

export default api;