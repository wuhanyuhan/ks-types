package kstypes

import "encoding/json"

// SecurityScheme 表示 A2A AgentCard 中的一种安全方案。
// 实现类型必须是 immutable struct（不暴露字段写入）；通过自定义 MarshalJSON 注入 type 判别字段。
//
// 设计说明：A2A v1.0 spec 在 schema 草案中用了 schemeType()（小写、未导出），但 keystone 实现统一改为
// SchemeType()（导出），便于消费方做 type routing 而无需 type assertion。这是相对 spec 的有意设计修正。
type SecurityScheme interface {
	SchemeType() string
}

// APIKeyScheme 表示 API Key 认证方案（header / query / cookie）。
type APIKeyScheme struct {
	In   string `json:"in"`   // header / query / cookie
	Name string `json:"name"` // 键名，如 X-API-Key
}

// 编译期接口合规检查。
var _ SecurityScheme = APIKeyScheme{}

// SchemeType 实现 SecurityScheme 接口，返回 type 判别值。
func (s APIKeyScheme) SchemeType() string { return "apiKey" }

// MarshalJSON 在序列化时注入 type 判别字段，满足 A2A v1.0 规范。
func (s APIKeyScheme) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Type string `json:"type"`
		In   string `json:"in"`
		Name string `json:"name"`
	}{Type: s.SchemeType(), In: s.In, Name: s.Name})
}

// HTTPAuthScheme 表示 HTTP 认证方案（bearer / basic）。
type HTTPAuthScheme struct {
	Scheme       string `json:"scheme"`                // bearer / basic
	BearerFormat string `json:"bearerFormat,omitempty"` // 仅 bearer 时有意义，如 JWT
}

// 编译期接口合规检查。
var _ SecurityScheme = HTTPAuthScheme{}

// SchemeType 实现 SecurityScheme 接口，返回 type 判别值。
func (s HTTPAuthScheme) SchemeType() string { return "http" }

// MarshalJSON 在序列化时注入 type 判别字段，满足 A2A v1.0 规范。
func (s HTTPAuthScheme) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Type         string `json:"type"`
		Scheme       string `json:"scheme"`
		BearerFormat string `json:"bearerFormat,omitempty"`
	}{Type: s.SchemeType(), Scheme: s.Scheme, BearerFormat: s.BearerFormat})
}

// OAuth2Scheme / OpenIDConnectScheme / MutualTLSScheme 是 α 最小占位实现：
// 类型按 A2A v1.0 规范完整定义，但 verifier 实现留待跨组织场景落地时再补。
// 序列化 / 反序列化能力完整，足够 AgentCard JSON 输出 / 解析。

// OAuth2Flow 表示 OAuth2 授权流程配置。
type OAuth2Flow struct {
	AuthorizationURL string            `json:"authorizationUrl,omitempty"`
	TokenURL         string            `json:"tokenUrl,omitempty"`
	Scopes           map[string]string `json:"scopes"`
}

// OAuth2Scheme 表示 OAuth2 认证方案。
type OAuth2Scheme struct {
	Flows map[string]OAuth2Flow `json:"flows"`
}

var _ SecurityScheme = OAuth2Scheme{}

func (s OAuth2Scheme) SchemeType() string { return "oauth2" }

func (s OAuth2Scheme) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Type  string                `json:"type"`
		Flows map[string]OAuth2Flow `json:"flows"`
	}{Type: s.SchemeType(), Flows: s.Flows})
}

// OpenIDConnectScheme 表示 OpenID Connect 认证方案。
type OpenIDConnectScheme struct {
	OpenIDConnectURL string `json:"openIdConnectUrl"`
}

var _ SecurityScheme = OpenIDConnectScheme{}

func (s OpenIDConnectScheme) SchemeType() string { return "openIdConnect" }

func (s OpenIDConnectScheme) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Type             string `json:"type"`
		OpenIDConnectURL string `json:"openIdConnectUrl"`
	}{Type: s.SchemeType(), OpenIDConnectURL: s.OpenIDConnectURL})
}

// MutualTLSScheme 表示双向 TLS 认证方案。
type MutualTLSScheme struct{}

var _ SecurityScheme = MutualTLSScheme{}

func (s MutualTLSScheme) SchemeType() string { return "mutualTLS" }

func (s MutualTLSScheme) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Type string `json:"type"`
	}{Type: s.SchemeType()})
}
