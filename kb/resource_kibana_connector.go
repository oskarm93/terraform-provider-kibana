// Manage connectors in Kibana
// API documentation: https://www.elastic.co/guide/en/kibana/current/actions-and-connectors-api.html
// Supported version:
//  - v8

package kb

import (
	"context"
	"fmt"

	kibana "github.com/disaster37/go-kibana-rest/v8"
	kbapi "github.com/disaster37/go-kibana-rest/v8/kbapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	log "github.com/sirupsen/logrus"
)

// Resource specification to handle connector in Kibana
func resourceKibanaConnector() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceKibanaConnectorCreate,
		ReadContext:   resourceKibanaConnectorRead,
		UpdateContext: resourceKibanaConnectorUpdate,
		DeleteContext: resourceKibanaConnectorDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"connector_type_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"is_preconfigured": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"is_deprecated": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"is_missing_secrets": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"referenced_by_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"config": {
				Type:     schema.TypeMap,
				Optional: true,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"secrets": {
				Type:      schema.TypeMap,
				Optional:  true,
				Computed:  true,
				Sensitive: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

// Create new connector in Kibana
func resourceKibanaConnectorCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	connectorTypeID := d.Get("connector_type_id").(string)
	config := (kbapi.KibanaConnectorConfig)(d.Get("config").(map[string]interface{}))
	secrets := (kbapi.KibanaConnectorSecrets)(d.Get("secrets").(map[string]interface{}))

	client := meta.(*kibana.Client)

	createParams := &kbapi.KibanaConnectorCreateParams{
		Name:            name,
		ConnectorTypeID: connectorTypeID,
		Config:          config,
		Secrets:         secrets,
	}

	connector, err := client.API.KibanaConnector.Create(createParams)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(connector.ID)
	d.Set("secrets", createParams.Secrets)

	log.Infof("Created connector %s (%s) successfully", connector.ID, name)
	fmt.Printf("[INFO] Created connector %s (%s) successfully", connector.ID, name)

	return resourceKibanaConnectorRead(ctx, d, meta)
}

// Read existing connector in Kibana
func resourceKibanaConnectorRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var err error
	id := d.Id()

	log.Debugf("Connector ID:  %s", id)

	client := meta.(*kibana.Client)

	connector, err := client.API.KibanaConnector.Get(id)
	if err != nil {
		return diag.FromErr(err)
	}

	if connector == nil {
		log.Warnf("Connector %s not found - removing from state", id)
		fmt.Printf("[WARN] Connector %s not found - removing from state", id)
		d.SetId("")
		return nil
	}

	log.Debugf("Get connector %s successfully:\n%s", id, connector)

	if err = d.Set("name", connector.Name); err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set("connector_type_id", connector.ConnectorTypeID); err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set("is_preconfigured", connector.IsPreconfigured); err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set("is_deprecated", connector.IsDeprecated); err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set("is_missing_secrets", connector.IsMissingSecrets); err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set("referenced_by_count", connector.ReferencedByCount); err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set("config", connector.Config); err != nil {
		return diag.FromErr(err)
	}

	log.Infof("Read connector %s successfully", id)
	fmt.Printf("[INFO] Read connector %s successfully", id)

	return nil
}

// Update existing connector in Kibana
func resourceKibanaConnectorUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	id := d.Id()
	name := d.Get("name").(string)
	config := (kbapi.KibanaConnectorConfig)(d.Get("config").(map[string]interface{}))
	secrets := (kbapi.KibanaConnectorSecrets)(d.Get("secrets").(map[string]interface{}))

	client := meta.(*kibana.Client)

	createParams := &kbapi.KibanaConnectorCreateParams{
		Name:    name,
		Config:  config,
		Secrets: secrets,
	}

	connector, err := client.API.KibanaConnector.Update(id, createParams)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(connector.ID)
	d.Set("secrets", createParams.Secrets)

	log.Infof("Updated connector %s (%s) successfully", connector.ID, name)
	fmt.Printf("[INFO] Updated connector %s (%s) successfully", connector.ID, name)

	return resourceKibanaConnectorRead(ctx, d, meta)
}

// Delete existing connector in Kibana
func resourceKibanaConnectorDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	id := d.Id()
	log.Debugf("Connector ID: %s", id)

	client := meta.(*kibana.Client)

	err := client.API.KibanaConnector.Delete(id)
	if err != nil {
		if err.(kbapi.APIError).Code == 404 {
			log.Warnf("Connector %s not found - removing from state", id)
			fmt.Printf("[WARN] Connector %s not found - removing from state", id)
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	d.SetId("")

	log.Infof("Deleted connector %s successfully", id)
	fmt.Printf("[INFO] Deleted connector %s successfully", id)
	return nil

}
