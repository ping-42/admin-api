package middleware

import (
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/kataras/iris/v12"
	"github.com/ping-42/42lib/db/models"
	"github.com/ping-42/admin-api/utils"
	"gorm.io/gorm"
)

// UserClaims represents the authenticated user and their permissions.
type UserClaims struct {
	UserId         uuid.UUID `json:"user_id"`
	OrganizationId uuid.UUID `json:"organization_id"`
	Permissions    []string  `json:"permissions"`
	UserGroupId    uint64    `json:"user_group_id"`
	jwt.RegisteredClaims
}

// GenerateJWT generates a JWT token for a given user.
func GenerateJWT(db *gorm.DB, user models.User) (string, error) {

	jwtSecret := getJwtSecret()

	var lvPermissions []models.LvPermission
	err := db.Table("users").
		Select("lv_permissions.permission").
		Joins("join lv_user_groups on lv_user_groups.id = users.user_group_id").
		Joins("join permission_to_user_groups on permission_to_user_groups.user_group_id = lv_user_groups.id").
		Joins("join lv_permissions on lv_permissions.id = permission_to_user_groups.permission_id").
		Where("users.id = ?", user.ID).
		Scan(&lvPermissions).Error

	if err != nil {
		return "", fmt.Errorf("selecting permissions error: %v", err)
	}

	permissions := make([]string, len(lvPermissions))
	for i, p := range lvPermissions {
		permissions[i] = p.Permission
	}

	userClaims := UserClaims{
		UserId:         user.ID,
		OrganizationId: user.OrganizationID,
		Permissions:    permissions,
		UserGroupId:    user.UserGroupID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 1)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, userClaims)
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", fmt.Errorf("error signing token: %v", err)
	}

	return tokenString, nil
}

// ValidateJWTMiddleware validates the JWT token from the request header.
func ValidateJWTMiddleware(ctx iris.Context) {

	jwtSecret := getJwtSecret()

	tokenString := ctx.GetHeader("Authorization")
	if tokenString == "" {
		ctx.StatusCode(iris.StatusUnauthorized)
		_ = ctx.JSON(iris.Map{"error": "Missing token"})
		return
	}

	token, err := jwt.ParseWithClaims(tokenString, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})

	if err != nil {
		ctx.StatusCode(iris.StatusUnauthorized)
		_ = ctx.JSON(iris.Map{"error": fmt.Sprintf("error parsing token: %v", err)})
		return
	}

	if claims, ok := token.Claims.(*UserClaims); ok && token.Valid {
		ctx.Values().Set("UserClaims", claims)
		ctx.Next()
	} else {
		ctx.StatusCode(iris.StatusUnauthorized)
		_ = ctx.JSON(iris.Map{"error": "invalid token"})
	}
}

// ValidateAdminMiddleware checks if the user is an admin.
func ValidateAdminMiddleware(ctx iris.Context) {
	userClaims, ok := ctx.Values().Get("UserClaims").(*UserClaims)
	if !ok {
		utils.RespondError(ctx, iris.StatusUnauthorized, "Unauthorized user", fmt.Errorf("ServeSensorsCreate casting to middleware.UserClaims error"))
		return
	}

	// check if the user is Root (UserGroupId == 1)
	if userClaims.UserGroupId != 1 {
		ctx.StatusCode(iris.StatusUnauthorized)
		_ = ctx.JSON(iris.Map{"error": "not root"})
		return
	}

	ctx.Next()
}

// PermissionMiddleware checks if the user has the required permissions.
func PermissionMiddleware(requiredPermissions ...string) iris.Handler {
	return func(ctx iris.Context) {
		userClaims, ok := ctx.Values().Get("UserClaims").(*UserClaims)
		if !ok {
			ctx.StatusCode(iris.StatusUnauthorized)
			_ = ctx.JSON(iris.Map{"error": "unauthorized user"})
			return
		}

		for _, perm := range requiredPermissions {
			if !hasPermission(userClaims.Permissions, perm) {
				ctx.StatusCode(iris.StatusForbidden)
				_ = ctx.JSON(iris.Map{"error": "you do not have permission to access this resource"})
				return
			}
		}
		ctx.Next()
	}
}

// Helper function to check if a slice contains a specific string
func hasPermission(permissions []string, requiredPermission string) bool {
	for _, perm := range permissions {
		if perm == requiredPermission {
			return true
		}
	}
	return false
}

// TODO move to config?
// Maybe we can pass context e.g. AdminApi to the 42lib.Config and to return only the needed data based on the cotext
func getJwtSecret() []byte {
	var jwtSecret = os.Getenv("ADMIN_API_JWT_SECRET")
	if jwtSecret == "" {
		panic("ADMIN_API_JWT_SECRET env var required.")
	}
	return []byte(jwtSecret)
}
