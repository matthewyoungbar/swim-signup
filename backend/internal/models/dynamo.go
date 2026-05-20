package models

import "time"

const PracticeSK = "PRACTICE"

type Practice struct {
	PK          string    `json:"-" dynamodbav:"pk"`
	SK          string    `json:"-" dynamodbav:"sk"`
	ID          string    `json:"id" dynamodbav:"id"`
	Title       string    `json:"title" dynamodbav:"title"`
	Description string    `json:"description" dynamodbav:"description"`
	Location    string    `json:"location" dynamodbav:"location"`
	StartTime   time.Time `json:"startTime" dynamodbav:"startTime"`
	EndTime     time.Time `json:"endTime" dynamodbav:"endTime"`
	Capacity    int       `json:"capacity" dynamodbav:"capacity"`
	SignupCount int       `json:"signupCount" dynamodbav:"signupCount"`
	TTL         int64     `json:"-" dynamodbav:"ttl"`
}

type Signup struct {
	PK           string    `json:"-" dynamodbav:"pk"`
	SK           string    `json:"-" dynamodbav:"sk"`
	PracticeID   string    `json:"practiceId" dynamodbav:"practiceId"`
	SwimmerEmail string    `json:"swimmerEmail" dynamodbav:"swimmerEmail"`
	SwimmerName  string    `json:"swimmerName" dynamodbav:"swimmerName"`
	RegisteredAt time.Time `json:"registeredAt" dynamodbav:"registeredAt"`
	Notes        string    `json:"notes,omitempty" dynamodbav:"notes,omitempty"`
}