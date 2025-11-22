//go:build local

package utils

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"math/big"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// Constants for gin context keys
const (
	// Exported key so local_server can reference it
	CtxAuthorizerKey = "_authorizer"
)

// JWKS-related types minimal for our usage
type jwksKey struct {
	Kty string `json:"kty"`
	Alg string `json:"alg"`
	Kid string `json:"kid"`
	Use string `json:"use"`
	N   string `json:"n"`
	E   string `json:"e"`
}

type jwks struct {
	Keys []jwksKey `json:"keys"`
}

var (
	jwksCache     = map[string]map[string]*rsa.PublicKey{} // issuer -> kid -> key
	jwksCacheLock sync.RWMutex
)

// fetchAndCacheJWKS fetches JWKS from issuer + '/.well-known/jwks.json' and caches rsa keys by kid.
func fetchAndCacheJWKS(issuer string) (map[string]*rsa.PublicKey, error) {
	jwksCacheLock.RLock()
	if m, ok := jwksCache[issuer]; ok && len(m) > 0 {
		jwksCacheLock.RUnlock()
		return m, nil
	}
	jwksCacheLock.RUnlock()

	url := strings.TrimRight(issuer, "/") + "/.well-known/jwks.json"
	r, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	// explicitly ignore Close error
	defer func() { _ = r.Body.Close() }()
	if r.StatusCode != http.StatusOK {
		return nil, errors.New("failed to fetch jwks")
	}
	b, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	var jw jwks
	if err := json.Unmarshal(b, &jw); err != nil {
		return nil, err
	}
	m := map[string]*rsa.PublicKey{}
	for _, k := range jw.Keys {
		if k.Kty != "RSA" {
			continue
		}
		// Build rsa.PublicKey from N/E (base64url)
		nBytes, err := base64.RawURLEncoding.DecodeString(k.N)
		if err != nil {
			continue
		}
		eBytes, err := base64.RawURLEncoding.DecodeString(k.E)
		if err != nil {
			continue
		}
		var eInt int
		for _, by := range eBytes {
			eInt = eInt<<8 + int(by)
		}
		pub := &rsa.PublicKey{
			N: new(big.Int).SetBytes(nBytes),
			E: eInt,
		}
		m[k.Kid] = pub
	}
	jwksCacheLock.Lock()
	jwksCache[issuer] = m
	jwksCacheLock.Unlock()
	return m, nil
}

// parseJWTHeaderKid returns the kid from the JWT header without validating signature
func parseJWTHeaderKid(token string) (string, error) {
	parts := strings.Split(token, ".")
	if len(parts) < 2 {
		return "", errors.New("token has not enough parts")
	}
	headB, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		// try standard base64 with padding
		headB, err = base64.StdEncoding.DecodeString(parts[0])
		if err != nil {
			return "", err
		}
	}
	var h map[string]interface{}
	if err := json.Unmarshal(headB, &h); err != nil {
		return "", err
	}
	kidI, ok := h["kid"]
	if !ok {
		return "", errors.New("kid not found in token header")
	}
	kid, ok := kidI.(string)
	if !ok {
		return "", errors.New("kid header not a string")
	}
	return kid, nil
}

// verifyRS256 verifies token signature using given RSA public key and returns payload bytes
func verifyRS256(token string, pub *rsa.PublicKey) ([]byte, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, errors.New("invalid token format")
	}
	sig, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		sig, err = base64.StdEncoding.DecodeString(parts[2])
		if err != nil {
			return nil, err
		}
	}
	msg := parts[0] + "." + parts[1]
	h := sha256.Sum256([]byte(msg))
	if err := rsa.VerifyPKCS1v15(pub, cryptoHashToHash(), h[:], sig); err != nil {
		return nil, err
	}
	// return payload
	payloadB, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		payloadB, err = base64.StdEncoding.DecodeString(parts[1])
		if err != nil {
			return nil, err
		}
	}
	return payloadB, nil
}

// cryptoHashToHash returns the crypto.Hash used for signature verification.
func cryptoHashToHash() crypto.Hash {
	return crypto.SHA256
}

