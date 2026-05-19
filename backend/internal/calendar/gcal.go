package calendar

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"

	"github.com/yourorg/swim-signup/internal/models"
)

const defaultCapacity = 20

// Client wraps the Google Calendar API.
type Client struct {
	svc        *calendar.Service
	calendarID string
}

// NewClient creates a Google Calendar client.
// Credentials are read from the GOOGLE_CREDENTIALS_JSON env var (service account JSON),
// or from GOOGLE_TOKEN_JSON for OAuth2 user tokens.
// Calendar ID comes from the GOOGLE_CALENDAR_ID env var.
func NewClient(ctx context.Context) (*Client, error) {
	calendarID := os.Getenv("GOOGLE_CALENDAR_ID")
	if calendarID == "" {
		return nil, fmt.Errorf("GOOGLE_CALENDAR_ID env var not set")
	}

	var svc *calendar.Service
	var err error

	// Try service account first (recommended for server-to-server)
	if credsJSON := os.Getenv("GOOGLE_CREDENTIALS_JSON"); credsJSON != "" {
		cfg, e := google.JWTConfigFromJSON([]byte(credsJSON), calendar.CalendarReadonlyScope)
		if e != nil {
			return nil, fmt.Errorf("parse service account credentials: %w", e)
		}
		svc, err = calendar.NewService(ctx, option.WithHTTPClient(cfg.Client(ctx)))
	} else if tokenJSON := os.Getenv("GOOGLE_TOKEN_JSON"); tokenJSON != "" {
		// Fall back to OAuth2 token
		var token oauth2.Token
		if e := json.Unmarshal([]byte(tokenJSON), &token); e != nil {
			return nil, fmt.Errorf("parse oauth token: %w", e)
		}
		oauthCfg := &oauth2.Config{
			Scopes:   []string{calendar.CalendarReadonlyScope},
			Endpoint: google.Endpoint,
		}
		svc, err = calendar.NewService(ctx, option.WithTokenSource(oauthCfg.TokenSource(ctx, &token)))
	} else {
		return nil, fmt.Errorf("no Google credentials found: set GOOGLE_CREDENTIALS_JSON or GOOGLE_TOKEN_JSON")
	}

	if err != nil {
		return nil, fmt.Errorf("create calendar service: %w", err)
	}

	return &Client{svc: svc, calendarID: calendarID}, nil
}

// FetchUpcomingPractices fetches swim practices from Google Calendar for the next N days.
func (c *Client) FetchUpcomingPractices(ctx context.Context, days int) ([]models.Practice, error) {
	now := time.Now().UTC()
	end := now.Add(time.Duration(days) * 24 * time.Hour)

	events, err := c.svc.Events.List(c.calendarID).
		Context(ctx).
		TimeMin(now.Format(time.RFC3339)).
		TimeMax(end.Format(time.RFC3339)).
		SingleEvents(true).
		OrderBy("startTime").
		MaxResults(250).
		Do()
	if err != nil {
		return nil, fmt.Errorf("fetch calendar events: %w", err)
	}

	practices := make([]models.Practice, 0, len(events.Items))
	for _, event := range events.Items {
		p, err := eventToPractice(event)
		if err != nil {
			// Skip malformed events
			continue
		}
		practices = append(practices, p)
	}
	return practices, nil
}

// eventToPractice converts a Google Calendar event to a Practice model.
func eventToPractice(event *calendar.Event) (models.Practice, error) {
	startTime, err := time.Parse(time.RFC3339, event.Start.DateTime)
	if err != nil {
		// Try date-only events
		startTime, err = time.Parse("2006-01-02", event.Start.Date)
		if err != nil {
			return models.Practice{}, fmt.Errorf("parse start time: %w", err)
		}
	}

	endTime, err := time.Parse(time.RFC3339, event.End.DateTime)
	if err != nil {
		endTime, err = time.Parse("2006-01-02", event.End.Date)
		if err != nil {
			return models.Practice{}, fmt.Errorf("parse end time: %w", err)
		}
	}

	capacity := defaultCapacity
	// Optionally parse capacity from description: "Capacity: 25"
	if event.Description != "" {
		var cap int
		if _, e := fmt.Sscanf(event.Description, "Capacity: %d", &cap); e == nil && cap > 0 {
			capacity = cap
		}
	}

	// TTL: expire 7 days after the practice ends
	ttl := endTime.Add(7 * 24 * time.Hour).Unix()

	return models.Practice{
		ID:          event.Id,
		Title:       event.Summary,
		Description: event.Description,
		Location:    event.Location,
		StartTime:   startTime,
		EndTime:     endTime,
		Capacity:    capacity,
		TTL:         ttl,
	}, nil
}
