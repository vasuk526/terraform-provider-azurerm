package domainservices_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"

	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/acceptance"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/acceptance/check"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/clients"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/services/domainservices/parse"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
)

type ActiveDirectoryDomainServiceResource struct {
	adminPassword string
}

func TestAccActiveDirectoryDomainService_complete(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_active_directory_domain_service", "test")
	r := ActiveDirectoryDomainServiceResource{
		adminPassword: fmt.Sprintf("%s%s", "p@$$Wd", acctest.RandString(6)),
	}

	data.ResourceTest(t, r, []resource.TestStep{
		{
			Config: r.complete(data),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				resource.TestCheckResourceAttr(data.ResourceName, "replica_set.0.domain_controller_ip_addresses.#", "2"),
			),
		},
		data.ImportStep("ldaps.0.pfx_certificate", "ldaps.0.pfx_certificate_password"),
		{
			Config:      r.requiresImport(data),
			ExpectError: acceptance.RequiresImportError("azurerm_active_directory_domain_service"),
		},
	})
}

func (ActiveDirectoryDomainServiceResource) Exists(ctx context.Context, client *clients.Client, state *terraform.InstanceState) (*bool, error) {
	id, err := parse.DomainServiceID(state.ID)
	if err != nil {
		return nil, err
	}

	resp, err := client.DomainServices.DomainServicesClient.Get(ctx, id.ResourceGroup, id.Name)
	if err != nil {
		return nil, fmt.Errorf("reading DomainService: %+v", err)
	}

	return utils.Bool(resp.ID != nil), nil
}

func (r ActiveDirectoryDomainServiceResource) template(data acceptance.TestData, location, replica, cidr string) string {
	return fmt.Sprintf(`
resource "azurerm_resource_group" "test_%[1]s" {
  name     = "acctestRG-aadds-%[1]s-%[4]d"
  location = "%[2]s"
}

resource "azurerm_virtual_network" "test_%[1]s" {
  name                = "acctestVnet-aadds-%[1]s-%[4]d"
  location            = azurerm_resource_group.test_%[1]s.location
  resource_group_name = azurerm_resource_group.test_%[1]s.name
  address_space       = ["%[3]s"]

  lifecycle {
    ignore_changes = [dns_servers]
  }
}

resource "azurerm_subnet" "aadds_%[1]s" {
  name                 = "acctestSubnet-aadds-%[1]s-%[4]d"
  resource_group_name  = azurerm_resource_group.test_%[1]s.name
  virtual_network_name = azurerm_virtual_network.test_%[1]s.name
  address_prefixes     = [cidrsubnet("%[3]s", 8, 0)]
}

resource "azurerm_subnet" "workload_%[1]s" {
  name                 = "acctestSubnet-aadds-%[1]s-%[4]d"
  resource_group_name  = azurerm_resource_group.test_%[1]s.name
  virtual_network_name = azurerm_virtual_network.test_%[1]s.name
  address_prefixes     = [cidrsubnet("%[3]s", 8, 1)]
}

resource "azurerm_network_security_group" "aadds_%[1]s" {
  name                = "acctestNSG-aadds-%[1]s-%[4]d"
  location            = azurerm_resource_group.test_%[1]s.location
  resource_group_name = azurerm_resource_group.test_%[1]s.name

  security_rule {
    name                       = "AllowSyncWithAzureAD"
    priority                   = 101
    direction                  = "Inbound"
    access                     = "Allow"
    protocol                   = "Tcp"
    source_port_range          = "*"
    destination_port_range     = "443"
    source_address_prefix      = "AzureActiveDirectoryDomainServices"
    destination_address_prefix = "*"
  }

  security_rule {
    name                       = "AllowRD"
    priority                   = 201
    direction                  = "Inbound"
    access                     = "Allow"
    protocol                   = "Tcp"
    source_port_range          = "*"
    destination_port_range     = "3389"
    source_address_prefix      = "CorpNetSaw"
    destination_address_prefix = "*"
  }

  security_rule {
    name                       = "AllowPSRemoting"
    priority                   = 301
    direction                  = "Inbound"
    access                     = "Allow"
    protocol                   = "Tcp"
    source_port_range          = "*"
    destination_port_range     = "5986"
    source_address_prefix      = "AzureActiveDirectoryDomainServices"
    destination_address_prefix = "*"
  }

  security_rule {
    name                       = "AllowLDAPS"
    priority                   = 401
    direction                  = "Inbound"
    access                     = "Allow"
    protocol                   = "Tcp"
    source_port_range          = "*"
    destination_port_range     = "636"
    source_address_prefix      = "*"
    destination_address_prefix = "*"
  }
}

resource azurerm_subnet_network_security_group_association "test_%[1]s" {
  subnet_id                 = azurerm_subnet.aadds_%[1]s.id
  network_security_group_id = azurerm_network_security_group.aadds_%[1]s.id
}
`, replica, location, cidr, data.RandomInteger)
}

