package middleware

import (
	"net/http"
	"strings"

	"sitchi/internal/common"
	"sitchi/internal/common/model"

	"github.com/gin-gonic/gin"
)

const (
	CtxUserIDKey     = "user_id"
	CtxModuleCodeKey = "module_code"
	CtxRolesKey      = "roles"
	CtxClaimsKey     = "jwt_claims"
)

type AuthMiddleware struct {
	JwtAuthService *common.JwtAuthService
}

func NewAuthMiddleware(jwtAuthService *common.JwtAuthService) *AuthMiddleware {
	return &AuthMiddleware{JwtAuthService: jwtAuthService}
}

// RequireAuth 解析并校验 access token，成功后将用户信息写入 gin context。
func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		if m == nil || m.JwtAuthService == nil {
			appErr := model.ErrInternal.WithDetail("jwt service not initialized")
			c.JSON(appErr.HTTPStatus(), model.ApiErrorResponse(appErr.Code, appErr.Message, appErr))
			c.Abort()
			return
		}

		authHeader := strings.TrimSpace(c.GetHeader("Authorization"))
		if authHeader == "" {
			appErr := model.ErrUnauthorized.WithDetail("missing Authorization header")
			c.JSON(appErr.HTTPStatus(), model.ApiErrorResponse(appErr.Code, appErr.Message, appErr))
			c.Abort()
			return
		}

		tokenString := extractBearerToken(authHeader)
		if tokenString == "" {
			appErr := model.ErrUnauthorized.WithDetail("invalid Authorization header format")
			c.JSON(appErr.HTTPStatus(), model.ApiErrorResponse(appErr.Code, appErr.Message, appErr))
			c.Abort()
			return
		}

		claims, err := m.JwtAuthService.ParseAndVerifyAccessToken(tokenString)
		if err != nil {
			appErr := model.ErrUnauthorized.WithDetail(err.Error())
			c.JSON(appErr.HTTPStatus(), model.ApiErrorResponse(appErr.Code, appErr.Message, appErr))
			c.Abort()
			return
		}

		c.Set(CtxUserIDKey, claims.UserID)
		c.Set(CtxModuleCodeKey, claims.ModuleCode)
		c.Set(CtxRolesKey, claims.Roles)
		c.Set(CtxClaimsKey, claims)
		c.Next()
	}
}

func extractBearerToken(authHeader string) string {
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 {
		return ""
	}
	if !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1])
}

// OptionalAuth 可选鉴权：有 token 就解析，无 token 则放行。
func (m *AuthMiddleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := strings.TrimSpace(c.GetHeader("Authorization"))
		if authHeader == "" {
			c.Next()
			return
		}

		tokenString := extractBearerToken(authHeader)
		if tokenString == "" || m == nil || m.JwtAuthService == nil {
			c.Next()
			return
		}

		claims, err := m.JwtAuthService.ParseAndVerifyAccessToken(tokenString)
		if err != nil {
			c.Next()
			return
		}

		c.Set(CtxUserIDKey, claims.UserID)
		c.Set(CtxModuleCodeKey, claims.ModuleCode)
		c.Set(CtxRolesKey, claims.Roles)
		c.Set(CtxClaimsKey, claims)
		c.Next()
	}
}

// WriteUnauthorized 是给非中间件场景的快捷返回。
func WriteUnauthorized(c *gin.Context, detail any) {
	appErr := model.ErrUnauthorized.WithDetail(detail)
	c.JSON(http.StatusUnauthorized, model.ApiErrorResponse(appErr.Code, appErr.Message, appErr))
}
