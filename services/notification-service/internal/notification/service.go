package notification

import (
	"context"
	"fmt"
	"time"

	"github.com/Aarav-S2005/flight-booking-microservices/services/notification-service/internal/database"
	"github.com/Aarav-S2005/flight-booking-microservices/services/notification-service/internal/email"
	httpclient "github.com/Aarav-S2005/flight-booking-microservices/services/notification-service/internal/http-client"
)

type Service interface {
	Notify(ctx context.Context, job Job) error
}

type service struct {
	authClient httpclient.HttpClient
	sender     email.EmailSender
	templates  TemplateRenderer
	repo       database.Repository
}

func NewService(authClient httpclient.HttpClient, sender email.EmailSender, templates TemplateRenderer, repo database.Repository) Service {
	return &service{
		authClient: authClient,
		sender:     sender,
		templates:  templates,
		repo:       repo,
	}
}

func (s *service) Notify(ctx context.Context, job Job) error {
	userEmail, err := s.authClient.GetEmailFromAuthService(job.UserID)
	if err != nil {
		return fmt.Errorf("fetch email: %w", err)
	}
	subject, body, err := s.templates.Render(job.TemplateName, job.Data)
	if err != nil {
		return fmt.Errorf("render template: %w", err)
	}

	err = s.sender.Send(userEmail, subject, body)
	if err != nil {
		return err
	}
	rec := database.Notification{
		RecipientEmail: userEmail,
		Subject:        subject,
		Body:           body,
		SentAt:         time.Now().UTC(),
		Success:        err == nil,
	}
	if err := s.repo.Save(ctx, rec); err != nil {
		return fmt.Errorf("save notification: %w", err)
	}
	return nil
}
