package services

import (
	"context"
	"fmt"
	"time"

	"dwell/internal/aws"
	"dwell/internal/config"
	"dwell/internal/domain"

	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type AuthService struct {
	awsClients *aws.Clients
	config     *config.Config
}

type AuthRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
	UserID       string `json:"user_id"`
	UserType     string `json:"user_type"`
}

type SignUpRequest struct {
	Email       string `json:"email" binding:"required,email"`
	Password    string `json:"password" binding:"required,min=8"`
	FirstName   string `json:"first_name" binding:"required"`
	LastName    string `json:"last_name" binding:"required"`
	Phone       string `json:"phone"`
	CompanyName string `json:"company_name"`
	UserType    string `json:"user_type" binding:"required,oneof=landlord tenant"`
}

type SignUpResponse struct {
	UserID      string `json:"user_id"`
	UserType    string `json:"user_type"`
	Message     string `json:"message"`
	ConfirmCode string `json:"confirm_code,omitempty"`
}

func NewAuthService(awsClients *aws.Clients, config *config.Config) *AuthService {
	return &AuthService{
		awsClients: awsClients,
		config:     config,
	}
}

// SignUp creates a new user account in Cognito
func (s *AuthService) SignUp(ctx context.Context, req *SignUpRequest) (*SignUpResponse, error) {
	// Prepare Cognito signup request
	signUpInput := &cognitoidentityprovider.SignUpInput{
		ClientId: aws.String(s.config.AWS.Cognito.ClientID),
		Username: aws.String(req.Email),
		Password: aws.String(req.Password),
		UserAttributes: []types.AttributeType{
			{
				Name:  aws.String("email"),
				Value: aws.String(req.Email),
			},
			{
				Name:  aws.String("given_name"),
				Value: aws.String(req.FirstName),
			},
			{
				Name:  aws.String("family_name"),
				Value: aws.String(req.LastName),
			},
			{
				Name:  aws.String("phone_number"),
				Value: aws.String(req.Phone),
			},
			{
				Name:  aws.String("custom:user_type"),
				Value: aws.String(req.UserType),
			},
			{
				Name:  aws.String("custom:company_name"),
				Value: aws.String(req.CompanyName),
			},
		},
	}

	// Call Cognito SignUp
	result, err := s.awsClients.GetCognitoClient().SignUp(ctx, signUpInput)
	if err != nil {
		return nil, fmt.Errorf("failed to sign up user: %w", err)
	}

	return &SignUpResponse{
		UserID:      *result.UserSub,
		UserType:    req.UserType,
		Message:     "User registered successfully. Please check your email for confirmation code.",
		ConfirmCode: "", // Cognito will send this via email
	}, nil
}

// ConfirmSignUp confirms user registration with confirmation code
func (s *AuthService) ConfirmSignUp(ctx context.Context, email, confirmationCode string) error {
	confirmInput := &cognitoidentityprovider.ConfirmSignUpInput{
		ClientId:         aws.String(s.config.AWS.Cognito.ClientID),
		Username:         aws.String(email),
		ConfirmationCode: aws.String(confirmationCode),
	}

	_, err := s.awsClients.GetCognitoClient().ConfirmSignUp(ctx, confirmInput)
	if err != nil {
		return fmt.Errorf("failed to confirm signup: %w", err)
	}

	return nil
}

// SignIn authenticates user and returns tokens
func (s *AuthService) SignIn(ctx context.Context, req *AuthRequest) (*AuthResponse, error) {
	// Prepare Cognito signin request
	authInput := &cognitoidentityprovider.InitiateAuthInput{
		ClientId: aws.String(s.config.AWS.Cognito.ClientID),
		AuthFlow: types.AuthFlowTypeUserPasswordAuth,
		AuthParameters: map[string]string{
			"USERNAME": req.Email,
			"PASSWORD": req.Password,
		},
	}

	// Call Cognito InitiateAuth
	result, err := s.awsClients.GetCognitoClient().InitiateAuth(ctx, authInput)
	if err != nil {
		return nil, fmt.Errorf("failed to sign in: %w", err)
	}

	// Extract tokens and user info
	accessToken := *result.AuthenticationResult.AccessToken
	refreshToken := *result.AuthenticationResult.RefreshToken
	expiresIn := int(*result.AuthenticationResult.ExpiresIn)

	// Get user attributes to determine user type
	userInfo, err := s.getUserInfo(ctx, accessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	return &AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    expiresIn,
		TokenType:    "Bearer",
		UserID:       userInfo.UserID,
		UserType:     userInfo.UserType,
	}, nil
}

// RefreshToken refreshes the access token using refresh token
func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*AuthResponse, error) {
	authInput := &cognitoidentityprovider.InitiateAuthInput{
		ClientId: aws.String(s.config.AWS.Cognito.ClientID),
		AuthFlow: types.AuthFlowTypeRefreshTokenAuth,
		AuthParameters: map[string]string{
			"REFRESH_TOKEN": refreshToken,
		},
	}

	result, err := s.awsClients.GetCognitoClient().InitiateAuth(ctx, authInput)
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	accessToken := *result.AuthenticationResult.AccessToken
	expiresIn := int(*result.AuthenticationResult.ExpiresIn)

	// Get user info from the new access token
	userInfo, err := s.getUserInfo(ctx, accessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	return &AuthResponse{
		AccessToken: accessToken,
		ExpiresIn:   expiresIn,
		TokenType:   "Bearer",
		UserID:      userInfo.UserID,
		UserType:    userInfo.UserType,
	}, nil
}

// SignOut signs out the user
func (s *AuthService) SignOut(ctx context.Context, accessToken string) error {
	signOutInput := &cognitoidentityprovider.GlobalSignOutInput{
		AccessToken: aws.String(accessToken),
	}

	_, err := s.awsClients.GetCognitoClient().GlobalSignOut(ctx, signOutInput)
	if err != nil {
		return fmt.Errorf("failed to sign out: %w", err)
	}

	return nil
}

// ValidateToken validates the JWT token and returns user claims
func (s *AuthService) ValidateToken(tokenString string) (*domain.UserClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.config.JWT.SecretKey), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userID, _ := claims["user_id"].(string)
		userType, _ := claims["user_type"].(string)
		landlordID, _ := claims["landlord_id"].(string)

		var landlordUUID *uuid.UUID
		if landlordID != "" {
			if id, err := uuid.Parse(landlordID); err == nil {
				landlordUUID = &id
			}
		}

		return &domain.UserClaims{
			UserID:     userID,
			UserType:   userType,
			LandlordID: landlordUUID,
			ExpiresAt:  time.Unix(int64(claims["exp"].(float64)), 0),
		}, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// getUserInfo retrieves user information from Cognito
func (s *AuthService) getUserInfo(ctx context.Context, accessToken string) (*domain.UserInfo, error) {
	getUserInput := &cognitoidentityprovider.GetUserInput{
		AccessToken: aws.String(accessToken),
	}

	result, err := s.awsClients.GetCognitoClient().GetUser(ctx, getUserInput)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	userInfo := &domain.UserInfo{
		UserID: *result.Username,
	}

	// Extract user attributes
	for _, attr := range result.UserAttributes {
		switch *attr.Name {
		case "custom:user_type":
			userInfo.UserType = *attr.Value
		case "custom:landlord_id":
			if *attr.Value != "" {
				if id, err := uuid.Parse(*attr.Value); err == nil {
					userInfo.LandlordID = &id
				}
			}
		}
	}

	return userInfo, nil
}

