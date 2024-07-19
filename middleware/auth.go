// package middleware

// import (
// 	"admin-api/utils"
// 	"fmt"
// 	"os"
// 	"reflect"
// 	"time"

// 	"github.com/golang-jwt/jwt/v5"
// 	"github.com/google/uuid"
// 	"github.com/kataras/iris/v12"
// 	"github.com/ping-42/42lib/db/models"
// 	"gorm.io/gorm"
// )

// // User represents the authenticated user and their permissions.
// type UserClaims struct {
// 	UserId      uuid.UUID `json:"user_id"`
// 	Permissions []string  `json:"permissions"`
// 	UserGroupId uint16
// }

// var jwtSecret = []byte(os.Getenv("ADMIN_API_JWT_SECRET")) // TODO mv to config

// // GenerateJWT generates a JWT token for a given Ethereum address.
// func GenerateJWT(db *gorm.DB, user models.User) (string, error) {

// 	// get the permissions
// 	var lvPermissions []models.LvPermission
// 	err := db.Table("users").
// 		Select("lv_permissions.permission").
// 		Joins("join lv_user_groups on lv_user_groups.id = users.user_group_id").
// 		Joins("join permission_to_user_groups on permission_to_user_groups.user_group_id = lv_user_groups.id").
// 		Joins("join lv_permissions on lv_permissions.id = permission_to_user_groups.permission_id").
// 		Where("users.id = ?", user.ID).
// 		Scan(&lvPermissions).Error

// 	if err != nil {
// 		return "", fmt.Errorf("selecting permissions error: %v", err)
// 	}

// 	permissions := make([]string, len(lvPermissions))
// 	for i, p := range lvPermissions {
// 		permissions[i] = p.Permission
// 	}

// 	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{ //TODO can i cast it directly from ?
// 		"userId":      user.ID,
// 		"permissions": permissions,
// 		"userGroupId": user.UserGroupID,
// 		"exp":         time.Now().Add(time.Hour * 1).Unix(), // TODO need to implement refresh logic and the exp here to be e.g. 5mins
// 	})

// 	tokenString, err := token.SignedString(jwtSecret)
// 	if err != nil {
// 		return "", fmt.Errorf("error signing token: %v", err)
// 	}

// 	return tokenString, nil
// }

// // ValidateJWTMiddleware validates the JWT token from the request header.
// func ValidateJWTMiddleware(ctx iris.Context) {
// 	tokenString := ctx.GetHeader("Authorization")
// 	if tokenString == "" {
// 		ctx.StatusCode(iris.StatusUnauthorized)
// 		ctx.JSON(iris.Map{"error": "Missing token"})
// 		return
// 	}

// 	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
// 		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
// 			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
// 		}
// 		return jwtSecret, nil
// 	})

// 	if err != nil {
// 		ctx.StatusCode(iris.StatusUnauthorized)
// 		ctx.JSON(iris.Map{"error": fmt.Sprintf("error parsing token: %v", err)})
// 		return
// 	}

// 	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
// 		user := UserClaims{}
// 		if err := mapClaimsToStruct(claims, &user); err != nil { //TODO can i cast it directly?
// 			fmt.Println(err) //TODO handle the errors
// 			ctx.StatusCode(iris.StatusUnauthorized)
// 			ctx.JSON(iris.Map{"error": "invalid token claims"})
// 			return
// 		}

// 		ctx.Values().Set("UserClaims", user)
// 		ctx.Next()
// 	} else {
// 		ctx.StatusCode(iris.StatusUnauthorized)
// 		ctx.JSON(iris.Map{"error": "invalid token"})
// 	}
// }

// // ValidateAdminMiddleware is admin check
// func ValidateAdminMiddleware(ctx iris.Context) {
// 	userClaims, ok := ctx.Values().Get("UserClaims").(UserClaims)
// 	if !ok {
// 		utils.RespondError(ctx, iris.StatusUnauthorized, "Unauthorized user", fmt.Errorf("ServeSensorsCreate casting to middleware.UserClaims error"))
// 		return
// 	}

// 	// Is the user Admin(UserGroupId==1)?
// 	if userClaims.UserGroupId != 1 {
// 		ctx.StatusCode(iris.StatusUnauthorized)
// 		ctx.JSON(iris.Map{"error": "not admin"})
// 		return
// 	}

// 	ctx.Next()
// }

// // PermissionMiddleware checks if the user has the required permissions.
// func PermissionMiddleware(requiredPermissions ...string) iris.Handler {
// 	return func(ctx iris.Context) {
// 		user, ok := ctx.Values().Get("UserClaims").(UserClaims)
// 		if !ok {
// 			ctx.StatusCode(iris.StatusUnauthorized)
// 			ctx.JSON(iris.Map{"error": "unauthorized user"})
// 			return
// 		}

