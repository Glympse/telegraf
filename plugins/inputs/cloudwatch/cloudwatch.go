package cloudwatch

import (
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/ec2rolecreds"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type CloudWatch struct {
	Region  string    `toml:"region"`
	Metrics []*Metric `toml:"metrics"`
	Period  string    `toml:"period"`
	client  *cloudwatch.CloudWatch
}

type Metric struct {
	Tag        string       `toml:"tag"`
	MetricName string       `toml:"name"`
	Namespace  string       `toml:"namespace"`
	Stats      []string     `toml:"statistics"`
	Dimensions []*Dimension `toml:"dimensions"`
	Unit       string       `toml:"unit"`
	period     time.Duration
}

type Dimension struct {
	Name  string `toml:"name"`
	Value string `toml:"value"`
}

func (c *CloudWatch) SampleConfig() string {
	return `
[[inputs.cloudwatch]]
## Amazon Region
region = 'us-east-1'

## Requested CloudWatch aggregation Period (must be a multiple of 60s)
## Recomended: use metric 'interval' that is a multiple of 'period'
## to avoid gaps or overlap in pulled data
period = '1m'
interval = '1m'

## Array of Metric Statistics to pull
  [[inputs.cloudwatch.metrics]]
  statistics = ['Average', 'Maximum']
  namespace = 'AWS/ELB'
  name = 'Latency'
  unit = 'Seconds'
  tag = 'p-example-latency'
  
    [[inputs.cloudwatch.metrics.dimensions]]
    name = 'LoadBalancerName'
    value = 'p-example'
	
  [[inputs.cloudwatch.metrics]]
  statistics = ['Sum']
  namespace = 'AWS/ELB'
  name = 'RequestCount'
  tag = 'p-example-req'
  unit = 'Count'
    [[inputs.cloudwatch.metrics.dimensions]]
      name = 'LoadBalancerName'
      value = 'p-example'
	`
}

func (c *CloudWatch) Description() string {
	return "Pull Metric Statistics from Amazon CloudWatch"
}

func (c *CloudWatch) Gather(acc telegraf.Accumulator) error {
	if c.client == nil {
		c.intializeClient()
	}

	metricCount := len(c.Metrics)
	var errChan = make(chan error, metricCount)

	now := time.Now()

	for _, m := range c.Metrics {
		go c.gatherMetric(acc, m, now, errChan)
	}

	for i := 1; i <= metricCount; i++ {
		err := <-errChan
		if err != nil {
			return err
		}
	}
	return nil
}

func init() {
	inputs.Add("cloudwatch", func() telegraf.Input {
		return &CloudWatch{}
	})
}

/*
 * Initialize CloudWatch client
 */
func (c *CloudWatch) intializeClient() {
	config := &aws.Config{
		Region: aws.String(c.Region),
		Credentials: credentials.NewChainCredentials(
			[]credentials.Provider{
				&ec2rolecreds.EC2RoleProvider{Client: ec2metadata.New(session.New())},
				&credentials.EnvProvider{},
				&credentials.SharedCredentialsProvider{},
			}),
	}

	c.client = cloudwatch.New(session.New(config))
}

/*
 * Gather given Metric and emit any error
 */
func (c *CloudWatch) gatherMetric(acc telegraf.Accumulator, metric *Metric, now time.Time, errChan chan error) {
	period, err := time.ParseDuration(c.Period)
	if err != nil {
		errChan <- err
		return
	}
	metric.period = period

	params := metric.getStatisticsInput(now)
	resp, err := c.client.GetMetricStatistics(params)
	if err != nil {
		errChan <- err
		return
	}

	for _, point := range resp.Datapoints {
		tags := map[string]string{
			"name": metric.MetricName,
			"unit": *point.Unit,
		}
		if metric.Tag != "" {
			tags["tag"] = metric.Tag
		}

		// record field for each statistic
		fields := map[string]interface{}{}
		for _, stat := range metric.Stats {
			var v interface{}
			switch stat {
			case cloudwatch.StatisticAverage:
				v = *point.Average
			case cloudwatch.StatisticMaximum:
				v = *point.Maximum
			case cloudwatch.StatisticMinimum:
				v = *point.Minimum
			case cloudwatch.StatisticSampleCount:
				v = *point.SampleCount
			case cloudwatch.StatisticSum:
				v = *point.Sum
			}

			fields[stat] = v
		}

		acc.AddFields("cloudwatch", fields, tags, *point.Timestamp)
	}

	errChan <- nil
}

/*
 * Map Metric to *cloudwatch.GetMetricStatisticsInput for given timeframe
 */
func (m *Metric) getStatisticsInput(end time.Time) *cloudwatch.GetMetricStatisticsInput {
	dimensions := make([]*cloudwatch.Dimension, len(m.Dimensions))
	for i := 0; i < len(m.Dimensions); i++ {
		dimensions[i] = &cloudwatch.Dimension{
			Name:  aws.String(m.Dimensions[i].Name),
			Value: aws.String(m.Dimensions[i].Value),
		}
	}

	stats := make([]*string, len(m.Stats))
	for i := 0; i < len(m.Stats); i++ {
		stats[i] = aws.String(m.Stats[i])
	}

	return &cloudwatch.GetMetricStatisticsInput{
		StartTime:  aws.Time(end.Add(-time.Duration(m.period))),
		EndTime:    aws.Time(end),
		MetricName: aws.String(m.MetricName),
		Namespace:  aws.String(m.Namespace),
		Period:     aws.Int64(int64(m.period.Seconds())),
		Statistics: stats,
		Dimensions: dimensions,
		Unit:       aws.String(m.Unit),
	}
}
