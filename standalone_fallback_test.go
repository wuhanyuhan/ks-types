package kstypes

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestStandaloneFallbackSpec_YAML_RoundTrip(t *testing.T) {
	input := `mysql:
  host: localhost
  port: 3306
  database: dev_db
  user: root
  password: root
object_storage:
  type: filesystem
  root: ./var/storage
vector_store:
  url: http://qdrant:6333
  collection_prefix: dev-
cache:
  type: inmem
storage:
  root: ./var
`
	var spec StandaloneFallbackSpec
	require.NoError(t, yaml.Unmarshal([]byte(input), &spec))

	require.NotNil(t, spec.MySQL)
	assert.Equal(t, "localhost", spec.MySQL.Host)
	assert.Equal(t, 3306, spec.MySQL.Port)
	assert.Equal(t, "dev_db", spec.MySQL.Database)
	assert.Equal(t, "root", spec.MySQL.User)
	assert.Equal(t, "root", spec.MySQL.Password)

	require.NotNil(t, spec.ObjectStorage)
	assert.Equal(t, "filesystem", spec.ObjectStorage.Type)
	assert.Equal(t, "./var/storage", spec.ObjectStorage.Root)

	require.NotNil(t, spec.VectorStore)
	assert.Equal(t, "http://qdrant:6333", spec.VectorStore.URL)
	assert.Equal(t, "dev-", spec.VectorStore.CollectionPrefix)

	require.NotNil(t, spec.Cache)
	assert.Equal(t, "inmem", spec.Cache.Type)

	require.NotNil(t, spec.Storage)
	assert.Equal(t, "./var", spec.Storage.Root)
}

func TestStandaloneFallbackSpec_Validate_MySQL_Required(t *testing.T) {
	spec := StandaloneFallbackSpec{
		MySQL: &StandaloneMySQLFallback{}, // 缺所有字段
	}
	err := spec.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "mysql")
}

func TestStandaloneFallbackSpec_Validate_OK(t *testing.T) {
	spec := StandaloneFallbackSpec{
		MySQL: &StandaloneMySQLFallback{
			Host: "localhost", Port: 3306, Database: "dev_db",
			User: "root", Password: "root",
		},
	}
	require.NoError(t, spec.Validate())
}
