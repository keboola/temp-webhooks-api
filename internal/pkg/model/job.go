package model

type JobError struct {
	Message string `json:"message" validate:"required"`
}

// Job - Storage API job.
type Job struct {
	Id      int                    `json:"id" validate:"required"`
	Error   JobError               `json:"error" validate:"required"`
	Status  string                 `json:"status" validate:"required"`
	Url     string                 `json:"url" validate:"required"`
	Results map[string]interface{} `json:"results"`
}
