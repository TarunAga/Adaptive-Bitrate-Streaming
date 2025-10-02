package auth

import (
    "encoding/json"
    "net/http"
    "strings"
    "log"

    "gorm.io/gorm"
)

type Handler struct {
    authService *Service
}

func NewAuthHandler(db *gorm.DB) *Handler {
    return &Handler{
        authService: NewAuthService(db),
    }
}

func (h *Handler) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req RegisterRequest 
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.Printf("Failed to decode register request: %v", err)
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	if req.UserName == "" || req.Password == "" || req.Email == "" {
		respondWithError(w, http.StatusBadRequest, "userName, password and email are required")
		return
	}
	resp, err := h.authService.Register(&req)
	if err != nil {
		log.Printf("Registration failed: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Registration failed")
		return
	}
	respondWithJSON(w, http.StatusOK, resp)
}

func (h *Handler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req LoginRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.Printf("Failed to decode login request: %v", err)
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if req.UserName == "" || req.Password == "" {
		respondWithError(w, http.StatusBadRequest, "userName and password are required")
		return
	}

	resp, err := h.authService.Login(&req)
	if err != nil {
		log.Printf("Login failed: %v", err)
		respondWithError(w, http.StatusUnauthorized, "Login failed")
		return
	}

	respondWithJSON(w, http.StatusOK, resp)
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}
func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		http.Error(w, "Failed to marshal response", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func (h *Handler) AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
        
        if r.Method == "OPTIONS" {
            w.WriteHeader(http.StatusOK)
            return
        }

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			respondWithError(w, http.StatusUnauthorized, "Missing Authorization header")
			return
		}
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			respondWithError(w, http.StatusUnauthorized, "Invalid Authorization header format")
			return
		}
		tokenString := parts[1]
		claims, err := h.authService.ValidateToken(tokenString)
		if err != nil {
			log.Printf("Token validation failed: %v", err)
			respondWithError(w, http.StatusUnauthorized, "Invalid or expired token")
			return
		}
		// Add user info to request context
        r.Header.Set("X-User-ID", string(rune(claims.UserID)))
        r.Header.Set("X-User-Name", claims.UserName)
        r.Header.Set("X-User-Email", claims.Email) // NEW: Added email

        next(w,r)
	})
}

