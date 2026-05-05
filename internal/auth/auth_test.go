package auth

import (
	"net/http"
	"testing"
	"time"
)

func TestValidateBearerAcceptsWellFormedUnsignedJWT(t *testing.T) {
	t.Parallel()

	token := MakeTestToken(t, time.Now().Add(time.Hour))
	req, err := http.NewRequest(http.MethodGet, "/mobile/home", nil)
	if err != nil {
		t.Fatalf("リクエストを作成できませんでした: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	claims, err := ValidateBearer(req)
	if err != nil {
		t.Fatalf("有効なBearerトークンでエラーになりました: %v", err)
	}
	if claims.Subject != "user-123" {
		t.Fatalf("subject = %q, want user-123", claims.Subject)
	}
}

func TestValidateBearerRejectsExpiredToken(t *testing.T) {
	t.Parallel()

	token := MakeTestToken(t, time.Now().Add(-time.Minute))
	req, err := http.NewRequest(http.MethodGet, "/mobile/home", nil)
	if err != nil {
		t.Fatalf("リクエストを作成できませんでした: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	_, err = ValidateBearer(req)
	if err == nil {
		t.Fatal("期限切れトークンでエラーになりませんでした")
	}
}

func TestValidateTokenRejectsInvalidSignature(t *testing.T) {
	t.Parallel()

	token, err := NewToken("user-123", time.Now().Add(time.Hour), []byte("secret-a"))
	if err != nil {
		t.Fatalf("トークンを作成できませんでした: %v", err)
	}

	_, err = ValidateToken(token, []byte("secret-b"), time.Now())
	if err == nil {
		t.Fatal("不正な署名でエラーになりませんでした")
	}
}

func TestValidateTokenRejectsMalformedToken(t *testing.T) {
	t.Parallel()

	_, err := ValidateToken("not-a-jwt", []byte("secret"), time.Now())
	if err == nil {
		t.Fatal("不正な形式のトークンでエラーになりませんでした")
	}
}
