package db

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/matthewyoungbar/swim-attendance-app/internal/models"
)

func (c *Client) GetAttendance(ctx context.Context, practiceID string) (*models.Attendance, error) {
	out, err := c.ddb.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(c.table),
		Key: map[string]types.AttributeValue{
			"pk": &types.AttributeValueMemberS{Value: "ATTENDANCE#" + practiceID},
			"sk": &types.AttributeValueMemberS{Value: "ATTENDANCE"},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("get attendance: %w", err)
	}
	if out.Item == nil {
		return nil, nil
	}
	var a models.Attendance
	if err := attributevalue.UnmarshalMap(out.Item, &a); err != nil {
		return nil, fmt.Errorf("unmarshal attendance: %w", err)
	}
	return &a, nil
}

func (c *Client) SaveAttendance(ctx context.Context, a models.Attendance) error {
	a.PK = "ATTENDANCE#" + a.PracticeID
	a.SK = "ATTENDANCE"
	a.RecordType = models.RecordTypeAttendance
	a.UpdatedAt = time.Now().UTC()
	if a.Attendees == nil {
		a.Attendees = []models.AttendeeEntry{}
	}
	item, err := attributevalue.MarshalMap(a)
	if err != nil {
		return fmt.Errorf("marshal attendance: %w", err)
	}
	_, err = c.ddb.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(c.table),
		Item:      item,
	})
	return err
}
