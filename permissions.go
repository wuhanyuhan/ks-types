package kstypes

import "fmt"

// PermissionDecl 应用在 manifest 中声明的单个权限维度
type PermissionDecl struct {
	Level          string   `yaml:"level" json:"level"`
	AllowedDomains []string `yaml:"allowed_domains,omitempty" json:"allowed_domains,omitempty"`
	Paths          []string `yaml:"paths,omitempty" json:"paths,omitempty"`
}

// PermissionDimension 权限维度的注册信息
type PermissionDimension struct {
	DisplayName string   `json:"display_name"`
	Description string   `json:"description,omitempty"`
	Levels      []string `json:"levels"`
	RiskWeight  int      `json:"risk_weight"`
}

// PermissionWarning 校验时产生的警告（非致命错误）
type PermissionWarning struct {
	Dimension string
	Message   string
}

// PermissionRegistry 权限维度注册表
type PermissionRegistry struct {
	dimensions map[string]PermissionDimension
}

// NewPermissionRegistry 创建空的权限注册表
func NewPermissionRegistry() *PermissionRegistry {
	return &PermissionRegistry{dimensions: make(map[string]PermissionDimension)}
}

// Register 注册一个权限维度
func (r *PermissionRegistry) Register(key string, dim PermissionDimension) {
	r.dimensions[key] = dim
}

// Validate 校验应用声明的权限。返回 warnings（未知维度）和 error（非法 level）。
func (r *PermissionRegistry) Validate(perms map[string]PermissionDecl) ([]PermissionWarning, error) {
	var warnings []PermissionWarning

	for key, decl := range perms {
		dim, known := r.dimensions[key]
		if !known {
			warnings = append(warnings, PermissionWarning{
				Dimension: key,
				Message:   fmt.Sprintf("未知的权限维度 %q，需要人工审核", key),
			})
			continue
		}

		valid := false
		for _, l := range dim.Levels {
			if l == decl.Level {
				valid = true
				break
			}
		}
		if !valid {
			return nil, fmt.Errorf("权限 %q 的 level %q 不在合法值 %v 中", key, decl.Level, dim.Levels)
		}
	}

	return warnings, nil
}

// HighRiskPermissions 返回风险权重超过阈值的已声明权限维度
func (r *PermissionRegistry) HighRiskPermissions(perms map[string]PermissionDecl, threshold int) []string {
	var result []string
	for key := range perms {
		if dim, ok := r.dimensions[key]; ok && dim.RiskWeight > threshold {
			result = append(result, key)
		}
	}
	return result
}

// DefaultPermissionRegistry 返回预置权限维度的注册表
func DefaultPermissionRegistry() *PermissionRegistry {
	r := NewPermissionRegistry()
	r.Register("network", PermissionDimension{
		DisplayName: "网络访问",
		Levels:      []string{"none", "restricted", "unrestricted"},
		RiskWeight:  5,
	})
	r.Register("llm", PermissionDimension{
		DisplayName: "LLM 调用",
		Levels:      []string{"none", "host_proxy", "self_managed"},
		RiskWeight:  3,
	})
	r.Register("filesystem", PermissionDimension{
		DisplayName: "文件系统",
		Levels:      []string{"none", "read_scoped", "scoped", "full"},
		RiskWeight:  8,
	})
	r.Register("user_context", PermissionDimension{
		DisplayName: "用户数据",
		Levels:      []string{"none", "read", "write"},
		RiskWeight:  7,
	})
	return r
}
