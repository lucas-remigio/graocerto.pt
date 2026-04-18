package notification

import (
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/SherClockHolmes/webpush-go"
	"github.com/lucas-remigio/wallet-tracker/config"
	"github.com/lucas-remigio/wallet-tracker/types"
)

type PushService struct {
	publicKey  string
	privateKey string
}

func NewPushService(publicKey, privateKey string) *PushService {
	return &PushService{
		publicKey:  publicKey,
		privateKey: privateKey,
	}
}

type PushNotificationPayload struct {
	Title string `json:"title"`
	Body  string `json:"body"`
	Icon  string `json:"icon"`
	Data  any    `json:"data"`
}

func (s *PushService) SendPush(sub *types.PushSubscription, payload PushNotificationPayload) error {
	if s.publicKey == "" || s.privateKey == "" {
		return fmt.Errorf("VAPID keys not configured")
	}

	slog.Info("sending push notification", "endpoint", sub.Endpoint)

	// Decode subscription info
	sSub := &webpush.Subscription{
		Endpoint: sub.Endpoint,
		Keys: webpush.Keys{
			P256dh: sub.P256dh,
			Auth:   sub.Auth,
		},
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal push payload: %w", err)
	}

	// Send notification
	resp, err := webpush.SendNotification(payloadBytes, sSub, &webpush.Options{
		Subscriber:      config.Envs.SMTPFromEmail,
		VAPIDPublicKey:  s.publicKey,
		VAPIDPrivateKey: s.privateKey,
		TTL:             86400, // 24 hours
	})
	if err != nil {
		return fmt.Errorf("failed to send push notification: %w", err)
	}
	defer resp.Body.Close()

	slog.Info("push notification response", "status", resp.StatusCode)

	if resp.StatusCode >= 400 {
		return fmt.Errorf("push notification failed with status: %d", resp.StatusCode)
	}

	return nil
}

func (s *PushService) NotifyUser(userID int, store types.NotificationStore, payload PushNotificationPayload) {
	subs, err := store.GetPushSubscriptionsByUserID(userID)
	if err != nil {
		slog.Error("failed to get push subscriptions for user", "userID", userID, "error", err)
		return
	}

	slog.Info("notifying user via push", "userID", userID, "subsCount", len(subs))

	for _, sub := range subs {
		if err := s.SendPush(sub, payload); err != nil {
			slog.Error("failed to send push notification to subscription", "userID", userID, "endpoint", sub.Endpoint, "error", err)
		}
	}
}
