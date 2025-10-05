package seeders

import (
	"template-golang/pkg/logger"

	"gorm.io/gorm"
)

const (
	EXAMPLE_PDF        = "https://is3.cloudhost.id/uts/brochures/1758074703488556600.pdf"
	EXAMPLE_ALUMNI     = "https://is3.cloudhost.id/uts/example/alumni.jpg"
	EXAMPLE_THUMBNAIL  = "https://is3.cloudhost.id/uts/example/thumbnail.jpg"
	EXAMPLE_FACILITIES = "https://is3.cloudhost.id/uts/example/facilities.jpg"
	EXAMPLE_LOGO       = "https://is3.cloudhost.id/uts/example/logo.png"

	LANG_ID = "id_ID"
	LANG_EN = "en_US"
)

func Seed(db *gorm.DB) error {
	logger.L().Info("seeding database")

	if err := SeedUsers(db); err != nil {
		logger.L().Errorf("failed to seed users: %v", err)
		return err
	}

	logger.L().Info("seeding database completed")
	return nil
}
