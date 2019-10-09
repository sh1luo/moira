package redis

import (
	"fmt"
	"github.com/moira-alert/moira"
)

// Compatibility with moira < v2.6.0
// TODO(litleleprikon): Remove this file in moira v2.8.0

const firstTarget = "t1"

// checkDataDidUnmarshal is a function that adds to CheckData metrics
// a Values map.
func checkDataDidUnmarshal(checkData *moira.CheckData) {
	for metricName, metricState := range checkData.Metrics {
		if metricState.Values == nil {
			metricState.Values = make(map[string]float64)
		}
		if metricState.Value != nil {
			metricState.Values[firstTarget] = *metricState.Value
			metricState.Value = nil
		}
		checkData.Metrics[metricName] = metricState
	}
	if checkData.MetricsToTargetRelation == nil {
		checkData.MetricsToTargetRelation = make(map[string]string)
	}
}

// checkDataWillMarshal is a function that fill Value field
// from Values map.
func checkDataWillMarshal(checkData *moira.CheckData) {
	for metricName, metricState := range checkData.Metrics {
		if metricState.Value == nil {
			if value, ok := metricState.Values[firstTarget]; ok {
				metricState.Value = &value
				checkData.Metrics[metricName] = metricState
			}
		}
	}
}

// notificationEventDidUnmarshal is a function that adds to NotificationEvent
// a Values map.
func notificationEventDidUnmarshal(event *moira.NotificationEvent) {
	if event.Values == nil {
		event.Values = make(map[string]float64)
	}
	if event.Value != nil {
		event.Values[firstTarget] = *event.Value
		event.Value = nil
	}
}

// notificationEventWillMarshal is a function that fill Value field
// from Values map.
func notificationEventWillMarshal(event *moira.NotificationEvent) {
	if event.Value == nil {
		if value, ok := event.Values[firstTarget]; ok {
			event.Value = &value
		}
	}
}

// triggerDidUnmarshal is a function that fills AloneMetrics map for trigger.
func triggerDidUnmarshal(trigger *moira.Trigger) {
	if trigger.AloneMetrics == nil {
		aloneMetricsLen := len(trigger.Targets)
		trigger.AloneMetrics = make(map[string]bool, aloneMetricsLen)
		for i := 2; i <= aloneMetricsLen; i++ {
			targetName := fmt.Sprintf("t%d", i)
			trigger.AloneMetrics[targetName] = true
		}
	}
}
