package kstypes

// PermissionDecl 应用在 manifest 中声明的单个权限维度
type PermissionDecl struct {
	Level          string   `yaml:"level" json:"level"`
	AllowedDomains []string `yaml:"allowed_domains,omitempty" json:"allowed_domains,omitempty"`
	Paths          []string `yaml:"paths,omitempty" json:"paths,omitempty"`
}
