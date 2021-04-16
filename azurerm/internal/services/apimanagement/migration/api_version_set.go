package migration

import (
	"log"
	"strings"

	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/azure"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/services/apimanagement/schemaz"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/tf/pluginsdk"
)

func ApiVersionSetUpgradeV0Schema() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Schema: map[string]*pluginsdk.Schema{
			"name": schemaz.SchemaApiManagementChildName(),

			"resource_group_name": azure.SchemaResourceGroupName(),

			"api_management_name": schemaz.SchemaApiManagementName(),

			"display_name": {
				Type:     pluginsdk.TypeString,
				Required: true,
			},

			"versioning_scheme": {
				Type:     pluginsdk.TypeString,
				Required: true,
			},

			"description": {
				Type:     pluginsdk.TypeString,
				Optional: true,
			},

			"version_header_name": {
				Type:     pluginsdk.TypeString,
				Optional: true,
			},

			"version_query_name": {
				Type:     pluginsdk.TypeString,
				Optional: true,
			},
		},
	}
}

func ApiVersionSetUpgradeV0ToV1(rawState map[string]interface{}, meta interface{}) (map[string]interface{}, error) {
	oldId := rawState["id"].(string)
	newId := strings.Replace(rawState["id"].(string), "/api-version-set/", "/apiVersionSets/", 1)

	log.Printf("[DEBUG] Updating ID from %q to %q", oldId, newId)

	rawState["id"] = newId

	return rawState, nil
}
