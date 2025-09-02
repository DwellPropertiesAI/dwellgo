# 🚀 EC2 Deployment Guide for Dwell Backend

## 📋 **Prerequisites**

- AWS Account with EC2 access
- AWS CLI configured
- Domain name (optional but recommended)
- SSL certificate (for HTTPS)

## 🏗️ **Option 1: Simple EC2 Instance (MVP)**

### **Step 1: Create EC2 Instance**

1. **Launch EC2 Instance:**
   - AMI: Amazon Linux 2023 (or Ubuntu 22.04 LTS)
   - Instance Type: `t3.medium` (2 vCPU, 4GB RAM)
   - Storage: 20GB GP3
   - Security Group: Create new with these rules:
     - SSH (22): Your IP
     - HTTP (80): 0.0.0.0/0
     - HTTPS (443): 0.0.0.0/0
     - Custom TCP (8080): 0.0.0.0/0

2. **Key Pair:** Create or select existing key pair

### **Step 2: Connect and Setup**

```bash
# Connect to your EC2 instance
ssh -i your-key.pem ec2-user@your-ec2-public-ip

# Update system
sudo yum update -y  # For Amazon Linux
# OR
sudo apt update && sudo apt upgrade -y  # For Ubuntu

# Install Docker
sudo yum install -y docker  # Amazon Linux
sudo systemctl start docker
sudo systemctl enable docker
sudo usermod -a -G docker ec2-user

# Install Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# Logout and login again for group changes
exit
ssh -i your-key.pem ec2-user@your-ec2-public-ip
```

### **Step 3: Deploy Application**

```bash
# Clone your repository
git clone https://github.com/yourusername/dwell.git
cd dwell

# Create production environment file
cp env.example .env.prod

# Edit production environment
nano .env.prod
```

**Production Environment Variables:**
```bash
# Server Configuration
SERVER_PORT=8080
SERVER_HOST=0.0.0.0
ENVIRONMENT=production

# Database Configuration (Use RDS in production)
DB_HOST=your-rds-endpoint
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your-secure-password
DB_NAME=dwell
DB_SSLMODE=require

# AWS Configuration
AWS_REGION=us-east-1
AWS_ACCESS_KEY_ID=your-access-key
AWS_SECRET_ACCESS_KEY=your-secret-key

# Other AWS services...
```

### **Step 4: Run Application**

```bash
# Build and run with Docker Compose
docker-compose -f docker-compose.prod.yml up -d

# Check status
docker-compose -f docker-compose.prod.yml ps

# View logs
docker-compose -f docker-compose.prod.yml logs -f
```

## 🌐 **Option 2: Production Setup with Load Balancer**

### **Step 1: Create RDS Database**

```bash
# Use AWS CLI or Console to create Aurora PostgreSQL
aws rds create-db-cluster \
  --db-cluster-identifier dwell-prod-cluster \
  --engine aurora-postgresql \
  --engine-version 15.4 \
  --master-username postgres \
  --master-user-password your-secure-password \
  --db-cluster-instance-class db.t3.medium \
  --vpc-security-group-ids sg-xxxxxxxxx
```

### **Step 2: Create Application Load Balancer**

```bash
# Create ALB
aws elbv2 create-load-balancer \
  --name dwell-alb \
  --subnets subnet-xxxxxxxxx subnet-yyyyyyyyy \
  --security-groups sg-xxxxxxxxx
```

### **Step 3: Create Auto Scaling Group**

```bash
# Create launch template
aws ec2 create-launch-template \
  --launch-template-name dwell-launch-template \
  --version-description v1 \
  --launch-template-data '{
    "ImageId": "ami-xxxxxxxxx",
    "InstanceType": "t3.medium",
    "SecurityGroupIds": ["sg-xxxxxxxxx"],
    "UserData": "base64-encoded-user-data"
  }'

# Create auto scaling group
aws autoscaling create-auto-scaling-group \
  --auto-scaling-group-name dwell-asg \
  --launch-template LaunchTemplateName=dwell-launch-template,Version=\$Latest \
  --min-size 2 \
  --max-size 5 \
  --desired-capacity 2 \
  --vpc-zone-identifier "subnet-xxxxxxxxx,subnet-yyyyyyyyy"
```

## 🔒 **Security Configuration**

### **Security Groups**

```bash
# Application Security Group
aws ec2 create-security-group \
  --group-name dwell-app-sg \
  --description "Security group for Dwell application"

# Add rules
aws ec2 authorize-security-group-ingress \
  --group-id sg-xxxxxxxxx \
  --protocol tcp \
  --port 8080 \
  --source-group sg-xxxxxxxxx  # ALB security group

# Database Security Group
aws ec2 create-security-group \
  --group-name dwell-db-sg \
  --description "Security group for Dwell database"

aws ec2 authorize-security-group-ingress \
  --group-id sg-yyyyyyyyy \
  --protocol tcp \
  --port 5432 \
  --source-group sg-xxxxxxxxx  # App security group
```

