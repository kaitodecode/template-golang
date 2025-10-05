package helper

import (
	"encoding/json"
	"runtime/debug"
	"strings"
	"time"

	"template-golang/internal/db/model"
	"template-golang/pkg/apperror"
	"template-golang/pkg/config"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

func BoolPtr(v bool) *bool {
	return &v
}

func stringPtr(v string) *string {
	return &v
}

func Int64Ptr(v int64) *int64 {
	return &v
}

func Float64Ptr(v float64) *float64 {
	return &v
}

func StringPtr(v string) *string {
	return &v
}

func MapKeys[M ~map[K]V, K comparable, V any](m M) []K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func FormatData[T any](v string) (T, error) {
	var result T
	if err := json.Unmarshal([]byte(v), &result); err != nil {
		return result, apperror.New("helper", "failed to unmarshal JSON", 500, err, string(debug.Stack()))
	}
	return result, nil
}

func FormatWord(word string) string {
	firstWord := strings.ToUpper(word[:1])
	return firstWord + strings.ToLower(word[1:])
}

func Hash(str string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(str), bcrypt.DefaultCost)
	if err != nil {
		return "", apperror.New("helper", "failed to hash password", 500, err, string(debug.Stack()))
	}
	return string(hash), nil
}

func CompareHashAndPassword(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

func GenerateJwtToken(user model.User) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["user_id"] = user.ID
	claims["user"] = user
	claims["role"] = user.Role
	claims["exp"] = time.Now().Add(time.Hour * 24).Unix()
	conf := config.GetConfig()
	tokenString, err := token.SignedString([]byte(conf.JwtSecret))
	if err != nil {
		return "", apperror.New("helper", "failed to generate jwt token", 500, err, string(debug.Stack()))
	}

	return tokenString, nil
}

func ParseJwtToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.GetConfig().JwtSecret), nil
	})
	if err != nil {
		return nil, apperror.New("helper", "failed to parse jwt token", 400, err, string(debug.Stack()))
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, apperror.New("helper", "invalid jwt token", 400, err, string(debug.Stack()))
	}
	return claims, nil
}

func GetUserIDFromCtx(c *fiber.Ctx) (string, error) {
	tokenString, err := GetTokenFromHeader(c)
	if err != nil {
		return "", err
	}
	return GetUserIDFromToken(tokenString)
}
func GetUserIDFromToken(tokenString string) (string, error) {
	claims, err := ParseJwtToken(tokenString)
	if err != nil {
		return "", err
	}
	userID, ok := claims["user_id"].(string)
	if !ok {
		return "", apperror.New("helper", "invalid user_id in jwt token", 400, nil, string(debug.Stack()))
	}
	return userID, nil
}

func VerifyJwtToken(tokenString string) (jwt.MapClaims, error) {
	claims, err := ParseJwtToken(tokenString)
	if err != nil {
		return nil, err
	}
	return claims, nil
}

func IsExpiredToken(tokenString string) (bool, error) {
	claims, err := ParseJwtToken(tokenString)
	if err != nil {
		return false, err
	}
	exp, ok := claims["exp"].(float64)
	if !ok {
		return false, apperror.New("helper", "invalid exp in jwt token", 400, nil, string(debug.Stack()))
	}
	return time.Unix(int64(exp), 0).Before(time.Now()), nil
}

func GetRoleFromToken(tokenString string) (string, error) {
	claims, err := ParseJwtToken(tokenString)
	if err != nil {
		return "", err
	}
	role, ok := claims["role"].(string)
	if !ok {
		return "", apperror.New("helper", "invalid role in jwt token", 400, nil, string(debug.Stack()))
	}
	return role, nil
}

func GetTokenFromHeader(ctx *fiber.Ctx) (string, error) {
	token := ctx.Get("Authorization")
	if token == "" {
		return "", apperror.New("helper", "missing authorization header", 400, nil, string(debug.Stack()))
	}
	if strings.HasPrefix(token, "Bearer ") {
		token = strings.TrimPrefix(token, "Bearer ")
	} else {
		return "", apperror.New("helper", "invalid authorization format", 400, nil, string(debug.Stack()))
	}
	if token == "" {
		return "", apperror.New("helper", "missing token", 400, nil, string(debug.Stack()))
	}
	return token, nil
}
