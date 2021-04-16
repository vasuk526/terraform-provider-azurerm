package migration

import (
	"log"
	"strings"

	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/tf/pluginsdk"
)

func TimeSeriesInsightsAccessPolicyV0() pluginsdk.StateUpgrader {
	return pluginsdk.StateUpgrader{
		Type:    timeSeriesInsightsAccessPolicyV0StateMigration().CoreConfigSchema().ImpliedType(),
		Upgrade: timeSeriesInsightsAccessPolicyV0StateUpgradeV0ToV1,
		Version: 0,
	}
}

func timeSeriesInsightsAccessPolicyV0StateMigration() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Schema: map[string]*pluginsdk.Schema{
			"name": {
				Type:     pluginsdk.TypeString,
				Required: true,
				ForceNew: true,
			},

			"time_series_insights_environment_id": {
				Type:     pluginsdk.TypeString,
				Required: true,
				ForceNew: true,
			},

			"principal_object_id": {
				Type:     pluginsdk.TypeString,
				Required: true,
				ForceNew: true,
			},

			"description": {
				Type:     pluginsdk.TypeString,
				Optional: true,
			},

			"roles": {
				Type:     pluginsdk.TypeSet,
				Required: true,
				Elem: &pluginsdk.Schema{
					Type: pluginsdk.TypeString,
				},
			},
		},
	}
}

func timeSeriesInsightsAccessPolicyV0StateUpgradeV0ToV1(rawState map[string]interface{}, meta interface{}) (map[string]interface{}, error) {
	log.Println("[DEBUG] Migrating ResourceType from v0 to v1 format")
	oldId := rawState["id"].(string)
	newId := strings.Replace(oldId, "/accesspolicies/", "/accessPolicies/", 1)

	log.Printf("[DEBUG] Updating ID from %q to %q", oldId, newId)

	rawState["id"] = newId

	return rawState, nil
}