// 		for _, perm := range requiredPermissions {
// 			if !hasPermission(user.Permissions, perm) {
// 				ctx.StatusCode(iris.StatusForbidden)
// 				ctx.JSON(iris.Map{"error": "you do not have permission to access this resource"})
// 				return
// 			}
// 		}
// 		ctx.Next()
// 	}
// }

// // Helper function to map MapClaims to User struct
// func mapClaimsToStruct(claims jwt.MapClaims, user *UserClaims) error {
// 	userIdStr, ok := claims["userId"].(string)
// 	if !ok {
// 		return fmt.Errorf("userId is not a string in token claims: %v/%s", claims["userId"], reflect.TypeOf(claims["userId"]))
// 	}

// 	userId, err := uuid.Parse(userIdStr)
// 	if err != nil {
// 		return fmt.Errorf("invalid userId address in token claims: %+v", err)
// 	}

// 	// for some reason userGroupId is considered as float64...
// 	userGroupIdFloat, ok := claims["userGroupId"].(float64)
// 	if !ok {
// 		return fmt.Errorf("userGroupId is not float64 in token claims: %v/%s", claims["userGroupId"], reflect.TypeOf(claims["userGroupId"]))
// 	}
// 	userGroupId := uint16(userGroupIdFloat)

// 	permissions, err := convertInterfaceToStringSlice(claims["permissions"])
// 	if err != nil {
// 		return fmt.Errorf("invalid permissions in token claims:%+v", claims["permissions"])
// 	}

// 	user.UserId = userId
// 	user.Permissions = permissions
// 	user.UserGroupId = userGroupId

// 	return nil
// }

// // Helper function to convert interface{} to []string
// func convertInterfaceToStringSlice(input interface{}) ([]string, error) {
// 	interfaces, ok := input.([]interface{})
// 	if !ok {
// 		return nil, fmt.Errorf("type assertion to []interface{} failed")
// 	}
// 	strings := make([]string, len(interfaces))
// 	for i, v := range interfaces {
// 		str, ok := v.(string)
// 		if !ok {
// 			return nil, fmt.Errorf("type assertion to string failed")
// 		}
// 		strings[i] = str
// 	}
// 	return strings, nil
// }

// // Helper function to check if a slice contains a specific string
//
//	func hasPermission(permissions []string, requiredPermission string) bool {
//		for _, perm := range permissions {
//			if perm == requiredPermission {
//				return true
//			}
//		}
//		return false
//	}
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
	UserId      uuid.UUID `json:"user_id"`
	Permissions []string  `json:"permissions"`
	UserGroupId uint64    `json:"user_group_id"`
	jwt.RegisteredClaims
}

var jwtSecret = []byte(os.Getenv("ADMIN_API_JWT_SECRET"))

// GenerateJWT generates a JWT token for a given user.
func GenerateJWT(db *gorm.DB, user models.User) (string, error) {
	// Get the permissions
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
		UserId:      user.ID,
		Permissions: permissions,
		UserGroupId: user.UserGroupID,
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
		ctx.JSON(iris.Map{"error": fmt.Sprintf("error parsing token: %v", err)})
		return
	}

	if claims, ok := token.Claims.(*UserClaims); ok && token.Valid {
		ctx.Values().Set("UserClaims", claims)
		ctx.Next()
	} else {
		ctx.StatusCode(iris.StatusUnauthorized)
		ctx.JSON(iris.Map{"error": "invalid token"})
	}
}

// ValidateAdminMiddleware checks if the user is an admin.
func ValidateAdminMiddleware(ctx iris.Context) {
	userClaims, ok := ctx.Values().Get("UserClaims").(*UserClaims)
	if !ok {
		utils.RespondError(ctx, iris.StatusUnauthorized, "Unauthorized user", fmt.Errorf("ServeSensorsCreate casting to middleware.UserClaims error"))
		return
	}

	// Check if the user is Admin (UserGroupId == 1)
	if userClaims.UserGroupId != 1 {
		ctx.StatusCode(iris.StatusUnauthorized)
		ctx.JSON(iris.Map{"error": "not admin"})
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
			ctx.JSON(iris.Map{"error": "unauthorized user"})
			return
		}

		for _, perm := range requiredPermissions {
			if !hasPermission(userClaims.Permissions, perm) {
				ctx.StatusCode(iris.StatusForbidden)
				ctx.JSON(iris.Map{"error": "you do not have permission to access this resource"})
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
