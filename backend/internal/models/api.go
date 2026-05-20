package models

type SignupRequest struct {
	PracticeID   string `json:"practiceId"`
	SwimmerEmail string `json:"swimmerEmail"`
	SwimmerName  string `json:"swimmerName"`
	Notes        string `json:"notes,omitempty"`
}

type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

type PracticeWithSignups struct {
	Practice
	IsSignedUp bool     `json:"isSignedUp"`
	Signups    []Signup `json:"signups,omitempty"`
}