// normalizeToken strips the Bearer prefix and trims whitespace.
func normalizeToken(token string) (string, error) {
	if token == "" {
		return "", errors.New("empty token")
	}
	token = strings.TrimSpace(token)
	if strings.HasPrefix(strings.ToLower(token), "bearer ") {
		token = token[len("bearer "):]
	}
	if token == "" {
		return "", errors.New("empty token")
	}
	return token, nil
}

// getJWKSForIssuer reads LOCAL_COGNITO_ISSUER and fetches its JWKS.
func getJWKSForIssuer() (map[string]*rsa.PublicKey, error) {
	issuer := os.Getenv("LOCAL_COGNITO_ISSUER")
	if issuer == "" {
		return nil, errors.New("LOCAL_COGNITO_ISSUER not set")
	}
	return fetchAndCacheJWKS(issuer)
}

// getPublicKeyForKid returns the public key for the given kid from the jwks map.
func getPublicKeyForKid(m map[string]*rsa.PublicKey, kid string) (*rsa.PublicKey, error) {
	if m == nil {
		return nil, errors.New("jwks map is nil")
	}
	pub, ok := m[kid]
	if !ok || pub == nil {
		return nil, errors.New("unable to find key for kid")
	}
	return pub, nil
}

// parseClaimsFromToken verifies the token signature with pub and unmarshals the payload.
func parseClaimsFromToken(token string, pub *rsa.PublicKey) (map[string]interface{}, error) {
	payloadB, err := verifyRS256(token, pub)
	if err != nil {
		return nil, err
	}
	var claims map[string]interface{}
	if err := json.Unmarshal(payloadB, &claims); err != nil {
		return nil, err
	}
	return claims, nil
}

// validateExp checks the exp claim against current time.
func validateExp(claims map[string]interface{}) error {
	if expI, ok := claims["exp"]; ok {
		var expInt int64
		switch v := expI.(type) {
		case float64:
			expInt = int64(v)
		case int64:
			expInt = v
		case json.Number:
			if n, err := v.Int64(); err == nil {
				expInt = n
			}
		}
		if expInt > 0 {
			if time.Now().Unix() > expInt {
				return errors.New("token expired")
			}
		}
	}
	return nil
}

// audMatches checks whether the aud claim matches the expected audience.
func audMatches(audClaim interface{}, expected string) bool {
	switch v := audClaim.(type) {
	case string:
		return v == expected
	case []interface{}:
		for _, ai := range v {
			if s, ok := ai.(string); ok && s == expected {
				return true
			}
		}
	}
	return false
}

// validateAud checks the aud claim against LOCAL_COGNITO_AUDIENCE if set.
func validateAud(claims map[string]interface{}) error {
	aud := os.Getenv("LOCAL_COGNITO_AUDIENCE")
	if aud == "" {
		return nil
	}
	if aVal, ok := claims["aud"]; ok {
		if !audMatches(aVal, aud) {
			return errors.New("invalid audience")
		}
	}
	return nil
}

// ValidateToken validates a JWT from the Authorization header using Cognito JWKS
// configured via LOCAL_COGNITO_ISSUER and LOCAL_COGNITO_AUDIENCE env vars.
// On success it returns the claims map.
func ValidateToken(token string) (map[string]interface{}, error) {
	// normalize and strip Bearer
	tok, err := normalizeToken(token)
	if err != nil {
		return nil, err
	}

	// get kid from token header
	kid, err := parseJWTHeaderKid(tok)
	if err != nil {
		return nil, err
	}

	// fetch jwks for issuer
	m, err := getJWKSForIssuer()
	if err != nil {
		return nil, err
	}

	// lookup public key
	pub, err := getPublicKeyForKid(m, kid)
	if err != nil {
		return nil, err
	}

	// verify signature and parse claims
	claims, err := parseClaimsFromToken(tok, pub)
	if err != nil {
		return nil, err
	}

	// validate exp
	if err := validateExp(claims); err != nil {
		return nil, err
	}

	// validate aud if configured
	if err := validateAud(claims); err != nil {
		return nil, err
	}

	return claims, nil
}
