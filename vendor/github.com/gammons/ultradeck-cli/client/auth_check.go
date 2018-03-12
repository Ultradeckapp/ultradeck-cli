package client

import "encoding/json"

type AuthCheck struct{}

type AuthCheckResponse struct {
	IsSignedIn       bool   `json:"is_signed_in"`
	Name             string `json:"name"`
	Username         string `json:"username"`
	ImageUrl         string `json:"image_url"`
	Email            string `json:"email"`
	SubscriptionName string `json:"subscriptionName"`
	Token            string
}

func (a *AuthCheck) CheckAuth(token string) *AuthCheckResponse {
	httpClient := NewHttpClient(token)
	bodyBytes := httpClient.GetRequest("api/v1/auth/me")

	resp := &AuthCheckResponse{}
	if err := json.Unmarshal(bodyBytes, &resp); err != nil {
		panic(err)
	}
	return resp
}