### **IAM Roles**

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "ec2:DescribeInstances",
        "ec2:DescribeTags",
        "autoscaling:DescribeAutoScalingGroups",
        "autoscaling:DescribeAutoScalingInstances"
      ],
      "Resource": "*"
    }
  ]
}
```

## 📦 **Deployment Scripts**

### **Deploy Script**

```bash
#!/bin/bash
# deploy.sh

set -e

echo "🚀 Deploying Dwell Backend to EC2..."

# Pull latest code
git pull origin main

# Build new image
docker-compose -f docker-compose.prod.yml build

# Stop existing services
docker-compose -f docker-compose.prod.yml down

# Start with new image
docker-compose -f docker-compose.prod.yml up -d

# Health check
sleep 10
curl -f http://localhost:8080/api/v1/health

echo "✅ Deployment complete!"
```

### **Production Docker Compose**

```yaml
# docker-compose.prod.yml
version: '3.8'

services:
  dwell-api:
    build: .
    container_name: dwell-api-prod
    restart: unless-stopped
    ports:
      - "8080:8080"
    env_file:
      - .env.prod
    environment:
      - ENVIRONMENT=production
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/api/v1/health"]
      interval: 30s
      timeout: 10s
      retries: 3
    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "3"
```

## 🌍 **Domain & SSL Setup**

### **Route 53 Configuration**

```bash
# Create hosted zone
aws route53 create-hosted-zone \
  --name yourdomain.com \
  --caller-reference $(date +%s)

# Add A record pointing to ALB
aws route53 change-resource-record-sets \
  --hosted-zone-id Z1234567890 \
  --change-batch '{
    "Changes": [{
      "Action": "CREATE",
      "ResourceRecordSet": {
        "Name": "api.yourdomain.com",
        "Type": "A",
        "AliasTarget": {
          "HostedZoneId": "Z35SXDOTRQ7X7K",
          "DNSName": "your-alb-dns-name",
          "EvaluateTargetHealth": true
        }
      }
    }]
  }'
```

### **SSL Certificate with ACM**

```bash
# Request certificate
aws acm request-certificate \
  --domain-name api.yourdomain.com \
  --validation-method DNS

# Add validation records to Route 53
# (AWS will provide the records to add)
```

## 📊 **Monitoring & Logging**

### **CloudWatch Setup**

```bash
# Create log group
aws logs create-log-group --log-group-name /aws/ec2/dwell

# Create metric filter
aws logs put-metric-filter \
  --log-group-name /aws/ec2/dwell \
  --filter-name dwell-error-filter \
  --filter-pattern "ERROR" \
  --metric-transformations '[
    {
      "metricName": "ErrorCount",
      "metricNamespace": "Dwell/Backend",
      "metricValue": "1"
    }
  ]'
```

### **Health Check Endpoint**

Your application should have a health check endpoint:

```go
// In your router
router.GET("/health", func(c *gin.Context) {
    c.JSON(200, gin.H{
        "status": "healthy",
        "timestamp": time.Now().Unix(),
        "version": "1.0.0",
    })
})
```

## 🚀 **Quick Deployment Commands**

### **Initial Setup:**
```bash
# Create EC2 instance
# Configure security groups
# Install Docker & Docker Compose
# Clone repository
# Configure environment
# Run application
```

### **Update Deployment:**
```bash
# SSH to EC2
ssh -i your-key.pem ec2-user@your-ec2-public-ip

# Pull and deploy
cd dwell
git pull origin main
docker-compose -f docker-compose.prod.yml up -d --build
```

### **Rollback:**
```bash
# Revert to previous version
git checkout HEAD~1
docker-compose -f docker-compose.prod.yml up -d --build
```

## 🔍 **Testing Your Deployment**

### **Health Check:**
```bash
curl http://your-ec2-public-ip:8080/api/v1/health
curl http://api.yourdomain.com/health
```

### **API Endpoints:**
```bash
# Test authentication
curl -X POST http://api.yourdomain.com/api/v1/auth/signup \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"Test123!","userType":"tenant"}'

# Test AI chatbot
curl -X POST http://api.yourdomain.com/api/v1/ai/query \
  -H "Content-Type: application/json" \
  -d '{"question":"Hello, how are you?"}'
```

## 🚨 **Important Notes**

1. **Never commit `.env.prod` files** to version control
2. **Use RDS for production** instead of local PostgreSQL
3. **Enable CloudWatch monitoring** for production instances
4. **Set up automated backups** for your database
5. **Use HTTPS in production** with proper SSL certificates
6. **Monitor costs** and set up billing alerts

## 📞 **Need Help?**

- Check CloudWatch logs for debugging
- Use AWS Systems Manager for easy EC2 access
- Set up CloudTrail for API call logging
- Monitor with AWS X-Ray for distributed tracing

---

**Next Steps:** After deployment, update your frontend to point to your new EC2 endpoint! 🎉

