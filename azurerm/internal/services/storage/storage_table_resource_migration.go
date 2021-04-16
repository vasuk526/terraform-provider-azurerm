package storage

import (
	"fmt"
	"log"

	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/azure"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/clients"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/tf/pluginsdk"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/tf/validation"
)

// the schema schema was used for both V0 and V1
func ResourceStorageTableStateResourceV0V1() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Schema: map[string]*pluginsdk.Schema{
			"name": {
				Type:     pluginsdk.TypeString,
				Required: true,
				ForceNew: true,
			},
			"storage_account_name": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: ValidateStorageAccountName,
			},
			"resource_group_name": azure.SchemaResourceGroupName(),
			"acl": {
				Type:     pluginsdk.TypeSet,
				Optional: true,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"id": {
							Type:         pluginsdk.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(1, 64),
						},
						"access_policy": {
							Type:     pluginsdk.TypeList,
							Optional: true,
							Elem: &pluginsdk.Resource{
								Schema: map[string]*pluginsdk.Schema{
									"start": {
										Type:         pluginsdk.TypeString,
										Required:     true,
										ValidateFunc: validation.StringIsNotEmpty,
									},
									"expiry": {
										Type:         pluginsdk.TypeString,
										Required:     true,
										ValidateFunc: validation.StringIsNotEmpty,
									},
									"permissions": {
										Type:         pluginsdk.TypeString,
										Required:     true,
										ValidateFunc: validation.StringIsNotEmpty,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func ResourceStorageTableStateUpgradeV0ToV1(rawState map[string]interface{}, meta interface{}) (map[string]interface{}, error) {
	tableName := rawState["name"].(string)
	accountName := rawState["storage_account_name"].(string)
	environment := meta.(*clients.Client).Account.Environment

	id := rawState["id"].(string)
	newResourceID := fmt.Sprintf("https://%s.table.%s/%s", accountName, environment.StorageEndpointSuffix, tableName)
	log.Printf("[DEBUG] Updating ID from %q to %q", id, newResourceID)

	rawState["id"] = newResourceID
	return rawState, nil
}

func ResourceStorageTableStateUpgradeV1ToV2(rawState map[string]interface{}, meta interface{}) (map[string]interface{}, error) {
	tableName := rawState["name"].(string)
	accountName := rawState["storage_account_name"].(string)
	environment := meta.(*clients.Client).Account.Environment

	id := rawState["id"].(string)
	newResourceID := fmt.Sprintf("https://%s.table.%s/Tables('%s')", accountName, environment.StorageEndpointSuffix, tableName)
	log.Printf("[DEBUG] Updating ID from %q to %q", id, newResourceID)

	rawState["id"] = newResourceID
	return rawState, nil
}
