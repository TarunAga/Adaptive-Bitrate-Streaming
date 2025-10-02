package repository

import (
    "github.com/TarunAga/adaptive-bitrate-streaming/pkg/entities"
    "gorm.io/gorm"
)

type UserRepository struct {
    db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
    return &UserRepository{db: db}
}

// CreateUser creates a new user
func (r *UserRepository) CreateUserWithPassword(user *entities.User) (*entities.User, error) {
    result := r.db.Create(user)
    if result.Error != nil {
        return nil, result.Error
    }
    return user, nil
}



func (r *UserRepository) GetUserByEmail(email string) (*entities.User, error) {
	var user entities.User
	result := r.db.Where("email = ?", email).First(&user)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}


// GetUserByUserName finds user by username
func (r *UserRepository) GetUserByUserName(userName string) (*entities.User, error) {
    var user entities.User
    result := r.db.Where("user_name = ?", userName).First(&user)
    
    if result.Error != nil {
        return nil, result.Error
    }
    
    return &user, nil
}

// GetUserByID finds user by ID
func (r *UserRepository) GetUserByID(userID uint) (*entities.User, error) {
    var user entities.User
    result := r.db.First(&user, userID)
    
    if result.Error != nil {
        return nil, result.Error
    }
    
    return &user, nil
}
