package eventhub

import (
	"fmt"
	"time"

	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/azure"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/clients"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/tf/pluginsdk"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/timeouts"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
)

func dataSourceEventHub() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Read: dataSourceEventHubRead,

		Timeouts: &pluginsdk.ResourceTimeout{
			Read: pluginsdk.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*pluginsdk.Schema{
			"name": {
				Type:     pluginsdk.TypeString,
				Required: true,
			},

			"namespace_name": {
				Type:     pluginsdk.TypeString,
				Required: true,
			},

			"resource_group_name": azure.SchemaResourceGroupNameForDataSource(),

			"partition_count": {
				Type:     pluginsdk.TypeInt,
				Computed: true,
			},

			"partition_ids": {
				Type:     pluginsdk.TypeList,
				Computed: true,
				Elem:     &pluginsdk.Schema{Type: pluginsdk.TypeString},
			},
		},
	}
}

func dataSourceEventHubRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Eventhub.EventHubsClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	name := d.Get("name").(string)
	resourceGroup := d.Get("resource_group_name").(string)
	namespaceName := d.Get("namespace_name").(string)

	resp, err := client.Get(ctx, resourceGroup, namespaceName, name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			return fmt.Errorf("Error: EventHub %q (Resource Group %q / Namespace Name %q) was not found", name, resourceGroup, namespaceName)
		}

		return fmt.Errorf("Error making Read request on EventHub %q (Resource Group %q / Namespace Name %q): %+v", name, resourceGroup, namespaceName, err)
	}

	d.SetId(*resp.ID)

	d.Set("name", name)
	d.Set("namespace_name", namespaceName)
	d.Set("resource_group_name", resourceGroup)

	if props := resp.Properties; props != nil {
		d.Set("partition_count", props.PartitionCount)
		d.Set("partition_ids", props.PartitionIds)
	}

	return nil
}
