package common

import (
	"github.com/Azure/azure-sdk-for-go/services/preview/cosmos-db/mgmt/2020-04-01-preview/documentdb"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/services/cosmos/validate"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/tf/pluginsdk"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/tf/suppress"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/tf/validation"
)

func CassandraTableSchemaPropertySchema() *pluginsdk.Schema {
	return &pluginsdk.Schema{
		Type:     pluginsdk.TypeList,
		Required: true,
		MaxItems: 1,
		Elem: &pluginsdk.Resource{
			Schema: map[string]*pluginsdk.Schema{
				"column": {
					Type:     pluginsdk.TypeList,
					Required: true,
					MinItems: 1,
					Elem: &pluginsdk.Resource{
						Schema: map[string]*pluginsdk.Schema{
							"name": {
								Required:     true,
								Type:         pluginsdk.TypeString,
								ValidateFunc: validation.StringIsNotEmpty,
							},
							"type": {
								Required:     true,
								Type:         pluginsdk.TypeString,
								ValidateFunc: validation.StringIsNotEmpty,
							},
						},
					},
				},
				"partition_key": {
					Type:     pluginsdk.TypeList,
					Required: true,
					Elem: &pluginsdk.Resource{
						Schema: map[string]*pluginsdk.Schema{
							"name": {
								Required:     true,
								Type:         pluginsdk.TypeString,
								ValidateFunc: validation.StringIsNotEmpty,
							},
						},
					},
				},
				"cluster_key": {
					Optional: true,
					Type:     pluginsdk.TypeList,
					Elem: &pluginsdk.Resource{
						Schema: map[string]*pluginsdk.Schema{
							"name": {
								Type:         pluginsdk.TypeString,
								Required:     true,
								ValidateFunc: validation.StringIsNotEmpty,
							},
							"order_by": {
								Type:     pluginsdk.TypeString,
								Required: true,
								ValidateFunc: validation.StringInSlice([]string{
									"Asc",
									"Desc",
								}, false),
							},
						},
					},
				},
			},
		},
	}
}

func DatabaseAutoscaleSettingsSchema() *pluginsdk.Schema {
	return &pluginsdk.Schema{
		Type:     pluginsdk.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &pluginsdk.Resource{
			Schema: map[string]*pluginsdk.Schema{
				"max_throughput": {
					Type:          pluginsdk.TypeInt,
					Optional:      true,
					Computed:      true,
					ConflictsWith: []string{"throughput"},
					ValidateFunc:  validate.CosmosMaxThroughput,
				},
			},
		},
	}
}

func ContainerAutoscaleSettingsSchema() *pluginsdk.Schema {
	autoscaleSettingsDatabaseSchema := DatabaseAutoscaleSettingsSchema()
	autoscaleSettingsDatabaseSchema.RequiredWith = []string{"partition_key_path"}

	return autoscaleSettingsDatabaseSchema
}

func MongoCollectionAutoscaleSettingsSchema() *pluginsdk.Schema {
	autoscaleSettingsDatabaseSchema := DatabaseAutoscaleSettingsSchema()
	autoscaleSettingsDatabaseSchema.RequiredWith = []string{"shard_key"}

	return autoscaleSettingsDatabaseSchema
}

func CosmosDbIndexingPolicySchema() *pluginsdk.Schema {
	return &pluginsdk.Schema{
		Type:     pluginsdk.TypeList,
		Optional: true,
		Computed: true,
		MaxItems: 1,
		Elem: &pluginsdk.Resource{
			Schema: map[string]*pluginsdk.Schema{
				// `automatic` is excluded as it is deprecated; see https://stackoverflow.com/a/58721386
				"indexing_mode": {
					Type:             pluginsdk.TypeString,
					Optional:         true,
					Default:          documentdb.Consistent,
					DiffSuppressFunc: suppress.CaseDifference, // Open issue https://github.com/Azure/azure-sdk-for-go/issues/6603
					ValidateFunc: validation.StringInSlice([]string{
						string(documentdb.Consistent),
						string(documentdb.None),
					}, false),
				},

				"included_path": {
					Type:     pluginsdk.TypeList,
					Optional: true,
					Computed: true,
					Elem: &pluginsdk.Resource{
						Schema: map[string]*pluginsdk.Schema{
							"path": {
								Type:         pluginsdk.TypeString,
								Required:     true,
								ValidateFunc: validation.StringIsNotEmpty,
							},
						},
					},
				},
				"excluded_path": {
					Type:     pluginsdk.TypeList,
					Optional: true,
					Computed: true,
					Elem: &pluginsdk.Resource{
						Schema: map[string]*pluginsdk.Schema{
							"path": {
								Type:         pluginsdk.TypeString,
								Required:     true,
								ValidateFunc: validation.StringIsNotEmpty,
							},
						},
					},
				},
				"composite_index": {
					Type:     pluginsdk.TypeList,
					Optional: true,
					Elem: &pluginsdk.Resource{
						Schema: map[string]*pluginsdk.Schema{
							"index": {
								Type:     pluginsdk.TypeList,
								MinItems: 1,
								Required: true,
								Elem: &pluginsdk.Resource{
									Schema: map[string]*pluginsdk.Schema{
										"path": {
											Type:         pluginsdk.TypeString,
											Required:     true,
											ValidateFunc: validation.StringIsNotEmpty,
										},
										"order": {
											Type:     pluginsdk.TypeString,
											Required: true,
											// Workaround for Azure/azure-rest-api-specs#11222
											DiffSuppressFunc: suppress.CaseDifference,
											ValidateFunc: validation.StringInSlice(
												[]string{
													string(documentdb.Ascending),
													string(documentdb.Descending),
												}, false),
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}
