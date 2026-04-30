package ginmw

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	kstypes "github.com/wuhanyuhan/ks-types"
)

func init() { gin.SetMode(gin.TestMode) }

// signCustomAudienceJWT 用任意 audience 签一个有效 JWT，绕过 SignInstanceJWT 内置 aud 列表，
// 用于验证中间件的 RequireAudience 行为。
func signCustomAudienceJWT(t *testing.T, privPEM []byte, instanceID string, aud []string) string {
	t.Helper()
	priv, err := kstypes.ParseEd25519PrivateKeyPEM(privPEM)
	if err != nil {
		t.Fatalf("parse private: %v", err)
	}
	now := time.Now().UTC()
	claims := kstypes.InstanceClaims{
		InstanceID: instanceID,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   instanceID,
			Issuer:    "ks-admin",
			Audience:  jwt.ClaimStrings(aud),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour)),
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims).SignedString(priv)
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	return token
}

func setupRouter(pubPEM []byte) *gin.Engine {
	r := gin.New()
	r.Use(InstanceJWTMiddleware(pubPEM, nil))
	r.GET("/test", func(c *gin.Context) {
		info := GetInstanceInfo(c)
		c.JSON(200, gin.H{
			"instance_id": info.InstanceID,
			"name":        info.Name,
			"group":       info.Group,
		})
	})
	return r
}

func TestMiddleware_ValidToken(t *testing.T) {
	priv, _ := os.ReadFile("../testdata/test_private.pem")
	pub, _ := os.ReadFile("../testdata/test_public.pem")

	claims := kstypes.InstanceClaims{
		InstanceID: "inst_ok",
		Name:       "测试实例",
		Group:      "test",
	}
	token, _ := kstypes.SignInstanceJWT(claims, priv, time.Hour)

	r := setupRouter(pub)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("status: got %d, body: %s", w.Code, w.Body.String())
	}
}

func TestMiddleware_MissingHeader(t *testing.T) {
	pub, _ := os.ReadFile("../testdata/test_public.pem")
	r := setupRouter(pub)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != 401 {
		t.Fatalf("status: got %d, want 401", w.Code)
	}
}

func TestMiddleware_ExpiredToken(t *testing.T) {
	priv, _ := os.ReadFile("../testdata/test_private.pem")
	pub, _ := os.ReadFile("../testdata/test_public.pem")

	claims := kstypes.InstanceClaims{InstanceID: "inst_exp"}
	token, _ := kstypes.SignInstanceJWT(claims, priv, -1*time.Hour)

	r := setupRouter(pub)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	if w.Code != 401 {
		t.Fatalf("status: got %d, want 401", w.Code)
	}
}

func TestMiddleware_RevokedInstance(t *testing.T) {
	priv, _ := os.ReadFile("../testdata/test_private.pem")
	pub, _ := os.ReadFile("../testdata/test_public.pem")

	claims := kstypes.InstanceClaims{InstanceID: "inst_revoked"}
	token, _ := kstypes.SignInstanceJWT(claims, priv, time.Hour)

	// 创建吊销检查函数
	isRevoked := func(instanceID string) bool {
		return instanceID == "inst_revoked"
	}

	r := gin.New()
	r.Use(InstanceJWTMiddleware(pub, isRevoked))
	r.GET("/test", func(c *gin.Context) { c.Status(200) })

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	if w.Code != 401 {
		t.Fatalf("status: got %d, want 401", w.Code)
	}
}

// Phase A —— RequireAudience option 行为验证

// 默认（不传 RequireAudience option）：即使 token aud 是异常值（如旧 2-aud 或完全无关 aud），
// 中间件也放行。理由：现网历史 token 仍可能是旧 aud 列表，本期不强制升级，由 Phase B/C 各
// 服务显式启用 RequireAudience。
func TestMiddleware_RequireAudience_DefaultPassThrough(t *testing.T) {
	priv, _ := os.ReadFile("../testdata/test_private.pem")
	pub, _ := os.ReadFile("../testdata/test_public.pem")

	// 模拟一个异常 aud token（旧版只有 2 项，或完全无关的 aud）
	token := signCustomAudienceJWT(t, priv, "inst_legacy", []string{"foo", "bar"})

	r := gin.New()
	r.Use(InstanceJWTMiddleware(pub, nil)) // 不传 option
	r.GET("/test", func(c *gin.Context) { c.Status(200) })

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("status: got %d (body=%s), want 200 (default option must not enforce aud)", w.Code, w.Body.String())
	}
}

// 启用 RequireAudience("ks-llm-gateway")，token 的 aud 包含 ks-llm-gateway → 放行
func TestMiddleware_RequireAudience_Match(t *testing.T) {
	priv, _ := os.ReadFile("../testdata/test_private.pem")
	pub, _ := os.ReadFile("../testdata/test_public.pem")

	// 用真实 SignInstanceJWT 签发，aud 含 ks-llm-gateway
	claims := kstypes.InstanceClaims{InstanceID: "inst_aud_ok"}
	token, _ := kstypes.SignInstanceJWT(claims, priv, time.Hour)

	r := gin.New()
	r.Use(InstanceJWTMiddleware(pub, nil, RequireAudience("ks-llm-gateway")))
	r.GET("/test", func(c *gin.Context) { c.Status(200) })

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("status: got %d (body=%s), want 200", w.Code, w.Body.String())
	}
}

// 启用 RequireAudience("ks-llm-gateway")，token 的 aud 不含 ks-llm-gateway → 401
func TestMiddleware_RequireAudience_Mismatch(t *testing.T) {
	priv, _ := os.ReadFile("../testdata/test_private.pem")
	pub, _ := os.ReadFile("../testdata/test_public.pem")

	// 手工签一个不含 ks-llm-gateway 的 token
	token := signCustomAudienceJWT(t, priv, "inst_aud_bad", []string{"ks-admin", "ks-hub"})

	r := gin.New()
	r.Use(InstanceJWTMiddleware(pub, nil, RequireAudience("ks-llm-gateway")))
	r.GET("/test", func(c *gin.Context) { c.Status(200) })

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	if w.Code != 401 {
		t.Fatalf("status: got %d (body=%s), want 401", w.Code, w.Body.String())
	}
}
