package cmd

import "strings"

const callbackDefaultLang = "en"

// callbackPageMessages 保存 OAuth callback 页面上的固定文案。
// OAuth 服务端返回的错误内容不放入这里翻译，避免排障时和服务端日志无法对齐。
type callbackPageMessages struct {
	DocumentTitleSuccess string `json:"documentTitleSuccess"`
	DocumentTitleFailure string `json:"documentTitleFailure"`
	SuccessTitle         string `json:"successTitle"`
	FailureTitle         string `json:"failureTitle"`
	SuccessCopy          string `json:"successCopy"`
	FailureCopy          string `json:"failureCopy"`
	OAuthErrorLabel      string `json:"oauthErrorLabel"`
}

// callbackPageData 是注入 callback.html 的唯一动态数据入口。
// 页面端只负责用 textContent 渲染这些值，避免把错误内容当作 HTML 执行。
type callbackPageData struct {
	Lang         string               `json:"lang"`
	ErrorMessage string               `json:"errorMessage"`
	Messages     callbackPageMessages `json:"messages"`
}

var callbackMessagesByLang = map[string]callbackPageMessages{
	"en": {
		DocumentTitleSuccess: "BytePlus Authentication Successful",
		DocumentTitleFailure: "BytePlus Authentication Failed",
		SuccessTitle:         "Authentication successful",
		FailureTitle:         "Authentication failed",
		SuccessCopy:          "You can close this page and return to\nthe terminal.",
		FailureCopy:          "Please return to the terminal.",
		OAuthErrorLabel:      "OAuth error",
	},
	"zh": {
		DocumentTitleSuccess: "BytePlus 认证成功",
		DocumentTitleFailure: "BytePlus 认证失败",
		SuccessTitle:         "认证成功",
		FailureTitle:         "认证失败",
		SuccessCopy:          "你可以关闭此页面并返回\n终端继续操作。",
		FailureCopy:          "请返回终端继续操作。",
		OAuthErrorLabel:      "OAuth 错误",
	},
	"zh-Hant-TW": {
		DocumentTitleSuccess: "BytePlus 驗證成功",
		DocumentTitleFailure: "BytePlus 驗證失敗",
		SuccessTitle:         "驗證成功",
		FailureTitle:         "驗證失敗",
		SuccessCopy:          "你可以關閉此頁面並返回\n終端機繼續操作。",
		FailureCopy:          "請返回終端機繼續操作。",
		OAuthErrorLabel:      "OAuth 錯誤",
	},
	"ja-JP": {
		DocumentTitleSuccess: "BytePlus 認証成功",
		DocumentTitleFailure: "BytePlus 認証失敗",
		SuccessTitle:         "認証に成功しました",
		FailureTitle:         "認証に失敗しました",
		SuccessCopy:          "このページを閉じて\nターミナルに戻れます。",
		FailureCopy:          "ターミナルに戻って操作を続けてください。",
		OAuthErrorLabel:      "OAuth エラー",
	},
	"ko-KR": {
		DocumentTitleSuccess: "BytePlus 인증 성공",
		DocumentTitleFailure: "BytePlus 인증 실패",
		SuccessTitle:         "인증에 성공했습니다",
		FailureTitle:         "인증에 실패했습니다",
		SuccessCopy:          "이 페이지를 닫고\n터미널로 돌아가세요.",
		FailureCopy:          "터미널로 돌아가 계속 진행하세요.",
		OAuthErrorLabel:      "OAuth 오류",
	},
	"id-ID": {
		DocumentTitleSuccess: "Autentikasi BytePlus Berhasil",
		DocumentTitleFailure: "Autentikasi BytePlus Gagal",
		SuccessTitle:         "Autentikasi berhasil",
		FailureTitle:         "Autentikasi gagal",
		SuccessCopy:          "Anda dapat menutup halaman ini dan kembali\nke terminal.",
		FailureCopy:          "Silakan kembali ke terminal.",
		OAuthErrorLabel:      "Kesalahan OAuth",
	},
	"vi-VN": {
		DocumentTitleSuccess: "Xác thực BytePlus thành công",
		DocumentTitleFailure: "Xác thực BytePlus thất bại",
		SuccessTitle:         "Xác thực thành công",
		FailureTitle:         "Xác thực thất bại",
		SuccessCopy:          "Bạn có thể đóng trang này và quay lại\nterminal.",
		FailureCopy:          "Vui lòng quay lại terminal.",
		OAuthErrorLabel:      "Lỗi OAuth",
	},
	"th-TH": {
		DocumentTitleSuccess: "การยืนยันตัวตน BytePlus สำเร็จ",
		DocumentTitleFailure: "การยืนยันตัวตน BytePlus ล้มเหลว",
		SuccessTitle:         "ยืนยันตัวตนสำเร็จ",
		FailureTitle:         "ยืนยันตัวตนล้มเหลว",
		SuccessCopy:          "คุณสามารถปิดหน้านี้และกลับไปที่\nเทอร์มินัลได้",
		FailureCopy:          "โปรดกลับไปที่เทอร์มินัล",
		OAuthErrorLabel:      "ข้อผิดพลาด OAuth",
	},
	"pt": {
		DocumentTitleSuccess: "Autenticação BytePlus bem-sucedida",
		DocumentTitleFailure: "Falha na autenticação BytePlus",
		SuccessTitle:         "Autenticação bem-sucedida",
		FailureTitle:         "Falha na autenticação",
		SuccessCopy:          "Você pode fechar esta página e voltar\nao terminal.",
		FailureCopy:          "Volte ao terminal.",
		OAuthErrorLabel:      "Erro OAuth",
	},
	"es": {
		DocumentTitleSuccess: "Autenticación de BytePlus correcta",
		DocumentTitleFailure: "Error de autenticación de BytePlus",
		SuccessTitle:         "Autenticación correcta",
		FailureTitle:         "Error de autenticación",
		SuccessCopy:          "Puedes cerrar esta página y volver\na la terminal.",
		FailureCopy:          "Vuelve a la terminal.",
		OAuthErrorLabel:      "Error de OAuth",
	},
	"fr": {
		DocumentTitleSuccess: "Authentification BytePlus réussie",
		DocumentTitleFailure: "Échec de l'authentification BytePlus",
		SuccessTitle:         "Authentification réussie",
		FailureTitle:         "Échec de l'authentification",
		SuccessCopy:          "Vous pouvez fermer cette page et revenir\nau terminal.",
		FailureCopy:          "Veuillez revenir au terminal.",
		OAuthErrorLabel:      "Erreur OAuth",
	},
	"de-DE": {
		DocumentTitleSuccess: "BytePlus-Authentifizierung erfolgreich",
		DocumentTitleFailure: "BytePlus-Authentifizierung fehlgeschlagen",
		SuccessTitle:         "Authentifizierung erfolgreich",
		FailureTitle:         "Authentifizierung fehlgeschlagen",
		SuccessCopy:          "Sie können diese Seite schließen und zum\nTerminal zurückkehren.",
		FailureCopy:          "Bitte kehren Sie zum Terminal zurück.",
		OAuthErrorLabel:      "OAuth-Fehler",
	},
}

