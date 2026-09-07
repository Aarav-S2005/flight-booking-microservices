package notification

import "github.com/Aarav-S2005/flight-booking-microservices/services/notification-service/internal/email"

type Job struct {
	UserID       string
	TemplateName email.TemplateName
	Data         any
}

type TemplateRenderer interface {
	Render(name email.TemplateName, data any) (subject, body string, err error)
}
