package metricSource

import (
	"math"
	"reflect"
	"testing"

	"github.com/moira-alert/moira"
	. "github.com/smartystreets/goconvey/convey"
)

func Test_newSetHelperFromTriggerPatternMetrics(t *testing.T) {
	type args struct {
		metrics TriggerPatternMetrics
	}
	tests := []struct {
		name string
		args args
		want setHelper
	}{
		{
			name: "is empty",
			args: args{
				metrics: TriggerPatternMetrics{},
			},
			want: setHelper{},
		},
		{
			name: "is not empty",
			args: args{
				metrics: TriggerPatternMetrics{
					"metric.test.1": {Name: "metric.name.1"},
				},
			},
			want: setHelper{"metric.test.1": true},
		},
	}

	Convey("TriggerPatterMetrics", t, func() {
		for _, tt := range tests {
			Convey(tt.name, func() {
				actual := newSetHelperFromTriggerPatternMetrics(tt.args.metrics)
				So(actual, ShouldResemble, tt.want)
			})
		}
	})
}

func Test_setHelper_union(t *testing.T) {
	type args struct {
		other setHelper
	}
	tests := []struct {
		name string
		h    setHelper
		args args
		want setHelper
	}{
		{
			name: "Both empty",
			h:    setHelper{},
			args: args{
				other: setHelper{},
			},
			want: setHelper{},
		},
		{
			name: "Target is empty, other is not empty",
			h:    setHelper{},
			args: args{
				other: setHelper{"metric.test.1": true},
			},
			want: setHelper{"metric.test.1": true},
		},
		{
			name: "Target is not empty, other is empty",
			h:    setHelper{"metric.test.1": true},
			args: args{
				other: setHelper{},
			},
			want: setHelper{"metric.test.1": true},
		},
		{
			name: "Both are not empty",
			h:    setHelper{"metric.test.1": true},
			args: args{
				other: setHelper{"metric.test.2": true},
			},
			want: setHelper{"metric.test.1": true, "metric.test.2": true},
		},
		{
			name: "Both are not empty and have same names",
			h:    setHelper{"metric.test.1": true, "metric.test.2": true},
			args: args{
				other: setHelper{"metric.test.2": true, "metric.test.3": true},
			},
			want: setHelper{"metric.test.1": true, "metric.test.2": true, "metric.test.3": true},
		},
	}
	Convey("union", t, func() {
		for _, tt := range tests {
			Convey(tt.name, func() {
				actual := tt.h.union(tt.args.other)
				So(actual, ShouldResemble, tt.want)
			})
		}
	})
}

func Test_setHelper_diff(t *testing.T) {
	type args struct {
		other setHelper
	}
	tests := []struct {
		name string
		h    setHelper
		args args
		want setHelper
	}{
		{
			name: "both have same elements",
			h:    setHelper{"t1": true, "t2": true},
			args: args{
				other: setHelper{"t1": true, "t2": true},
			},
			want: setHelper{},
		},
		{
			name: "other have additional values",
			h:    setHelper{"t1": true, "t2": true},
			args: args{
				other: setHelper{"t1": true, "t2": true, "t3": true},
			},
			want: setHelper{"t3": true},
		},
		{
			name: "origin have additional values",
			h:    setHelper{"t1": true, "t2": true, "t3": true},
			args: args{
				other: setHelper{"t1": true, "t2": true},
			},
			want: setHelper{},
		},
	}
	Convey("diff", t, func() {
		for _, tt := range tests {
			Convey(tt.name, func() {
				actual := tt.h.diff(tt.args.other)
				So(actual, ShouldResemble, tt.want)
			})
		}
	})
}

func Test_isOneMetricMap(t *testing.T) {
	type args struct {
		metrics map[string]MetricData
	}
	tests := []struct {
		name  string
		args  args
		want  bool
		want1 string
	}{
		{
			name: "is one metric map",
			args: args{
				metrics: map[string]MetricData{
					"metric.test.1": {},
				},
			},
			want:  true,
			want1: "metric.test.1",
		},
		{
			name: "is not one metric map",
			args: args{
				metrics: map[string]MetricData{
					"metric.test.1": {},
					"metric.test.2": {},
				},
			},
			want:  false,
			want1: "",
		},
	}
	Convey("metrics map", t, func() {
		for _, tt := range tests {
			Convey(tt.name, func() {
				ok, metricName := isOneMetricMap(tt.args.metrics)
				So(ok, ShouldResemble, tt.want)
				So(metricName, ShouldResemble, tt.want1)
			})
		}
	})
}