var callbackLangAliases = map[string]string{
	"en":         "en",
	"en-us":      "en",
	"en-gb":      "en",
	"zh":         "zh",
	"zh-cn":      "zh",
	"zh-hans":    "zh",
	"zh-hans-cn": "zh",
	"zh-tw":      "zh-Hant-TW",
	"zh-hk":      "zh-Hant-TW",
	"zh-mo":      "zh-Hant-TW",
	"zh-hant":    "zh-Hant-TW",
	"zh-hant-tw": "zh-Hant-TW",
	"ja":         "ja-JP",
	"ja-jp":      "ja-JP",
	"ko":         "ko-KR",
	"ko-kr":      "ko-KR",
	"id":         "id-ID",
	"id-id":      "id-ID",
	"vi":         "vi-VN",
	"vi-vn":      "vi-VN",
	"th":         "th-TH",
	"th-th":      "th-TH",
	"pt":         "pt",
	"pt-br":      "pt",
	"pt-pt":      "pt",
	"es":         "es",
	"es-es":      "es",
	"es-mx":      "es",
	"fr":         "fr",
	"fr-fr":      "fr",
	"de":         "de-DE",
	"de-de":      "de-DE",
}

// normalizeCallbackLang 将 URL 参数中的语言码归一成页面支持的规范语言码。
// 只接受已明确支持的语言，未知值统一回退英文，避免页面渲染空文案。
func normalizeCallbackLang(lang string) string {
	normalized := strings.ToLower(strings.ReplaceAll(strings.TrimSpace(lang), "_", "-"))
	if normalized == "" {
		return callbackDefaultLang
	}

	if canonical, ok := callbackLangAliases[normalized]; ok {
		return canonical
	}

	// 对地区变体做保守回退，例如 es-AR -> es、fr-CA -> fr。
	if base := strings.Split(normalized, "-")[0]; base != "" {
		if canonical, ok := callbackLangAliases[base]; ok {
			return canonical
		}
	}

	return callbackDefaultLang
}

func callbackMessagesForLang(lang string) callbackPageMessages {
	messages, ok := callbackMessagesByLang[normalizeCallbackLang(lang)]
	if ok {
		return messages
	}
	return callbackMessagesByLang[callbackDefaultLang]
}

func newCallbackPageData(errorMessage, lang string) callbackPageData {
	normalizedLang := normalizeCallbackLang(lang)
	return callbackPageData{
		Lang:         normalizedLang,
		ErrorMessage: errorMessage,
		Messages:     callbackMessagesForLang(normalizedLang),
	}
}
