package domainservices

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"

	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/tags"

	"github.com/Azure/azure-sdk-for-go/services/domainservices/mgmt/2020-01-01/aad"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/azure"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/clients"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
)

func dataSourceArmActiveDirectoryDomainService() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceArmActiveDirectoryDomainServiceRead,

		Schema: map[string]*schema.Schema{
			"domain_configuration_type": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"domain_name": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"filtered_sync_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"ldaps": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enabled": {
							Type:     schema.TypeBool,
							Computed: true,
						},

						"external_access_enabled": {
							Type:     schema.TypeBool,
							Computed: true,
						},

						"external_access_ip_address": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"pfx_certificate": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"pfx_certificate_password": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},

			"location": azure.SchemaLocationForDataSource(),

			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotWhiteSpace,
			},

			"notifications": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"additional_recipients": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"notify_dc_admins": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"notify_global_admins": {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},

			"replica_sets": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						// TODO: add health-related attributes

						"domain_controller_ip_addresses": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},

						"external_access_ip_address": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"location": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"replica_set_id": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"service_status": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"subnet_id": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"vnet_site_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},

			"resource_forest": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"resource_forest": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"forest_trust": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:     schema.TypeString,
										Computed: true,
									},

									"remote_dns_ips": {
										Type:     schema.TypeList,
										Computed: true,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
									},

									"trust_direction": {
										Type:     schema.TypeString,
										Computed: true,
									},

									"trust_password": {
										Type:     schema.TypeString,
										Computed: true,
									},

									"trusted_domain_fqdn": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
					},
				},
			},

			"resource_group_name": azure.SchemaResourceGroupNameForDataSource(),

			"security": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ntlm_v1_enabled": {
							Type:     schema.TypeBool,
							Computed: true,
						},

						"sync_kerberos_passwords": {
							Type:     schema.TypeBool,
							Computed: true,
						},

						"sync_ntlm_passwords": {
							Type:     schema.TypeBool,
							Computed: true,
						},

						"sync_on_prem_passwords": {
							Type:     schema.TypeBool,
							Computed: true,
						},

						"tls_v1_enabled": {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},

			"sku": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"tags": tags.SchemaDataSource(),
		},
	}
}

func dataSourceArmActiveDirectoryDomainServiceRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).DomainServices.DomainServicesClient
	ctx := meta.(*clients.Client).StopContext

	name := d.Get("name").(string)
	resourceGroup := d.Get("resource_group_name").(string)

	resp, err := client.Get(ctx, resourceGroup, name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			return nil
		}
		return err
	}

	d.SetId(*resp.ID)

	d.Set("name", name)
	d.Set("resource_group_name", resourceGroup)

	if location := resp.Location; location != nil {
		d.Set("location", azure.NormalizeLocation(*location))
	}

	if props := resp.DomainServiceProperties; props != nil {
		domainConfigType := ""
		if v := props.DomainConfigurationType; v != nil {
			domainConfigType = *v
		}
		d.Set("domain_configuration_type", domainConfigType)

		d.Set("domain_name", props.DomainName)

		d.Set("filtered_sync_enabled", false)
		if props.FilteredSync == aad.FilteredSyncEnabled {
			d.Set("filtered_sync_enabled", true)
		}

		d.Set("sku", props.Sku)

		if err := d.Set("ldaps", flattenDomainServiceLdaps(props.LdapsSettings)); err != nil {
			return fmt.Errorf("setting `ldaps`: %+v", err)
		}

		if err := d.Set("notifications", flattenDomainServiceNotifications(props.NotificationSettings)); err != nil {
			return fmt.Errorf("setting `notifications`: %+v", err)
		}

		if err := d.Set("replica_sets", flattenDomainServiceReplicaSets(props.ReplicaSets)); err != nil {
			return fmt.Errorf("setting `replica_sets`: %+v", err)
		}

		if err := d.Set("resource_forest", flattenDomainServiceResourceForest(props.ResourceForestSettings)); err != nil {
			return fmt.Errorf("setting `resource_forest`: %+v", err)
		}

		if err := d.Set("security", flattenDomainServiceSecurity(props.DomainSecuritySettings)); err != nil {
			return fmt.Errorf("setting `security`: %+v", err)
		}
	}

	return tags.FlattenAndSet(d, resp.Tags)
}
