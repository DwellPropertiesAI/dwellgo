-- Dwell Property Management Database Schema
-- Multi-tenant aware with proper indexing for performance

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Landlords table
CREATE TABLE landlords (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE NOT NULL,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    phone VARCHAR(20),
    company_name VARCHAR(255),
    business_address TEXT,
    tax_id VARCHAR(50),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Properties table
CREATE TABLE properties (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    landlord_id UUID NOT NULL REFERENCES landlords(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    address TEXT NOT NULL,
    city VARCHAR(100) NOT NULL,
    state VARCHAR(50) NOT NULL,
    zip_code VARCHAR(20) NOT NULL,
    property_type VARCHAR(50) NOT NULL,
    bedrooms INTEGER,
    bathrooms INTEGER,
    square_footage INTEGER,
    year_built INTEGER,
    monthly_rent DECIMAL(10,2),
    security_deposit DECIMAL(10,2),
    is_available BOOLEAN DEFAULT true,
    current_tenant_id UUID,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Tenants table
CREATE TABLE tenants (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    landlord_id UUID NOT NULL REFERENCES landlords(id) ON DELETE CASCADE,
    email VARCHAR(255) NOT NULL,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    phone VARCHAR(20),
    emergency_contact TEXT,
    lease_start_date DATE NOT NULL,
    lease_end_date DATE NOT NULL,
    monthly_rent DECIMAL(10,2) NOT NULL,
    security_deposit DECIMAL(10,2),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Contractors table
CREATE TABLE contractors (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    landlord_id UUID NOT NULL REFERENCES landlords(id) ON DELETE CASCADE,
    company_name VARCHAR(255) NOT NULL,
    contact_person VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    phone VARCHAR(20),
    specialization VARCHAR(100) NOT NULL,
    license_number VARCHAR(100),
    insurance_info TEXT,
    hourly_rate DECIMAL(10,2),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Maintenance requests table
CREATE TABLE maintenance_requests (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    landlord_id UUID NOT NULL REFERENCES landlords(id) ON DELETE CASCADE,
    property_id UUID NOT NULL REFERENCES properties(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    priority VARCHAR(20) NOT NULL CHECK (priority IN ('low', 'medium', 'high', 'emergency')),
    status VARCHAR(20) NOT NULL DEFAULT 'open' CHECK (status IN ('open', 'in_progress', 'completed', 'cancelled')),
    category VARCHAR(50) NOT NULL,
    requested_date TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    completed_date TIMESTAMP WITH TIME ZONE,
    estimated_cost DECIMAL(10,2),
    actual_cost DECIMAL(10,2),
    contractor_id UUID REFERENCES contractors(id),
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Maintenance photos table
CREATE TABLE maintenance_photos (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    maintenance_request_id UUID NOT NULL REFERENCES maintenance_requests(id) ON DELETE CASCADE,
    photo_url TEXT NOT NULL,
    photo_key VARCHAR(500) NOT NULL,
    description TEXT,
    is_before BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Payments table
CREATE TABLE payments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    landlord_id UUID NOT NULL REFERENCES landlords(id) ON DELETE CASCADE,
    property_id UUID NOT NULL REFERENCES properties(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    amount DECIMAL(10,2) NOT NULL,
    payment_type VARCHAR(50) NOT NULL,
    payment_method VARCHAR(50) NOT NULL,
    due_date DATE NOT NULL,
    paid_date DATE,
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'paid', 'overdue', 'cancelled')),
    reference_number VARCHAR(100),
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- AI chat messages table
CREATE TABLE ai_chat_messages (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    landlord_id UUID NOT NULL REFERENCES landlords(id) ON DELETE CASCADE,
    tenant_id UUID REFERENCES tenants(id),
    user_type VARCHAR(20) NOT NULL CHECK (user_type IN ('landlord', 'tenant')),
    question TEXT NOT NULL,
    answer TEXT NOT NULL,
    model_used VARCHAR(100) NOT NULL,
    tokens_used INTEGER,
    cost DECIMAL(10,6),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Notifications table
CREATE TABLE notifications (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    landlord_id UUID NOT NULL REFERENCES landlords(id) ON DELETE CASCADE,
    recipient_id UUID NOT NULL,
    recipient_type VARCHAR(20) NOT NULL CHECK (recipient_type IN ('landlord', 'tenant', 'contractor')),
    type VARCHAR(50) NOT NULL,
    title VARCHAR(255) NOT NULL,
    message TEXT NOT NULL,
    is_read BOOLEAN DEFAULT false,
    read_at TIMESTAMP WITH TIME ZONE,
    related_entity_id UUID,
    related_entity_type VARCHAR(50),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for performance (multi-tenant aware)
CREATE INDEX idx_properties_landlord_id ON properties(landlord_id);
CREATE INDEX idx_properties_current_tenant_id ON properties(current_tenant_id);
CREATE INDEX idx_properties_is_available ON properties(is_available);

CREATE INDEX idx_tenants_landlord_id ON tenants(landlord_id);
CREATE INDEX idx_tenants_email ON tenants(email);
CREATE INDEX idx_tenants_is_active ON tenants(is_active);

CREATE INDEX idx_contractors_landlord_id ON contractors(landlord_id);
CREATE INDEX idx_contractors_specialization ON contractors(specialization);
CREATE INDEX idx_contractors_is_active ON contractors(is_active);

CREATE INDEX idx_maintenance_requests_landlord_id ON maintenance_requests(landlord_id);
CREATE INDEX idx_maintenance_requests_property_id ON maintenance_requests(property_id);
CREATE INDEX idx_maintenance_requests_tenant_id ON maintenance_requests(tenant_id);
CREATE INDEX idx_maintenance_requests_status ON maintenance_requests(status);
CREATE INDEX idx_maintenance_requests_priority ON maintenance_requests(priority);
CREATE INDEX idx_maintenance_requests_category ON maintenance_requests(category);
CREATE INDEX idx_maintenance_requests_contractor_id ON maintenance_requests(contractor_id);

CREATE INDEX idx_maintenance_photos_maintenance_request_id ON maintenance_photos(maintenance_request_id);

CREATE INDEX idx_payments_landlord_id ON payments(landlord_id);
CREATE INDEX idx_payments_property_id ON payments(property_id);
CREATE INDEX idx_payments_tenant_id ON payments(tenant_id);
CREATE INDEX idx_payments_status ON payments(status);
CREATE INDEX idx_payments_due_date ON payments(due_date);
CREATE INDEX idx_payments_payment_type ON payments(payment_type);

CREATE INDEX idx_ai_chat_messages_landlord_id ON ai_chat_messages(landlord_id);
CREATE INDEX idx_ai_chat_messages_tenant_id ON ai_chat_messages(tenant_id);
CREATE INDEX idx_ai_chat_messages_user_type ON ai_chat_messages(user_type);

CREATE INDEX idx_notifications_landlord_id ON notifications(landlord_id);
CREATE INDEX idx_notifications_recipient_id ON notifications(recipient_id);
CREATE INDEX idx_notifications_recipient_type ON notifications(recipient_type);
CREATE INDEX idx_notifications_is_read ON notifications(is_read);
CREATE INDEX idx_notifications_type ON notifications(type);

-- Composite indexes for common query patterns
CREATE INDEX idx_maintenance_requests_landlord_status ON maintenance_requests(landlord_id, status);
CREATE INDEX idx_maintenance_requests_landlord_priority ON maintenance_requests(landlord_id, priority);
CREATE INDEX idx_payments_landlord_status ON payments(landlord_id, status);
CREATE INDEX idx_payments_landlord_due_date ON payments(landlord_id, due_date);

-- Update triggers for updated_at timestamps
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Apply update triggers to all tables
CREATE TRIGGER update_landlords_updated_at BEFORE UPDATE ON landlords FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_properties_updated_at BEFORE UPDATE ON properties FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_tenants_updated_at BEFORE UPDATE ON tenants FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_contractors_updated_at BEFORE UPDATE ON contractors FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_maintenance_requests_updated_at BEFORE UPDATE ON maintenance_requests FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_maintenance_photos_updated_at BEFORE UPDATE ON maintenance_photos FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_payments_updated_at BEFORE UPDATE ON payments FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_ai_chat_messages_updated_at BEFORE UPDATE ON ai_chat_messages FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_notifications_updated_at BEFORE UPDATE ON notifications FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

