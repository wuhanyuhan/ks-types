package ginmw

import (
	"net/http"
	"slices"
	"strings"

	"github.com/gin-gonic/gin"
	kstypes "github.com/wuhanyuhan/ks-types"
)

const keyInstanceInfo = "instance_info"

// IsRevokedFunc 检查实例是否被吊销。nil 表示不检查。
type IsRevokedFunc func(instanceID string) bool

// Option 控制 InstanceJWTMiddleware 行为的可选项。
type Option func(*middlewareConfig)

// middlewareConfig 中间件运行参数。零值表示"全部默认行为"，与无 option 调用兼容。
type middlewareConfig struct {
	requireAudience string // 空字符串：不校验 aud（向后兼容）
}

// RequireAudience 让中间件强制要求 JWT 的 aud 包含给定服务名。
// 不调用此 Option 时，默认不校验 aud（保持与旧版 2-aud token 的兼容性）。
// 各服务在 Phase B/C 完成切换后显式启用此 option 以获得更严格的最小授权检查。
func RequireAudience(svc string) Option {
	return func(c *middlewareConfig) { c.requireAudience = svc }
}

// InstanceJWTMiddleware 验证 Authorization: Bearer <jwt> 中的实例 JWT。
// publicPEM 是 Ed25519 公钥的 PEM 编码。
// isRevoked 可选，用于检查吊销缓存。传 nil 跳过吊销检查。
// opts 可选，按需启用 RequireAudience 等行为；不传保持向后兼容。
func InstanceJWTMiddleware(publicPEM []byte, isRevoked IsRevokedFunc, opts ...Option) gin.HandlerFunc {
	cfg := &middlewareConfig{}
	for _, opt := range opts {
		opt(cfg)
	}

	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    kstypes.ErrTokenInvalid.Code,
				"message": "缺少 Authorization Bearer 头",
			})
			return
		}

		tokenStr := strings.TrimPrefix(auth, "Bearer ")
		claims, err := kstypes.VerifyInstanceJWT(tokenStr, publicPEM)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    kstypes.ErrTokenInvalid.Code,
				"message": err.Error(),
			})
			return
		}

		if cfg.requireAudience != "" {
			if !slices.Contains([]string(claims.Audience), cfg.requireAudience) {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"code":    kstypes.ErrTokenInvalid.Code,
					"message": "token audience 不包含本服务（" + cfg.requireAudience + "）",
				})
				return
			}
		}

		if isRevoked != nil && isRevoked(claims.InstanceID) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    kstypes.ErrInstanceRevoked.Code,
				"message": kstypes.ErrInstanceRevoked.Message,
			})
			return
		}

		c.Set(keyInstanceInfo, claims)
		c.Next()
	}
}

// GetInstanceInfo 从 Gin context 获取已验证的实例信息
func GetInstanceInfo(c *gin.Context) *kstypes.InstanceClaims {
	v, exists := c.Get(keyInstanceInfo)
	if !exists {
		return nil
	}
	return v.(*kstypes.InstanceClaims)
}
