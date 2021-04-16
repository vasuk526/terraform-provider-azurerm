package network

import (
	"fmt"
	"log"

	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/azure"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/tf/pluginsdk"
)

func resourceNetworkInterfaceApplicationSecurityGroupAssociationUpgradeV0Schema() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Schema: map[string]*pluginsdk.Schema{
			"network_interface_id": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: azure.ValidateResourceID,
			},

			"application_security_group_id": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: azure.ValidateResourceID,
			},
		},
	}
}

func resourceNetworkInterfaceApplicationSecurityGroupAssociationUpgradeV0ToV1(rawState map[string]interface{}, _ interface{}) (map[string]interface{}, error) {
	// after shipping support for this Resource Azure's since changed the behaviour to require that all IP Configurations
	// are connected to the same Application Security Group
	applicationSecurityGroupId := rawState["application_security_group_id"].(string)
	networkInterfaceId := rawState["network_interface_id"].(string)

	oldID := rawState["id"].(string)
	newID := fmt.Sprintf("%s|%s", networkInterfaceId, applicationSecurityGroupId)
	log.Printf("[DEBUG] Updating ID from %q to %q", oldID, newID)

	rawState["id"] = newID
	return rawState, nil
}
