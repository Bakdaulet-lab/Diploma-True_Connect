package provider

import (
	"context"
	"fmt"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"
)

// PushProvider defines the interface for sending push notifications.
type PushProvider interface {
	SendTargeted(ctx context.Context, fcmToken, title, body string, data map[string]string) error
}

// FCMPushProvider implements PushProvider using Firebase Cloud Messaging.
type FCMPushProvider struct {
	client *messaging.Client
}

// NewFCMPushProvider creates a new FCM push provider.
// Provide an empty credentials file path to fall back to Google Application Default Credentials.
func NewFCMPushProvider(ctx context.Context, credentialsFile string) (*FCMPushProvider, error) {
	var opts []option.ClientOption
	if credentialsFile != "" {
		opts = append(opts, option.WithCredentialsFile(credentialsFile))
	}

	app, err := firebase.NewApp(ctx, nil, opts...)
	if err != nil {
		return nil, fmt.Errorf("initializing firebase app: %w", err)
	}

	client, err := app.Messaging(ctx)
	if err != nil {
		return nil, fmt.Errorf("initializing firebase messaging client: %w", err)
	}

	return &FCMPushProvider{client: client}, nil
}

func (p *FCMPushProvider) SendTargeted(ctx context.Context, fcmToken, title, body string, data map[string]string) error {
	msg := &messaging.Message{
		Token: fcmToken,
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Data: data,
	}

	_, err := p.client.Send(ctx, msg)
	if err != nil {
		return fmt.Errorf("sending fcm message: %w", err)
	}

	return nil
}

// MockPushProvider simulates pushing notifications to console.

type MockPushProvider struct {
}

func NewMockPushProvider() *MockPushProvider {
	return &MockPushProvider{}
}

func (m *MockPushProvider) SendTargeted(ctx context.Context, fcmToken, title, body string, data map[string]string) error {
	fmt.Printf("MOCK PUSH [Token: %s]: %s - %s (Data: %v)\n", fcmToken, title, body, data)
	return nil
}
