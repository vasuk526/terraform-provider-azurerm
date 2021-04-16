package migration

import (
	"log"

	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/azure"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/tags"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/tf/validation"

	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/services/msi/parse"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/tf/pluginsdk"
)

func UserAssignedIdentityV0ToV1() pluginsdk.StateUpgrader {
	return pluginsdk.StateUpgrader{
		Version: 0,
		Type:    userAssignedIdentityV0Schema().CoreConfigSchema().ImpliedType(),
		Upgrade: userAssignedIdentityUpgradeV0ToV1,
	}
}

func userAssignedIdentityV0Schema() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Schema: map[string]*pluginsdk.Schema{
			"name": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(3, 128),
			},

			"resource_group_name": azure.SchemaResourceGroupName(),

			"location": azure.SchemaLocation(),

			"tags": tags.Schema(),

			"principal_id": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},

			"client_id": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},
		},
	}
}

func userAssignedIdentityUpgradeV0ToV1(rawState map[string]interface{}, meta interface{}) (map[string]interface{}, error) {
	oldId := rawState["id"].(string)
	id, err := parse.UserAssignedIdentityID(oldId)
	if err != nil {
		return rawState, err
	}

	newId := id.ID()
	log.Printf("Updating `id` from %q to %q", oldId, newId)
	rawState["id"] = newId
	return rawState, nil
}
