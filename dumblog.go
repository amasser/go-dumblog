package main

import (
  "github.com/aws/aws-sdk-go/service/cloudwatchlogs"
  "github.com/aws/aws-sdk-go/aws"
  "github.com/aws/aws-sdk-go/aws/session"
  "time"
  "fmt"
  "flag"
  "strings"
  "os"
)

type Config struct {
  AwsSession   *session.Session
  LogsService  *cloudwatchlogs.CloudWatchLogs
  LogStream     string
  LogGroup      string
}

func (c *Config) SetupStream() {
  _, err := c.LogsService.CreateLogStream(&cloudwatchlogs.CreateLogStreamInput{
    LogGroupName: &c.LogGroup,
	  LogStreamName: &c.LogStream,
  })

  if err != nil {
    fmt.Println(err)
  }
}

func (c *Config) SetupGroup() {
  _, err := c.LogsService.CreateLogGroup(&cloudwatchlogs.CreateLogGroupInput{
    LogGroupName: &c.LogGroup,
  })

  if err != nil {
    fmt.Println(err)
  }
}

func (c *Config) NextToken() (*string, error){
  var sequence *string

  limit := int64(1)
  resp, e := c.LogsService.DescribeLogStreams(
    &cloudwatchlogs.DescribeLogStreamsInput{
       LogGroupName: &c.LogGroup,
       LogStreamNamePrefix: &c.LogStream,
       Limit: &limit,
    },
  )

  for _, e := range resp.LogStreams{
   sequence = e.UploadSequenceToken
  }

  return sequence, e
}

func NewConfig(region string, group string, stream string) *Config {
  sess := session.Must(session.NewSession(&aws.Config{
    Region: aws.String("us-west-2"),
  }))

  svc := cloudwatchlogs.New(sess)

  return &Config{
    AwsSession: sess,
    LogsService: svc,
    LogGroup: group,
    LogStream: stream,
  }
}

func unixTimeNow() *int64 {
  t := int64(time.Now().UnixNano() / 1000000)
  return &t
}

func main() {
  group  := flag.String("group", "", "Log group to write to")
  stream := flag.String("stream", "", "Log stream to write to")
  region := flag.String("region", "us-west-2", "AWS region")
  flag.Parse()

  if *stream == "" {
    fmt.Println("stream argument is required")
    os.Exit(255)
  }
  if *group == "" {
    fmt.Println("group argument is required")
    os.Exit(255)
  }

  // Every single position argument is just part of the message
  message := strings.Join(flag.Args()[:]," ")

  conf := NewConfig(*region, *group, *stream)
  conf.SetupGroup()
  conf.SetupStream()
  sequence, seqErr := conf.NextToken()
  if seqErr != nil { fmt.Println(seqErr) }

  e := cloudwatchlogs.InputLogEvent{
    Message: &message,
    Timestamp: unixTimeNow(),
  }

  var events []*cloudwatchlogs.InputLogEvent
  events = append(events, &e)

  eventInput := &cloudwatchlogs.PutLogEventsInput{
    LogEvents:     events,
    LogStreamName: &conf.LogStream,
    LogGroupName:  &conf.LogGroup,
  }

  // if we aren't the first event we have to provide the sequence token
  if sequence != nil { eventInput.SetSequenceToken(*sequence) }

  output, err := conf.LogsService.PutLogEvents(eventInput)

  if err != nil {
    fmt.Println(err)
  }

  fmt.Println(output)
}
