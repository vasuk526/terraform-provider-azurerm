package domainservices_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"

	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/azure"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/acceptance"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/acceptance/check"
)

type ActiveDirectoryDomainServiceDataSource struct{}

func TestAccDataSourceActiveDirectoryDomainService_complete(t *testing.T) {
	data := acceptance.BuildTestData(t, "data.azurerm_active_directory_domain_service", "test")
	r := ActiveDirectoryDomainServiceDataSource{}

	data.DataSourceTest(t, []resource.TestStep{
		{
			Config: r.complete(data),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).Key("filtered_sync_enabled").HasValue("false"),
				check.That(data.ResourceName).Key("ldaps.#").HasValue("1"),
				check.That(data.ResourceName).Key("ldaps.0.enabled").HasValue("false"),
				check.That(data.ResourceName).Key("ldaps.0.external_access_enabled").HasValue("false"),
				check.That(data.ResourceName).Key("location").HasValue(azure.NormalizeLocation(data.Locations.Primary)),
				check.That(data.ResourceName).Key("notifications.#").HasValue("1"),
				check.That(data.ResourceName).Key("notifications.0.additional_recipients.#").HasValue("2"),
				check.That(data.ResourceName).Key("notifications.0.notify_dc_admins").HasValue("true"),
				check.That(data.ResourceName).Key("notifications.0.notify_global_admins").HasValue("true"),
				check.That(data.ResourceName).Key("replica_sets.#").HasValue("2"),
				check.That(data.ResourceName).Key("replica_sets.0.domain_controller_ip_addresses.#").HasValue("2"),
				check.That(data.ResourceName).Key("replica_sets.0.location").HasValue(azure.NormalizeLocation(data.Locations.Primary)),
				check.That(data.ResourceName).Key("replica_sets.0.replica_set_id").Exists(),
				check.That(data.ResourceName).Key("replica_sets.0.service_status").Exists(),
				check.That(data.ResourceName).Key("replica_sets.0.subnet_id").Exists(),
				check.That(data.ResourceName).Key("replica_sets.0.vnet_site_id").Exists(),
				check.That(data.ResourceName).Key("replica_sets.1.domain_controller_ip_addresses.#").HasValue("2"),
				check.That(data.ResourceName).Key("replica_sets.1.location").HasValue(azure.NormalizeLocation(data.Locations.Secondary)),
				check.That(data.ResourceName).Key("replica_sets.1.replica_set_id").Exists(),
				check.That(data.ResourceName).Key("replica_sets.1.service_status").Exists(),
				check.That(data.ResourceName).Key("replica_sets.1.subnet_id").Exists(),
				check.That(data.ResourceName).Key("replica_sets.1.vnet_site_id").Exists(),
				check.That(data.ResourceName).Key("security.#").HasValue("1"),
				check.That(data.ResourceName).Key("security.0.ntlm_v1_enabled").HasValue("true"),
				check.That(data.ResourceName).Key("security.0.sync_kerberos_passwords").HasValue("true"),
				check.That(data.ResourceName).Key("security.0.sync_ntlm_passwords").HasValue("true"),
				check.That(data.ResourceName).Key("security.0.sync_on_prem_passwords").HasValue("true"),
				check.That(data.ResourceName).Key("security.0.tls_v1_enabled").HasValue("true"),
				check.That(data.ResourceName).Key("sku").HasValue("Enterprise"),
			),
		},
	})
}

func (r ActiveDirectoryDomainServiceDataSource) complete(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

data "azurerm_active_directory_domain_service" "test" {
  name                = "acctest-jdlux.net"
  resource_group_name = "acctestRG-aadds-210302023734044644"
}
`)
}
