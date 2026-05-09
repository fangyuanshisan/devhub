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
	"regexp"
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
	Sub       int64  `json:"sub"`
	Exp       int64  `json:"exp"`
	Iat       int64  `json:"iat"`
	Typ       string `json:"typ"`
	TokenType string `json:"token_type,omitempty"`
	Audience  string `json:"aud,omitempty"`
}

func hashPassword(password string) (string, error) {
	out, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(out), err
}

func checkPassword(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

func newAccessToken(userID int64) (string, int64, error) {
	return newScopedAccessToken(userID, "user", "devhub_frontend")
}

func newAdminAccessToken(adminUserID int64) (string, int64, error) {
	return newScopedAccessToken(adminUserID, "admin", "devhub_admin")
}

func newScopedAccessToken(subjectID int64, tokenType, audience string) (string, int64, error) {
	now := time.Now()
	claims := accessClaims{Sub: subjectID, Iat: now.Unix(), Exp: now.Add(accessTokenTTL).Unix(), Typ: "access", TokenType: tokenType, Audience: audience}
	header := map[string]string{"alg": "HS256", "typ": "JWT"}
	headerJSON, _ := json.Marshal(header)
	claimsJSON, _ := json.Marshal(claims)
	unsigned := base64.RawURLEncoding.EncodeToString(headerJSON) + "." + base64.RawURLEncoding.EncodeToString(claimsJSON)
	sig := signJWT(unsigned)
	return unsigned + "." + sig, int64(accessTokenTTL.Seconds()), nil
}

func parseAccessToken(token string) (int64, error) {
	return parseScopedAccessToken(token, "user", "devhub_frontend")
}

func parseAdminAccessToken(token string) (int64, error) {
	return parseScopedAccessToken(token, "admin", "devhub_admin")
}

func parseScopedAccessToken(token, expectedType, expectedAudience string) (int64, error) {
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
	if claims.TokenType == "" {
		return 0, errors.New("token 缺少身份类型")
	}
	if claims.TokenType != expectedType {
		return 0, errors.New("token 身份类型不匹配")
	}
	if expectedAudience != "" && claims.Audience != "" && claims.Audience != expectedAudience {
		return 0, errors.New("token 受众不匹配")
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

func roleCodeByID(id int64) string {
	switch id {
	case 1:
		return "super_admin"
	case 2:
		return "site_admin"
	case 3:
		return "editor"
	case 4:
		return "moderator"
	case 5:
		return "user"
	default:
		return ""
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

var adminLogTargetPattern = regexp.MustCompile(`^([A-Za-z_]+)[#:](\d+)`)

func enrichAdminLog(log domain.AdminLog, users []domain.AdminUser) domain.AdminLog {
	if log.ActorType == "" {
		if log.Actor == "system" || log.Actor == "" {
			log.ActorType = "system"
		} else if log.Role == "moderator" {
			log.ActorType = "moderator"
		} else {
			log.ActorType = "admin_user"
		}
	}
	target := strings.TrimSpace(log.Target)
	if log.TargetType == "" {
		if index := strings.IndexAny(target, "#:"); index > 0 {
			log.TargetType = target[:index]
		}
	}
	if match := adminLogTargetPattern.FindStringSubmatch(target); len(match) == 3 {
		log.TargetType = match[1]
		if id, err := strconv.ParseInt(match[2], 10, 64); err == nil {
			log.TargetID = id
		}
	}
	if log.ActorUserID == 0 && strings.TrimSpace(log.Actor) != "" {
		for _, user := range users {
			if user.Username == log.Actor || user.Nickname == log.Actor {
				log.ActorUserID = user.ID
				break
			}
		}
	}
	if log.ActorID == 0 {
		log.ActorID = log.ActorUserID
	}
	if log.ActorUserID == 0 && log.ActorID > 0 {
		log.ActorUserID = log.ActorID
	}
	if log.CommunityID == 0 {
		log.CommunityID = communityIDBySite(normalizeSiteScope(log.Site))
	}
	return log
}