func (r ActiveDirectoryDomainServiceResource) complete(data acceptance.TestData) string {
	template1 := r.template(data, data.Locations.Primary, "one", "10.10.0.0/16")
	template2 := r.template(data, data.Locations.Secondary, "two", "10.20.0.0/16")

	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

provider "azuread" {}

%[1]s
%[2]s

resource "azurerm_virtual_network_peering" "test" {
  name                      = "acctestVnet-aadds-%[4]d"
  resource_group_name       = azurerm_virtual_network.test_one.resource_group_name
  virtual_network_name      = azurerm_virtual_network.test_one.name
  remote_virtual_network_id = azurerm_virtual_network.test_two.id

  allow_forwarded_traffic      = true
  allow_gateway_transit        = false
  allow_virtual_network_access = true
  use_remote_gateways          = false
}

resource "azurerm_resource_group" "aadds" {
  name     = "acctestRG-aadds-%[4]d"
  location = "%[3]s"
}

data "azuread_domains" "test" {
  only_initial = true
}

resource "azuread_service_principal" "test" {
  application_id = "2565bd9d-da50-47d4-8b85-4c97f669dc36" // published app for domain services
}

resource "azuread_group" "test" {
  name = "AAD DC Administrators"
}

resource "azuread_user" "test" {
  user_principal_name = "acctestAADDSAdminUser-%[4]d@${data.azuread_domains.test.domains.0.domain_name}"
  display_name        = "acctestAADDSAdminUser-%[4]d"
  password            = "%[6]s"
}

resource "azuread_group_member" "test" {
  group_object_id  = azuread_group.test.object_id
  member_object_id = azuread_user.test.object_id
}

resource "azurerm_active_directory_domain_service" "test" {
  name                = "acctest-%[5]s.net"
  location            = azurerm_resource_group.aadds.location
  resource_group_name = azurerm_resource_group.aadds.name
  filtered_sync       = false

  //ldaps {
  //  external_access          = true
  //  pfx_certificate          = "TODO Generate a dummy pfx key+cert (https://docs.microsoft.com/en-us/azure/active-directory-domain-services/tutorial-configure-ldaps)"
  //  pfx_certificate_password = "test"
  //}

  replica_set {
    location  = azurerm_virtual_network.test_one.location
    subnet_id = azurerm_subnet.aadds_one.id
  }

  replica_set {
    location  = azurerm_virtual_network.test_two.location
    subnet_id = azurerm_subnet.aadds_two.id
  }

  security {
    ntlm_v1             = true
    tls_v1              = true
    sync_ntlm_passwords = true
  }

  depends_on = [
    //azuread_service_principal.test,
    azurerm_subnet_network_security_group_association.test_one,
    azurerm_subnet_network_security_group_association.test_two,
    azurerm_virtual_network_peering.test,
  ]
}
`, template1, template2, data.Locations.Primary, data.RandomInteger, data.RandomString, r.adminPassword)
}

func (r ActiveDirectoryDomainServiceResource) requiresImport(data acceptance.TestData) string {
	return fmt.Sprintf(`
%[1]s

resource "azurerm_active_directory_domain_service" "import" {
  name                = azurerm_active_directory_domain_service.test.name
  location            = azurerm_resource_group.aadds.location
  resource_group_name = azurerm_resource_group.aadds.name
}
`, r.complete(data))
}
