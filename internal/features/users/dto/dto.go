package dto

import (
	"time"

	"template-golang/internal/db/model"
	"template-golang/pkg/pagination"
	"template-golang/pkg/response"
)

// LoginRequest represents the login request data structure
// @Description Login request payload
type LoginRequest struct {
	// @Description User email address
	// @Example john.doe@example.com
	Email string `json:"email" validate:"required,email"`
	// @Description User password
	// @Example secretpassword123
	Password string `json:"password" validate:"required"`
}

// LoginResponse represents the login response data structure
// @Description Login response payload
type LoginResponse struct {
	// @Description User ID
	// @Example 123e4567-e89b-12d3-a456-426614174000
	ID string `json:"id"`
	// @Description User full name
	// @Example John Doe
	Name string `json:"name"`
	// @Description User email address
	// @Example john.doe@example.com
	Email string `json:"email"`
	// @Description User role
	// @Example admin
	Role model.UserRole `json:"role"`
	// @Description JWT access token
	// @Example eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
	Token string `json:"token"`
	// @Description User creation timestamp
	// @Example 2024-03-15T10:00:00Z
	CreatedAt time.Time `json:"created_at"`
}

// CreateUserRequest represents the create user request data structure
// @Description Create user request payload
type CreateUserRequest struct {
	// @Description User full name
	// @Example John Doe
	Name string `json:"name" validate:"required"`
	// @Description User email address
	// @Example john.doe@example.com
	Email string `json:"email" validate:"required,email"`
	// @Description User password (minimum 6 characters)
	// @Example secretpassword123
	Password string `json:"password" validate:"required,min=6"`
	// @Description User learning point ID
	// @Example 123
	LearningPointId string `json:"learning_point_id" validate:"required"`
}

// UpdateUserRequest represents the update user request data structure
// @Description Update user request payload
type UpdateUserRequest struct {
	// @Description User full name
	// @Example John Doe
	Name *string `json:"name,omitempty"`
	// @Description User email address
	// @Example john.doe@example.com
	Email *string `json:"email,omitempty" validate:"omitempty,email"`
	// @Description User password (minimum 6 characters, optional)
	// @Example newpassword123
	Password *string `json:"password,omitempty" validate:"omitempty,min=6"`
}

// UserResponse represents the user response data structure
// @Description User response payload
type UserResponse struct {
	// @Description User ID
	// @Example 123e4567-e89b-12d3-a456-426614174000
	ID string `json:"id"`
	// @Description User full name
	// @Example John Doe
	Name string `json:"name"`
	// @Description User email address
	// @Example john.doe@example.com
	Email string `json:"email"`
	// @Description User role
	// @Example admin
	Role model.UserRole `json:"role"`
	// @Description User creation timestamp
	// @Example 2024-03-15T10:00:00Z
	CreatedAt time.Time `json:"created_at"`
	// @Description User last update timestamp
	// @Example 2024-03-15T15:30:00Z
	UpdatedAt time.Time `json:"updated_at"`
	// @Description User deletion timestamp
	// @Example 2024-03-15T16:00:00Z
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

// UserListResponse represents a list of users response
// @Description List of users response payload
type UserListResponse struct {
	// @Description Array of user objects
	Users []UserResponse `json:"users"`
	// @Description Total number of users
	// @Example 42
	Total int64 `json:"total"`
}

type SwaggerPaginationResponse struct {
	response.Response[pagination.PaginationResponse[UserResponse]]
}
