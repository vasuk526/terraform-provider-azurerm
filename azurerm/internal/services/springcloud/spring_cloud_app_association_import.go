package springcloud

import (
	"fmt"

	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/clients"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/services/springcloud/parse"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/tf/pluginsdk"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/timeouts"
)

const (
	springCloudAppAssociationTypeMysql = "Microsoft.DBforMySQL"
	springCloudAppAssociationTypeRedis = "Microsoft.Cache"
)

func importSpringCloudAppAssociation(resourceType string) func(d *pluginsdk.ResourceData, meta interface{}) (data []*pluginsdk.ResourceData, err error) {
	return func(d *pluginsdk.ResourceData, meta interface{}) (data []*pluginsdk.ResourceData, err error) {
		id, err := parse.SpringCloudAppAssociationID(d.Id())
		if err != nil {
			return []*pluginsdk.ResourceData{}, err
		}

		client := meta.(*clients.Client).AppPlatform.BindingsClient
		ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
		defer cancel()

		resp, err := client.Get(ctx, id.ResourceGroup, id.SpringName, id.AppName, id.BindingName)
		if err != nil {
			return []*pluginsdk.ResourceData{}, fmt.Errorf("retrieving %s: %+v", id, err)
		}

		if resp.Properties == nil || resp.Properties.ResourceType == nil {
			return []*pluginsdk.ResourceData{}, fmt.Errorf("retrieving %s: `properties` or `properties.resourceType` was nil", id)
		}

		if *resp.Properties.ResourceType != resourceType {
			return []*pluginsdk.ResourceData{}, fmt.Errorf(`spring Cloud App Association "type" mismatch, expected "%s", got "%s"`, resourceType, *resp.Properties.ResourceType)
		}

		return []*pluginsdk.ResourceData{d}, nil
	}
}
