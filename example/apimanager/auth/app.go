package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gofiber/fiber/v3"
	"github.com/lvlcn-t/go-kit/apimanager"
	"github.com/lvlcn-t/go-kit/apimanager/middleware"
	"github.com/lvlcn-t/loggerhead/logger"
	"golang.org/x/oauth2"
)

type Claims struct {
	Issuer  string   `json:"iss"`
	Subject string   `json:"sub"`
	Email   string   `json:"email"`
	Roles   []string `json:"roles"`
}

func main() {
	ctx, cancel := logger.NewContextWithLogger(context.Background())
	defer cancel()
	log := logger.FromContext(ctx)

	// Create an auth provider with the required configuration
	provider, err := middleware.NewAuthProvider(ctx, &middleware.AuthConfig{
		Config: oidc.Config{
			ClientID: os.Getenv("CLIENT_ID"),
		},
		ProviderURL:  os.Getenv("PROVIDER_URL"),
		ClientSecret: os.Getenv("CLIENT_SECRET"),
		RedirectURL:  os.Getenv("REDIRECT_URL"),
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	})
	if err != nil {
		log.FatalContext(ctx, "Failed to create auth provider", err)
	}

	// Create a new server and mount a route with the auth middleware
	server := apimanager.New(&apimanager.Config{
		Address: ":8080",
	})
	err = server.Mount(apimanager.Route{
		Path:    "/",
		Methods: []string{http.MethodGet},
		Handler: func(c fiber.Ctx) error {
			return c.Status(http.StatusOK).JSON(fiber.Map{
				"message": "Hello, World!",
			})
		},
		Middlewares: []fiber.Handler{
			middleware.AuthenticateWithClaims[Claims](provider),
		},
	})
	if err != nil {
		log.FatalContext(ctx, "Failed to mount route", err)
	}

	// Run the server (runs in a goroutine because it'd block the main goroutine)
	cErr := make(chan error, 1)
	go func() {
		if err = server.Run(ctx); err != nil {
			cErr <- err
		}
	}()

	// Simulate obtaining an authorization code and send a request to the server
	err = sendRequestFromClient(ctx)
	if err != nil {
		_ = server.Shutdown(ctx)
		log.FatalContext(ctx, "Failed to request token", err)
	}

	// Block until the server stops running
	err = <-cErr
	if err != nil {
		log.FatalContext(ctx, "Failed to run server", err)
	}
}

// sendRequestFromClient sends a request to the server with a token from the auth provider.
func sendRequestFromClient(ctx context.Context) error {
	log := logger.FromContext(ctx)

	// Simulate obtaining an authorization code
	code, err := obtainAuthCode(ctx)
	if err != nil {
		log.ErrorContext(ctx, "Failed to obtain auth code", err)
		return err
	}

	// Request a token from the auth provider using the authorization code
	tk, err := requestToken(ctx, code)
	if err != nil {
		log.ErrorContext(ctx, "Failed to request token", err)
		return err
	}
	log.InfoContext(ctx, "Token received", "token", tk)

	// Send a request to the server with the token
	return requestServer(ctx, tk)
}

type Token struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

// obtainAuthCode simulates obtaining an authorization code.
func obtainAuthCode(_ context.Context) (string, error) {
	// Simulate user login and authorization process
	// In real scenarios, this would be handled by redirecting the user to the provider's login page
	// and then redirecting back to your application with the authorization code.
	return "example_auth_code", nil
}

// requestToken requests a new access token from the auth provider using the authorization code.
func requestToken(ctx context.Context, code string) (*Token, error) {
	// Get the required environment variables
	clientID := os.Getenv("CLIENT_ID")
	clientSecret := os.Getenv("CLIENT_SECRET")
	redirectURL := os.Getenv("REDIRECT_URL")
	providerURL := os.Getenv("PROVIDER_URL")

	// Create the OAuth2 configuration
	conf := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes:       []string{"openid", "profile", "email"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  fmt.Sprintf("%s/auth", providerURL),
			TokenURL: fmt.Sprintf("%s/token", providerURL),
		},
	}

	// Exchange the authorization code for a token
	tok, err := conf.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange token: %w", err)
	}

	token := &Token{
		AccessToken: tok.AccessToken,
		TokenType:   tok.TokenType,
		ExpiresIn:   int(time.Until(tok.Expiry).Seconds()),
	}
	return token, nil
}

type errorResponse struct {
	Error struct {
		Message string `json:"message"`
		Code    int    `json:"code"`
	} `json:"error"`
}

// requestServer sends a request to the server with the given token.
func requestServer(ctx context.Context, tk *Token) error {
	log := logger.FromContext(ctx)

	// Create a request with the token in the Authorization header
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost:8080", http.NoBody)
	if err != nil {
		log.ErrorContext(ctx, "Failed to create request", err)
		return err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tk.AccessToken))

	// Send the request to the server
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.ErrorContext(ctx, "Failed to send request", err)
		return err
	}
	defer func() {
		err = errors.Join(err, resp.Body.Close())
	}()

	// Check the response status code and handle accordingly
	var errResp errorResponse
	switch resp.StatusCode {
	case http.StatusOK:
		log.InfoContext(ctx, "Request successful")
		return nil
	case http.StatusUnauthorized, http.StatusForbidden, http.StatusInternalServerError:
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			log.ErrorContext(ctx, "Failed to decode error response", err)
			return err
		}
		return errors.New(errResp.Error.Message)
	default:
		return errors.New("unexpected status code")
	}
}
