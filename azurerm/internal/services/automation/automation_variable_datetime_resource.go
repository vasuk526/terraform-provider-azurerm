package automation

import (
	"time"

	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/tf/pluginsdk"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/tf/validation"
)

func resourceAutomationVariableDateTime() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourceAutomationVariableDateTimeCreateUpdate,
		Read:   resourceAutomationVariableDateTimeRead,
		Update: resourceAutomationVariableDateTimeCreateUpdate,
		Delete: resourceAutomationVariableDateTimeDelete,

		Importer: &pluginsdk.ResourceImporter{
			State: pluginsdk.ImportStatePassthrough,
		},

		Timeouts: &pluginsdk.ResourceTimeout{
			Create: pluginsdk.DefaultTimeout(30 * time.Minute),
			Read:   pluginsdk.DefaultTimeout(5 * time.Minute),
			Update: pluginsdk.DefaultTimeout(30 * time.Minute),
			Delete: pluginsdk.DefaultTimeout(30 * time.Minute),
		},

		Schema: resourceAutomationVariableCommonSchema(pluginsdk.TypeString, validation.IsRFC3339Time),
	}
}

func resourceAutomationVariableDateTimeCreateUpdate(d *pluginsdk.ResourceData, meta interface{}) error {
	return resourceAutomationVariableCreateUpdate(d, meta, "Datetime")
}

func resourceAutomationVariableDateTimeRead(d *pluginsdk.ResourceData, meta interface{}) error {
	return resourceAutomationVariableRead(d, meta, "Datetime")
}

func resourceAutomationVariableDateTimeDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	return resourceAutomationVariableDelete(d, meta, "Datetime")
}
