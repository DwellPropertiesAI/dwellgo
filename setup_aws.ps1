# 🚀 Dwell AWS Setup Script for Windows
# This script helps you set up AWS services for the Dwell Property Management MVP

param(
    [switch]$SkipTests
)

# Set error action preference
$ErrorActionPreference = "Stop"

# Colors for output
$Red = "Red"
$Green = "Green"
$Yellow = "Yellow"
$Blue = "Blue"

# Function to print colored output
function Write-Status {
    param([string]$Message)
    Write-Host "[INFO] $Message" -ForegroundColor $Blue
}

function Write-Success {
    param([string]$Message)
    Write-Host "[SUCCESS] $Message" -ForegroundColor $Green
}

function Write-Warning {
    param([string]$Message)
    Write-Host "[WARNING] $Message" -ForegroundColor $Yellow
}

function Write-Error {
    param([string]$Message)
    Write-Host "[ERROR] $Message" -ForegroundColor $Red
}

Write-Host "🚀 Welcome to Dwell AWS Setup!" -ForegroundColor Cyan
Write-Host "==================================" -ForegroundColor Cyan

# Check if AWS CLI is installed
function Test-AWSCLI {
    Write-Status "Checking if AWS CLI is installed..."
    
    try {
        $awsVersion = aws --version 2>$null
        if ($awsVersion) {
            Write-Success "AWS CLI is installed: $awsVersion"
        } else {
            throw "AWS CLI not found"
        }
    } catch {
        Write-Error "AWS CLI is not installed. Please install it first:"
        Write-Host "  https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html" -ForegroundColor White
        exit 1
    }
}

# Check AWS credentials
function Test-AWSCredentials {
    Write-Status "Checking AWS credentials..."
    
    try {
        $identity = aws sts get-caller-identity 2>$null | ConvertFrom-Json
        if ($identity.Arn) {
            Write-Success "AWS credentials are configured"
            Write-Host "Account: $($identity.Account)" -ForegroundColor White
            Write-Host "User: $($identity.UserId)" -ForegroundColor White
            Write-Host "ARN: $($identity.Arn)" -ForegroundColor White
        } else {
            throw "No valid credentials found"
        }
    } catch {
        Write-Error "AWS credentials are not configured or invalid"
        Write-Host "Please run: aws configure" -ForegroundColor White
        exit 1
    }
}

# Create environment file
function New-EnvironmentFile {
    Write-Status "Creating environment file..."
    
    if (Test-Path ".env") {
        Write-Warning ".env file already exists. Backing up to .env.backup"
        Copy-Item ".env" ".env.backup"
    }
    
    if (Test-Path "env.example") {
        Copy-Item "env.example" ".env"
        Write-Success "Created .env file from env.example"
        Write-Warning "Please edit .env file with your actual AWS credentials and configuration"
    } else {
        Write-Error "env.example file not found"
        exit 1
    }
}

# Generate JWT secret
function New-JWTSecret {
    Write-Status "Generating JWT secret..."
    
    # Generate a random 32-character string
    $jwtSecret = -join ((33..126) | Get-Random -Count 32 | ForEach-Object {[char]$_})
    
    if (Test-Path ".env") {
        # Replace the JWT_SECRET_KEY in .env file
        $content = Get-Content ".env" -Raw
        $content = $content -replace "JWT_SECRET_KEY=.*", "JWT_SECRET_KEY=$jwtSecret"
        Set-Content ".env" $content -NoNewline
        Write-Success "Generated and updated JWT secret in .env file"
    } else {
        Write-Warning "JWT secret generated: $jwtSecret"
        Write-Warning "Please add this to your .env file: JWT_SECRET_KEY=$jwtSecret"
    }
}

# Test AWS services
function Test-AWSServices {
    if ($SkipTests) {
        Write-Warning "Skipping AWS service tests"
        return
    }
    
    Write-Status "Testing AWS services..."
    
    # Test Cognito
    Write-Status "Testing Cognito..."
    try {
        $cognito = aws cognito-idp list-user-pools --max-items 1 2>$null | ConvertFrom-Json
        if ($cognito.UserPools) {
            Write-Success "Cognito access confirmed"
        } else {
            Write-Warning "Cognito access failed - check IAM permissions"
        }
    } catch {
        Write-Warning "Cognito access failed - check IAM permissions"
    }
    
    # Test S3
    Write-Status "Testing S3..."
    try {
        $s3 = aws s3 ls 2>$null
        if ($s3) {
            Write-Success "S3 access confirmed"
        } else {
            Write-Warning "S3 access failed - check IAM permissions"
        }
    } catch {
        Write-Warning "S3 access failed - check IAM permissions"
    }
    
    # Test Bedrock
    Write-Status "Testing Bedrock..."
    try {
        $bedrock = aws bedrock list-foundation-models 2>$null | ConvertFrom-Json
        if ($bedrock.modelSummaries) {
            Write-Success "Bedrock access confirmed"
        } else {
            Write-Warning "Bedrock access failed - check IAM permissions or model access"
        }
    } catch {
        Write-Warning "Bedrock access failed - check IAM permissions or model access"
    }
    
    # Test SNS
    Write-Status "Testing SNS..."
    try {
        $sns = aws sns list-topics 2>$null | ConvertFrom-Json
        if ($sns.Topics) {
            Write-Success "SNS access confirmed"
        } else {
            Write-Warning "SNS access failed - check IAM permissions"
        }
    } catch {
        Write-Warning "SNS access failed - check IAM permissions"
    }
    
    # Test SES
    Write-Status "Testing SES..."
    try {
        $ses = aws ses get-send-quota 2>$null | ConvertFrom-Json
        if ($ses.Max24HourSend) {
            Write-Success "SES access confirmed"
        } else {
            Write-Warning "SES access failed - check IAM permissions"
        }
    } catch {
        Write-Warning "SES access failed - check IAM permissions"
    }
}

# Main setup function
function Start-Setup {
    Write-Status "Starting AWS setup..."
    
    Test-AWSCLI
    Test-AWSCredentials
    New-EnvironmentFile
    New-JWTSecret
    Test-AWSServices
    
    Write-Host ""
    Write-Host "==================================" -ForegroundColor Cyan
    Write-Success "Setup completed!"
    Write-Host ""
    Write-Host "Next steps:" -ForegroundColor White
    Write-Host "1. Edit .env file with your AWS configuration" -ForegroundColor White
    Write-Host "2. Follow the AWS_SETUP_GUIDE.md for detailed service setup" -ForegroundColor White
    Write-Host "3. Run 'make docker-build' to test your configuration" -ForegroundColor White
    Write-Host ""
    Write-Host "Need help? Check AWS_SETUP_GUIDE.md for detailed instructions" -ForegroundColor White
}

# Run main setup
Start-Setup

