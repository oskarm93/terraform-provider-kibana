// Big props to the authors of this repo for doing a lot of the legwork
// https://github.com/qonto/terraform-provider-kibana/blob/main/internal/provider/resource_alert_rule.go

// Manage alert rules in Kibana
// API documentation: https://www.elastic.co/guide/en/kibana/current/alerting-apis.html
// Supported version:
//  - v8

package kb

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	kibana "github.com/disaster37/go-kibana-rest/v8"
	"github.com/disaster37/go-kibana-rest/v8/kbapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"
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
				Description: "A name to reference and search.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"tags": {
				Description: "A list of keywords to reference and search.",
				Type:        schema.TypeList,
				Optional:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"rule_type_id": {
				Description: "The ID of the rule type that you want to call when the rule is scheduled to run.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"schedule": {
				Description: "The schedule specifying when this rule should be run, using one of the available schedule formats.",
				Type:        schema.TypeMap,
				Required:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"throttle": {
				Description: "How often this rule should fire the same actions. This will prevent the rule from sending out the same notification over and over. For example, if a rule with a schedule of 1 minute stays in a triggered state for 90 minutes, setting a throttle of 10m or 1h will prevent it from sending 90 notifications during this period.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"notify_when": {
				Description: "The condition for throttling the notification: onActionGroupChange, onActiveAlert, or onThrottleInterval.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"enabled": {
				Description: "Indicates if you want to run the rule on an interval basis after it is created.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
			},
			"consumer": {
				Description: "The name of the application that owns the rule. This name has to match the Kibana Feature name, as that dictates the required RBAC privileges.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"params": {
				Description:      "The parameters to pass to the rule type executor params value. This will also validate against the rule type params validator, if defined.",
				Type:             schema.TypeString,
				Required:         true,
				DiffSuppressFunc: rawJsonEqual,
			},
			"actions": {
				Description: "An array of the following action objects.",
				Type:        schema.TypeList,
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Description: "The ID of the connector saved object to execute.",
							Type:        schema.TypeString,
							Required:    true,
						},
						"group": {
							Description: "Grouping actions is recommended for escalations for different types of alerts. If you don’t need this, set this value to default.",
							Type:        schema.TypeString,
							Required:    true,
						},
						"params": {
							Description:      "The map to the params that the connector type will receive. ` params` are handled as Mustache templates and passed a default set of context.",
							Type:             schema.TypeString,
							Required:         true,
							DiffSuppressFunc: rawJsonEqual,
						},
					},
				},
			},
		},
	}
}

// Create new alert rule in Kibana
func resourceKibanaAlertRuleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*kibana.Client)
	createParams := &kbapi.KibanaAlertRuleCreateParams{
		Name:       d.Get("name").(string),
		Consumer:   d.Get("consumer").(string),
		Throttle:   d.Get("throttle").(string),
		RuleTypeID: d.Get("rule_type_id").(string),
		NotifyWhen: d.Get("notify_when").(string),
		Enabled:    d.Get("enabled").(bool),
	}

	tags := d.Get("tags").([]interface{})
	for _, tag := range tags {
		createParams.Tags = append(createParams.Tags, tag.(string))
	}

	schedule := d.Get("schedule").(map[string]interface{})
	scheduleInterval := schedule["interval"].(string)
	createParams.Schedule = kbapi.KibanaAlertRuleSchedule{
		Interval: scheduleInterval,
	}

	params := d.Get("params").(string)
	createParams.Params = json.RawMessage([]byte(params))

	actionsInterface := d.Get("actions").([]interface{})
	actionsList := make([]map[string]interface{}, 0, len(actionsInterface))
	for _, action := range actionsInterface {
		actionsList = append(actionsList, action.(map[string]interface{}))
	}

	var err error
	createParams.Actions, err = deflateActions(actionsList)
	if err != nil {
		return diag.FromErr(err)
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

	if err = d.Set("name", alert_rule.Name); err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set("tags", alert_rule.Tags); err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set("consumer", alert_rule.Consumer); err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set("enabled", alert_rule.Enabled); err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set("rule_type_id", alert_rule.RuleTypeID); err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set("schedule", map[string]string{
		"interval": alert_rule.Schedule.Interval,
	}); err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set("throttle", alert_rule.Throttle); err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set("notify_when", alert_rule.NotifyWhen); err != nil {
		return diag.FromErr(err)
	}

	paramsBytes, err := json.Marshal(alert_rule.Params)
	if err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set("params", string(paramsBytes)); err != nil {
		return diag.FromErr(err)
	}

	flattenedActions, err := flattenActions(alert_rule.Actions)
	if err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set("actions", flattenedActions); err != nil {
		return diag.FromErr(err)
	}

	log.Infof("Read alert rule %s successfully", id)
	fmt.Printf("[INFO] Read alert rule %s successfully", id)

	return nil
}

