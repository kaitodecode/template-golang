package model

import (
	"time"

	"github.com/nrednav/cuid2"
	"gorm.io/gorm"
)

type BaseModel struct {
	ID        string    `json:"id" gorm:"primaryKey;type:varchar(255)"`
	CreatedAt time.Time `json:"created_at" gorm:"type:timestamptz;default:now()"`
	UpdatedAt time.Time `json:"updated_at" gorm:"type:timestamptz;default:now()"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"type:timestamptz" swaggerignore:"true"`
}

func (b *BaseModel) BeforeCreate(tx *gorm.DB) error {
	if b.ID == "" {
		b.ID = cuid2.Generate()
	}
	return nil
}
