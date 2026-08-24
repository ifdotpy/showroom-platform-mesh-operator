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

	rendered := strings.ReplaceAll(string(contents), "{{ .apiExportRootTenancyKcpIoIdentityHash }}", "test-identity-hash")
	var export map[string]interface{}
	require.NoError(t, yaml.Unmarshal([]byte(rendered), &export))
	require.Equal(t, "apis.kcp.io/v1alpha2", export["apiVersion"])
	metadata := export["metadata"].(map[string]interface{})
	annotations := metadata["annotations"].(map[string]interface{})
	compatibilityClaims := annotations["apis.v1alpha2.kcp.io/v1alpha1-permission-claims"].(string)
	require.Equal(t, 7, strings.Count(compatibilityClaims, `"all":true`))
	require.NotContains(t, compatibilityClaims, `"all":false`)

	resources, found, err := unstructured.NestedSlice(export, "spec", "resources")
	require.NoError(t, err)
	require.True(t, found)
	require.Len(t, resources, 7)

	var providerMetadataResources []map[string]interface{}
	resourceKeys := make(map[string]struct{}, len(resources))
	for _, resource := range resources {
		entry := resource.(map[string]interface{})
		key := entry["group"].(string) + "/" + entry["name"].(string)
		_, duplicate := resourceKeys[key]
		require.False(t, duplicate, "duplicate APIExport resource %s", key)
		resourceKeys[key] = struct{}{}
		require.Equal(t, map[string]interface{}{"crd": map[string]interface{}{}}, entry["storage"])
		if entry["group"] == "ui.platform-mesh.io" && entry["name"] == "providermetadatas" {
			providerMetadataResources = append(providerMetadataResources, entry)
		}
	}

	require.Len(t, providerMetadataResources, 1)
	require.Equal(t, providerMetadataSchemaName, providerMetadataResources[0]["schema"])

	permissionClaims, found, err := unstructured.NestedSlice(export, "spec", "permissionClaims")
	require.NoError(t, err)
	require.True(t, found)
	require.Len(t, permissionClaims, 7)
	for _, permissionClaim := range permissionClaims {
		claim := permissionClaim.(map[string]interface{})
		require.Equal(t, []interface{}{"*"}, claim["verbs"])
	}
}

func readManifest(t *testing.T, path string) map[string]interface{} {
	t.Helper()

	contents, err := os.ReadFile(path)
	require.NoError(t, err)

	var manifest map[string]interface{}
	require.NoError(t, yaml.Unmarshal(contents, &manifest))
	return manifest
}
