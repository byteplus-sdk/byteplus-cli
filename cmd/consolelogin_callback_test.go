package cmd

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

func TestHandleCallbackDoesNotDoubleDecode(t *testing.T) {
	server, err := NewCallbackServer()
	if err != nil {
		t.Fatalf("failed to create callback server: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/oauth/callback?error=invalid_request&error_description=%2B", nil)
	recorder := httptest.NewRecorder()

	server.handleCallback(recorder, req)

	select {
	case result := <-server.result:
		if result.ErrorDescription != "+" {
			t.Fatalf("unexpected error description: got %q, want %q", result.ErrorDescription, "+")
		}
	default:
		t.Fatalf("callback result was not delivered")
	}
}

func TestHandleCallbackErrorPriority(t *testing.T) {
	tests := []struct {
		name                 string
		query                string
		wantError            string
		wantErrorDescription string
	}{
		{
			name:                 "error has highest priority",
			query:                "/oauth/callback?error=from_error&Error=from_Error&error_description=from_desc",
			wantError:            "from_error",
			wantErrorDescription: "from_desc",
		},
		{
			name:                 "Error used when error missing",
			query:                "/oauth/callback?Error=from_Error&error_description=from_desc",
			wantError:            "from_Error",
			wantErrorDescription: "from_desc",
		},
		{
			name:                 "error_description used as fallback when both missing",
			query:                "/oauth/callback?error_description=from_desc",
			wantError:            "from_desc",
			wantErrorDescription: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			server, err := NewCallbackServer()
			if err != nil {
				t.Fatalf("failed to create callback server: %v", err)
			}

			req := httptest.NewRequest(http.MethodGet, tc.query, nil)
			recorder := httptest.NewRecorder()
			server.handleCallback(recorder, req)

			select {
			case result := <-server.result:
				if result.Error != tc.wantError {
					t.Fatalf("unexpected error: got %q, want %q", result.Error, tc.wantError)
				}
				if result.ErrorDescription != tc.wantErrorDescription {
					t.Fatalf("unexpected error description: got %q, want %q", result.ErrorDescription, tc.wantErrorDescription)
				}
			default:
				t.Fatalf("callback result was not delivered")
			}
		})
	}
}

func TestRenderCallbackPageInjectsServerErrorMessageSafely(t *testing.T) {
	maliciousError := "</script><script>alert(1)</script>"
	page, err := renderCallbackPage(maliciousError)
	if err != nil {
		t.Fatalf("failed to render callback page: %v", err)
	}

	got := string(page)
	if strings.Contains(got, maliciousError) {
		t.Fatalf("rendered page must not inject raw server-side oauth error text")
	}
	if !strings.Contains(got, `const callbackError = "<\/script><script>alert(1)<\/script>";`) {
		t.Fatalf("rendered page should inject escaped server-side oauth error text")
	}
	if !strings.Contains(got, "textContent = title") {
		t.Fatalf("rendered page should write oauth error text through textContent")
	}
}

func TestRenderCallbackPageContainsDefaultSuccessState(t *testing.T) {
	page, err := renderCallbackPage("")
	if err != nil {
		t.Fatalf("failed to render callback page: %v", err)
	}

	if !strings.Contains(string(page), `Authentication successful`) {
		t.Fatalf("rendered page should contain default success state")
	}
}

func TestRenderCallbackPageUsesServerErrorForFailureState(t *testing.T) {
	page, err := renderCallbackPage("invalid_request: denied")
	if err != nil {
		t.Fatalf("failed to render callback page: %v", err)
	}

	got := string(page)
	if !strings.Contains(got, `const callbackError = "invalid_request: denied";`) {
		t.Fatalf("rendered page should receive the normalized oauth error from the callback server")
	}
	if !strings.Contains(got, `document.documentElement.dataset.state = hasError ? "error" : "success";`) {
		t.Fatalf("rendered page should switch success and failure states from the server error")
	}
}

func TestHandleCallbackFallbackEscapesErrorDetails(t *testing.T) {
	server, err := NewCallbackServer()
	if err != nil {
		t.Fatalf("failed to create callback server: %v", err)
	}

	// Force renderCallbackPage to fail so that fallback HTML is used.
	savedOnce := callbackTemplateOnce
	savedTemplate := callbackTemplate
	savedErr := callbackTemplateErr
	callbackTemplateOnce = sync.Once{}
	callbackTemplateOnce.Do(func() {})
	callbackTemplate = nil
	callbackTemplateErr = errors.New(`render failure </script><script>alert("xss")</script>`)
	defer func() {
		callbackTemplateOnce = savedOnce
		callbackTemplate = savedTemplate
		callbackTemplateErr = savedErr
	}()

	req := httptest.NewRequest(http.MethodGet, "/oauth/callback?error=invalid_request&error_description=%3Cscript%3Eboom%3C%2Fscript%3E", nil)
	recorder := httptest.NewRecorder()

	server.handleCallback(recorder, req)
	body := recorder.Body.String()

	if !strings.Contains(body, "Authentication failed") {
		t.Fatalf("fallback page should indicate authentication failure")
	}
	if !strings.Contains(body, "OAuth error: invalid_request: &lt;script&gt;boom&lt;/script&gt;") {
		t.Fatalf("fallback page should include escaped oauth error details")
	}
	if strings.Contains(body, "Page render error:") {
		t.Fatalf("fallback page should not expose render errors")
	}
	if strings.Contains(body, "<script>boom</script>") || strings.Contains(body, `</script><script>alert("xss")</script>`) || strings.Contains(body, "render failure") {
		t.Fatalf("fallback page must not contain unescaped script content")
	}
}