func Test_newTriggerPatternMetricsWithCapacity(t *testing.T) {
	Convey("newTriggerPatternMetricsWithCapacity", t, func() {
		Convey("call", func() {
			capacity := 10
			actual := newTriggerPatternMetricsWithCapacity(capacity)
			So(actual, ShouldNotBeNil)
			So(actual, ShouldHaveLength, 0)
		})
	})
}

func TestNewTriggerPatternMetrics(t *testing.T) {
	Convey("NewTriggerPatternMetrics", t, func() {
		fetched := FetchedPatternMetrics{
			{Name: "metric.test.1"},
			{Name: "metric.test.2"},
		}
		actual := NewTriggerPatternMetrics(fetched)
		So(actual, ShouldHaveLength, 2)
		So(actual["metric.test.1"].Name, ShouldResemble, "metric.test.1")
		So(actual["metric.test.2"].Name, ShouldResemble, "metric.test.2")
	})
}

func TestTriggerPatternMetrics_Populate(t *testing.T) {
	type args struct {
		lastMetrics map[string]bool
		from        int64
		to          int64
	}
	tests := []struct {
		name string
		m    TriggerPatternMetrics
		args args
		want TriggerPatternMetrics
	}{
		{
			name: "origin do not have missing metrics",
			m: TriggerPatternMetrics{
				"metric.test.1": {Name: "metric.test.1", StartTime: 17, StopTime: 67, StepTime: 60, Values: []float64{0}},
				"metric.test.2": {Name: "metric.test.2", StartTime: 17, StopTime: 67, StepTime: 60, Values: []float64{0}},
			},
			args: args{
				lastMetrics: map[string]bool{
					"metric.test.1": true,
					"metric.test.2": true,
				},
				from: 17,
				to:   67,
			},
			want: TriggerPatternMetrics{
				"metric.test.1": {Name: "metric.test.1", StartTime: 17, StopTime: 67, StepTime: 60, Values: []float64{0}},
				"metric.test.2": {Name: "metric.test.2", StartTime: 17, StopTime: 67, StepTime: 60, Values: []float64{0}},
			},
		},
		{
			name: "origin have missing metrics",
			m: TriggerPatternMetrics{
				"metric.test.1": {Name: "metric.test.1", StartTime: 17, StopTime: 67, StepTime: 60, Values: []float64{0}},
			},
			args: args{
				lastMetrics: map[string]bool{
					"metric.test.1": true,
					"metric.test.2": true,
				},
				from: 17,
				to:   67,
			},
			want: TriggerPatternMetrics{
				"metric.test.1": {Name: "metric.test.1", StartTime: 17, StopTime: 67, StepTime: 60, Values: []float64{0}},
				"metric.test.2": {Name: "metric.test.2", StartTime: 17, StopTime: 67, StepTime: 60, Values: []float64{math.NaN()}},
			},
		},
	}
	Convey("Populate", t, func() {
		for _, tt := range tests {
			Convey(tt.name, func() {
				actual := tt.m.Populate(tt.args.lastMetrics, tt.args.from, tt.args.to)
				So(actual, ShouldHaveLength, len(tt.want))
				for metricName, actualMetric := range actual {
					wantMetric, ok := tt.want[metricName]
					So(ok, ShouldBeTrue)
					So(actualMetric.StartTime, ShouldResemble, wantMetric.StartTime)
					So(actualMetric.StopTime, ShouldResemble, wantMetric.StopTime)
					So(actualMetric.StepTime, ShouldResemble, wantMetric.StepTime)
					So(actualMetric.Values, ShouldHaveLength, len(wantMetric.Values))
				}
			})
		}
	})
}
func TestNewTriggerMetricsWithCapacity(t *testing.T) {
	Convey("NewTriggerMetricsWithCapacity", t, func() {
		capacity := 10
		actual := NewTriggerMetricsWithCapacity(capacity)
		So(actual, ShouldNotBeNil)
		So(actual, ShouldHaveLength, 0)
	})
}

