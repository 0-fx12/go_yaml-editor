package model

import "time"

type VNFInstance struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Name      string         `gorm:"size:255;not null" json:"name"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	Defs      []VNFDefinition `gorm:"constraint:OnDelete:CASCADE" json:"-"`
}

type VNFDefinition struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	VNFID           uint      `gorm:"index;not null" json:"vnfId"`
	ParameterName   string    `gorm:"size:255;not null" json:"parameterName"`
	DefaultValue    string    `gorm:"size:1024;not null" json:"defaultValue"`
	DescriptionText string    `gorm:"size:1024;not null" json:"descriptionTxt"`
	Type            string    `gorm:"size:64;not null" json:"type"`
	CanBeUpdated    bool      `gorm:"default:false" json:"canBeUpdated"`
	HiddenCondition string    `gorm:"size:64" json:"hidenCondition"`
	Optional        *bool     `json:"optional"`
	Constraints     string    `gorm:"size:1024" json:"constraints"`
	CurrentValue    string    `gorm:"size:1024" json:"currentValue"`
	Modified        bool      `gorm:"default:false" json:"modified"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}


