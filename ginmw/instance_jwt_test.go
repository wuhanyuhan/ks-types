package ginmw

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	kstypes "github.com/wuhanyuhan/ks-types"
)

func init() { gin.SetMode(gin.TestMode) }

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