func TestTriggerMetrics_Populate(t *testing.T) {
	type args struct {
		lastCheck moira.CheckData
		from      int64
		to        int64
	}
	tests := []struct {
		name string
		m    TriggerMetrics
		args args
		want TriggerMetrics
	}{
		{
			name: "origin do not have missing metrics",
			m: TriggerMetrics{
				"t1": TriggerPatternMetrics{
					"metric.test.1": {Name: "metric.test.1"},
					"metric.test.2": {Name: "metric.test.2"},
				},
			},
			args: args{
				lastCheck: moira.CheckData{
					Metrics: map[string]moira.MetricState{
						"metric.test.1": {Values: map[string]float64{"t1": 0}},
						"metric.test.2": {Values: map[string]float64{"t1": 0}},
					},
				},
				from: 17,
				to:   67,
			},
			want: TriggerMetrics{
				"t1": TriggerPatternMetrics{
					"metric.test.1": {Name: "metric.test.1"},
					"metric.test.2": {Name: "metric.test.2"},
				},
			},
		},
		{
			name: "origin have missing alone metric",
			m: TriggerMetrics{
				"t1": TriggerPatternMetrics{
					"metric.test.1": {Name: "metric.test.1"},
					"metric.test.2": {Name: "metric.test.2"},
				},
			},
			args: args{
				lastCheck: moira.CheckData{
					MetricsToTargetRelation: map[string]string{"t2": "metric.test.3"},
					Metrics: map[string]moira.MetricState{
						"metric.test.1": {Values: map[string]float64{"t1": 0, "t2": 0}},
						"metric.test.2": {Values: map[string]float64{"t1": 0, "t2": 0}},
					},
				},
				from: 17,
				to:   67,
			},
			want: TriggerMetrics{
				"t1": TriggerPatternMetrics{
					"metric.test.1": {Name: "metric.test.1"},
					"metric.test.2": {Name: "metric.test.2"},
				},
				"t2": {
					"metric.test.3": {Name: "metric.test.3", StartTime: 17, StopTime: 67, StepTime: 60, Values: []float64{math.NaN()}},
				},
			},
		},
		{
			name: "origin have missing metrics",
			m: TriggerMetrics{
				"t1": TriggerPatternMetrics{
					"metric.test.1": {Name: "metric.test.1"},
				},
			},
			args: args{
				lastCheck: moira.CheckData{
					Metrics: map[string]moira.MetricState{
						"metric.test.1": {Values: map[string]float64{"t1": 0}},
						"metric.test.2": {Values: map[string]float64{"t1": 0}},
					},
				},
				from: 17,
				to:   67,
			},
			want: TriggerMetrics{
				"t1": TriggerPatternMetrics{
					"metric.test.1": {Name: "metric.test.1"},
					"metric.test.2": {Name: "metric.test.2", StartTime: 17, StopTime: 67, StepTime: 60, Values: []float64{math.NaN()}},
				},
			},
		},
		{
			name: "origin have missing metrics and alone metrics",
			m: TriggerMetrics{
				"t1": TriggerPatternMetrics{
					"metric.test.1": {Name: "metric.test.1"},
				},
				"t2": TriggerPatternMetrics{
					"metric.test.3": {Name: "metric.test.3"},
				},
			},
			args: args{
				lastCheck: moira.CheckData{
					MetricsToTargetRelation: map[string]string{"t2": "metric.test.3"},
					Metrics: map[string]moira.MetricState{
						"metric.test.1": {Values: map[string]float64{"t1": 0, "t2": 0}},
						"metric.test.2": {Values: map[string]float64{"t1": 0, "t2": 0}},
					},
				},
				from: 17,
				to:   67,
			},
			want: TriggerMetrics{
				"t1": TriggerPatternMetrics{
					"metric.test.1": {Name: "metric.test.1"},
					"metric.test.2": {Name: "metric.test.2", StartTime: 17, StopTime: 67, StepTime: 60, Values: []float64{math.NaN()}},
				},
				"t2": TriggerPatternMetrics{
					"metric.test.3": {Name: "metric.test.3"},
				},
			},
		},
		{
			name: "origin have target with missing metrics and alone metrics",
			m: TriggerMetrics{
				"t1": TriggerPatternMetrics{
					"metric.test.1": {Name: "metric.test.1"},
					"metric.test.2": {Name: "metric.test.2"},
					"metric.test.3": {Name: "metric.test.3"},
				},
				"t2": TriggerPatternMetrics{
					"metric.test.4": {Name: "metric.test.4"},
				},
				"t3": TriggerPatternMetrics{
					"metric.test.1": {Name: "metric.test.1"},
					"metric.test.2": {Name: "metric.test.2"},
				},
			},
			args: args{
				lastCheck: moira.CheckData{
					MetricsToTargetRelation: map[string]string{"t2": "metric.test.4"},
					Metrics: map[string]moira.MetricState{
						"metric.test.1": {Values: map[string]float64{"t1": 0, "t2": 0, "t3": 0}},
						"metric.test.2": {Values: map[string]float64{"t1": 0, "t2": 0, "t3": 0}},
						"metric.test.3": {Values: map[string]float64{"t1": 0, "t2": 0, "t3": 0}},
						"metric.test.4": {Values: map[string]float64{"t1": 0, "t2": 0, "t3": 0}},
					},
				},
				from: 17,
				to:   67,
			},
			want: TriggerMetrics{
				"t1": TriggerPatternMetrics{
					"metric.test.1": {Name: "metric.test.1"},
					"metric.test.2": {Name: "metric.test.2"},
					"metric.test.3": {Name: "metric.test.3"},
				},
				"t2": TriggerPatternMetrics{
					"metric.test.4": {Name: "metric.test.4"},
				},
				"t3": TriggerPatternMetrics{
					"metric.test.1": {Name: "metric.test.1"},
					"metric.test.2": {Name: "metric.test.2"},
					"metric.test.3": {Name: "metric.test.3", StartTime: 17, StopTime: 67, StepTime: 60, Values: []float64{math.NaN()}},
				},
			},
		},
	}
	Convey("Populate", t, func() {
		for _, tt := range tests {
			Convey(tt.name, func() {
				actual := tt.m.Populate(tt.args.lastCheck, tt.args.from, tt.args.to)
				So(actual, ShouldHaveLength, len(tt.want))
				for targetName, metrics := range actual {
					wantMetrics, ok := tt.want[targetName]
					So(metrics, ShouldHaveLength, len(wantMetrics))
					So(ok, ShouldBeTrue)
					for metricName, actualMetric := range metrics {
						wantMetric, ok := wantMetrics[metricName]
						So(ok, ShouldBeTrue)
						So(actualMetric.Name, ShouldResemble, wantMetric.Name)
						So(actualMetric.StartTime, ShouldResemble, wantMetric.StartTime)
						So(actualMetric.StopTime, ShouldResemble, wantMetric.StopTime)
						So(actualMetric.StepTime, ShouldResemble, wantMetric.StepTime)
						So(actualMetric.Values, ShouldHaveLength, len(wantMetric.Values))
					}
				}
			})
		}
	})
}

