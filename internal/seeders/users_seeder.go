package seeders

import (
	"fmt"

	"template-golang/internal/db/model"
	"template-golang/pkg/logger"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func SeedUsers(db *gorm.DB) error {
	pass, err := hashPassword("admin")
	if err != nil {
		logger.L().Errorf("failed to hash password: %v", err)
		return err
	}

	user := model.User{
		Name:     "admin",
		Password: pass,
		Email:    "admin@gmail.com",
		Role:     "superadmin",
	}

	err = db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&user).Error; err != nil {
			logger.L().Errorf("failed to create user: %v", err)
			return err
		}
		return nil
	})

	if err != nil {
		logger.L().Errorf("failed to seed users: %v", err)
		return err
	}

	return nil
}

func hashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hashedBytes), nil
}
