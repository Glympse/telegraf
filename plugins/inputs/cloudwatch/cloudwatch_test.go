package cloudwatch

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateStatisticsInputParams(t *testing.T) {
	d := &Dimension{
		Name:  "LoadBalancerName",
		Value: "p-example",
	}

	m := &Metric{
		Tag:        "production_example_latency",
		MetricName: "Latency",
		Namespace:  "AWS/ELB",
		Unit:       "Seconds",
		Stats:      []string{"Average", "Maximum"},
		Dimensions: []*Dimension{d},
	}

	interval, err := time.ParseDuration("1m")
	require.NoError(t, err)
	m.period = interval

	now := time.Now()

	params := m.getStatisticsInput(now)

	assert.EqualValues(t, *params.EndTime, now)
	assert.EqualValues(t, *params.StartTime, now.Add(-m.period))
	assert.Len(t, params.Dimensions, 1)
	assert.Len(t, params.Statistics, 2)
	assert.EqualValues(t, *params.Period, 60)
}

func TestGatherMetricsInvalidConfig(t *testing.T) {
	d := &Dimension{
		Name:  "LoadBalancerName",
		Value: "p-example",
	}

	m := &Metric{
		Tag:        "production_example_latency",
		MetricName: "Latency",
		Namespace:  "AWS/ELB",
		Unit:       "Seconds",
		Stats:      []string{"Average", "Maximum"},
		Dimensions: []*Dimension{d},
	}

	c := &CloudWatch{
		Region:  "us-east-1",
		Period:  "1mins",
		Metrics: []*Metric{m},
	}

	var acc testutil.Accumulator

	err := c.Gather(&acc)
	require.Error(t, err)
}
