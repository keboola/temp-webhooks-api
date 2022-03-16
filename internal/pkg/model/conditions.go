package model

type Conditions struct {
	Count int
	Time  string
	Size  string
}

func NewConditions() Conditions {
	return Conditions{
		Count: 1000,
		Time:  "30s",
		Size:  "10MB",
	}
}
