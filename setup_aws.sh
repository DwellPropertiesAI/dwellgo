#!/bin/bash

# 🚀 Dwell AWS Setup Script
# This script helps you set up AWS services for the Dwell Property Management MVP

set -e

echo "🚀 Welcome to Dwell AWS Setup!"
echo "=================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if AWS CLI is installed
check_aws_cli() {
    if ! command -v aws &> /dev/null; then
        print_error "AWS CLI is not installed. Please install it first:"
        echo "  https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html"
        exit 1
    fi
    print_success "AWS CLI is installed"
}

# Check AWS credentials
check_aws_credentials() {
    print_status "Checking AWS credentials..."
    
    if aws sts get-caller-identity &> /dev/null; then
        print_success "AWS credentials are configured"
        aws sts get-caller-identity
    else
        print_error "AWS credentials are not configured or invalid"
        echo "Please run: aws configure"
        exit 1
    fi
}

# Create environment file
create_env_file() {
    print_status "Creating environment file..."
    
    if [ -f .env ]; then
        print_warning ".env file already exists. Backing up to .env.backup"
        cp .env .env.backup
    fi
    
    cp env.example .env
    print_success "Created .env file from env.example"
    print_warning "Please edit .env file with your actual AWS credentials and configuration"
}

# Test AWS services
test_aws_services() {
    print_status "Testing AWS services..."
    
    # Test Cognito
    print_status "Testing Cognito..."
    if aws cognito-idp list-user-pools --max-items 1 &> /dev/null; then
        print_success "Cognito access confirmed"
    else
        print_warning "Cognito access failed - check IAM permissions"
    fi
    
    # Test S3
    print_status "Testing S3..."
    if aws s3 ls &> /dev/null; then
        print_success "S3 access confirmed"
    else
        print_warning "S3 access failed - check IAM permissions"
    fi
    
    # Test Bedrock
    print_status "Testing Bedrock..."
    if aws bedrock list-foundation-models &> /dev/null; then
        print_success "Bedrock access confirmed"
    else
        print_warning "Bedrock access failed - check IAM permissions or model access"
    fi
    
    # Test SNS
    print_status "Testing SNS..."
    if aws sns list-topics &> /dev/null; then
        print_success "SNS access confirmed"
    else
        print_warning "SNS access failed - check IAM permissions"
    fi
    
    # Test SES
    print_status "Testing SES..."
    if aws ses get-send-quota &> /dev/null; then
        print_success "SES access confirmed"
    else
        print_warning "SES access failed - check IAM permissions"
    fi
}

# Generate JWT secret
generate_jwt_secret() {
    print_status "Generating JWT secret..."
    
    # Generate a random 32-character string
    JWT_SECRET=$(openssl rand -base64 32 | tr -d "=+/" | cut -c1-32)
    
    if [ -f .env ]; then
        # Replace the JWT_SECRET_KEY in .env file
        if [[ "$OSTYPE" == "darwin"* ]]; then
            # macOS
            sed -i '' "s/JWT_SECRET_KEY=.*/JWT_SECRET_KEY=$JWT_SECRET/" .env
        else
            # Linux
            sed -i "s/JWT_SECRET_KEY=.*/JWT_SECRET_KEY=$JWT_SECRET/" .env
        fi
        print_success "Generated and updated JWT secret in .env file"
    else
        print_warning "JWT secret generated: $JWT_SECRET"
        print_warning "Please add this to your .env file: JWT_SECRET_KEY=$JWT_SECRET"
    fi
}

# Main setup function
main() {
    echo ""
    print_status "Starting AWS setup..."
    
    check_aws_cli
    check_aws_credentials
    create_env_file
    generate_jwt_secret
    test_aws_services
    
    echo ""
    echo "=================================="
    print_success "Setup completed!"
    echo ""
    echo "Next steps:"
    echo "1. Edit .env file with your AWS configuration"
    echo "2. Follow the AWS_SETUP_GUIDE.md for detailed service setup"
    echo "3. Run 'make docker-build' to test your configuration"
    echo ""
    echo "Need help? Check AWS_SETUP_GUIDE.md for detailed instructions"
}

# Run main function
main "$@"

