package kstypes

import "fmt"

// StandaloneFallbackSpec 声明应用在 standalone（非 keystone 托管）模式下使用的本地资源 fallback。
// 与 ManagedResourcesSpec 一一对应：每个声明了的 managed resource 都必须在 standalone 模式下
// 通过此处的 fallback 提供本地配置，否则 framework 在启动期 fail-fast。
//
// 设计原则：manifest 是资源声明的单一权威，不允许通过 config.yaml 或其他途径绕过。
type StandaloneFallbackSpec struct {
	MySQL         *StandaloneMySQLFallback         `yaml:"mysql,omitempty" json:"mysql,omitempty"`
	ObjectStorage *StandaloneObjectStorageFallback `yaml:"object_storage,omitempty" json:"object_storage,omitempty"`
	VectorStore   *StandaloneVectorStoreFallback   `yaml:"vector_store,omitempty" json:"vector_store,omitempty"`
	Storage       *StandaloneStorageFallback       `yaml:"storage,omitempty" json:"storage,omitempty"`
	Cache         *StandaloneCacheFallback         `yaml:"cache,omitempty" json:"cache,omitempty"`
}

// StandaloneMySQLFallback 本地 MySQL 连接信息。
type StandaloneMySQLFallback struct {
	Host     string `yaml:"host" json:"host"`
	Port     int    `yaml:"port" json:"port"`
	Database string `yaml:"database" json:"database"`
	User     string `yaml:"user" json:"user"`
	Password string `yaml:"password" json:"password"`
}

// StandaloneObjectStorageFallback 本地对象存储 fallback。
// Type 当前仅支持 "filesystem"，对应 Root 下的目录结构。
type StandaloneObjectStorageFallback struct {
	Type string `yaml:"type" json:"type"`
	Root string `yaml:"root" json:"root"`
}

// StandaloneVectorStoreFallback 本地向量存储 fallback。
type StandaloneVectorStoreFallback struct {
	URL              string `yaml:"url" json:"url"`
	CollectionPrefix string `yaml:"collection_prefix" json:"collection_prefix"`
}

// StandaloneCacheFallback 本地缓存 fallback。
// Type 支持 "inmem"（进程内）或 "redis"（需配 Addr）。
type StandaloneCacheFallback struct {
	Type      string `yaml:"type,omitempty" json:"type,omitempty"`
	Addr      string `yaml:"addr,omitempty" json:"addr,omitempty"`
	Password  string `yaml:"password,omitempty" json:"password,omitempty"`
	DB        int    `yaml:"db,omitempty" json:"db,omitempty"`
	KeyPrefix string `yaml:"key_prefix,omitempty" json:"key_prefix,omitempty"`
}

// StandaloneStorageFallback 本地私有文件系统 fallback。
// Root 下按 manifest.managed_resources.storage 声明的 scope 创建子目录。
type StandaloneStorageFallback struct {
	Root string `yaml:"root" json:"root"`
}

// Validate 校验 standalone fallback 各资源必填字段。
// 注意：本方法**不**检查与 ManagedResourcesSpec 的"一一对应关系"——那是 framework 端的工作。
// 此处只做结构本身的字段完整性校验。
func (s StandaloneFallbackSpec) Validate() error {
	if s.MySQL != nil {
		if err := s.MySQL.validate(); err != nil {
			return fmt.Errorf("standalone_fallback.mysql: %w", err)
		}
	}
	if s.ObjectStorage != nil {
		if err := s.ObjectStorage.validate(); err != nil {
			return fmt.Errorf("standalone_fallback.object_storage: %w", err)
		}
	}
	if s.VectorStore != nil {
		if err := s.VectorStore.validate(); err != nil {
			return fmt.Errorf("standalone_fallback.vector_store: %w", err)
		}
	}
	if s.Storage != nil {
		if err := s.Storage.validate(); err != nil {
			return fmt.Errorf("standalone_fallback.storage: %w", err)
		}
	}
	if s.Cache != nil {
		if err := s.Cache.validate(); err != nil {
			return fmt.Errorf("standalone_fallback.cache: %w", err)
		}
	}
	return nil
}

func (m *StandaloneMySQLFallback) validate() error {
	if m.Host == "" {
		return fmt.Errorf("host 为必填项")
	}
	if m.Port == 0 {
		return fmt.Errorf("port 为必填项")
	}
	if m.Database == "" {
		return fmt.Errorf("database 为必填项")
	}
	if m.User == "" {
		return fmt.Errorf("user 为必填项")
	}
	if m.Password == "" {
		return fmt.Errorf("password 为必填项")
	}
	return nil
}

func (o *StandaloneObjectStorageFallback) validate() error {
	switch o.Type {
	case "":
		return fmt.Errorf("type 为必填项")
	case "filesystem":
		// valid
	default:
		return fmt.Errorf("type 当前仅支持 filesystem，收到 %q", o.Type)
	}
	if o.Root == "" {
		return fmt.Errorf("root 为必填项")
	}
	return nil
}

func (v *StandaloneVectorStoreFallback) validate() error {
	if v.URL == "" {
		return fmt.Errorf("url 为必填项")
	}
	if v.CollectionPrefix == "" {
		return fmt.Errorf("collection_prefix 为必填项")
	}
	return nil
}

func (c *StandaloneCacheFallback) validate() error {
	switch c.Type {
	case "inmem":
		// inmem 不需要其他字段
		return nil
	case "redis":
		if c.Addr == "" {
			return fmt.Errorf("redis fallback 必须提供 addr")
		}
		return nil
	default:
		return fmt.Errorf("type 仅支持 inmem 或 redis，收到 %q", c.Type)
	}
}

func (s *StandaloneStorageFallback) validate() error {
	if s.Root == "" {
		return fmt.Errorf("root 为必填项")
	}
	return nil
}
