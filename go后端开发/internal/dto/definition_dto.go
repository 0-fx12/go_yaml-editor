package dto

type DefinitionCreateRequest struct {
	ParameterName   string  `json:"parameterName" binding:"required"`
	DefaultValue    string  `json:"defaultValue" binding:"required"`
	DescriptionText string  `json:"descriptionTxt" binding:"required"`
	Type            string  `json:"type" binding:"required"`
	CanBeUpdated    bool    `json:"canBeUpdated"`
	HiddenCondition string  `json:"hidenCondition"`
	Optional        *bool   `json:"optional"`
	Constraints     string  `json:"constraints"`
	CurrentValue    *string `json:"currentValue"`
}

type DefinitionUpdateRequest struct {
	DefaultValue    *string `json:"defaultValue"`
	DescriptionText *string `json:"descriptionTxt"`
	Type            *string `json:"type"`
	CanBeUpdated    *bool   `json:"canBeUpdated"`
	HiddenCondition *string `json:"hidenCondition"`
	Optional        *bool   `json:"optional"`
	Constraints     *string `json:"constraints"`
	CurrentValue    *string `json:"currentValue"`
}


