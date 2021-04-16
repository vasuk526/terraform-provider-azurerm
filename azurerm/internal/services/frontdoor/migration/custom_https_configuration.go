package migration

import (
	"fmt"
	"log"
	"strings"

	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/azure"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/services/frontdoor/parse"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/tf/pluginsdk"
)

func CustomHttpsConfigurationV0Schema() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Schema: map[string]*pluginsdk.Schema{
			"frontend_endpoint_id": {
				Type:     pluginsdk.TypeString,
				Required: true,
				ForceNew: true,
			},

			"custom_https_provisioning_enabled": {
				Type:     pluginsdk.TypeBool,
				Required: true,
			},

			"custom_https_configuration": {
				Type:     pluginsdk.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"certificate_source": {
							Type:     pluginsdk.TypeString,
							Optional: true,
						},
						"minimum_tls_version": {
							Type:     pluginsdk.TypeString,
							Computed: true,
						},
						"provisioning_state": {
							Type:     pluginsdk.TypeString,
							Computed: true,
						},
						"provisioning_substate": {
							Type:     pluginsdk.TypeString,
							Computed: true,
						},
						"azure_key_vault_certificate_secret_name": {
							Type:     pluginsdk.TypeString,
							Optional: true,
						},
						"azure_key_vault_certificate_secret_version": {
							Type:     pluginsdk.TypeString,
							Optional: true,
						},
						"azure_key_vault_certificate_vault_id": {
							Type:     pluginsdk.TypeString,
							Optional: true,
						},
					},
				},
			},

			"resource_group_name": {
				Type:     pluginsdk.TypeString,
				Required: true,
			},
		},
	}
}

func CustomHttpsConfigurationV0ToV1(rawState map[string]interface{}, _ interface{}) (map[string]interface{}, error) {
	// this was: fmt.Sprintf("%s/customHttpsConfiguration/%s", frontEndEndpointId, frontendEndpointName
	oldId := rawState["id"].(string)
	oldParsedId, err := azure.ParseAzureResourceID(oldId)
	if err != nil {
		return rawState, err
	}

	resourceGroup := oldParsedId.ResourceGroup
	frontdoorName := ""
	frontendEndpointName := ""
	for key, value := range oldParsedId.Path {
		if strings.EqualFold(key, "frontdoors") {
			frontdoorName = value
			continue
		}

		if strings.EqualFold(key, "frontendEndpoints") {
			frontendEndpointName = value
			continue
		}
	}

	if frontdoorName == "" {
		return rawState, fmt.Errorf("couldn't find the `frontdoors` segment in the old resource id %q", oldId)
	}

	if frontendEndpointName == "" {
		return rawState, fmt.Errorf("couldn't find the `frontendEndpoints` segment in the old resource id %q", oldId)
	}

	newId := parse.NewFrontendEndpointID(oldParsedId.SubscriptionID, resourceGroup, frontdoorName, frontendEndpointName)
	newIdStr := newId.ID()

	log.Printf("[DEBUG] Updating ID from %q to %q", oldId, newIdStr)

	rawState["id"] = newIdStr

	return rawState, nil
}
