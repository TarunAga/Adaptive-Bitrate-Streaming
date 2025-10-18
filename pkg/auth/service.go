package auth
import (
	"time"
	"errors"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"github.com/google/uuid"
	"github.com/TarunAga/adaptive-bitrate-streaming/pkg/entities"
	"github.com/TarunAga/adaptive-bitrate-streaming/pkg/repository"
)

var jwtSecret = []byte("your_secret_key") // Change this to a secure key
type Service struct {
	userRepo *repository.UserRepository
}
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	UserName string `json:"user_name"` // Added UserName field
}

type RegisterRequest struct {
	UserName string `json:"username"`
	Password string `json:"password"`
	Email	string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type AuthResponse struct {
	Token string `json:"token"`
	User  *entities.User `json:"user"`
	Success bool `json:"success"`
	Message string `json:"message"`
}

type Claims  struct {
	jwt.RegisteredClaims
    UserID   uuid.UUID `json:"user_id"`   
	UserName string `json:"user_name"`
	Email    string `json:"email"`
}

func NewAuthService(db *gorm.DB) *Service {
	return &Service{
		userRepo: repository.NewUserRepository(db),
	}
}

func (s *Service) Register(req *RegisterRequest) (*AuthResponse, error) {
	// Check if user already exists
	existingUser, _ := s.userRepo.GetUserByUserName(req.UserName)
	if existingUser != nil {
		return &AuthResponse{
			Success: false,
			Message: "Username already taken",
		}, nil
	}
	existingEmailUser, _ := s.userRepo.GetUserByEmail(req.Email)
	if existingEmailUser != nil {
		return &AuthResponse{
			Success: false,
			Message: "Email already registered",
		}, nil
	}
	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	// Create new user
	user := &entities.User{
		UserName:     req.UserName,
		PasswordHash: string(hashedPassword),
		Email:        req.Email,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		IsActive:     true,
	}
	createdUser, err := s.userRepo.CreateUserWithPassword(user)
	if err != nil {
		return nil, err
	}
	// Generate JWT token
	token, err := generateJWT(createdUser)
	if err != nil {
		return nil, err
	}
	return &AuthResponse{
		Token:   token,
		User:    createdUser,
		Success: true,
		Message: "Registration successful",
	}, nil
}

func (s *Service) Login(req *LoginRequest) (*AuthResponse, error) {
	// Find user by email
	user, err := s.userRepo.GetUserByEmail(req.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &AuthResponse{
				Success: false,
				Message: "Invalid email or password",
			}, nil
		}
		return nil, err
	}
	// Check password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		return &AuthResponse{
			Success: false,
			Message: "Invalid email or password",
		}, nil
	}
	// Generate JWT token
	token, err := generateJWT(user)
	if err != nil {
		return nil, err
	}
	return &AuthResponse{
		Token:   token,
		User:    user,
		Success: true,
		Message: "Login successful",
	}, nil
}

func generateJWT(user *entities.User) (string, error) {
	claims := &Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(72 * time.Hour)), // Token expires in 72 hours
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "adaptive-bitrate-streaming",
		},
		UserID:   user.UserID,
		UserName: user.UserName,
		Email:    user.Email,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func (s *Service) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, errors.New("invalid token")
}

// GetUserFromToken retrieves the user associated with the given JWT token
func (s *Service) GetUserFromToken(tokenString string) (*entities.User, error) {
	claims, err := s.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}
	user, err := s.userRepo.GetUserByID(claims.UserID)
	if err != nil {
		return nil, err
	}
	return user, nil
}
