package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const (
	// defaultConsoleEndpoint is the default Byteplus console sign-in endpoint.
	defaultConsoleEndpoint = "https://signin.byteplus.com"

	// consoleAuthorizePath is the path appended to the endpoint for the authorization URL.
	consoleAuthorizePath = "/authorize/oauth/authorize"

	// consoleTokenPath is the path appended to the endpoint for the token URL.
	consoleTokenPath = "/authorize/oauth/token"

	// consoleTokenRequestTimeout is the HTTP timeout for console token exchange requests.
	consoleTokenRequestTimeout = 30 * time.Second

	// consoleTokenRetryAttempts is the number of retry attempts for token exchange.
	consoleTokenRetryAttempts = 3

	// ConsoleClientIDSameDevice is the public client ID for local/same-device login mode.
	ConsoleClientIDSameDevice = "trn:signin:::devtools/same-device"

	// ConsoleClientIDCrossDevice is the public client ID for remote/cross-device login mode.
	ConsoleClientIDCrossDevice = "trn:signin:::devtools/cross-device"
)

// ---------------------------------------------------------------------------
// Error types (independent from SSO OAuthAPIError)
// ---------------------------------------------------------------------------

type ConsoleOAuthErrorResponse struct {
	State            string `json:"state,omitempty"`
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description,omitempty"`
	ErrorURI         string `json:"error_uri,omitempty"`
}

type ConsoleOAuthAPIError struct {
	StatusCode int
	Response   ConsoleOAuthErrorResponse
	RawBody    string
	RequestID  string // X-Tt-Logid header
}

func (e *ConsoleOAuthAPIError) Error() string {
	if e == nil {
		return ""
	}

	var parts []string
	if e.Response.Error != "" {
		parts = append(parts, e.Response.Error)
	}
	if e.Response.ErrorDescription != "" {
		parts = append(parts, e.Response.ErrorDescription)
	}

	msg := strings.Join(parts, ": ")
	if msg == "" {
		if e.RawBody != "" {
			msg = e.RawBody
		} else {
			msg = "unknown error"
		}
	}

	suffix := fmt.Sprintf("[status %d", e.StatusCode)
	if e.RequestID != "" {
		suffix += ", requestId: " + e.RequestID
	}
	suffix += "]"

	return fmt.Sprintf("console oauth request failed: %s %s", msg, suffix)
}

func (e *ConsoleOAuthAPIError) IsRetryable() bool {
	if e == nil {
		return false
	}
	return e.StatusCode == http.StatusTooManyRequests ||
		e.StatusCode == http.StatusRequestTimeout ||
		e.StatusCode/100 == 5
}

// ---------------------------------------------------------------------------
// Client config & types
// ---------------------------------------------------------------------------

type ConsoleOAuthClientConfig struct {
	EndpointURL string
	HTTPClient  *http.Client
}

type ConsoleOAuthClient struct {
	endpointURL  string
	authorizeURL string
	tokenURL     string
	httpClient   *http.Client
}

type AuthorizeParams struct {
	ClientID            string
	RedirectURI         string
	Scope               string
	State               string
	CodeChallenge       string
	CodeChallengeMethod string // e.g. "S256"
}

type ConsoleTokenRequest struct {
	GrantType    string // "authorization_code" or "refresh_token"
	Code         string // authorization code (for auth_code grant)
	RedirectURI  string // must match the one used in the authorize request
	ClientID     string
	Scope        string
	CodeVerifier string // PKCE verifier (for auth_code grant)
	RefreshToken string // for refresh_token grant
}

type ConsoleTokenResponse struct {
	AccessToken  string `json:"access_token"`  // JSON string containing STS credentials
	TokenType    string `json:"token_type"`    // e.g. "urn:ietf:params:oauth:token-type:access_token_sts"
	ExpiresIn    int    `json:"expires_in"`    // seconds, e.g. 900
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
	IDToken      string `json:"id_token"` // JWT
}

type STSCredentials struct {
	AccessKeyID     string `json:"access_key_id"`
	SecretAccessKey string `json:"secret_access_key"`
	SessionToken    string `json:"session_token"`
}

// ---------------------------------------------------------------------------
// Constructor
// ---------------------------------------------------------------------------

func NewConsoleOAuthClient(cfg *ConsoleOAuthClientConfig) *ConsoleOAuthClient {
	endpoint := defaultConsoleEndpoint
	if cfg != nil && strings.TrimSpace(cfg.EndpointURL) != "" {
		endpoint = strings.TrimSpace(cfg.EndpointURL)
	}
	endpoint = strings.TrimRight(endpoint, "/")

	client := &http.Client{Timeout: consoleTokenRequestTimeout}
	if cfg != nil && cfg.HTTPClient != nil {
		client = cfg.HTTPClient
	}

	return &ConsoleOAuthClient{
		endpointURL:  endpoint,
		authorizeURL: endpoint + consoleAuthorizePath,
		tokenURL:     endpoint + consoleTokenPath,
		httpClient:   client,
	}
}

// ---------------------------------------------------------------------------
// BuildAuthorizeURL
// ---------------------------------------------------------------------------

