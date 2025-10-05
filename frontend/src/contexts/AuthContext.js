import React, { createContext, useState, useContext, useEffect } from 'react';

const AuthContext = createContext();

export const useAuth = () => {
    const context = useContext(AuthContext);
    if (!context) {
        throw new Error('useAuth must be used within an AuthProvider');
    }
    return context;
};

export const AuthProvider = ({ children }) => {
    const [user, setUser] = useState(null);
    const [token, setToken] = useState(null);
    const [loading, setLoading] = useState(true);

    // Check if user is already logged in on app start
    useEffect(() => {
        const initializeAuth = () => {
            try {
                const savedToken = localStorage.getItem('authToken');
                const savedUser = localStorage.getItem('user');

                console.log('🔍 Checking saved auth data...');
                console.log('Token exists:', !!savedToken);
                console.log('User exists:', !!savedUser);

                if (savedToken && savedUser) {
                    const parsedUser = JSON.parse(savedUser);
                    console.log('✅ Restoring auth state for user:', parsedUser.user_name || parsedUser.email);
                    
                    setToken(savedToken);
                    setUser(parsedUser);
                } else {
                    console.log('❌ No saved auth data found');
                }
            } catch (error) {
                console.error('❌ Error parsing saved auth data:', error);
                // Clear corrupted data
                localStorage.removeItem('authToken');
                localStorage.removeItem('user');
            } finally {
                setLoading(false);
            }
        };

        initializeAuth();
    }, []);

    // ✅ FIX: Updated login function to handle your backend response format
    const login = (userData, authToken) => {
        console.log('🔄 AuthContext login called with:', { userData, token: !!authToken });
        
        try {
            // Store in localStorage
            localStorage.setItem('authToken', authToken);
            localStorage.setItem('user', JSON.stringify(userData));
            
            // Update state
            setToken(authToken);
            setUser(userData);
            
            console.log('✅ AuthContext state updated successfully');
        } catch (error) {
            console.error('❌ Error storing auth data:', error);
            throw new Error('Failed to save authentication data');
        }
    };

    const logout = () => {
        console.log('🚪 Logging out user');
        
        // Clear localStorage
        localStorage.removeItem('authToken');
        localStorage.removeItem('user');
        
        // Clear state
        setToken(null);
        setUser(null);
        
        console.log('✅ User logged out successfully');
    };

    const isAuthenticated = () => {
        return !!(user && token);
    };

    const value = {
        user,
        token,
        loading,
        login,
        logout,
        isAuthenticated: isAuthenticated()
    };

    return (
        <AuthContext.Provider value={value}>
            {children}
        </AuthContext.Provider>
    );
};