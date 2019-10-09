package redis

import (
	. "github.com/glycerine/goconvey/convey"
	"github.com/moira-alert/moira"
	"testing"
)

func Test_checkDataDidUnmarshal(t *testing.T) {
	Convey("checkDataDidUnmarshal", t, func() {
		checkData := &moira.CheckData{}
		Convey("metrics are empty and metrics to target relation is empty", func() {
			checkDataDidUnmarshal(checkData)
			So(checkData, ShouldResemble, &moira.CheckData{
				MetricsToTargetRelation: map[string]string{},
			})
		})
		Convey("metrics are not empty and metrics to target relation is empty", func() {
			checkData.Metrics = map[string]moira.MetricState{
				"metric.test.1": {},
			}
			checkDataDidUnmarshal(checkData)
			So(checkData, ShouldResemble, &moira.CheckData{
				Metrics: map[string]moira.MetricState{
					"metric.test.1": {
						Values: map[string]float64{},
					},
				},
				MetricsToTargetRelation: map[string]string{},
			})
		})
		Convey("metrics are not empty with filled value and metrics to target relation is empty", func() {
			value := float64(10)
			checkData.Metrics = map[string]moira.MetricState{
				"metric.test.1": {Value: &value},
			}
			checkDataDidUnmarshal(checkData)
			So(checkData, ShouldResemble, &moira.CheckData{
				Metrics: map[string]moira.MetricState{
					"metric.test.1": {
						Values: map[string]float64{"t1": 10},
					},
				},
				MetricsToTargetRelation: map[string]string{},
			})
		})
	})
}

func Test_checkDataWillMarshal(t *testing.T) {
	Convey("checkDataWillMarshal", t, func() {
		checkData := &moira.CheckData{}
		Convey("metrics are empty", func() {
			checkDataWillMarshal(checkData)
			So(checkData, ShouldResemble, &moira.CheckData{})
		})
		Convey("metrics are not empty and values is empty", func() {
			checkData.Metrics = map[string]moira.MetricState{
				"metric.test.1": {},
			}
			checkDataWillMarshal(checkData)
			So(checkData, ShouldResemble, &moira.CheckData{
				Metrics: map[string]moira.MetricState{
					"metric.test.1": {},
				},
			})
		})
		Convey("metrics are not empty and values is not empty", func() {
			value := float64(10)
			checkData.Metrics = map[string]moira.MetricState{
				"metric.test.1": {
					Values: map[string]float64{"t1": 10},
				},
			}
			checkDataWillMarshal(checkData)
			So(checkData, ShouldResemble, &moira.CheckData{
				Metrics: map[string]moira.MetricState{
					"metric.test.1": {
						Value:  &value,
						Values: map[string]float64{"t1": 10},
					},
				},
			})
		})
	})
}

func Test_notificationEventDidUnmarshal(t *testing.T) {
	Convey("notificationEventDidUnmarshal", t, func() {
		event := &moira.NotificationEvent{}
		Convey("value is empty", func() {
			notificationEventDidUnmarshal(event)
			So(event, ShouldResemble, &moira.NotificationEvent{
				Values: map[string]float64{},
			})
		})
		Convey("value is not empty", func() {
			value := float64(10)
			event.Value = &value
			notificationEventDidUnmarshal(event)
			So(event, ShouldResemble, &moira.NotificationEvent{
				Values: map[string]float64{
					"t1": 10,
				},
			})
		})
	})
}

func Test_notificationEventWillMarshal(t *testing.T) {
	Convey("notificationEventWillMarshal", t, func() {
		event := &moira.NotificationEvent{Values: map[string]float64{},}
		Convey("values is empty", func() {
			notificationEventWillMarshal(event)
			So(event, ShouldResemble, &moira.NotificationEvent{
				Values: map[string]float64{},
			})
		})
		Convey("values is not empty", func() {
			value := float64(10)
			event.Values = map[string]float64{
				"t1": 10,
			}
			notificationEventWillMarshal(event)
			So(event, ShouldResemble, &moira.NotificationEvent{
				Values: map[string]float64{
					"t1": 10,
				},
				Value: &value,
			})
		})
	})
}

func Test_triggerDidUnmarshal(t *testing.T) {
	Convey("triggerDidUnmarshal", t, func() {
		trigger := &moira.Trigger{}
		Convey("have one target", func() {
			trigger.Targets = []string{"target.test.1"}
			triggerDidUnmarshal(trigger)
			So(trigger, ShouldResemble, &moira.Trigger{
				Targets:      []string{"target.test.1"},
				AloneMetrics: map[string]bool{},
			})
		})
		Convey("have more than one targets", func() {
			trigger.Targets = []string{"target.test.1", "target.test.2"}
			triggerDidUnmarshal(trigger)
			So(trigger, ShouldResemble, &moira.Trigger{
				Targets:      []string{"target.test.1", "target.test.2"},
				AloneMetrics: map[string]bool{"t2": true},
			})
		})
	})
}
