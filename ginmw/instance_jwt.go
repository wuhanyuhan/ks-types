package ginmw

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	kstypes "github.com/wuhanyuhan/ks-types"
)

const keyInstanceInfo = "instance_info"

// IsRevokedFunc 检查实例是否被吊销。nil 表示不检查。
type IsRevokedFunc func(instanceID string) bool

// InstanceJWTMiddleware 验证 Authorization: Bearer <jwt> 中的实例 JWT。
// publicPEM 是 Ed25519 公钥的 PEM 编码。
// isRevoked 可选，用于检查吊销缓存。传 nil 跳过吊销检查。
func InstanceJWTMiddleware(publicPEM []byte, isRevoked IsRevokedFunc) gin.HandlerFunc {
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
