package subroutines

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/yaml"
)

const providerMetadataSchemaName = "v260817-a47fd3e4.providermetadatas.ui.platform-mesh.io"

func TestProviderMetadataDetailViewExtensionsSchema(t *testing.T) {
	schema := readManifest(t, "../../manifests/kcp/01-platform-mesh-system/apiresourceschema-providermetadatas.ui.platform-mesh.io.yaml")

	metadata := schema["metadata"].(map[string]interface{})
	require.Equal(t, providerMetadataSchemaName, metadata["name"])

	versions, found, err := unstructured.NestedSlice(schema, "spec", "versions")
	require.NoError(t, err)
	require.True(t, found)
	require.Len(t, versions, 1)

	version := versions[0].(map[string]interface{})
	extensions, found, err := unstructured.NestedMap(version, "schema", "properties", "spec", "properties", "detailViewExtensions")
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, "array", extensions["type"])

	items := extensions["items"].(map[string]interface{})
	require.Equal(t, "object", items["type"])
	require.Equal(t, []interface{}{"url"}, items["required"])

	url, found, err := unstructured.NestedMap(items, "properties", "url")
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, "string", url["type"])
}

func TestCoreAPIExportUsesDetailViewExtensionsSchema(t *testing.T) {
	contents, err := os.ReadFile("../../manifests/kcp/01-platform-mesh-system/apiexport-core.platform-mesh.io.yaml")
	require.NoError(t, err)

	manifest := string(contents)
	require.Contains(t, manifest, "- "+providerMetadataSchemaName)
	require.False(t, strings.Contains(manifest, "v250725-732d200.providermetadatas.ui.platform-mesh.io"))
}

func readManifest(t *testing.T, path string) map[string]interface{} {
	t.Helper()

	contents, err := os.ReadFile(path)
	require.NoError(t, err)

	var manifest map[string]interface{}
	require.NoError(t, yaml.Unmarshal(contents, &manifest))
	return manifest
}
