// Manage alert rules in Kibana
// API documentation: https://www.elastic.co/guide/en/kibana/current/alerting-apis.html
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

// Resource specification to handle alert rule in Kibana
func resourceKibanaAlertRule() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceKibanaAlertRuleCreate,
		ReadContext:   resourceKibanaAlertRuleRead,
		UpdateContext: resourceKibanaAlertRuleUpdate,
		DeleteContext: resourceKibanaAlertRuleDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"consumer": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "alerts",
			},
			"tags": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Default: nil,
				// Default: []string{},
			},
			"throttle": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  nil,
			},
			"enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"schedule": {
				Type:     schema.TypeMap,
				Required: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"params": {
				Type:     schema.TypeString,
				Required: true,
				// DiffSuppressFunc: rawJsonEqual,
				// Elem:     &schema.Schema{
				// Type: schema.TypeString,
				// },
			},
			"actions": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeMap,
				},
				Default: nil, //[]kbapi.KibanaAlertRuleAction{},
			},
			"rule_type_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"created_by": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"updated_by": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"updated_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"api_key_owner": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"notify_when": {
				Type:     schema.TypeString,
				Required: true,
			},
			"mute_alert_ids": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"mute_all": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"scheduled_task_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"execution_status": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

// Create new alert rule in Kibana
func resourceKibanaAlertRuleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*kibana.Client)
	createParams := &kbapi.KibanaAlertRuleCreateParams{
		Name:     d.Get("name").(string),
		Consumer: d.Get("consumer").(string),
		Tags:     []string{},
		// Tags:       ([]string)(d.Get("tags").([]interface{}),
		Throttle: d.Get("throttle").(string),
		Enabled:  d.Get("enabled").(bool),
		Schedule: kbapi.KibanaAlertRuleSchedule{
			Interval: "1m",
		}, // (kbapi.KibanaAlertRuleSchedule)(d.Get("schedule").(map[string]interface{})),
		Params:     d.Get("params").(kbapi.KibanaAlertRuleParams),
		RuleTypeID: d.Get("rule_type_id").(string),
		NotifyWhen: d.Get("notify_when").(string),
		Actions:    d.Get("actions").([]kbapi.KibanaAlertRuleAction),
	}

	alertRule, err := client.API.KibanaAlertRule.Create(createParams)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(alertRule.ID)

	log.Infof("Created alert rule %s (%s) successfully", alertRule.ID, alertRule.Name)
	fmt.Printf("[INFO] Created alert rule %s (%s) successfully", alertRule.ID, alertRule.Name)

	return resourceKibanaAlertRuleRead(ctx, d, meta)
}

// Read existing alert rule in Kibana
func resourceKibanaAlertRuleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var err error
	id := d.Id()

	log.Debugf("AlertRule ID:  %s", id)

	client := meta.(*kibana.Client)

	alert_rule, err := client.API.KibanaAlertRule.Get(id)
	if err != nil {
		return diag.FromErr(err)
	}

	if alert_rule == nil {
		log.Warnf("AlertRule %s not found - removing from state", id)
		fmt.Printf("[WARN] AlertRule %s not found - removing from state", id)
		d.SetId("")
		return nil
	}

	log.Debugf("Get alert rule %s successfully:\n%s", id, alert_rule)

	// if err = d.Set("name", alert_rule.Name); err != nil {
	// 	return diag.FromErr(err)
	// }
	// if err = d.Set("connector_type_id", alert_rule.AlertRuleTypeID); err != nil {
	// 	return diag.FromErr(err)
	// }
	// if err = d.Set("is_preconfigured", alert_rule.IsPreconfigured); err != nil {
	// 	return diag.FromErr(err)
	// }
	// if err = d.Set("is_deprecated", alert_rule.IsDeprecated); err != nil {
	// 	return diag.FromErr(err)
	// }
	// if err = d.Set("is_missing_secrets", alert_rule.IsMissingSecrets); err != nil {
	// 	return diag.FromErr(err)
	// }
	// if err = d.Set("referenced_by_count", alert_rule.ReferencedByCount); err != nil {
	// 	return diag.FromErr(err)
	// }
	// if err = d.Set("config", alert_rule.Config); err != nil {
	// 	return diag.FromErr(err)
	// }

	log.Infof("Read alert rule %s successfully", id)
	fmt.Printf("[INFO] Read alert rule %s successfully", id)

	return nil
}

// Update existing alert rule in Kibana
func resourceKibanaAlertRuleUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// id := d.Id()
	// name := d.Get("name").(string)
	// config := (kbapi.KibanaAlertRuleConfig)(d.Get("config").(map[string]interface{}))
	// secrets := (kbapi.KibanaAlertRuleSecrets)(d.Get("secrets").(map[string]interface{}))

	// client := meta.(*kibana.Client)

	// createParams := &kbapi.KibanaAlertRuleCreateParams{
	// 	Name:    name,
	// 	Config:  config,
	// 	Secrets: secrets,
	// }

	// connector, err := client.API.KibanaAlertRule.Update(id, createParams)
	// if err != nil {
	// 	return diag.FromErr(err)
	// }

	// d.SetId(connector.ID)
	// d.Set("secrets", createParams.Secrets)

	// log.Infof("Updated alert rule %s (%s) successfully", connector.ID, name)
	// fmt.Printf("[INFO] Updated alert rule %s (%s) successfully", connector.ID, name)

	return resourceKibanaAlertRuleRead(ctx, d, meta)
}

// Delete existing alert rule in Kibana
func resourceKibanaAlertRuleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// id := d.Id()
	// log.Debugf("AlertRule ID: %s", id)

	// client := meta.(*kibana.Client)

	// err := client.API.KibanaAlertRule.Delete(id)
	// if err != nil {
	// 	if err.(kbapi.APIError).Code == 404 {
	// 		log.Warnf("AlertRule %s not found - removing from state", id)
	// 		fmt.Printf("[WARN] AlertRule %s not found - removing from state", id)
	// 		d.SetId("")
	// 		return nil
	// 	}
	// 	return diag.FromErr(err)
	// }

	// d.SetId("")

	// log.Infof("Deleted alert rule %s successfully", id)
	// fmt.Printf("[INFO] Deleted alert rule %s successfully", id)
	return nil

}
