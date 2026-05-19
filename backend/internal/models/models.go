package models

import "time"

// Practice represents a swim practice fetched from Google Calendar
type Practice struct {
	ID          string    `json:"id" dynamodbav:"id"`
	Title       string    `json:"title" dynamodbav:"title"`
	Description string    `json:"description" dynamodbav:"description"`
	Location    string    `json:"location" dynamodbav:"location"`
	StartTime   time.Time `json:"startTime" dynamodbav:"startTime"`
	EndTime     time.Time `json:"endTime" dynamodbav:"endTime"`
	Capacity    int       `json:"capacity" dynamodbav:"capacity"`
	SignupCount int       `json:"signupCount" dynamodbav:"signupCount"`
	// TTL for DynamoDB auto-expiry (Unix timestamp)
	TTL int64 `json:"-" dynamodbav:"ttl"`
}

// Signup represents a swimmer's registration for a practice
type Signup struct {
	// Partition key: practiceId, Sort key: swimmerEmail
	PracticeID  string    `json:"practiceId" dynamodbav:"practiceId"`
	SwimmerEmail string   `json:"swimmerEmail" dynamodbav:"swimmerEmail"`
	SwimmerName string    `json:"swimmerName" dynamodbav:"swimmerName"`
	RegisteredAt time.Time `json:"registeredAt" dynamodbav:"registeredAt"`
	Notes       string    `json:"notes,omitempty" dynamodbav:"notes,omitempty"`
}

// SignupRequest is the request body for POST /signups
type SignupRequest struct {
	PracticeID   string `json:"practiceId"`
	SwimmerEmail string `json:"swimmerEmail"`
	SwimmerName  string `json:"swimmerName"`
	Notes        string `json:"notes,omitempty"`
}

// APIResponse wraps all API responses
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// PracticeWithSignups bundles a practice with the current user's signup status
type PracticeWithSignups struct {
	Practice
	IsSignedUp bool     `json:"isSignedUp"`
	Signups    []Signup `json:"signups,omitempty"` // admin only
}
