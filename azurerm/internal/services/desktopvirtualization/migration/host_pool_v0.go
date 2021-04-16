package migration

import (
	"log"

	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/azure"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/services/desktopvirtualization/parse"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/tags"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/tf/pluginsdk"
)

func HostPoolUpgradeV0Schema() *pluginsdk.Resource {
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

			"load_balancer_type": {
				Type:     pluginsdk.TypeString,
				Required: true,
				ForceNew: true,
			},

			"friendly_name": {
				Type:     pluginsdk.TypeString,
				Optional: true,
			},

			"description": {
				Type:     pluginsdk.TypeString,
				Optional: true,
			},

			"validate_environment": {
				Type:     pluginsdk.TypeBool,
				Optional: true,
				Default:  false,
			},

			"personal_desktop_assignment_type": {
				Type:     pluginsdk.TypeString,
				Optional: true,
				ForceNew: true,
			},

			"maximum_sessions_allowed": {
				Type:     pluginsdk.TypeInt,
				Optional: true,
				Default:  999999,
			},

			"preferred_app_group_type": {
				Type:        pluginsdk.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "Preferred App Group type to display",
			},

			"registration_info": {
				Type:     pluginsdk.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"expiration_date": {
							Type:     pluginsdk.TypeString,
							Required: true,
						},

						"reset_token": {
							Type:     pluginsdk.TypeBool,
							Computed: true,
						},

						"token": {
							Type:      pluginsdk.TypeString,
							Sensitive: true,
							Computed:  true,
						},
					},
				},
			},

			"tags": tags.Schema(),
		},
	}
}

func HostPoolUpgradeV0ToV1(rawState map[string]interface{}, _ interface{}) (map[string]interface{}, error) {
	oldId := rawState["id"].(string)

	id, err := parse.HostPoolIDInsensitively(oldId)
	if err != nil {
		return nil, err
	}
	newId := id.ID()

	log.Printf("[DEBUG] Updating ID from %q to %q", oldId, newId)
	rawState["id"] = newId

	return rawState, nil
}