func TestTriggerMetrics_FilterAloneMetrics(t *testing.T) {
	tests := []struct {
		name  string
		m     TriggerMetrics
		want  TriggerMetrics
		want1 MetricsToCheck
	}{
		{
			name: "origin does not have alone metrics",
			m: TriggerMetrics{
				"t1": TriggerPatternMetrics{
					"metric.test.1": {Name: "metric.test.1"},
					"metric.test.2": {Name: "metric.test.2"},
				},
			},
			want: TriggerMetrics{
				"t1": TriggerPatternMetrics{
					"metric.test.1": {Name: "metric.test.1"},
					"metric.test.2": {Name: "metric.test.2"},
				},
			},
			want1: MetricsToCheck{},
		},
		{
			name: "origin has alone metrics",
			m: TriggerMetrics{
				"t1": TriggerPatternMetrics{
					"metric.test.1": {Name: "metric.test.1"},
					"metric.test.2": {Name: "metric.test.2"},
				},
				"t2": TriggerPatternMetrics{
					"metric.test.3": {Name: "metric.test.3"},
				},
			},
			want: TriggerMetrics{
				"t1": TriggerPatternMetrics{
					"metric.test.1": {Name: "metric.test.1"},
					"metric.test.2": {Name: "metric.test.2"},
				},
			},
			want1: MetricsToCheck{"t2": {Name: "metric.test.3"}},
		},
	}
	Convey("FilterAloneMetrics", t, func() {
		for _, tt := range tests {
			Convey(tt.name, func() {
				filtered, alone := tt.m.FilterAloneMetrics()
				So(filtered, ShouldResemble, tt.want)
				So(alone, ShouldResemble, tt.want1)
			})
		}
	})
}

