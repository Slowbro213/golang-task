package user

import (
	"context"
	"fmt"

	"echo-app/internal/repositories"
	"echo-app/internal/requests"
	"echo-app/internal/server/builders"

	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	userRepository *repositories.UserRepository
}

func NewService(userRepository *repositories.UserRepository) *Service {
	return &Service{userRepository: userRepository}
}

// Register handles both traditional registration and OIDC user creation
func (s *Service) Register(ctx context.Context, request *requests.RegisterRequest) error {
	encryptedPassword, err := bcrypt.GenerateFromPassword(
		[]byte(request.Password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return fmt.Errorf("encrypt password: %w", err)
	}

	user := builders.NewUserBuilder().
		SetEmail(request.Email).
		SetName(request.Name).
		SetPassword(string(encryptedPassword)).
		Build()

	if err := s.userRepository.Create(ctx, user); err != nil {
		return fmt.Errorf("create user in repository: %w", err)
	}

	return nil
}

// GetOrCreateUserFromOIDC handles OIDC user authentication
