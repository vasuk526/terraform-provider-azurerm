package migration

import (
	"log"

	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/location"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/tf/pluginsdk"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/tf/suppress"

	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/azure"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/services/cdn/parse"
)

func CdnProfileV0Schema() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Schema: map[string]*pluginsdk.Schema{
			"name": {
				Type:     pluginsdk.TypeString,
				Required: true,
				ForceNew: true,
			},

			"location": {
				Type:             pluginsdk.TypeString,
				Required:         true,
				ForceNew:         true,
				StateFunc:        location.StateFunc,
				DiffSuppressFunc: location.DiffSuppressFunc,
			},

			"resource_group_name": {
				Type:     pluginsdk.TypeString,
				Required: true,
				ForceNew: true,
			},

			"sku": {
				Type:             pluginsdk.TypeString,
				Required:         true,
				ForceNew:         true,
				DiffSuppressFunc: suppress.CaseDifference,
			},

			"tags": {
				Type:     pluginsdk.TypeMap,
				Optional: true,
				Elem: &pluginsdk.Schema{
					Type: pluginsdk.TypeString,
				},
			},
		},
	}
}

func CdnProfileV0ToV1(rawState map[string]interface{}, _ interface{}) (map[string]interface{}, error) {
	// old
	// 	/subscriptions/{subscriptionId}/resourcegroups/{resourceGroupName}/providers/Microsoft.Cdn/profiles/{profileName}
	// new:
	// 	/subscriptions/{subscriptionId}/resourceGroups/{resourceGroupName}/providers/Microsoft.Cdn/profiles/{profileName}
	// summary:
	// resourcegroups -> resourceGroups
	oldId := rawState["id"].(string)
	oldParsedId, err := azure.ParseAzureResourceID(oldId)
	if err != nil {
		return rawState, err
	}

	resourceGroup := oldParsedId.ResourceGroup
	name, err := oldParsedId.PopSegment("profiles")
	if err != nil {
		return rawState, err
	}

	newId := parse.NewProfileID(oldParsedId.SubscriptionID, resourceGroup, name)
	newIdStr := newId.ID()

	log.Printf("[DEBUG] Updating ID from %q to %q", oldId, newIdStr)

	rawState["id"] = newIdStr

	return rawState, nil
}