func TestTriggerMetrics_Diff(t *testing.T) {
	tests := []struct {
		name string
		m    TriggerMetrics
		want map[string]map[string]bool
	}{
		{
			name: "all targets have same metrics",
			m: TriggerMetrics{
				"t1": TriggerPatternMetrics{
					"metric.test.1": {Name: "metric.test.1"},
					"metric.test.2": {Name: "metric.test.2"},
					"metric.test.3": {Name: "metric.test.3"},
				},
				"t2": TriggerPatternMetrics{
					"metric.test.1": {Name: "metric.test.1"},
					"metric.test.2": {Name: "metric.test.2"},
					"metric.test.3": {Name: "metric.test.3"},
				},
			},
			want: map[string]map[string]bool{},
		},
		{
			name: "one target have missed metric",
			m: TriggerMetrics{
				"t1": TriggerPatternMetrics{
					"metric.test.1": {Name: "metric.test.1"},
					"metric.test.2": {Name: "metric.test.2"},
					"metric.test.3": {Name: "metric.test.3"},
				},
				"t2": TriggerPatternMetrics{
					"metric.test.1": {Name: "metric.test.1"},
					"metric.test.2": {Name: "metric.test.2"},
				},
			},
			want: map[string]map[string]bool{"t2": {"metric.test.3": true}},
		},
		{
			name: "one target is alone metric",
			m: TriggerMetrics{
				"t1": TriggerPatternMetrics{
					"metric.test.1": {Name: "metric.test.1"},
					"metric.test.2": {Name: "metric.test.2"},
					"metric.test.3": {Name: "metric.test.3"},
				},
				"t2": TriggerPatternMetrics{
					"metric.test.1": {Name: "metric.test.1"},
				},
			},
			want: map[string]map[string]bool{},
		},
	}
	Convey("Diff", t, func() {
		for _, tt := range tests {
			Convey(tt.name, func() {
				actual := tt.m.Diff()
				So(actual, ShouldResemble, tt.want)
			})
		}
	})
}

