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

func TestNormalizeCallbackLang(t *testing.T) {
	tests := []struct {
		name string
		lang string
		want string
	}{
		{name: "default empty lang", lang: "", want: "en"},
		{name: "keeps supported simplified Chinese", lang: "zh", want: "zh"},
		{name: "maps Chinese mainland alias", lang: "zh-CN", want: "zh"},
		{name: "maps traditional Chinese alias", lang: "zh-TW", want: "zh-Hant-TW"},
		{name: "maps Japanese base language", lang: "ja", want: "ja-JP"},
		{name: "maps Korean base language", lang: "ko", want: "ko-KR"},
		{name: "maps German base language", lang: "de", want: "de-DE"},
		{name: "falls back unsupported lang", lang: "unsupported", want: "en"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := normalizeCallbackLang(tc.lang); got != tc.want {
				t.Fatalf("normalizeCallbackLang(%q) = %q, want %q", tc.lang, got, tc.want)
			}
		})
	}
}

func TestRenderCallbackPageInjectsServerErrorMessageSafely(t *testing.T) {
	maliciousError := "</script><script>alert(1)</script>"
	page, err := renderCallbackPage(maliciousError, "en")
	if err != nil {
		t.Fatalf("failed to render callback page: %v", err)
	}

	got := string(page)
	if strings.Contains(got, maliciousError) {
		t.Fatalf("rendered page must not inject raw server-side oauth error text")
	}
	if !strings.Contains(got, `\u003c/script\u003e\u003cscript\u003ealert(1)\u003c/script\u003e`) {
		t.Fatalf("rendered page should inject JSON-escaped server-side oauth error text")
	}
	if !strings.Contains(got, "textContent = title") {
		t.Fatalf("rendered page should write oauth error text through textContent")
	}
}

func TestRenderCallbackPageContainsDefaultSuccessState(t *testing.T) {
	page, err := renderCallbackPage("", "en")
	if err != nil {
		t.Fatalf("failed to render callback page: %v", err)
	}

	if !strings.Contains(string(page), `Authentication successful`) {
		t.Fatalf("rendered page should contain default success state")
	}
}

func TestRenderCallbackPageLocalizesChineseSuccessState(t *testing.T) {
	page, err := renderCallbackPage("", "zh")
	if err != nil {
		t.Fatalf("failed to render callback page: %v", err)
	}

	got := string(page)
	for _, want := range []string{`"lang":"zh"`, `认证成功`, `你可以关闭此页面并返回`} {
		if !strings.Contains(got, want) {
			t.Fatalf("rendered page does not contain localized Chinese success snippet %q", want)
		}
	}
}

func TestRenderCallbackPageFallsBackUnsupportedLangToEnglish(t *testing.T) {
	page, err := renderCallbackPage("", "unsupported")
	if err != nil {
		t.Fatalf("failed to render callback page: %v", err)
	}

	got := string(page)
	if !strings.Contains(got, `"lang":"en"`) {
		t.Fatalf("unsupported lang should fall back to en")
	}
	if !strings.Contains(got, `Authentication successful`) {
		t.Fatalf("unsupported lang should use English messages")
	}
}

func TestRenderCallbackPageUsesLocalizedFailureTitleAndOriginalServerError(t *testing.T) {
	page, err := renderCallbackPage("invalid_request: denied", "ja-JP")
	if err != nil {
		t.Fatalf("failed to render callback page: %v", err)
	}

	got := string(page)
	if !strings.Contains(got, `認証に失敗しました`) {
		t.Fatalf("rendered page should contain localized failure title")
	}
	if !strings.Contains(got, `invalid_request: denied`) {
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

	req := httptest.NewRequest(http.MethodGet, "/oauth/callback?lang=zh&error=invalid_request&error_description=%3Cscript%3Eboom%3C%2Fscript%3E", nil)
	recorder := httptest.NewRecorder()

	server.handleCallback(recorder, req)
	body := recorder.Body.String()

	if !strings.Contains(body, "认证失败") {
		t.Fatalf("fallback page should indicate authentication failure")
	}
	if !strings.Contains(body, "OAuth 错误: invalid_request: &lt;script&gt;boom&lt;/script&gt;") {
		t.Fatalf("fallback page should include escaped oauth error details")
	}
	if strings.Contains(body, "Page render error:") {
		t.Fatalf("fallback page should not expose render errors")
	}
	if strings.Contains(body, "<script>boom</script>") || strings.Contains(body, `</script><script>alert("xss")</script>`) || strings.Contains(body, "render failure") {
		t.Fatalf("fallback page must not contain unescaped script content")
	}
}
