package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"
)

const defaultSecret = "local-development-secret"

// Claims はJWTから取り出した最小限の認証情報です。
type Claims struct {
	Subject   string
	ExpiresAt time.Time
}

type tokenHeader struct {
	Algorithm string `json:"alg"`
	Type      string `json:"typ"`
}

type tokenPayload struct {
	Subject string `json:"sub"`
	Expires int64  `json:"exp"`
}

// ValidateBearer はAuthorizationヘッダーのBearer JWTを検証します。
func ValidateBearer(req *http.Request) (Claims, error) {
	header := req.Header.Get("Authorization")
	if header == "" {
		return Claims{}, errors.New("authorization header is required")
	}
	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return Claims{}, errors.New("authorization header must use bearer scheme")
	}
	return ValidateToken(strings.TrimPrefix(header, prefix), []byte(defaultSecret), time.Now())
}

// ValidateToken はHS256署名付きJWTを検証します。
func ValidateToken(token string, secret []byte, now time.Time) (Claims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return Claims{}, errors.New("token must have three parts")
	}

	var hdr tokenHeader
	if err := decodePart(parts[0], &hdr); err != nil {
		return Claims{}, fmt.Errorf("decode token header: %w", err)
	}
	if hdr.Algorithm != "HS256" || hdr.Type != "JWT" {
		return Claims{}, errors.New("token header must be HS256 JWT")
	}

	signingInput := parts[0] + "." + parts[1]
	want := sign(signingInput, secret)
	if !hmac.Equal([]byte(parts[2]), []byte(want)) {
		return Claims{}, errors.New("token signature is invalid")
	}

	var payload tokenPayload
	if err := decodePart(parts[1], &payload); err != nil {
		return Claims{}, fmt.Errorf("decode token payload: %w", err)
	}
	if payload.Subject == "" {
		return Claims{}, errors.New("token subject is required")
	}
	expiresAt := time.Unix(payload.Expires, 0)
	if !expiresAt.After(now) {
		return Claims{}, errors.New("token is expired")
	}

	return Claims{Subject: payload.Subject, ExpiresAt: expiresAt}, nil
}

// MakeTestToken はテスト用のHS256 JWTを作成します。
func MakeTestToken(t testing.TB, expiresAt time.Time) string {
	t.Helper()
	token, err := NewToken("user-123", expiresAt, []byte(defaultSecret))
	if err != nil {
		t.Fatalf("テスト用トークンを作成できませんでした: %v", err)
	}
	return token
}

// NewToken は指定したsubjectと期限でHS256 JWTを作成します。
func NewToken(subject string, expiresAt time.Time, secret []byte) (string, error) {
	headerJSON, err := json.Marshal(tokenHeader{Algorithm: "HS256", Type: "JWT"})
	if err != nil {
		return "", fmt.Errorf("marshal token header: %w", err)
	}
	payloadJSON, err := json.Marshal(tokenPayload{Subject: subject, Expires: expiresAt.Unix()})
	if err != nil {
		return "", fmt.Errorf("marshal token payload: %w", err)
	}

	headerPart := base64.RawURLEncoding.EncodeToString(headerJSON)
	payloadPart := base64.RawURLEncoding.EncodeToString(payloadJSON)
	signingInput := headerPart + "." + payloadPart
	return signingInput + "." + sign(signingInput, secret), nil
}

func decodePart(part string, dst any) error {
	data, err := base64.RawURLEncoding.DecodeString(part)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dst)
}

func sign(input string, secret []byte) string {
	mac := hmac.New(sha256.New, secret)
	_, _ = mac.Write([]byte(input))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
