package store

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"devhub-gin-backend/internal/domain"

	"golang.org/x/crypto/bcrypt"
)

const (
	accessTokenTTL  = 2 * time.Hour
	refreshTokenTTL = 14 * 24 * time.Hour
)

type accessClaims struct {
	Sub int64  `json:"sub"`
	Exp int64  `json:"exp"`
	Iat int64  `json:"iat"`
	Typ string `json:"typ"`
}

func hashPassword(password string) (string, error) {
	out, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(out), err
}

func checkPassword(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

func newAccessToken(userID int64) (string, int64, error) {
	now := time.Now()
	claims := accessClaims{Sub: userID, Iat: now.Unix(), Exp: now.Add(accessTokenTTL).Unix(), Typ: "access"}
	header := map[string]string{"alg": "HS256", "typ": "JWT"}
	headerJSON, _ := json.Marshal(header)
	claimsJSON, _ := json.Marshal(claims)
	unsigned := base64.RawURLEncoding.EncodeToString(headerJSON) + "." + base64.RawURLEncoding.EncodeToString(claimsJSON)
	sig := signJWT(unsigned)
	return unsigned + "." + sig, int64(accessTokenTTL.Seconds()), nil
}

func parseAccessToken(token string) (int64, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return 0, errors.New("token 格式不合法")
	}
	unsigned := parts[0] + "." + parts[1]
	if !hmac.Equal([]byte(signJWT(unsigned)), []byte(parts[2])) {
		return 0, errors.New("token 签名无效")
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return 0, err
	}
	var claims accessClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return 0, err
	}
	if claims.Typ != "access" || claims.Sub <= 0 {
		return 0, errors.New("token 声明无效")
	}
	if time.Now().Unix() >= claims.Exp {
		return 0, errors.New("token 已过期")
	}
	return claims.Sub, nil
}

func signJWT(unsigned string) string {
	mac := hmac.New(sha256.New, []byte(jwtSecret()))
	_, _ = mac.Write([]byte(unsigned))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func jwtSecret() string {
	if v := os.Getenv("JWT_SECRET"); v != "" {
		return v
	}
	return "devhub-local-change-me"
}

func newRefreshToken() (string, string, time.Time, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", "", time.Time{}, err
	}
	token := base64.RawURLEncoding.EncodeToString(buf)
	return token, tokenHash(token), time.Now().Add(refreshTokenTTL), nil
}

func tokenHash(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func roleID(code string) int64 {
	switch code {
	case "super_admin":
		return 1
	case "site_admin":
		return 2
	case "editor":
		return 3
	case "moderator":
		return 4
	case "user":
		return 5
	default:
		return 0
	}
}

func userIDString(id int64) string {
	return strconv.FormatInt(id, 10)
}

func invalidAuth(msg string) error {
	return fmt.Errorf(msg)
}

func normalizeSiteScope(site string) string {
	site = strings.TrimSpace(site)
	if site == "" || site == "all" {
		return "portal"
	}
	return site
}

func notificationInSite(n domain.Notification, site string) bool {
	site = normalizeSiteScope(site)
	return site == "portal" || normalizeSiteScope(n.Site) == site || normalizeSiteScope(n.Site) == "portal"
}

func logInSite(log domain.AdminLog, site string) bool {
	site = normalizeSiteScope(site)
	return site == "portal" || normalizeSiteScope(log.Site) == site
}
