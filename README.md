# Dwell - AI-Powered Property Management API

Dwell is a comprehensive, scalable backend API for property management systems, leveraging AWS services for authentication, AI-powered assistance, file storage, and notifications.

## 🚀 Features

### Core Functionality
- **Multi-tenant Architecture**: Secure data isolation per landlord
- **AI-Powered Chatbot**: AWS Bedrock integration for property management assistance
- **File Management**: S3 integration for maintenance photos and documents
- **Real-time Notifications**: SNS/SES integration for email and SMS alerts
- **Comprehensive API**: RESTful endpoints with Swagger documentation

### User Types
- **Landlords**: Property management, tenant oversight, financial tracking
- **Tenants**: Maintenance requests, payment tracking, communication
- **Contractors**: Service coordination, work orders, billing

### AWS Service Integration
- **AWS Cognito**: User authentication and management
- **AWS S3**: File storage and management
- **AWS Bedrock**: AI chatbot and property management tips
- **AWS SNS**: SMS notifications and alerts
- **AWS SES**: Email notifications and templates
- **AWS Aurora**: PostgreSQL database (production)

## 🏗️ Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Frontend      │    │   API Gateway   │    │   Load Balancer │
│   (React/Vue)   │◄──►│   (Optional)    │◄──►│   (Optional)    │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                │
                                ▼
                       ┌─────────────────┐
                       │   Dwell API     │
                       │   (Go/Gin)      │
                       └─────────────────┘
                                │
                ┌────────────────┼────────────────┐
                ▼                ▼                ▼
        ┌─────────────┐ ┌─────────────┐ ┌─────────────┐
        │   Cognito   │ │     S3      │ │  Bedrock    │
        │  (Auth)     │ │ (Files)     │ │    (AI)     │
        └─────────────┘ └─────────────┘ └─────────────┘
                │                │                │
                ▼                ▼                ▼
        ┌─────────────┐ ┌─────────────┐ ┌─────────────┐
        │     SNS     │ │     SES     │ │   Aurora    │
        │   (SMS)     │ │  (Email)    │ │(Database)   │
        └─────────────┘ └─────────────┘ └─────────────┘
```

## 🛠️ Tech Stack

- **Backend**: Go 1.21+ with Gin framework
- **Database**: PostgreSQL with Aurora Serverless (production)
- **Authentication**: AWS Cognito with JWT tokens
- **File Storage**: AWS S3 with presigned URLs
- **AI Services**: AWS Bedrock (Claude 3 Sonnet)
- **Notifications**: AWS SNS (SMS) + SES (Email)
- **Containerization**: Docker + Docker Compose
- **Documentation**: Swagger/OpenAPI 3.0

## 📋 Prerequisites

- Go 1.21 or higher
- Docker and Docker Compose
- AWS Account with appropriate permissions
- PostgreSQL (for local development)

## 🚀 Quick Start

### 1. Clone the Repository
```bash
git clone https://github.com/yourusername/dwell.git
cd dwell
```

### 2. Set Up Environment Variables
```bash
cp env.example .env
# Edit .env with your AWS credentials and configuration
```

### 3. Start with Docker Compose
```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f dwell-api

# Stop services
docker-compose down
```

### 4. Manual Setup (Alternative)
```bash
# Install dependencies
go mod download

