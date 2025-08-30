package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"dwell/internal/aws"
	"dwell/internal/config"
	"dwell/internal/domain"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/aws/aws-sdk-go-v2/service/ses/types"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	snstypes "github.com/aws/aws-sdk-go-v2/service/sns/types"
	"github.com/google/uuid"
)

type NotificationService struct {
	awsClients *aws.Clients
	config     *config.Config
}

type NotificationRequest struct {
	Type              string     `json:"type" binding:"required"`
	Title             string     `json:"title" binding:"required"`
	Message           string     `json:"message" binding:"required"`
	LandlordID        string     `json:"landlord_id" binding:"required"`
	RecipientID       string     `json:"recipient_id" binding:"required"`
	RecipientType     string     `json:"recipient_type" binding:"required,oneof=landlord tenant contractor"`
	RecipientEmail    string     `json:"recipient_email" binding:"required,email"`
	RecipientPhone    string     `json:"recipient_phone,omitempty"`
	RelatedEntityID   *uuid.UUID `json:"related_entity_id,omitempty"`
	RelatedEntityType string     `json:"related_entity_type,omitempty"`
	Priority          string     `json:"priority,omitempty"` // low, medium, high, urgent
}

type NotificationResponse struct {
	NotificationID string    `json:"notification_id"`
	Status         string    `json:"status"`
	SentAt         time.Time `json:"sent_at"`
	Channel        string    `json:"channel"` // email, sms, push
}

type EmailTemplate struct {
	Subject   string            `json:"subject"`
	HTMLBody  string            `json:"html_body"`
	TextBody  string            `json:"text_body"`
	Variables map[string]string `json:"variables"`
}

type SMSTemplate struct {
	Message   string            `json:"message"`
	Variables map[string]string `json:"variables"`
}

func NewNotificationService(awsClients *aws.Clients, config *config.Config) *NotificationService {
	return &NotificationService{
		awsClients: awsClients,
		config:     config,
	}
}

// SendNotification sends a notification through the appropriate channel
func (s *NotificationService) SendNotification(ctx context.Context, req *NotificationRequest) (*NotificationResponse, error) {
	// Create notification record
	_ = &domain.Notification{
		LandlordID:        uuid.MustParse(req.LandlordID),
		RecipientID:       uuid.MustParse(req.RecipientID),
		RecipientType:     req.RecipientType,
		Type:              req.Type,
		Title:             req.Title,
		Message:           req.Message,
		RelatedEntityID:   req.RelatedEntityID,
		RelatedEntityType: req.RelatedEntityType,
		IsRead:            false,
	}

	// TODO: Save notification to database
	// This would typically be done through a repository layer

	// Send notification based on priority and recipient type
	var response *NotificationResponse
	var err error

	switch req.Priority {
	case "urgent":
		// Send both email and SMS for urgent notifications
		response, err = s.sendUrgentNotification(ctx, req)
	case "high":
		// Send email and optionally SMS
		response, err = s.sendHighPriorityNotification(ctx, req)
	default:
		// Send email only for regular notifications
		response, err = s.sendEmailNotification(ctx, req)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to send notification: %w", err)
	}

	return response, nil
}

// sendEmailNotification sends an email notification using AWS SES
func (s *NotificationService) sendEmailNotification(ctx context.Context, req *NotificationRequest) (*NotificationResponse, error) {
	// Get email template
	template := s.getEmailTemplate(req.Type, req)

	// Replace variables in template
	subject := s.replaceVariables(template.Subject, template.Variables)
	htmlBody := s.replaceVariables(template.HTMLBody, template.Variables)
	textBody := s.replaceVariables(template.TextBody, template.Variables)

	// Prepare SES email input
	emailInput := &ses.SendEmailInput{
		Source: awssdk.String(s.config.AWS.SES.FromEmail),
		Destination: &types.Destination{
			ToAddresses: []string{req.RecipientEmail},
		},
		Message: &types.Message{
			Subject: &types.Content{
				Data:    awssdk.String(subject),
				Charset: awssdk.String("UTF-8"),
			},
			Body: &types.Body{
				Html: &types.Content{
					Data:    awssdk.String(htmlBody),
					Charset: awssdk.String("UTF-8"),
				},
				Text: &types.Content{
					Data:    awssdk.String(textBody),
					Charset: awssdk.String("UTF-8"),
				},
			},
		},
	}

	// Send email
	_, err := s.awsClients.GetSESClient().SendEmail(ctx, emailInput)
	if err != nil {
		return nil, fmt.Errorf("failed to send email: %w", err)
	}

	return &NotificationResponse{
		NotificationID: uuid.New().String(),
		Status:         "sent",
		SentAt:         time.Now(),
		Channel:        "email",
	}, nil
}

