package model

// FileResource .
type FileResource struct {
	Id           int    `json:"id" validate:"required"`
	Name         string `json:"name" validate:"required"`
	Url          string `json:"url" validate:"required"`
	Provider     string `json:"provider" validate:"required"`
	Region       string `json:"region" validate:"required"`
	UploadParams struct {
		Key         string `json:"key" validate:"required"`
		Bucket      string `json:"bucket" validate:"required"`
		Acl         string `json:"private" validate:"required"`
		Credentials struct {
			AccessKeyId     string `json:"accessKeyId" validate:"required"`
			SecretAccessKey string `json:"secretAccessKey" validate:"required"`
			SessionToken    string `json:"sessionToken" validate:"required"`
			Expiration      string `json:"expiration" validate:"required"`
		}
	}
}
