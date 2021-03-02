---
subcategory: "Active Directory Domain Services"
layout: "azurerm"
page_title: "Azure Resource Manager: azurerm_active_directory_domain_service"
description: |-
  Manages an Active Directory Domain Service.
---

# azurerm_active_directory_domain_service

Manages an Active Directory Domain Service.

~> Implementation Note: Before using this resource, there must exist in your tenant a service principal for the Domain Services published application. This service principal cannot be easily managed by Terraform and it's recommended to create this manually, as it does not exist by default. See [official documentation](https://docs.microsoft.com/en-us/azure/active-directory-domain-services/powershell-create-instance#create-required-azure-ad-resources) for details.

## Example Usage

```hcl
resource "azurerm_resource_group" "deploy" {
  name     = "deploy-rg"
  location = "westeurope"
}

resource "azurerm_virtual_network" "deploy" {
  name                = "deploy-vnet"
  location            = azurerm_resource_group.deploy.location
  resource_group_name = azurerm_resource_group.deploy.name
  address_space       = ["10.0.1.0/16"]

  lifecycle {
    ignore_changes = [dns_servers]
  }
}

resource "azurerm_subnet" "deploy" {
  name                 = "deploy-subnet"
  resource_group_name  = azurerm_resource_group.deploy.name
  virtual_network_name = azurerm_virtual_network.deploy.name
  address_prefixes     = ["10.0.1.0/24"]
}

resource "azurerm_network_security_group" "deploy" {
  name                = "deploy-nsg"
  location            = azurerm_resource_group.deploy.location
  resource_group_name = azurerm_resource_group.deploy.name

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

resource azurerm_subnet_network_security_group_association "deploy" {
  subnet_id                 = azurerm_subnet.deploy.id
  network_security_group_id = azurerm_network_security_group.deploy.id
}

resource "azuread_group" "dc_admins" {
  name = "AAD DC Administrators"
}

resource "azuread_user" "admin" {
  user_principal_name = "dc-admin@$hashicorp-example.net"
  display_name        = "DC Administrator"
  password            = "Pa55w0Rd!!1"
}

resource "azuread_group_member" "admin" {
  group_object_id  = azuread_group.dc_admins.object_id
  member_object_id = azuread_user.admin.object_id
}

resource "azuread_service_principal" "example" {
  application_id = "2565bd9d-da50-47d4-8b85-4c97f669dc36" // published app for domain services
}

resource "azurerm_resource_group" "aadds" {
  name     = "aadds-rg"
  location = "westeurope"
}

resource "azurerm_active_directory_domain_service" "example" {
  name                = "domain.hashicorp-example.net"
  location            = azurerm_resource_group.aadds.location
  resource_group_name = azurerm_resource_group.aadds.name
  filtered_sync       = false

  replica_set {
    location  = azurerm_virtual_network.deploy.location
    subnet_id = azurerm_subnet.deploy.id
  }

  security {
    ntlm_v1             = true
    tls_v1              = true
    sync_ntlm_passwords = true
  }

  depends_on = [azurerm_subnet_network_security_group_association.deploy]
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The domain name for your managed Active Directory domain.

* `location` - (Required) The Azure location where the Domain Service exists. Changing this forces a new resource to be created.

* `resource_group_name` - (Required) The name of the Resource Group in which the Domain Service should exist. Changing this forces a new resource to be created.

* `filtered_sync` - (Optional) BLAH BLAH. Defaults to `false`.

---

* `ldaps` - (Optional) An `ldaps` block as defined below.

* `notifications` - (Optional) A `notifications` block as defined below.

* `replica_set` - (Required) One or more `replica_set` blocks as defined below.

* `security` - (Optional) A `security` block as defined below.

* `tags` - (Optional) A mapping of tags assigned to the resource.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the Domain Service.

## Timeouts

The `timeouts` block allows you to specify [timeouts](https://www.terraform.io/docs/configuration/resources.html#timeouts) for certain actions:

* `create` - (Defaults to 2 hours) Used when creating the Domain Service.
* `update` - (Defaults to 2 hours) Used when updating the Domain Service.
* `read` - (Defaults to 5 minutes) Used when retrieving the Domain Service.
* `delete` - (Defaults to 30 minutes) Used when deleting the Domain Service.

## Import

Domain Services can be imported using the resource ID, e.g.

```shell
terraform import azurerm_active_directory_domain_service.example /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/mygroup1/providers/Microsoft.AAD/domainServices/instance1
```