// sendSMSNotification sends an SMS notification using AWS SNS
func (s *NotificationService) sendSMSNotification(ctx context.Context, req *NotificationRequest) (*NotificationResponse, error) {
	if req.RecipientPhone == "" {
		return nil, fmt.Errorf("recipient phone number is required for SMS notifications")
	}

	// Get SMS template
	template := s.getSMSTemplate(req.Type, req)

	// Replace variables in template
	message := s.replaceVariables(template.Message, template.Variables)

	// Prepare SNS SMS input
	smsInput := &sns.PublishInput{
		Message:     awssdk.String(message),
		PhoneNumber: awssdk.String(req.RecipientPhone),
		MessageAttributes: map[string]snstypes.MessageAttributeValue{
			"AWS.SNS.SMS.SMSType": {
				DataType:    awssdk.String("String"),
				StringValue: awssdk.String("Transactional"),
			},
		},
	}

	// Send SMS
	_, err := s.awsClients.GetSNSClient().Publish(ctx, smsInput)
	if err != nil {
		return nil, fmt.Errorf("failed to send SMS: %w", err)
	}

	return &NotificationResponse{
		NotificationID: uuid.New().String(),
		Status:         "sent",
		SentAt:         time.Now(),
		Channel:        "sms",
	}, nil
}

// sendUrgentNotification sends both email and SMS for urgent notifications
func (s *NotificationService) sendUrgentNotification(ctx context.Context, req *NotificationRequest) (*NotificationResponse, error) {
	// Send email first
	emailResp, err := s.sendEmailNotification(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to send urgent email: %w", err)
	}

	// Send SMS if phone number is available
	if req.RecipientPhone != "" {
		_, err := s.sendSMSNotification(ctx, req)
		if err != nil {
			// Log SMS failure but don't fail the entire operation
			// TODO: Log error
		}
	}

	return emailResp, nil
}

// sendHighPriorityNotification sends email and optionally SMS
func (s *NotificationService) sendHighPriorityNotification(ctx context.Context, req *NotificationRequest) (*NotificationResponse, error) {
	// Send email
	emailResp, err := s.sendEmailNotification(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to send high priority email: %w", err)
	}

	// Send SMS for high priority if phone number is available and it's a critical type
	if req.RecipientPhone != "" && s.isCriticalNotificationType(req.Type) {
		_, err := s.sendSMSNotification(ctx, req)
		if err != nil {
			// Log SMS failure but don't fail the entire operation
			// TODO: Log error
		}
	}

	return emailResp, nil
}

// isCriticalNotificationType determines if a notification type is critical enough for SMS
func (s *NotificationService) isCriticalNotificationType(notificationType string) bool {
	criticalTypes := map[string]bool{
		"maintenance_emergency": true,
		"payment_overdue":       true,
		"lease_violation":       true,
		"property_damage":       true,
		"security_breach":       true,
	}
	return criticalTypes[notificationType]
}

// getEmailTemplate returns the appropriate email template for the notification type
func (s *NotificationService) getEmailTemplate(notificationType string, req *NotificationRequest) *EmailTemplate {
	// Base template variables
	variables := map[string]string{
		"recipient_name": req.RecipientType,
		"landlord_name":  "Property Management",
		"date":           time.Now().Format("January 2, 2006"),
		"time":           time.Now().Format("3:04 PM"),
	}

	switch notificationType {
	case "maintenance_request":
		return &EmailTemplate{
			Subject: "New Maintenance Request - {{title}}",
			HTMLBody: `
				<!DOCTYPE html>
				<html>
				<head>
					<style>
						body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
						.container { max-width: 600px; margin: 0 auto; padding: 20px; }
						.header { background-color: #f8f9fa; padding: 20px; border-radius: 5px; }
						.content { padding: 20px; }
						.button { display: inline-block; padding: 10px 20px; background-color: #007bff; color: white; text-decoration: none; border-radius: 5px; }
					</style>
				</head>
				<body>
					<div class="container">
						<div class="header">
							<h2>Maintenance Request</h2>
						</div>
						<div class="content">
							<p>Hello {{recipient_name}},</p>
							<p>A new maintenance request has been submitted:</p>
							<h3>{{title}}</h3>
							<p>{{message}}</p>
							<p><strong>Priority:</strong> {{priority}}</p>
							<p><strong>Category:</strong> {{category}}</p>
							<p><strong>Date:</strong> {{date}} at {{time}}</p>
							<p>Please review and take appropriate action.</p>
						</div>
					</div>
				</body>
				</html>`,
			TextBody: `
Maintenance Request

Hello {{recipient_name}},

A new maintenance request has been submitted:

{{title}}

{{message}}

Priority: {{priority}}
Category: {{category}}
Date: {{date}} at {{time}}

Please review and take appropriate action.`,
			Variables: variables,
		}

	case "payment_due":
		return &EmailTemplate{
			Subject: "Payment Due Reminder",
			HTMLBody: `
				<!DOCTYPE html>
				<html>
				<head>
					<style>
						body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
						.container { max-width: 600px; margin: 0 auto; padding: 20px; }
						.header { background-color: #fff3cd; padding: 20px; border-radius: 5px; border: 1px solid #ffeaa7; }
						.content { padding: 20px; }
						.amount { font-size: 24px; font-weight: bold; color: #d63031; }
					</style>
				</head>
				<body>
					<div class="container">
						<div class="header">
							<h2>Payment Due Reminder</h2>
						</div>
						<div class="content">
							<p>Hello {{recipient_name}},</p>
							<p>This is a friendly reminder that your payment is due:</p>
							<p class="amount">Amount: ${{amount}}</p>
							<p><strong>Due Date:</strong> {{due_date}}</p>
							<p><strong>Property:</strong> {{property_name}}</p>
							<p>Please ensure your payment is submitted on time to avoid any late fees.</p>
						</div>
					</div>
				</body>
				</html>`,
			TextBody: `
Payment Due Reminder

Hello {{recipient_name}},

This is a friendly reminder that your payment is due:

Amount: ${{amount}}
Due Date: {{due_date}}
Property: {{property_name}}

Please ensure your payment is submitted on time to avoid any late fees.`,
			Variables: variables,
		}

	default:
		// Generic template
		return &EmailTemplate{
			Subject: "{{title}}",
			HTMLBody: `
				<!DOCTYPE html>
				<html>
				<head>
					<style>
						body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
						.container { max-width: 600px; margin: 0 auto; padding: 20px; }
						.header { background-color: #f8f9fa; padding: 20px; border-radius: 5px; }
						.content { padding: 20px; }
					</style>
				</head>
				<body>
					<div class="container">
						<div class="header">
							<h2>{{title}}</h2>
						</div>
						<div class="content">
							<p>Hello {{recipient_name}},</p>
							<p>{{message}}</p>
							<p>Date: {{date}} at {{time}}</p>
						</div>
					</div>
				</body>
				</html>`,
			TextBody: `
{{title}}

Hello {{recipient_name}},

{{message}}

Date: {{date}} at {{time}}`,
			Variables: variables,
		}
	}
}

