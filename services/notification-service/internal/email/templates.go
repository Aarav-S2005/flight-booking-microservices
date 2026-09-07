package email

import (
	"bytes"
	"fmt"
	"html/template"
)

type TemplateName string

const (
	TemplateBookingConfirmed     TemplateName = "booking_confirmed"
	TemplateReservationConfirmed TemplateName = "reservation_confirmed"
	TemplatePaymentCompleted     TemplateName = "payment_completed"
	TemplatePaymentFailed        TemplateName = "payment_failed"
)

type emailTemplate struct {
	Subject string
	Body    *template.Template
}

type Templates struct {
	templates map[TemplateName]emailTemplate
}

func NewTemplates() *Templates {
	t := &Templates{templates: map[TemplateName]emailTemplate{}}
	raw := map[TemplateName]struct{ Subject, Body string }{
		TemplateBookingConfirmed:     {"Booking Confirmed", BookingConfirmedEmailTemplate},
		TemplateReservationConfirmed: {"Reservation Confirmed", ReservationConfirmedEmailTemplate},
		TemplatePaymentCompleted:     {"Payment Successful", PaymentCompletedEmailTemplate},
		TemplatePaymentFailed:        {"Payment Failed", PaymentFailedEmailTemplate},
	}
	for name, r := range raw {
		tmpl := template.Must(template.New(string(name)).Parse(r.Body))
		t.templates[name] = emailTemplate{Subject: r.Subject, Body: tmpl}
	}
	return t
}

func (t *Templates) Render(name TemplateName, data any) (string, string, error) {
	et, ok := t.templates[name]
	if !ok {
		return "", "", fmt.Errorf("email template %q not found", name)
	}
	var buf bytes.Buffer
	if err := et.Body.Execute(&buf, data); err != nil {
		return "", "", fmt.Errorf("render email template %q: %w", name, err)
	}
	return et.Subject, buf.String(), nil
}
