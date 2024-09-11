package server

type pingResponse struct {
	Pong string `json:"pong"`
}

type subscribeResponse struct {
	SubscriptionID string `json:"subscription_id"`
}
