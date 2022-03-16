package model

// File resource.
type FileResource struct {
	Id           int    `json:"id" validate:"required"`
	Name         int    `json:"name" validate:"required"`
	Url          string `json:"url" validate:"required"`
	Provider     string `json:"provider" validate:"required"`
	Region       string `json:"region" validate:"required"`
	UploadParams struct {
		Key         int `json:"key" validate:"required"`
		Bucket      int `json:"bucket" validate:"required"`
		Acl         int `json:"private" validate:"required"`
		Credentials struct {
			AccessKeyId     string `json:"provider" validate:"required"`
			SecretAccessKey string `json:"secretAccessKey" validate:"required"`
			SessionToken    string `json:"sessionToken" validate:"required"`
			Expiration      string `json:"expiration" validate:"required"`
		}
	}
}
