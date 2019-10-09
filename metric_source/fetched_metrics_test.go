package metricSource

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestNewFetchedPatternMetricsWithCapacity(t *testing.T) {
	Convey("NewFetchedPatternMetricsWithCapacity", t, func() {
		Convey("call", func() {
			capacity := 10
			actual := NewFetchedPatternMetricsWithCapacity(capacity)
			So(actual, ShouldNotBeNil)
			So(actual, ShouldHaveLength, 0)
			So(cap(actual), ShouldEqual, capacity)
		})
	})
}

func TestFetchedPatternMetrics_CleanWildcards(t *testing.T) {
	tests := []struct {
		name string
		m    FetchedPatternMetrics
		want FetchedPatternMetrics
	}{
		{
			name: "does not have wildcards",
			m: FetchedPatternMetrics{
				{Name: "metric.test.1", Wildcard: false},
			},
			want: FetchedPatternMetrics{
				{Name: "metric.test.1", Wildcard: false},
			},
		},
		{
			name: "has wildcards",
			m: FetchedPatternMetrics{
				{Name: "metric.test.1", Wildcard: false},
				{Name: "metric.test.2", Wildcard: true},
			},
			want: FetchedPatternMetrics{
				{Name: "metric.test.1", Wildcard: false},
			},
		},
	}
	Convey("FetchedPatternMetrics", t, func() {
		for _, tt := range tests {
			Convey(tt.name, func() {
				actual := tt.m.CleanWildcards()
				So(actual, ShouldResemble, tt.want)
			})
		}
	})
}

func TestFetchedPatternMetrics_Deduplicate(t *testing.T) {
	tests := []struct {
		name             string
		m                FetchedPatternMetrics
		wantDeduplicated FetchedPatternMetrics
		wantDuplicates   []string
	}{
		{
			name: "does not have duplicates",
			m: FetchedPatternMetrics{
				{Name: "metric.test.1"},
				{Name: "metric.test.2"},
			},
			wantDeduplicated: FetchedPatternMetrics{
				{Name: "metric.test.1"},
				{Name: "metric.test.2"},
			},
			wantDuplicates: nil,
		},
		{
			name: "has duplicates",
			m: FetchedPatternMetrics{
				{Name: "metric.test.1"},
				{Name: "metric.test.1"},
				{Name: "metric.test.2"},
			},
			wantDeduplicated: FetchedPatternMetrics{
				{Name: "metric.test.1"},
				{Name: "metric.test.2"},
			},
			wantDuplicates: []string{"metric.test.1"},
		},
		{
			name: "has multiple duplicates",
			m: FetchedPatternMetrics{
				{Name: "metric.test.1"},
				{Name: "metric.test.1"},
				{Name: "metric.test.1"},
				{Name: "metric.test.2"},
			},
			wantDeduplicated: FetchedPatternMetrics{
				{Name: "metric.test.1"},
				{Name: "metric.test.2"},
			},
			wantDuplicates: []string{"metric.test.1", "metric.test.1"},
		},
	}
	Convey("FetchedPatternMetrics", t, func() {
		for _, tt := range tests {
			Convey(tt.name, func() {
				deduplicated, duplicates := tt.m.Deduplicate()
				So(deduplicated, ShouldResemble, tt.wantDeduplicated)
				So(duplicates, ShouldResemble, tt.wantDuplicates)
			})
		}
	})
}

func TestFetchedPatternMetrics_HasOnlyWildcards(t *testing.T) {
	tests := []struct {
		name string
		m    FetchedPatternMetrics
		want bool
	}{
		{
			name: "does not have wildcards",
			m: FetchedPatternMetrics{
				{Name: "metric.test.1", Wildcard: false},
				{Name: "metric.test.2", Wildcard: false},
			},
			want: false,
		},
		{
			name: "has wildcards",
			m: FetchedPatternMetrics{
				{Name: "metric.test.1", Wildcard: false},
				{Name: "metric.test.2", Wildcard: true},
			},
			want: false,
		},
		{
			name: "has only wildcards",
			m: FetchedPatternMetrics{
				{Name: "metric.test.1", Wildcard: true},
				{Name: "metric.test.2", Wildcard: true},
			},
			want: true,
		},
	}
	Convey("FetchedPatternMetrics", t, func() {
		for _, tt := range tests {
			Convey(tt.name, func() {
				actual := tt.m.HasOnlyWildcards()
				So(actual, ShouldResemble, tt.want)
			})
		}
	})
}

func TestNewFetchedMetricsWithCapacity(t *testing.T) {
	Convey("NewNewFetchedMetricsWithCapacity", t, func() {
		Convey("call", func() {
			capacity := 10
			actual := NewFetchedMetricsWithCapacity(capacity)
			So(actual, ShouldNotBeNil)
			So(actual, ShouldHaveLength, 0)
		})
	})
}

func TestFetchedMetrics_AddMetrics(t *testing.T) {
	Convey("AddMetrics", t, func() {
		m := FetchedMetrics{}
		m.AddMetrics(1, FetchedPatternMetrics{{Name: "metric.test.1"}, {Name: "metric.test.2"}})
		m.AddMetrics(2, FetchedPatternMetrics{{Name: "metric.test.3"}})
		So(m, ShouldResemble, FetchedMetrics{
			"t1": FetchedPatternMetrics{{Name: "metric.test.1"}, {Name: "metric.test.2"}},
			"t2": FetchedPatternMetrics{{Name: "metric.test.3"}},
		})
	})
}

func TestFetchedMetrics_HasOnlyWildcards(t *testing.T) {
	tests := []struct {
		name string
		m    FetchedMetrics
		want bool
	}{
		{
			name: "does not have wildcards",
			m: FetchedMetrics{
				"t1": FetchedPatternMetrics{
					{Name: "metric.test.1", Wildcard: false},
					{Name: "metric.test.2", Wildcard: false},
				},
			},
			want: false,
		},
		{
			name: "one target has wildcards",
			m: FetchedMetrics{
				"t1": FetchedPatternMetrics{
					{Name: "metric.test.1", Wildcard: true},
					{Name: "metric.test.2", Wildcard: true},
				},
				"t2": FetchedPatternMetrics{
					{Name: "metric.test.1", Wildcard: false},
					{Name: "metric.test.2", Wildcard: true},
				},
			},
			want: false,
		},
		{
			name: "has only wildcards",
			m: FetchedMetrics{
				"t1": FetchedPatternMetrics{
					{Name: "metric.test.1", Wildcard: true},
					{Name: "metric.test.2", Wildcard: true},
				},
				"t2": FetchedPatternMetrics{
					{Name: "metric.test.1", Wildcard: true},
					{Name: "metric.test.2", Wildcard: true},
				},
			},
			want: true,
		}}
	Convey("FetchedMetrics", t, func() {
		for _, tt := range tests {
			Convey(tt.name, func() {
				actual := tt.m.HasOnlyWildcards()
				So(actual, ShouldResemble, tt.want)
			})
		}
	})
}
