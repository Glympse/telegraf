# Amazon CloudWatch Statistics Input

This plugin will pull Metric Statistics from Amazon CloudWatch.

### Amazon Authentication

This plugin uses a credential chain for Authentication with the CloudWatch
API endpoint. In the following order the plugin will attempt to authenticate.
1. [IAMS Role](http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/iam-roles-for-amazon-ec2.html)
2. [Environment Variables](https://github.com/aws/aws-sdk-go/wiki/configuring-sdk#environment-variables)
3. [Shared Credentials](https://github.com/aws/aws-sdk-go/wiki/configuring-sdk#shared-credentials-file)

### Configuration:

```toml
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
```
#### Requirements and Terminology

Plugin Configuration utilizes [CloudWatch concepts](http://docs.aws.amazon.com/AmazonCloudWatch/latest/DeveloperGuide/cloudwatch_concepts.html) and access pattern to allow monitoring of any CloudWatch Metric.

- `region` must be a valid AWS [Region](http://docs.aws.amazon.com/AmazonCloudWatch/latest/DeveloperGuide/cloudwatch_concepts.html#CloudWatchRegions) value
- `period` must be a valid CloudWatch [Period](http://docs.aws.amazon.com/AmazonCloudWatch/latest/DeveloperGuide/cloudwatch_concepts.html#CloudWatchPeriods) value
- `namespace` must be a valid CloudWatch [Namespace](http://docs.aws.amazon.com/AmazonCloudWatch/latest/DeveloperGuide/cloudwatch_concepts.html#Namespace) value
- `name` must be a valid CloudWatch [Metric](http://docs.aws.amazon.com/AmazonCloudWatch/latest/DeveloperGuide/cloudwatch_concepts.html#Metric) name
- `statistics` must contain valid CloudWatch [Statistic](http://docs.aws.amazon.com/AmazonCloudWatch/latest/DeveloperGuide/cloudwatch_concepts.html#Statistic) values
- `unit` must be a valid CloudWatch [Unit](http://docs.aws.amazon.com/AmazonCloudWatch/latest/DeveloperGuide/cloudwatch_concepts.html#Unit) value
- `dimensions` must be valid CloudWatch [Dimension](http://docs.aws.amazon.com/AmazonCloudWatch/latest/DeveloperGuide/cloudwatch_concepts.html#Dimension) name/value pairs

### Measurements & Fields:

There is a single measurement from this plugin with dynamic Tags to represent the varying data provided by CloudWatch for the Metric you are monitoring.

- cloudwatch

### Tags:

- All measurements have the following tags:
  - name        (CloudWatch Metric Name)
  - tag         (User-provided friendly tag)
  - unit        (CloudWatch Metric Unit)
  - Sum         (value - if 'Sum' Statistic requested)
  - Average     (value - if 'Average' Statistic requested)
  - Minimum     (value - if 'Minimum' Statistic requested)
  - Maximum     (value - if 'Maximum' Statistic requested)
  - SampleCount (value - if 'SampleCount' Statistic requested)

### Example Output:

```
$ ./telegraf -config telegraf.conf -input-filter cloudwatch -test
cloudwatch,name=RequestCount,tag=p-example-api-requests,unit=Count Sum=1041 1458939540000000000
```
