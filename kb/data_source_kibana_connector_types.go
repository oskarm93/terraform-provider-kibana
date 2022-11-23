// Return the connection settings of Kibana
// Supported version:
//  - v8

package kb

import (
	"context"

	kibana "github.com/disaster37/go-kibana-rest/v8"
	"github.com/disaster37/go-kibana-rest/v8/kbapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceKibanaConnectorTypes() *schema.Resource {
	return &schema.Resource{
		Description: "`kibana_connector_types` can be used to retrieve the list of supported Kibana connector types.",
		ReadContext: dataSourceKibanaConnectorTypesRead,

		Schema: map[string]*schema.Schema{
			"connector_types": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Connector Type ID",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Connector Type name",
						},
						"enabled": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Whether connector type is enabled",
						},
						"enabled_in_config": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Whether connector type is enabled in configuration",
						},
						"enabled_in_license": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Whether connector type is enabled in the current license",
						},
						"minimum_license_required": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Minimum license level requried to use this connector type",
						},
						"supported_feature_ids": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "Supported features where connector type can be used",
							Elem:        schema.TypeString,
						},
					},
				},
			},
		},
	}
}

func dataSourceKibanaConnectorTypesRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var err error

	client := m.(*kibana.Client)
	typesReply, err := client.API.KibanaConnectorTypes.List()
	if err != nil {
		return diag.FromErr(err)
	}

	types := flattenConnectorTypes(&typesReply)
	err = d.Set("connector_types", types)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("0")

	return nil
}

func flattenConnectorTypes(connectorTypes *kbapi.KibanaConnectorTypes) []interface{} {
	if connectorTypes != nil {
		cts := make([]interface{}, len(*connectorTypes), len(*connectorTypes))

		for i, connectorType := range *connectorTypes {
			ct := make(map[string]interface{})

			ct["id"] = connectorType.ID
			ct["name"] = connectorType.Name
			ct["enabled"] = connectorType.Enabled
			ct["enabled_in_config"] = connectorType.EnabledInConfig
			ct["enabled_in_license"] = connectorType.EnabledInLicense
			ct["minimum_license_required"] = connectorType.MinimumLicenseRequired
			ct["supported_feature_ids"] = connectorType.SupportedFeatureIDs
			cts[i] = ct
		}

		return cts
	}

	return make([]interface{}, 0)
}
