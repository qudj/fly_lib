package dynamodb

import (
	"net/http"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

type Option struct {
	Sts        bool   `form:"sts" json:"sts" xml:"sts" mapstructure:"sts"`
	Table      string `form:"table" json:"table" xml:"table" mapstructure:"table"`
	Region     string `form:"region" json:"region" xml:"region" mapstructure:"region"`
	Endpoint   string `form:"endpoint" json:"endpoint" xml:"endpoint" mapstructure:"endpoint"`
	HTTPClient *http.Client
}

func NewOption() *Option {
	return new(Option)
}

func (o *Option) WithSts(b bool) *Option {
	o.Sts = b
	return o
}

func (o *Option) WithTable(t string) *Option {
	o.Table = t
	return o
}

func (o *Option) WithRegion(r string) *Option {
	o.Region = r
	return o
}

func (o *Option) WithHttpClient(c *http.Client) *Option {
	o.HTTPClient = c
	return o
}

func NewDynamoDB(o *Option) *dynamodb.DynamoDB {
	sessCfg := aws.Config{}
	if o.HTTPClient != nil {
		sessCfg = aws.Config{
			HTTPClient: o.HTTPClient,
		}
	}
	instanceSession := session.Must(session.NewSession(&sessCfg))
	if o.Sts {
		instanceSession = session.Must(session.NewSessionWithOptions(session.Options{
			Profile: "sts",
			Config:  sessCfg,
		}))
	}

	awsConf := aws.NewConfig().WithRegion(o.Region)
	if o.Endpoint != "" {
		awsConf = awsConf.WithEndpoint(o.Endpoint) // for dynamodb local
	}
	return dynamodb.New(instanceSession, awsConf)
}
