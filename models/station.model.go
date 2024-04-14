package models

import (
	"time"

	"github.com/google/uuid"
)

type Station struct {
	ID        uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primary_key" json:"id,omitempty"`
	Name      string    `gorm:"not null" json:"name,omitempty"`
	LatLong   string    `gorm:"not null" json:"lat_long,omitempty"`
	UserId    uuid.UUID `gorm:"not null" json:"user_id,omitempty"`
	CreatedAt time.Time `gorm:"not null" json:"created_at,omitempty"`
	UpdatedAt time.Time `gorm:"not null" json:"updated_at,omitempty"`
}

type CreateStationRequest struct {
	Name      string    `json:"title"  binding:"required"`
	LatLong   string    `json:"lat_long,omitempty"`
	UserId    string    `json:"user_id,omitempty"`
	CreatedAt time.Time `json:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}

type UpdateStation struct {
	Name      string    `json:"title"  binding:"required"`
	LatLong   string    `json:"lat_long,omitempty"`
	UserId    string    `json:"user_id,omitempty"`
	CreatedAt time.Time `json:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}
