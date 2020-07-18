package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/signalsciences/go-sigsci"
	"sort"
)

type providerMetadata struct {
	Corp   string
	Client sigsci.Client
}

func flattenStringArray(entries []string) []interface{} {
	interfaceArray := make([]interface{}, len(entries))
	for i, val := range entries {
		interfaceArray[i] = val
	}
	return interfaceArray
}

func expandStringArray(entries *schema.Set) []string {
	listOfEntries := entries.List()
	strArray := make([]string, len(listOfEntries))
	for i, e := range listOfEntries {
		strArray[i] = e.(string)
	}
	return strArray
}

func flattenDetections(detections []sigsci.Detection) []interface{} {
	var detectionsSet = make([]interface{}, len(detections))
	for i, detection := range detections {
		fieldSet := make([]interface{}, len(detection.Fields))
		for j, field := range detection.Fields {
			fieldMap := map[string]interface{}{
				"name":  field.Name,
				"value": field.Value,
			}
			fieldSet[j] = fieldMap
		}
		detectionMap := map[string]interface{}{
			"id":      detection.ID,
			"name":    detection.Name,
			"enabled": detection.Enabled,
			"fields":  fieldSet,
		}
		detectionsSet[i] = detectionMap
	}
	return detectionsSet
}

func expandDetections(entries *schema.Set) []sigsci.Detection {
	var detections []sigsci.Detection
	for _, e := range entries.List() {
		cast := e.(map[string]interface{})
		fieldsI := cast["fields"].(*schema.Set)
		var fields []sigsci.ConfiguredDetectionField
		for _, v := range fieldsI.List() {
			castV := v.(map[string]interface{})
			fields = append(fields, sigsci.ConfiguredDetectionField{
				Name:  castV["name"].(string),
				Value: castV["value"],
			})
		}

		detections = append(detections, sigsci.Detection{
			DetectionUpdateBody: sigsci.DetectionUpdateBody{
				ID:      cast["id"].(string),
				Name:    cast["name"].(string),
				Enabled: cast["enabled"].(bool),
				Fields:  fields,
			},
		})
	}
	return detections
}

func flattenAlerts(alerts []sigsci.Alert) []interface{} {
	var alertsSet = make([]interface{}, len(alerts))
	for i, alert := range alerts {
		alertsSet[i] = map[string]interface{}{
			"id":                 alert.ID,
			"long_name":          alert.LongName,
			"interval":           alert.Interval,
			"threshold":          alert.Threshold,
			"skip_notifications": alert.SkipNotifications,
			"enabled":            alert.Enabled,
			"action":             alert.Action,
		}
	}
	return alertsSet
}

func getListAdditionsDeletions(existlist, newlist []string) (additions []string, deletions []string) {
	if len(existlist) == 0 {
		return newlist, []string{}
	}
	setExist := make(map[string]string, len(existlist))
	for _, exE := range existlist {
		setExist[exE] = ""
	}
	add := []string{}
	for _, nwE := range newlist {
		if _, ok := setExist[nwE]; !ok {
			add = append(add, nwE)
		}
	}
	setNew := make(map[string]string, len(newlist))
	for _, nwE := range newlist {
		setNew[nwE] = ""
	}
	del := []string{}
	for _, exE := range existlist {
		if _, ok := setNew[exE]; !ok {
			del = append(del, exE)
		}
	}

	return add, del
}

func diffTemplateDetections(orig, new []sigsci.Detection) ([]sigsci.DetectionUpdateBody, []sigsci.DetectionUpdateBody, []sigsci.DetectionUpdateBody) {
	return calcDetectionAdds(orig, new), calcDetectionUpdates(orig, new), calcDetectionDeletes(orig, new)
}

func calcDetectionAdds(old, new []sigsci.Detection) []sigsci.DetectionUpdateBody {
	var detectionAdds []sigsci.DetectionUpdateBody
	for _, newD := range new {
		exists := false
		for _, oldD := range old {
			if newD.Name == oldD.Name {
				exists = true
			}
		}
		if !exists {
			detectionAdds = append(detectionAdds, newD.DetectionUpdateBody)
		}
	}
	return detectionAdds
}

func calcDetectionDeletes(old, new []sigsci.Detection) []sigsci.DetectionUpdateBody {
	var detectionDeletes []sigsci.DetectionUpdateBody
	for _, oldD := range old {
		exists := false
		for _, newD := range new {
			if oldD.Name == newD.Name {
				exists = true
			}
		}
		if !exists {
			detectionDeletes = append(detectionDeletes, oldD.DetectionUpdateBody)
		}
	}
	return detectionDeletes
}

func calcDetectionUpdates(old, new []sigsci.Detection) []sigsci.DetectionUpdateBody {
	var detectionUpdates []sigsci.DetectionUpdateBody
	for _, oldD := range old {
		for _, newD := range new {
			if oldD.Name == newD.Name {
				if oldD.Enabled != newD.Enabled || !detectionFieldsEqual(newD.Fields, oldD.Fields) {
					detectionUpdates = append(detectionUpdates, newD.DetectionUpdateBody)
				}
			}
		}
	}
	return detectionUpdates
}

func detectionFieldsEqual(old, new []sigsci.ConfiguredDetectionField) bool {
	if len(old) != len(new) {
		return false
	}
	sort.Slice(old, func(i, j int) bool {
		return old[i].Name < old[j].Name
	})
	sort.Slice(new, func(i, j int) bool {
		return new[i].Name < new[j].Name
	})
	for i, _ := range old {
		if old[i].Name != new[i].Name {
			return false
		}
		if old[i].Value != new[i].Value {
			return false
		}
	}
	return true
}
