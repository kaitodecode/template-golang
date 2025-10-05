package model

// UserRole represents the user_role enum type
type UserRole string

const (
	RoleAdmin      UserRole = "admin"
	RoleSuperAdmin UserRole = "superadmin"
)

// User represents the users table in the database
type User struct {
	BaseModel
	LearningPointID *string       `json:"learning_point_id" gorm:"type:varchar(25);default:null"`
	Name           string         `json:"name" gorm:"type:varchar(255);not null"`
	Email          string         `json:"email" gorm:"type:varchar(100);not null;uniqueIndex:idx_users_email"`
	Password       string         `json:"password" gorm:"type:varchar(255);not null"`
	Role           UserRole       `json:"role" gorm:"type:user_role;not null;default:'admin'"`

}

// TableName specifies the table name for User model
func (User) TableName() string {
	return "users"
}