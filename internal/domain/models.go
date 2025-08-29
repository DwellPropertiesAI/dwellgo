package domain

import (
	"time"

	"github.com/google/uuid"
)

// Base entity with common fields
type BaseEntity struct {
	ID        uuid.UUID `json:"id" db:"id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// UserClaims represents JWT token claims
type UserClaims struct {
	UserID     string     `json:"user_id"`
	UserType   string     `json:"user_type"`
	LandlordID *uuid.UUID `json:"landlord_id,omitempty"`
	ExpiresAt  time.Time  `json:"expires_at"`
}

// UserInfo represents user information from Cognito
type UserInfo struct {
	UserID     string     `json:"user_id"`
	UserType   string     `json:"user_type"`
	LandlordID *uuid.UUID `json:"landlord_id,omitempty"`
}

// Landlord represents a property owner/manager
type Landlord struct {
	BaseEntity
	Email           string `json:"email" db:"email"`
	FirstName       string `json:"first_name" db:"first_name"`
	LastName        string `json:"last_name" db:"last_name"`
	Phone           string `json:"phone" db:"phone"`
	CompanyName     string `json:"company_name" db:"company_name"`
	BusinessAddress string `json:"business_address" db:"business_address"`
	TaxID           string `json:"tax_id" db:"tax_id"`
	IsActive        bool   `json:"is_active" db:"is_active"`
}

// Tenant represents a property renter
type Tenant struct {
	BaseEntity
	LandlordID      uuid.UUID `json:"landlord_id" db:"landlord_id"`
	Email           string    `json:"email" db:"email"`
	FirstName       string    `json:"first_name" db:"first_name"`
	LastName        string    `json:"last_name" db:"last_name"`
	Phone           string    `json:"phone" db:"phone"`
	EmergencyContact string   `json:"emergency_contact" db:"emergency_contact"`
	LeaseStartDate  time.Time `json:"lease_start_date" db:"lease_start_date"`
	LeaseEndDate    time.Time `json:"lease_end_date" db:"lease_end_date"`
	MonthlyRent     float64   `json:"monthly_rent" db:"monthly_rent"`
	SecurityDeposit float64   `json:"security_deposit" db:"security_deposit"`
	IsActive        bool      `json:"is_active" db:"is_active"`
}

// Property represents a real estate property
type Property struct {
	BaseEntity
	LandlordID      uuid.UUID `json:"landlord_id" db:"landlord_id"`
	Name            string    `json:"name" db:"name"`
	Address         string    `json:"address" db:"address"`
	City            string    `json:"city" db:"city"`
	State           string    `json:"state" db:"state"`
	ZipCode         string    `json:"zip_code" db:"zip_code"`
	PropertyType    string    `json:"property_type" db:"property_type"` // apartment, house, commercial, etc.
	Bedrooms        int       `json:"bedrooms" db:"bedrooms"`
	Bathrooms       int       `json:"bathrooms" db:"bathrooms"`
	SquareFootage   int       `json:"square_footage" db:"square_footage"`
	YearBuilt       int       `json:"year_built" db:"year_built"`
	MonthlyRent     float64   `json:"monthly_rent" db:"monthly_rent"`
	SecurityDeposit float64   `json:"security_deposit" db:"security_deposit"`
	IsAvailable     bool      `json:"is_available" db:"is_available"`
	CurrentTenantID *uuid.UUID `json:"current_tenant_id,omitempty" db:"current_tenant_id"`
}

// Contractor represents a service provider
type Contractor struct {
	BaseEntity
	LandlordID      uuid.UUID `json:"landlord_id" db:"landlord_id"`
	CompanyName     string    `json:"company_name" db:"company_name"`
	ContactPerson   string    `json:"contact_person" db:"contact_person"`
	Email           string    `json:"email" db:"email"`
	Phone           string    `json:"phone" db:"phone"`
	Specialization  string    `json:"specialization" db:"specialization"` // plumbing, electrical, HVAC, etc.
	LicenseNumber   string    `json:"license_number" db:"license_number"`
	InsuranceInfo   string    `json:"insurance_info" db:"insurance_info"`
	HourlyRate      float64   `json:"hourly_rate" db:"hourly_rate"`
	IsActive        bool      `json:"is_active" db:"is_active"`
}

// MaintenanceRequest represents a maintenance issue
type MaintenanceRequest struct {
	BaseEntity
	LandlordID      uuid.UUID `json:"landlord_id" db:"landlord_id"`
	PropertyID      uuid.UUID `json:"property_id" db:"property_id"`
	TenantID        uuid.UUID `json:"tenant_id" db:"tenant_id"`
	Title           string    `json:"title" db:"title"`
	Description     string    `json:"description" db:"description"`
	Priority        string    `json:"priority" db:"priority"` // low, medium, high, emergency
	Status          string    `json:"status" db:"status"`     // open, in_progress, completed, cancelled
	Category        string    `json:"category" db:"category"` // plumbing, electrical, HVAC, structural, etc.
	RequestedDate   time.Time `json:"requested_date" db:"requested_date"`
	CompletedDate   *time.Time `json:"completed_date,omitempty" db:"completed_date"`
	EstimatedCost   *float64  `json:"estimated_cost,omitempty" db:"estimated_cost"`
	ActualCost      *float64  `json:"actual_cost,omitempty" db:"actual_cost"`
	ContractorID    *uuid.UUID `json:"contractor_id,omitempty" db:"contractor_id"`
	Notes           string    `json:"notes" db:"notes"`
}

// MaintenancePhoto represents photos attached to maintenance requests
type MaintenancePhoto struct {
	BaseEntity
	MaintenanceRequestID uuid.UUID `json:"maintenance_request_id" db:"maintenance_request_id"`
	PhotoURL            string    `json:"photo_url" db:"photo_url"`
	PhotoKey            string    `json:"photo_key" db:"photo_key"` // S3 key
	Description         string    `json:"description" db:"description"`
	IsBefore            bool      `json:"is_before" db:"is_before"` // before/after photo indicator
}

// Payment represents rent or other payments
type Payment struct {
	BaseEntity
	LandlordID      uuid.UUID `json:"landlord_id" db:"landlord_id"`
	PropertyID      uuid.UUID `json:"property_id" db:"property_id"`
	TenantID        uuid.UUID `json:"tenant_id" db:"tenant_id"`
	Amount          float64   `json:"amount" db:"amount"`
	PaymentType     string    `json:"payment_type" db:"payment_type"` // rent, security_deposit, late_fee, etc.
	PaymentMethod   string    `json:"payment_method" db:"payment_method"` // bank_transfer, credit_card, cash, etc.
	DueDate         time.Time `json:"due_date" db:"due_date"`
	PaidDate        *time.Time `json:"paid_date,omitempty" db:"paid_date"`
	Status          string    `json:"status" db:"status"` // pending, paid, overdue, cancelled
	ReferenceNumber string    `json:"reference_number" db:"reference_number"`
	Notes           string    `json:"notes" db:"notes"`
}

// AI Chat Message represents a conversation with the AI chatbot
type AIChatMessage struct {
	BaseEntity
	LandlordID      uuid.UUID `json:"landlord_id" db:"landlord_id"`
	TenantID        *uuid.UUID `json:"tenant_id,omitempty" db:"tenant_id"`
	UserType        string    `json:"user_type" db:"user_type"` // landlord, tenant
	Question        string    `json:"question" db:"question"`
	Answer          string    `json:"answer" db:"answer"`
	ModelUsed       string    `json:"model_used" db:"model_used"`
	TokensUsed      int       `json:"tokens_used" db:"tokens_used"`
	Cost            float64   `json:"cost" db:"cost"`
}

// Notification represents system notifications
type Notification struct {
	BaseEntity
	LandlordID      uuid.UUID `json:"landlord_id" db:"landlord_id"`
	RecipientID     uuid.UUID `json:"recipient_id" db:"recipient_id"`
	RecipientType   string    `json:"recipient_type" db:"recipient_type"` // landlord, tenant, contractor
	Type            string    `json:"type" db:"type"`                     // maintenance_request, payment_due, payment_received, etc.
	Title           string    `json:"title" db:"title"`
	Message         string    `json:"message" db:"message"`
	IsRead          bool      `json:"is_read" db:"is_read"`
	ReadAt          *time.Time `json:"read_at,omitempty" db:"read_at"`
	RelatedEntityID *uuid.UUID `json:"related_entity_id,omitempty" db:"related_entity_id"`
	RelatedEntityType string   `json:"related_entity_type,omitempty" db:"related_entity_type"`
}
