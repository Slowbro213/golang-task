package repositories

import (
	"context"
	"errors"
	"fmt"

	"echo-app/internal/models"
	"echo-app/internal/server/builders"

	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	if err := r.db.WithContext(ctx).Create(user).Error; err != nil {
		return fmt.Errorf("execute insert user query: %w", err)
	}
	return nil
}

func (r *UserRepository) GetByID(ctx context.Context, id uint) (models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).Where("id = ?", id).Take(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.User{}, errors.Join(models.ErrUserNotFound, err)
	} else if err != nil {
		return models.User{}, fmt.Errorf("execute select user by id query: %w", err)
	}
	return user, nil
}

func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).Where("email = ?", email).Take(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.User{}, errors.Join(models.ErrUserNotFound, err)
	} else if err != nil {
		return models.User{}, fmt.Errorf("execute select user by email query: %w", err)
	}
	return user, nil
}

func (r *UserRepository) GetUserByOIDCSubject(ctx context.Context, sub string) (models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).Where("oidc_subject = ?", sub).Take(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.User{}, errors.Join(models.ErrUserNotFound, err)
	} else if err != nil {
		return models.User{}, fmt.Errorf("execute select user by OIDC subject query: %w", err)
	}
	return user, nil
}

func (r *UserRepository) Update(ctx context.Context, user *models.User) error {
	if err := r.db.WithContext(ctx).Save(user).Error; err != nil {
		return fmt.Errorf("execute update user query: %w", err)
	}
	return nil
}

func (r *UserRepository) GetOrCreateUserFromOIDC(ctx context.Context, claims *models.OIDCClaims) (models.User, error) {
	// First try to find by OIDC subject
	user, err := r.GetUserByOIDCSubject(ctx, claims.Sub)
	if err == nil {
		return user, nil
	}

	// If not found by subject, try by email
	user, err = r.GetUserByEmail(ctx, claims.Email)
	if err == nil {
		// Update existing user with OIDC subject
		user.OIDCSubject = claims.Sub
		if err := r.Update(ctx, &user); err != nil {
			return models.User{}, fmt.Errorf("update user with OIDC subject: %w", err)
		}
		return user, nil
	}

	// Create new user if not found
	newUser := builders.NewUserBuilder().
		SetEmail(claims.Email).
		SetName(claims.Name).
		SetOIDCSubject(claims.Sub).
		Build()

	if err := r.Create(ctx, newUser); err != nil {
		return models.User{}, fmt.Errorf("create OIDC user: %w", err)
	}

	return *newUser, nil
}