func TestTriggerMetrics_multiMetricsTarget(t *testing.T) {
	tests := []struct {
		name  string
		m     TriggerMetrics
		want  string
		want1 setHelper
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := tt.m.multiMetricsTarget()
			if got != tt.want {
				t.Errorf("TriggerMetrics.multiMetricsTarget() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("TriggerMetrics.multiMetricsTarget() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestTriggerMetrics_ConvertForCheck(t *testing.T) {
	tests := []struct {
		name string
		m    TriggerMetrics
		want TriggerMetricsToCheck
	}{
		{
			name: "origin is empty",
			m:    TriggerMetrics{},
			want: TriggerMetricsToCheck{},
		},
		{
			name: "origin have metrics",
			m: TriggerMetrics{
				"t1": TriggerPatternMetrics{
					"metric.test.1": {Name: "metric.test.1"},
					"metric.test.2": {Name: "metric.test.2"},
				},
			},
			want: TriggerMetricsToCheck{
				"metric.test.1": MetricsToCheck{
					"t1": {Name: "metric.test.1"},
				},
				"metric.test.2": MetricsToCheck{
					"t1": {Name: "metric.test.2"},
				},
			},
		},
		{
			name: "origin have metrics and target with empty metrics",
			m: TriggerMetrics{
				"t1": TriggerPatternMetrics{
					"metric.test.1": {Name: "metric.test.1"},
					"metric.test.2": {Name: "metric.test.2"},
				},
				"t2": TriggerPatternMetrics{
					"metric.test.3": {Name: "metric.test.3"},
				},
			},
			want: TriggerMetricsToCheck{
				"metric.test.1": MetricsToCheck{
					"t1": {Name: "metric.test.1"},
					"t2": {Name: "metric.test.3"},
				},
				"metric.test.2": MetricsToCheck{
					"t1": {Name: "metric.test.2"},
					"t2": {Name: "metric.test.3"},
				},
			},
		},
	}
	Convey("ConvertForCheck", t, func() {
		for _, tt := range tests {
			Convey(tt.name, func() {
				actual := tt.m.ConvertForCheck()
				So(actual, ShouldResemble, tt.want)
			})
		}
	})
}

func TestMetricsToCheck_MetricName(t *testing.T) {
	tests := []struct {
		name string
		m    MetricsToCheck
		want string
	}{
		{
			name: "origin is empty",
			m:    MetricsToCheck{},
			want: "",
		},
		{
			name: "origin is not empty and all metrics have same name",
			m: MetricsToCheck{
				"t1": MetricData{Name: "metric.test.1"},
				"t2": MetricData{Name: "metric.test.1"},
			},
			want: "metric.test.1",
		},
		{
			name: "origin is not empty and metrics have different names",
			m: MetricsToCheck{
				"t1": MetricData{Name: "metric.test.2"},
				"t2": MetricData{Name: "metric.test.1"},
			},
			want: "metric.test.1",
		},
	}
	Convey("MetricName", t, func() {
		for _, tt := range tests {
			Convey(tt.name, func() {
				actual := tt.m.MetricName()
				So(actual, ShouldResemble, tt.want)
			})
		}
	})
}

func TestMetricsToCheck_GetRelations(t *testing.T) {
	tests := []struct {
		name string
		m    MetricsToCheck
		want map[string]string
	}{
		{
			name: "origin is empty",
			m:    MetricsToCheck{},
			want: map[string]string{},
		},
		{
			name: "origin is not empty",
			m: MetricsToCheck{
				"t1": MetricData{Name: "metric.test.1"},
				"t2": MetricData{Name: "metric.test.2"},
			},
			want: map[string]string{
				"t1": "metric.test.1",
				"t2": "metric.test.2",
			},
		},
	}
	Convey("GetRelations", t, func() {
		for _, tt := range tests {
			Convey(tt.name, func() {
				actual := tt.m.GetRelations()
				So(actual, ShouldResemble, tt.want)
			})
		}
	})
}

func TestMetricsToCheck_Merge(t *testing.T) {
	type args struct {
		other MetricsToCheck
	}
	tests := []struct {
		name string
		m    MetricsToCheck
		args args
		want MetricsToCheck
	}{
		{
			name: "origin and other are empty",
			m:    MetricsToCheck{},
			args: args{
				other: MetricsToCheck{},
			},
			want: MetricsToCheck{},
		},
		{
			name: "origin is empty and other is not",
			m:    MetricsToCheck{},
			args: args{
				other: MetricsToCheck{"t1": MetricData{Name: "metric.test.1"}},
			},
			want: MetricsToCheck{"t1": MetricData{Name: "metric.test.1"}},
		},
		{
			name: "origin is not empty and other is empty",
			m:    MetricsToCheck{"t1": MetricData{Name: "metric.test.1"}},
			args: args{
				other: MetricsToCheck{},
			},
			want: MetricsToCheck{"t1": MetricData{Name: "metric.test.1"}},
		},
		{
			name: "origin and other have same targets",
			m:    MetricsToCheck{"t1": MetricData{Name: "metric.test.1"}},
			args: args{
				other: MetricsToCheck{"t1": MetricData{Name: "metric.test.2"}},
			},
			want: MetricsToCheck{"t1": MetricData{Name: "metric.test.2"}},
		},
	}
	Convey("Merge", t, func() {
		for _, tt := range tests {
			Convey(tt.name, func() {
				actual := tt.m.Merge(tt.args.other)
				So(actual, ShouldResemble, tt.want)
			})
		}
	})
}