// getSMSTemplate returns the appropriate SMS template for the notification type
func (s *NotificationService) getSMSTemplate(notificationType string, req *NotificationRequest) *SMSTemplate {
	variables := map[string]string{
		"recipient_name": req.RecipientType,
	}

	switch notificationType {
	case "maintenance_emergency":
		return &SMSTemplate{
			Message:   "URGENT: Emergency maintenance request at {{property_name}}. Please respond immediately.",
			Variables: variables,
		}
	case "payment_overdue":
		return &SMSTemplate{
			Message:   "Payment overdue: ${{amount}} due for {{property_name}}. Please contact us immediately.",
			Variables: variables,
		}
	default:
		return &SMSTemplate{
			Message:   "{{title}}: {{message}}",
			Variables: variables,
		}
	}
}

// replaceVariables replaces template variables with actual values
func (s *NotificationService) replaceVariables(template string, variables map[string]string) string {
	result := template
	for key, value := range variables {
		placeholder := "{{" + key + "}}"
		result = strings.ReplaceAll(result, placeholder, value)
	}
	return result
}

// SendBulkNotifications sends notifications to multiple recipients
func (s *NotificationService) SendBulkNotifications(ctx context.Context, requests []NotificationRequest) ([]NotificationResponse, error) {
	var responses []NotificationResponse

	for _, req := range requests {
		resp, err := s.SendNotification(ctx, &req)
		if err != nil {
			// Log error but continue with other notifications
			// TODO: Log error
			continue
		}
		responses = append(responses, *resp)
	}

	return responses, nil
}

// SendMaintenanceNotification sends a notification about a maintenance request
func (s *NotificationService) SendMaintenanceNotification(ctx context.Context, maintenanceReq *domain.MaintenanceRequest, recipientType, recipientEmail, recipientPhone string) error {
	req := &NotificationRequest{
		Type:              "maintenance_request",
		Title:             maintenanceReq.Title,
		Message:           maintenanceReq.Description,
		LandlordID:        maintenanceReq.LandlordID.String(),
		RecipientID:       maintenanceReq.TenantID.String(), // This would need to be adjusted based on recipient
		RecipientType:     recipientType,
		RecipientEmail:    recipientEmail,
		RecipientPhone:    recipientPhone,
		RelatedEntityID:   &maintenanceReq.ID,
		RelatedEntityType: "maintenance_request",
		Priority:          maintenanceReq.Priority,
	}

	_, err := s.SendNotification(ctx, req)
	return err
}

// SendPaymentNotification sends a notification about payment
func (s *NotificationService) SendPaymentNotification(ctx context.Context, payment *domain.Payment, recipientType, recipientEmail, recipientPhone string) error {
	req := &NotificationRequest{
		Type:              "payment_due",
		Title:             "Payment Due",
		Message:           fmt.Sprintf("Payment of $%.2f is due for your property", payment.Amount),
		LandlordID:        payment.LandlordID.String(),
		RecipientID:       payment.TenantID.String(),
		RecipientType:     recipientType,
		RecipientEmail:    recipientEmail,
		RecipientPhone:    recipientPhone,
		RelatedEntityID:   &payment.ID,
		RelatedEntityType: "payment",
		Priority:          "medium",
	}

	_, err := s.SendNotification(ctx, req)
	return err
}