func (c *ConsoleOAuthClient) BuildAuthorizeURL(params *AuthorizeParams) string {
	q := url.Values{}
	q.Set("response_type", "code")

	if params.ClientID != "" {
		q.Set("client_id", params.ClientID)
	}
	if params.RedirectURI != "" {
		q.Set("redirect_uri", params.RedirectURI)
	}
	if params.Scope != "" {
		q.Set("scope", params.Scope)
	}
	if params.State != "" {
		q.Set("state", params.State)
	}
	if params.CodeChallenge != "" {
		q.Set("code_challenge", params.CodeChallenge)
	}
	if params.CodeChallengeMethod != "" {
		q.Set("code_challenge_method", params.CodeChallengeMethod)
	}

	return c.authorizeURL + "?" + q.Encode()
}

// ---------------------------------------------------------------------------
// ExchangeToken
// ---------------------------------------------------------------------------

func (c *ConsoleOAuthClient) ExchangeToken(ctx context.Context, req *ConsoleTokenRequest) (*ConsoleTokenResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}
	if strings.TrimSpace(req.GrantType) == "" {
		return nil, fmt.Errorf("grant_type is required")
	}
	if strings.TrimSpace(req.ClientID) == "" {
		return nil, fmt.Errorf("client_id is required")
	}

	q := url.Values{}
	q.Set("grant_type", req.GrantType)
	q.Set("client_id", req.ClientID)

	if req.Scope != "" {
		q.Set("scope", req.Scope)
	}

	switch req.GrantType {
	case "authorization_code":
		if strings.TrimSpace(req.Code) == "" {
			return nil, fmt.Errorf("code is required for authorization_code grant")
		}
		if strings.TrimSpace(req.CodeVerifier) == "" {
			return nil, fmt.Errorf("code_verifier is required for authorization_code grant")
		}
		q.Set("code", req.Code)
		q.Set("code_verifier", req.CodeVerifier)
		if req.RedirectURI != "" {
			q.Set("redirect_uri", req.RedirectURI)
		}

	case "refresh_token":
		if strings.TrimSpace(req.RefreshToken) == "" {
			return nil, fmt.Errorf("refresh_token is required for refresh_token grant")
		}
		q.Set("refresh_token", req.RefreshToken)

	default:
		return nil, fmt.Errorf("unsupported grant_type: %s", req.GrantType)
	}

	requestBody := q.Encode()

	var tokenResp ConsoleTokenResponse
	err := doWithRetry(ctx, retryOptions{maxAttempts: consoleTokenRetryAttempts}, func() error {
		httpReq, reqErr := http.NewRequestWithContext(ctx, http.MethodPost, c.tokenURL, strings.NewReader(requestBody))
		if reqErr != nil {
			return fmt.Errorf("failed to build request: %w", reqErr)
		}
		httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		if customHeaders := os.Getenv("BYTEPLUS_LOGIN_HEADERS"); customHeaders != "" {
			for _, entry := range strings.Split(customHeaders, ";") {
				if idx := strings.Index(entry, "="); idx > 0 {
					httpReq.Header.Set(strings.TrimSpace(entry[:idx]), strings.TrimSpace(entry[idx+1:]))
				}
			}
		}
		resp, doErr := c.httpClient.Do(httpReq)
		if doErr != nil {
			return fmt.Errorf("request failed: %w", doErr)
		}
		defer resp.Body.Close()

		respBytes, readErr := ioutil.ReadAll(resp.Body)
		if readErr != nil {
			return fmt.Errorf("failed to read response: %w", readErr)
		}

		requestID := resp.Header.Get("X-Tt-Logid")

		if resp.StatusCode/100 != 2 {
			apiErr := &ConsoleOAuthAPIError{
				StatusCode: resp.StatusCode,
				RequestID:  requestID,
				RawBody:    string(respBytes),
			}

			if len(respBytes) > 0 {
				var errResp ConsoleOAuthErrorResponse
				if json.Unmarshal(respBytes, &errResp) == nil && errResp.Error != "" {
					apiErr.Response = errResp
				}
			}

			return apiErr
		}

		if len(respBytes) > 0 {
			if unmarshalErr := json.Unmarshal(respBytes, &tokenResp); unmarshalErr != nil {
				return fmt.Errorf(
					"failed to decode token response (status %d, requestId: %s): %w",
					resp.StatusCode, requestID, unmarshalErr,
				)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	if tokenResp.AccessToken == "" && tokenResp.TokenType == "" &&
		tokenResp.RefreshToken == "" && tokenResp.ExpiresIn == 0 {
		return nil, fmt.Errorf("ExchangeToken succeeded but response was empty")
	}

	return &tokenResp, nil
}

// ---------------------------------------------------------------------------
// ParseSTSCredentials
// ---------------------------------------------------------------------------

func ParseSTSCredentials(accessToken string) (*STSCredentials, error) {
	if strings.TrimSpace(accessToken) == "" {
		return nil, fmt.Errorf("access_token is empty")
	}

	var creds STSCredentials
	if err := json.Unmarshal([]byte(accessToken), &creds); err != nil {
		return nil, fmt.Errorf("failed to parse STS credentials from access_token: %w", err)
	}

	if creds.AccessKeyID == "" {
		return nil, fmt.Errorf("parsed STS credentials missing access_key_id")
	}
	if creds.SecretAccessKey == "" {
		return nil, fmt.Errorf("parsed STS credentials missing secret_access_key")
	}
	if creds.SessionToken == "" {
		return nil, fmt.Errorf("parsed STS credentials missing session_token")
	}

	return &creds, nil
}