# Run the application
go run main.go
```

## 🔧 Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `SERVER_PORT` | API server port | `8080` |
| `DB_HOST` | Database host | `localhost` |
| `DB_PORT` | Database port | `5432` |
| `AWS_REGION` | AWS region | `us-east-1` |
| `COGNITO_USER_POOL_ID` | Cognito User Pool ID | Required |
| `S3_BUCKET_NAME` | S3 bucket for files | Required |
| `BEDROCK_MODEL` | AI model identifier | `anthropic.claude-3-sonnet-20240229-v1:0` |

### AWS Service Setup

#### 1. AWS Cognito
- Create a User Pool
- Configure password policies
- Set up app client
- Enable email verification

#### 2. AWS S3
- Create bucket for file storage
- Configure CORS policy
- Set up IAM roles and policies

#### 3. AWS Bedrock
- Enable Claude 3 Sonnet model
- Configure IAM permissions
- Set up model access

#### 4. AWS SNS/SES
- Create SNS topic for notifications
- Verify SES email addresses
- Configure IAM permissions

## 📚 API Documentation

### Base URL
```
http://localhost:8080/api/v1
```

### Authentication Endpoints
- `POST /auth/signup` - User registration
- `POST /auth/confirm` - Confirm registration
- `POST /auth/signin` - User login
- `POST /auth/refresh` - Refresh token
- `POST /auth/signout` - User logout
- `GET /auth/profile` - Get user profile

### AI Chatbot Endpoints
- `POST /ai/query` - Ask AI question
- `GET /ai/tips` - Get property management tips
- `GET /ai/history` - Get chat history
- `GET /ai/analytics` - Get usage analytics

### File Management Endpoints
- `POST /files/upload` - Upload file to S3
- `DELETE /files/delete` - Delete file from S3
- `GET /files/list` - List files
- `GET /files/signed-url` - Get temporary access URL
- `GET /files/metadata` - Get file metadata

### Protected Routes
All endpoints except authentication require a valid JWT token in the Authorization header:
```
Authorization: Bearer <your-jwt-token>
```

### Swagger Documentation
Access the interactive API documentation at:
```
http://localhost:8080/swagger/index.html
```

## 🗄️ Database Schema

The application uses a multi-tenant PostgreSQL database with the following key tables:

- **landlords** - Property owners/managers
- **properties** - Real estate properties
- **tenants** - Property renters
- **maintenance_requests** - Maintenance issues
- **payments** - Rent and other payments
- **contractors** - Service providers
- **ai_chat_messages** - AI conversation history
- **notifications** - System notifications

## 🔒 Security Features

- **JWT Authentication**: Secure token-based authentication
- **Multi-tenant Isolation**: Data separation per landlord
- **Role-based Access Control**: Different permissions for landlords and tenants
- **Input Validation**: Comprehensive request validation
- **CORS Configuration**: Configurable cross-origin policies
- **File Type Validation**: Secure file upload restrictions

## 🧪 Testing

### Run Tests
```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific test
go test ./internal/services
```

### Test Coverage
```bash
# Generate coverage report
go test -coverprofile=coverage.out ./...

# View coverage in browser
go tool cover -html=coverage.out
```

## 🚀 Deployment

### Production Deployment
```bash
# Build production image
docker build -t dwell:latest .

# Run with production environment
docker run -d \
  --name dwell-api \
  -p 8080:8080 \
  --env-file .env.prod \
  dwell:latest
```

### AWS Deployment
- Use AWS ECS/Fargate for container orchestration
- Configure Aurora Serverless for database
- Set up Application Load Balancer
- Use AWS Secrets Manager for sensitive data

## 📊 Monitoring & Logging

### Health Checks
- Application health: `GET /api/v1/health`
- Database connectivity
- AWS service availability

### Logging
- Structured logging with Go's log package
- Request/response logging middleware
- Error tracking and monitoring

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🆘 Support

- **Documentation**: [API Docs](http://localhost:8080/swagger/index.html)
- **Issues**: [GitHub Issues](https://github.com/yourusername/dwell/issues)
- **Discussions**: [GitHub Discussions](https://github.com/yourusername/dwell/discussions)

## 🗺️ Roadmap

- [ ] Property management CRUD operations
- [ ] Tenant management system
- [ ] Maintenance request workflow
- [ ] Payment processing integration
- [ ] Real-time notifications
- [ ] Advanced analytics dashboard
- [ ] Mobile app API endpoints
- [ ] Third-party integrations

## 🙏 Acknowledgments

- AWS for comprehensive cloud services
- Gin framework for the excellent Go web framework
- PostgreSQL community for the robust database
- Open source contributors for inspiration and tools

---

**Built with ❤️ for the property management community**

