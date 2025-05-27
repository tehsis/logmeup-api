package middleware

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// JWK represents a JSON Web Key
type JWK struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	Use string `json:"use"`
	N   string `json:"n"`
	E   string `json:"e"`
}

// JWKS represents a JSON Web Key Set
type JWKS struct {
	Keys []JWK `json:"keys"`
}

// UserClaims represents the JWT claims we care about
type UserClaims struct {
	Sub       string   `json:"sub"`
	Email     string   `json:"email"`
	Name      string   `json:"name"`
	Picture   string   `json:"picture"`
	Audience  Audience `json:"aud"`
	Issuer    string   `json:"iss"`
	ExpiresAt int64    `json:"exp"`
	IssuedAt  int64    `json:"iat"`
	jwt.RegisteredClaims
}

// Audience represents the audience claim which can be either a string or array of strings
type Audience []string

// UnmarshalJSON implements custom unmarshaling for audience claim
func (a *Audience) UnmarshalJSON(data []byte) error {
	var single string
	if err := json.Unmarshal(data, &single); err == nil {
		*a = Audience{single}
		return nil
	}

	var multiple []string
	if err := json.Unmarshal(data, &multiple); err == nil {
		*a = Audience(multiple)
		return nil
	}

	return fmt.Errorf("audience must be either string or array of strings")
}

// Contains checks if the audience contains a specific value
func (a Audience) Contains(audience string) bool {
	for _, aud := range a {
		if aud == audience {
			return true
		}
	}
	return false
}

// String returns the first audience value or empty string if none
func (a Audience) String() string {
	if len(a) > 0 {
		return a[0]
	}
	return ""
}

var (
	jwksCache    *JWKS
	jwksCacheExp time.Time
)

// AuthMiddleware validates JWT tokens from Auth0
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>"
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Bearer token required"})
			c.Abort()
			return
		}

		// Parse and validate token
		claims, err := validateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": fmt.Sprintf("Invalid token: %v", err)})
			c.Abort()
			return
		}

		// Set user information in context
		c.Set("user_id", claims.Sub)
		c.Set("user_email", claims.Email)
		c.Set("user_name", claims.Name)
		c.Set("user_picture", claims.Picture)

		c.Next()
	}
}

// validateToken validates the JWT token using Auth0's public keys
func validateToken(tokenString string) (*UserClaims, error) {
	// Parse token without verification first to get the kid
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, &UserClaims{})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %v", err)
	}

	// Get the kid from token header
	kid, ok := token.Header["kid"].(string)
	if !ok {
		return nil, fmt.Errorf("kid not found in token header")
	}

	// Get the public key for this kid
	publicKey, err := getPublicKey(kid)
	if err != nil {
		return nil, fmt.Errorf("failed to get public key: %v", err)
	}

	// Parse and validate token with the public key
	parsedToken, err := jwt.ParseWithClaims(tokenString, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return publicKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token with claims: %v", err)
	}

	claims, ok := parsedToken.Claims.(*UserClaims)
	if !ok || !parsedToken.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	// Validate issuer and audience
	expectedIssuer := fmt.Sprintf("https://%s/", os.Getenv("AUTH0_DOMAIN"))
	if claims.Issuer != expectedIssuer {
		return nil, fmt.Errorf("invalid issuer: expected %s, got %s", expectedIssuer, claims.Issuer)
	}

	expectedAudience := os.Getenv("AUTH0_AUDIENCE")
	if expectedAudience != "" && !claims.Audience.Contains(expectedAudience) {
		return nil, fmt.Errorf("invalid audience: expected %s, got %v", expectedAudience, claims.Audience)
	}

	return claims, nil
}

// getPublicKey retrieves the public key for the given kid from Auth0's JWKS endpoint
func getPublicKey(kid string) (*rsa.PublicKey, error) {
	// Check cache first
	if jwksCache != nil && time.Now().Before(jwksCacheExp) {
		for _, key := range jwksCache.Keys {
			if key.Kid == kid {
				return jwkToRSAPublicKey(key)
			}
		}
	}

	// Fetch JWKS from Auth0
	jwksURL := fmt.Sprintf("https://%s/.well-known/jwks.json", os.Getenv("AUTH0_DOMAIN"))
	resp, err := http.Get(jwksURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch JWKS: %v", err)
	}
	defer resp.Body.Close()

	var jwks JWKS
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return nil, fmt.Errorf("failed to decode JWKS: %v", err)
	}

	// Cache JWKS for 1 hour
	jwksCache = &jwks
	jwksCacheExp = time.Now().Add(time.Hour)

	// Find the key with matching kid
	for _, key := range jwks.Keys {
		if key.Kid == kid {
			return jwkToRSAPublicKey(key)
		}
	}

	return nil, fmt.Errorf("key with kid %s not found", kid)
}

// jwkToRSAPublicKey converts a JWK to an RSA public key
func jwkToRSAPublicKey(jwk JWK) (*rsa.PublicKey, error) {
	// Decode the modulus
	nBytes, err := base64.RawURLEncoding.DecodeString(jwk.N)
	if err != nil {
		return nil, fmt.Errorf("failed to decode modulus: %v", err)
	}

	// Decode the exponent
	eBytes, err := base64.RawURLEncoding.DecodeString(jwk.E)
	if err != nil {
		return nil, fmt.Errorf("failed to decode exponent: %v", err)
	}

	// Convert to big integers
	n := new(big.Int).SetBytes(nBytes)
	e := new(big.Int).SetBytes(eBytes)

	// Create RSA public key
	publicKey := &rsa.PublicKey{
		N: n,
		E: int(e.Int64()),
	}

	return publicKey, nil
}

// GetUserID extracts the user ID from the Gin context
func GetUserID(c *gin.Context) (string, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return "", false
	}
	return userID.(string), true
}

// GetUserEmail extracts the user email from the Gin context
func GetUserEmail(c *gin.Context) (string, bool) {
	email, exists := c.Get("user_email")
	if !exists {
		return "", false
	}
	return email.(string), true
}
