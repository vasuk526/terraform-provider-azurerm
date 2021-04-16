package migration

import (
	"log"

	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/azure"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/services/desktopvirtualization/parse"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/tags"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/tf/pluginsdk"
)

func ApplicationGroupUpgradeV0Schema() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Schema: map[string]*pluginsdk.Schema{
			"name": {
				Type:     pluginsdk.TypeString,
				Required: true,
				ForceNew: true,
			},

			"location": azure.SchemaLocation(),

			"resource_group_name": azure.SchemaResourceGroupName(),

			"type": {
				Type:     pluginsdk.TypeString,
				Required: true,
				ForceNew: true,
			},

			"host_pool_id": {
				Type:     pluginsdk.TypeString,
				Required: true,
			},

			"friendly_name": {
				Type:     pluginsdk.TypeString,
				Optional: true,
			},

			"description": {
				Type:     pluginsdk.TypeString,
				Optional: true,
			},

			"tags": tags.Schema(),
		},
	}
}

func ApplicationGroupUpgradeV0ToV1(rawState map[string]interface{}, _ interface{}) (map[string]interface{}, error) {
	oldId := rawState["id"].(string)
	id, err := parse.ApplicationGroupIDInsensitively(oldId)
	if err != nil {
		return nil, err
	}
	newId := id.ID()
	log.Printf("[DEBUG] Updating ID from %q to %q", oldId, newId)
	rawState["id"] = newId

	oldHostPoolId := rawState["host_pool_id"].(string)
	hostPoolId, err := parse.HostPoolIDInsensitively(oldHostPoolId)
	if err != nil {
		return nil, err
	}
	newHostPoolId := hostPoolId.ID()
	log.Printf("[DEBUG] Updating Host Pool ID from %q to %q", oldHostPoolId, newHostPoolId)
	rawState["host_pool_id"] = newHostPoolId

	return rawState, nil
}
