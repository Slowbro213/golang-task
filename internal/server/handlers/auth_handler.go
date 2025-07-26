package handlers

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"time"

	"echo-app/internal/models"
	"echo-app/internal/repositories"
	"echo-app/internal/responses"
	s "echo-app/internal/server"
	"echo-app/internal/services/user"

	"echo-app/internal/config"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/labstack/echo/v4"
	"golang.org/x/oauth2"
)

type AuthHandler struct {
	oauth2Config   *oauth2.Config
	oidcProvider   *oidc.Provider
	tokenVerifier  *oidc.IDTokenVerifier
	server         *s.Server
	userRepository *repositories.UserRepository
	userGetter     *user.Service
}

func NewAuthHandler(server *s.Server, userGetter *user.Service, userRepository *repositories.UserRepository, aconf *config.Auth) (*AuthHandler, error) {
	// Initialize OIDC provider
	provider, err := oidc.NewProvider(context.Background(), server.Config.Auth.OIDCIssuer)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize OIDC provider: %w", err)
	}

	// Configure OAuth2
	oauth2Config := &oauth2.Config{
		ClientID:     aconf.OIDCClientID,
		ClientSecret: aconf.OIDCClientSecret,
		RedirectURL:  aconf.OIDCRedirectURL,
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}

	return &AuthHandler{
		oauth2Config:   oauth2Config,
		oidcProvider:   provider,
		tokenVerifier:  provider.Verifier(&oidc.Config{ClientID: server.Config.Auth.OIDCClientID}),
		server:         server,
		userGetter:     userGetter,
		userRepository: userRepository,
	}, nil
}

// InitiateLogin godoc
// @Summary Initiate OIDC login
// @Description Redirects to Authentik for authentication
// @ID initiate-login
// @Tags Authentication
// @Success 302
// @Router /login [get]
func (h *AuthHandler) InitiateLogin(c echo.Context) error {
	// Add these checks at the start
	if h == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "auth handler not initialized")
	}
	if h.oauth2Config == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "OAuth2 config not initialized")
	}

	state, err := generateRandomState()
	if err != nil {
		return responses.ErrorResponse(c, http.StatusInternalServerError, "Failed to generate state")
	}

	c.SetCookie(&http.Cookie{
		Name:     "auth_state",
		Value:    state,
		Path:     "/",
		MaxAge:   3600,
		HttpOnly: true,
		Secure:   false, // Use config
	})

	authURL := h.oauth2Config.AuthCodeURL(state)
	return c.Redirect(http.StatusFound, authURL)
}

// HandleCallback godoc
// @Summary OIDC callback handler
// @Description Handles the callback from Authentik after login
// @ID handle-callback
// @Tags Authentication
// @Success 302
// @Failure 400 {object} responses.Error
// @Failure 401 {object} responses.Error
// @Failure 500 {object} responses.Error
// @Router /callback [get]
func (h *AuthHandler) HandleCallback(c echo.Context) error {
	// Verify state
	stateCookie, err := c.Cookie("auth_state")
	if err != nil {
		return responses.ErrorResponse(c, http.StatusBadRequest, "State cookie missing")
	}

	if c.QueryParam("state") != stateCookie.Value {
		return responses.ErrorResponse(c, http.StatusBadRequest, "Invalid state parameter")
	}

	// Exchange code for token
	token, err := h.oauth2Config.Exchange(c.Request().Context(), c.QueryParam("code"))
	if err != nil {
		return responses.ErrorResponse(c, http.StatusUnauthorized, "Failed to exchange token: "+err.Error())
	}

	// Extract ID token
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		return responses.ErrorResponse(c, http.StatusInternalServerError, "No id_token in token response")
	}

	// Verify ID token
	idToken, err := h.tokenVerifier.Verify(c.Request().Context(), rawIDToken)
	if err != nil {
		return responses.ErrorResponse(c, http.StatusUnauthorized, "Failed to verify ID Token: "+err.Error())
	}

	// Extract claims
	var claims models.OIDCClaims
	if err := idToken.Claims(&claims); err != nil {
		return responses.ErrorResponse(c, http.StatusInternalServerError, "Failed to parse claims: "+err.Error())
	}

	// Get or create user in our system
	user, err := h.userRepository.GetOrCreateUserFromOIDC(c.Request().Context(), &claims)
	if err != nil {
		return responses.ErrorResponse(c, http.StatusInternalServerError, "Failed to process user: "+err.Error())
	}

	// Create session tokens (simplified - in production use proper session management)
	accessToken, err := h.createAccessToken(user)
	if err != nil {
		return responses.ErrorResponse(c, http.StatusInternalServerError, "Failed to create access token")
	}

	// Set cookies (simplified - consider HttpOnly, Secure flags in production)
	c.SetCookie(&http.Cookie{
		Name:     "access_token",
		Value:    accessToken,
		Path:     "/",
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
		Secure:   false,
	})

	// Redirect to frontend or return success
	return c.Redirect(http.StatusFound, h.server.Config.Auth.OIDCSuccessRedirect)
}

// HandleLogout godoc
// @Summary Logout user
// @Description Clears authentication cookies
// @ID handle-logout
// @Tags Authentication
// @Success 200
// @Router /logout [post]
func (h *AuthHandler) HandleLogout(c echo.Context) error {
	// Clear cookies
	c.SetCookie(&http.Cookie{
		Name:     "access_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})

	c.SetCookie(&http.Cookie{
		Name:     "auth_state",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})

	// You might also want to hit Authentik's logout endpoint
	// logoutURL := h.oidcProvider.Endpoint().LogoutURL
	// return c.Redirect(http.StatusFound, logoutURL)

	return c.JSON(http.StatusOK, responses.MessageResponse(c, 200, "Successfully logged out"))
}

func (h *AuthHandler) createAccessToken(user models.User) (string, error) {
	// In a real implementation, you might want to use JWT or keep using OIDC tokens
	// This is a simplified version - consider proper token generation
	return "generated-access-token", nil
}

func generateRandomState() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
