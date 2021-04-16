package migration

import (
	"fmt"
	"log"
	"strings"

	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/services/securitycenter/parse"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/tf/pluginsdk"
)

func AdvancedThreatProtectionV0ToV1() pluginsdk.StateUpgrader {
	return pluginsdk.StateUpgrader{
		Version: 0,
		Type:    advancedThreatProtectionV0Schema().CoreConfigSchema().ImpliedType(),
		Upgrade: advancedThreadProtectionV0toV1Upgrade,
	}
}

func advancedThreatProtectionV0Schema() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Schema: map[string]*pluginsdk.Schema{
			"target_resource_id": {
				Type:     pluginsdk.TypeString,
				Required: true,
				ForceNew: true,
			},

			"enabled": {
				Type:     pluginsdk.TypeBool,
				Required: true,
			},
		},
	}
}

func advancedThreadProtectionV0toV1Upgrade(rawState map[string]interface{}, meta interface{}) (map[string]interface{}, error) {
	oldId := rawState["id"].(string)

	// remove the existing `/` if it's present (2.42+) which'll do nothing if it wasn't (2.38)
	newId := fmt.Sprintf("/%s", strings.TrimPrefix(oldId, "/"))

	parsedId, err := parse.AdvancedThreatProtectionID(newId)
	if err != nil {
		return nil, err
	}

	newId = parsedId.ID()

	log.Printf("[DEBUG] Updating ID from %q to %q", oldId, newId)
	rawState["id"] = newId
	return rawState, nil
}
