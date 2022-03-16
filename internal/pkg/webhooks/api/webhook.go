package api

type Webhook struct {
	Token      string
	Table      string
	Hash       string
	Conditions struct {
		Count int
		Time  string
		Size  string
	}
}

type RegisteredWebhooks map[string]Webhook