// Update existing alert rule in Kibana
func resourceKibanaAlertRuleUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	id := d.Id()
	updateParams := &kbapi.KibanaAlertRuleUpdateParams{
		Name:       d.Get("name").(string),
		Throttle:   d.Get("throttle").(string),
		NotifyWhen: d.Get("notify_when").(string),
	}

	tags := d.Get("tags").([]interface{})
	for _, tag := range tags {
		updateParams.Tags = append(updateParams.Tags, tag.(string))
	}

	schedule := d.Get("schedule").(map[string]interface{})
	scheduleInterval := schedule["interval"].(string)
	updateParams.Schedule = kbapi.KibanaAlertRuleSchedule{
		Interval: scheduleInterval,
	}

	params := d.Get("params").(string)
	updateParams.Params = json.RawMessage([]byte(params))

	actionsInterface := d.Get("actions").([]interface{})
	actionsList := make([]map[string]interface{}, 0, len(actionsInterface))
	for _, action := range actionsInterface {
		actionsList = append(actionsList, action.(map[string]interface{}))
	}

	var err error
	updateParams.Actions, err = deflateActions(actionsList)
	if err != nil {
		return diag.FromErr(err)
	}

	client := meta.(*kibana.Client)

	alertRule, err := client.API.KibanaAlertRule.Update(id, updateParams)
	if err != nil {
		return diag.FromErr(err)
	}

	if d.HasChange("enabled") {
		if enabled := d.Get("enabled").(bool); enabled {
			err = client.API.KibanaAlertRule.Enable(id)
		} else {
			err = client.API.KibanaAlertRule.Disable(id)
		}
		if err != nil {
			return diag.FromErr(err)
		}
	}

	log.Infof("Updated alert rule %s (%s) successfully", alertRule.ID, alertRule.Name)
	fmt.Printf("[INFO] Updated alert rule %s (%s) successfully", alertRule.ID, alertRule.Name)

	return resourceKibanaAlertRuleRead(ctx, d, meta)
}

// Delete existing alert rule in Kibana
func resourceKibanaAlertRuleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	id := d.Id()
	log.Debugf("AlertRule ID: %s", id)

	client := meta.(*kibana.Client)

	err := client.API.KibanaAlertRule.Delete(id)
	if err != nil {
		if err.(kbapi.APIError).Code == 404 {
			log.Warnf("AlertRule %s not found - removing from state", id)
			fmt.Printf("[WARN] AlertRule %s not found - removing from state", id)
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	d.SetId("")

	log.Infof("Deleted alert rule %s successfully", id)
	fmt.Printf("[INFO] Deleted alert rule %s successfully", id)
	return nil

}

func deflateActions(actionArray []map[string]interface{}) ([]kbapi.KibanaAlertRuleAction, error) {
	actions := []kbapi.KibanaAlertRuleAction{}
	for _, flatAction := range actionArray {
		var action kbapi.KibanaAlertRuleAction
		id := flatAction["id"].(string)
		action.Id = id
		group := flatAction["group"].(string)
		action.Group = group
		params := flatAction["params"].(string)
		action.Params = json.RawMessage([]byte(params))
		actions = append(actions, action)
	}
	return actions, nil
}

func flattenActions(actions []kbapi.KibanaAlertRuleAction) ([]map[string]interface{}, error) {
	res := make([]map[string]interface{}, 0, len(actions))
	for _, a := range actions {
		action := make(map[string]interface{})
		action["id"] = a.Id
		action["group"] = a.Group
		paramsBytes, err := json.Marshal(a.Params)
		if err != nil {
			return nil, errors.Wrapf(err, "Failed to marshal Action")
		}
		action["params"] = string(paramsBytes)
		res = append(res, action)
	}
	return res, nil
}

func rawJsonEqual(k, oldValue, newValue string, d *schema.ResourceData) bool {
	var oldInterface, newInterface interface{}
	if err := json.Unmarshal([]byte(oldValue), &oldInterface); err != nil {
		return false
	}
	if err := json.Unmarshal([]byte(newValue), &newInterface); err != nil {
		return false
	}
	return reflect.DeepEqual(oldInterface, newInterface)
}
