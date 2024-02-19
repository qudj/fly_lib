package dynamodb

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/shopspring/decimal"
)

type Decimal struct {
	decimal.Decimal
}

func (m *Decimal) Load(d decimal.Decimal) {
	m.Decimal = d
}

func (m *Decimal) Origin() decimal.Decimal {
	return m.Decimal
}

func (m *Decimal) UnmarshalDynamoDBAttributeValue(value *dynamodb.AttributeValue) error {
	var (
		d   decimal.Decimal
		err error
	)
	switch {
	case value.NULL != nil && *value.NULL:
	case value.N != nil && len(*value.N) > 0:
		d, err = decimal.NewFromString(*value.N)
	case value.S != nil && len(*value.S) > 0:
		d, err = decimal.NewFromString(*value.S)
	default:
		err = fmt.Errorf("decimal unmarshal ddb only support N & S\n%v", value.String())
	}
	if err != nil {
		return err
	}
	m.Decimal = d
	return nil
}

func (m *Decimal) MarshalDynamoDBAttributeValue(value *dynamodb.AttributeValue) error {
	value.N = aws.String(m.Decimal.String())
	return nil
}
