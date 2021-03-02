package domainservices

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/location"

	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/tf"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/timeouts"

	"github.com/Azure/azure-sdk-for-go/services/domainservices/mgmt/2020-01-01/aad"
	"github.com/hashicorp/go-azure-helpers/response"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"

	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/azure"
	azValidate "github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/validate"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/clients"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/services/domainservices/parse"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
)

func resourceActiveDirectoryDomainService() *schema.Resource {
	return &schema.Resource{
		Create: resourceActiveDirectoryDomainServiceCreateUpdate,
		Read:   resourceActiveDirectoryDomainServiceRead,
		Update: resourceActiveDirectoryDomainServiceCreateUpdate,
		Delete: resourceActiveDirectoryDomainServiceDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(120 * time.Minute),
			Read:   schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(120 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{ // TODO: add computed attributes: deployment_id, sync_owner,
			"name": { // TODO: set domain_name separately
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringIsNotEmpty, // TODO: proper validation
			},

			"location": azure.SchemaLocation(),

			"resource_group_name": azure.SchemaResourceGroupNameDiffSuppress(),

			"filtered_sync": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"ldaps": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"external_access": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},

						"pfx_certificate": {
							Type:         schema.TypeString,
							Required:     true,
							Sensitive:    true,
							ValidateFunc: azValidate.Base64EncodedString,
						},

						"pfx_certificate_password": {
							Type:      schema.TypeString,
							Required:  true,
							Sensitive: true,
						},
					},
				},
			},

			"notifications": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"additional_recipients": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.StringIsNotWhiteSpace,
							},
						},

						"notify_dc_admins": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},

						"notify_global_admins": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},
					},
				},
			},

			"replica_set": {
				Type:     schema.TypeList,
				Optional: true, // TODO: make required
				MinItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						// TODO: add health-related attributes and other computed attributes (replicaSetId, vnetId etc)

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
							Type:             schema.TypeString,
							Required:         true,
							ForceNew:         true, // TODO: figure out if this is needed
							ValidateFunc:     location.EnhancedValidate,
							StateFunc:        location.StateFunc,
							DiffSuppressFunc: location.DiffSuppressFunc,
						},

						"service_status": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"subnet_id": {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true, // TODO: figure out if this is needed
							ValidateFunc: azure.ValidateResourceID,
						},
					},
				},
			},

			"security": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{ // TODO: add sync_kerberos and sync_on_prem properties
						"ntlm_v1": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},

						"sync_ntlm_passwords": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},

						"tls_v1": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},
					},
				},
			},
		}, // TODO: add forest settings, SKU
	}
}

func resourceActiveDirectoryDomainServiceCreateUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).DomainServices.DomainServicesClient
	ctx, cancel := timeouts.ForCreate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	name := d.Get("name").(string)
	resourceGroup := d.Get("resource_group_name").(string)
	id := parse.NewDomainServiceID(client.SubscriptionID, resourceGroup, name)

	if d.IsNewResource() {
		existing, err := client.Get(ctx, id.ResourceGroup, id.Name)
		if err != nil {
			if !utils.ResponseWasNotFound(existing.Response) {
				return fmt.Errorf("checking for presence of existing %s: %s", id.String(), err)
			}
		}

		if existing.ID != nil && *existing.ID != "" {
			return tf.ImportAsExistsError("azurerm_active_directory_domain_service", id.ID())
		}
	}

	location := azure.NormalizeLocation(d.Get("location").(string))
	filteredSync := aad.FilteredSyncDisabled
	if d.Get("filtered_sync").(bool) {
		filteredSync = aad.FilteredSyncDisabled
	}

	domainService := aad.DomainService{ // TODO: add tags
		Location: &location,
		DomainServiceProperties: &aad.DomainServiceProperties{
			DomainName:             utils.String(d.Get("domain_name").(string)),
			DomainSecuritySettings: expandDomainServiceSecurity(d.Get("security").([]interface{})),
			FilteredSync:           filteredSync,
			LdapsSettings:          expandDomainServiceLdaps(d.Get("ldaps").([]interface{})),
			NotificationSettings:   expandDomainServiceNotifications(d.Get("notifications").([]interface{})),
			ReplicaSets:            expandDomainServiceReplicaSets(d.Get("replica_set").([]interface{})),
		}, // TODO: look into DomainConfigurationType
	}

	future, err := client.CreateOrUpdate(ctx, id.ResourceGroup, id.Name, domainService)
	if err != nil {
		return fmt.Errorf("creating %s: %+v", id.String(), err)
	}
	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for creation of %s: %+v", id.String(), err)
	}

	// A fully deployed domain service has 2 domain controllers per replica set, but the create operation completes early before the DCs are online.
	// Further operations are blocked until both controllers are up.
	stateConf := &resource.StateChangeConf{
		Pending:      []string{"pending"},
		Target:       []string{"available"},
		Refresh:      domainServiceControllerRefreshFunc(ctx, client, id),
		Delay:        30 * time.Second,
		PollInterval: 10 * time.Second,
		Timeout:      1 * time.Hour,
	}

	if _, err := stateConf.WaitForState(); err != nil {
		return fmt.Errorf("waiting for both Domain Service controllers to become available: %+v", err)
	}

	d.SetId(id.ID())

	return resourceActiveDirectoryDomainServiceRead(d, meta)
}

func resourceActiveDirectoryDomainServiceRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).DomainServices.DomainServicesClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.DomainServiceID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.Get(ctx, id.ResourceGroup, id.Name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			return nil
		}
		return err
	}

	d.Set("name", id.Name)
	d.Set("resource_group_name", id.ResourceGroup)

	if location := resp.Location; location != nil {
		d.Set("location", azure.NormalizeLocation(*location))
	}

	if domainServiceProperties := resp.DomainServiceProperties; domainServiceProperties != nil {
		if err := d.Set("ldaps", flattenDomainServiceLdaps(domainServiceProperties.LdapsSettings)); err != nil {
			return fmt.Errorf("setting `ldaps`: %+v", err)
		}
		if err := d.Set("notifications", flattenDomainServiceNotifications(domainServiceProperties.NotificationSettings)); err != nil {
			return fmt.Errorf("setting `notifications`: %+v", err)
		}
		if err := d.Set("replica_set", flattenDomainServiceReplicaSets(domainServiceProperties.ReplicaSets)); err != nil {
			return fmt.Errorf("setting `replica_set`: %+v", err)
		}
		if err := d.Set("security", flattenDomainServiceSecurity(domainServiceProperties.DomainSecuritySettings)); err != nil {
			return fmt.Errorf("setting `security`: %+v", err)
		}

		d.Set("filtered_sync", false)
		if domainServiceProperties.FilteredSync == aad.FilteredSyncEnabled {
			d.Set("filtered_sync", true)
		}
	}

	return nil
}

func resourceActiveDirectoryDomainServiceDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).DomainServices.DomainServicesClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.DomainServiceID(d.Id())
	if err != nil {
		return err
	}

	future, err := client.Delete(ctx, id.ResourceGroup, id.Name)
	if err != nil {
		if response.WasNotFound(future.Response()) {
			return nil
		}
		return fmt.Errorf("deleting %s: %+v", id.String(), err)
	}

	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		if !response.WasNotFound(future.Response()) {
			return fmt.Errorf("waiting for deletion of %s: %+v", id.String(), err)
		}
	}

	return nil
}

func domainServiceControllerRefreshFunc(ctx context.Context, client *aad.DomainServicesClient, id parse.DomainServiceId) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		log.Printf("[DEBUG] Waiting for domain controllers to deploy...")
		resp, err := client.Get(ctx, id.ResourceGroup, id.Name)
		if err != nil {
			return nil, "error", err
		}
		if resp.ReplicaSets != nil {
			for _, repl := range *resp.ReplicaSets {
				// when a domain controller is online, its IP address will be returned
				if repl.DomainControllerIPAddress == nil || len(*repl.DomainControllerIPAddress) < 2 {
					return resp, "pending", nil
				}
			}
		}
		return resp, "available", nil
	}
}

func expandDomainServiceLdaps(input []interface{}) (ldaps *aad.LdapsSettings) {
	ldaps = &aad.LdapsSettings{
		Ldaps: aad.LdapsDisabled,
	}

	if len(input) == 0 {
		return
	}

	v := input[0].(map[string]interface{})
	ldaps.PfxCertificate = utils.String(v["pfx_certificate"].(string))
	ldaps.PfxCertificatePassword = utils.String(v["pfx_certificate_password"].(string))
	if v["external_access"].(bool) {
		ldaps.ExternalAccess = aad.Enabled
	} else {
		ldaps.ExternalAccess = aad.Disabled
	}

	return
}

func expandDomainServiceNotifications(input []interface{}) *aad.NotificationSettings {
	if len(input) == 0 {
		return nil
	}

	v := input[0].(map[string]interface{})

	additionalRecipients := make([]string, 0)
	if ar, ok := v["additional_recipients"]; ok {
		additionalRecipients = ar.([]string)
	}

	notifyDcAdmins := aad.NotifyDcAdminsDisabled
	if n, ok := v["notify_dc_admins"]; ok && n.(bool) {
		notifyDcAdmins = aad.NotifyDcAdminsEnabled
	}

	notifyGlobalAdmins := aad.NotifyGlobalAdminsDisabled
	if n, ok := v["notify_global_admins"]; ok && n.(bool) {
		notifyGlobalAdmins = aad.NotifyGlobalAdminsEnabled
	}

	return &aad.NotificationSettings{
		AdditionalRecipients: &additionalRecipients,
		NotifyDcAdmins:       notifyDcAdmins,
		NotifyGlobalAdmins:   notifyGlobalAdmins,
	}
}

func expandDomainServiceReplicaSets(input []interface{}) *[]aad.ReplicaSet {
	ret := make([]aad.ReplicaSet, 0)

	for _, replicaRaw := range input {
		replica := replicaRaw.(map[string]interface{})

		loc := ""
		if v, ok := replica["location"]; ok {
			loc = v.(string)
		}

		subnetId := ""
		if v, ok := replica["subnet_id"]; ok {
			subnetId = v.(string)
		}

		ret = append(ret, aad.ReplicaSet{
			Location: &loc,
			SubnetID: &subnetId,
		})
	}

	return &ret
}

func expandDomainServiceSecurity(input []interface{}) *aad.DomainSecuritySettings {
	if len(input) == 0 {
		return nil
	}
	v := input[0].(map[string]interface{})

	ntlmV1 := aad.NtlmV1Disabled
	syncNtlmPasswords := aad.SyncNtlmPasswordsDisabled
	tlsV1 := aad.TLSV1Disabled

	if v["ntlm_v1"].(bool) {
		ntlmV1 = aad.NtlmV1Enabled
	}
	if v["sync_ntlm_passwords"].(bool) {
		syncNtlmPasswords = aad.SyncNtlmPasswordsEnabled
	}
	if v["tls_v1"].(bool) {
		tlsV1 = aad.TLSV1Enabled
	}

	return &aad.DomainSecuritySettings{
		NtlmV1:            ntlmV1,
		SyncNtlmPasswords: syncNtlmPasswords,
		TLSV1:             tlsV1,
	}
}

func flattenDomainServiceLdaps(input *aad.LdapsSettings) []interface{} {
	if input == nil {
		return make([]interface{}, 0)
	}

	result := map[string]interface{}{
		"external_access": false,
		"ldaps":           false,
	}

	if input.ExternalAccess == aad.Enabled {
		result["external_access"] = true
	}
	if input.Ldaps == aad.LdapsEnabled {
		result["ldaps"] = true
	}
	if pfxCertificate := input.PfxCertificate; pfxCertificate != nil {
		result["pfx_certificate"] = *pfxCertificate
	}
	if pfxCertificatePassword := input.PfxCertificatePassword; pfxCertificatePassword != nil {
		result["pfx_certificate_password"] = *pfxCertificatePassword
	}

	return []interface{}{result}
}

func flattenDomainServiceNotifications(input *aad.NotificationSettings) []interface{} {
	if input == nil {
		return make([]interface{}, 0)
	}

	result := map[string]interface{}{
		"additional_recipients": make([]string, 0),
		"notify_dc_admins":      false,
		"notify_global_admins":  false,
	}
	if input.AdditionalRecipients != nil && len(*input.AdditionalRecipients) > 0 {
		result["additional_recipients"] = *input.AdditionalRecipients
	}
	if input.NotifyDcAdmins == aad.NotifyDcAdminsEnabled {
		result["notify_dc_admins"] = true
	}
	if input.NotifyGlobalAdmins == aad.NotifyGlobalAdminsEnabled {
		result["notify_global_admins"] = true
	}

	return []interface{}{result}
}

func flattenDomainServiceReplicaSets(input *[]aad.ReplicaSet) []interface{} {
	return nil
}

func flattenDomainServiceSecurity(input *aad.DomainSecuritySettings) []interface{} {
	if input == nil {
		return make([]interface{}, 0)
	}

	result := map[string]bool{
		"ntlm_v1":             false,
		"sync_ntlm_passwords": false,
		"tls_v1":              false,
	}
	if input.NtlmV1 == aad.NtlmV1Enabled {
		result["ntlm_v1"] = true
	}
	if input.SyncNtlmPasswords == aad.SyncNtlmPasswordsEnabled {
		result["sync_ntlm_passwords"] = true
	}
	if input.TLSV1 == aad.TLSV1Enabled {
		result["tls_v1"] = true
	}

	return []interface{}{result}
}
