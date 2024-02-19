package dynamodb

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/pkg/errors"
	"reflect"
	"strings"
)

type Table struct {
	Db    *dynamodb.DynamoDB
	Table string
}

func NewDynamoDBTable(o *Option) *Table {
	return &Table{
		Db:    NewDynamoDB(o),
		Table: o.Table,
	}
}

func (db *Table) Query(input *dynamodb.QueryInput) (*dynamodb.QueryOutput, error) {
	input.TableName = aws.String(db.Table)
	return db.Db.Query(input)
}
func (db *Table) QueryWithContext(ctx context.Context, input *dynamodb.QueryInput) (*dynamodb.QueryOutput, error) {
	input.TableName = aws.String(db.Table)
	return db.Db.QueryWithContext(ctx, input)
}
func (db *Table) BatchGetItem(keys []map[string]*dynamodb.AttributeValue) ([]map[string]*dynamodb.AttributeValue, error) {
	input := &dynamodb.BatchGetItemInput{
		RequestItems: map[string]*dynamodb.KeysAndAttributes{
			db.Table: {
				Keys: keys,
			},
		},
	}
	out, err := db.Db.BatchGetItem(input)
	if err != nil {
		return nil, err
	}
	res, _ := out.Responses[db.Table]
	return res, nil
}
func (db *Table) BatchGetItemWithContext(ctx context.Context, keys []map[string]*dynamodb.AttributeValue) ([]map[string]*dynamodb.AttributeValue, error) {
	input := &dynamodb.BatchGetItemInput{
		RequestItems: map[string]*dynamodb.KeysAndAttributes{
			db.Table: {
				Keys: keys,
			},
		},
	}
	out, err := db.Db.BatchGetItemWithContext(ctx, input)
	if err != nil {
		return nil, err
	}
	res, _ := out.Responses[db.Table]
	return res, nil
}

func (db *Table) GetItem(input *dynamodb.GetItemInput) (*dynamodb.GetItemOutput, error) {
	input.TableName = &db.Table
	return db.Db.GetItem(input)
}

func (db *Table) PutItem(item map[string]*dynamodb.AttributeValue) (*dynamodb.PutItemOutput, error) {
	input := &dynamodb.PutItemInput{
		Item:      item,
		TableName: &db.Table,
	}
	return db.Db.PutItem(input)
}

func (db *Table) SaveData(value interface{}) (*dynamodb.PutItemOutput, error) {
	item , err := GetStructAttribute(value)
	if err != nil {
		return nil, err
	}
	if len(item) == 0 {
		return nil, errors.New("no fields add")
	}
	input := &dynamodb.PutItemInput{
		Item:      item,
		TableName: &db.Table,
	}
	ret, err := db.Db.PutItem(input)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func GetStructAttribute(value interface{}) (map[string]*dynamodb.AttributeValue, error) {
	if reflect.TypeOf(value).Kind() != reflect.Struct {
		return nil, errors.New("data not struct")
	}
	item := map[string]*dynamodb.AttributeValue{}
	t := reflect.TypeOf(value)
	v := reflect.ValueOf(value)
	for i := 0; i < t.NumField(); i++ {
		keySet := t.Field(i).Tag.Get("dynamo")
		if keySet == "" {
			continue
		}
		fieldVue := v.Field(i).Interface()
		if IsNil(fieldVue) {
			continue
		}
		keySets := strings.Split(keySet, ",")
		key := keySets[0]
		setVe := GetDyAttributeValue(t.Field(i).Type.String(), fieldVue)
		item[key] = setVe
	}
	if len(item) == 0 {
		return nil, errors.New("no fields add")
	}
	return item, nil
}

func (db *Table) UpdateDyData(value interface{}, delFields []string) (*dynamodb.UpdateItemOutput, error) {
	if reflect.TypeOf(value).Kind() != reflect.Struct {
		return nil, errors.New("data not struct")
	}
	item := map[string]*dynamodb.AttributeValueUpdate{}
	dyKey := map[string]*dynamodb.AttributeValue{}
	delMap := map[string]bool{}
	for _, v := range delFields {
		delStr := "DELETE"
		delMap[v] = true
		item[v] = &dynamodb.AttributeValueUpdate{Action: &delStr}
	}
	t := reflect.TypeOf(value)
	v := reflect.ValueOf(value)
	for i := 0; i < t.NumField(); i++ {
		keySet := t.Field(i).Tag.Get("dynamo")
		if keySet == "" {
			continue
		}
		fieldVue := v.Field(i).Interface()
		if IsNil(fieldVue) {
			continue
		}
		keySets := strings.Split(keySet, ",")
		key := keySets[0]
		setVe := GetDyAttributeValue(t.Field(i).Type.String(), fieldVue)
		prkIndex := strings.Index(keySet, ",prk")
		srkIndex := strings.Index(keySet, ",srk")
		if prkIndex > 0 || srkIndex > 0 {
			dyKey[key] = setVe
		} else {
			item[key] = &dynamodb.AttributeValueUpdate{Value: setVe}
		}
	}
	if len(dyKey) == 0 {
		return nil, errors.New("no update key")
	}
	if len(item) == 0 {
		return nil, errors.New("no update fields")
	}
	input := &dynamodb.UpdateItemInput{
		TableName:        &db.Table,
		Key:              dyKey,
		AttributeUpdates: item,
	}
	ret, err := db.Db.UpdateItem(input)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func GetDyAttributeValue(typeStr string, value interface{}) *dynamodb.AttributeValue {
	var ret *dynamodb.AttributeValue
	switch typeStr {
	case "string":
		ve := value.(string)
		ret = &dynamodb.AttributeValue{S: &ve}
	case "int64", "int":
		ve := fmt.Sprintf("%d", value)
		ret = &dynamodb.AttributeValue{N: &ve}
	case "float64", "float32":
		ve := fmt.Sprintf("%f", value)
		ret = &dynamodb.AttributeValue{N: &ve}
	case "*string":
		ve := value.(*string)
		ret = &dynamodb.AttributeValue{S: ve}
	case "*int64":
		ve := fmt.Sprintf("%d", *value.(*int64))
		ret = &dynamodb.AttributeValue{N: &ve}
	case "*int":
		ve := fmt.Sprintf("%d", *value.(*int))
		ret = &dynamodb.AttributeValue{N: &ve}
	case "*float64":
		ve := fmt.Sprintf("%f", *value.(*float64))
		ret = &dynamodb.AttributeValue{N: &ve}
	case "*float32":
		ve := fmt.Sprintf("%f", *value.(*float32))
		ret = &dynamodb.AttributeValue{N: &ve}
	case "interface {}":
		m, _ := GetStructAttribute(value)
		ret = &dynamodb.AttributeValue{M: m}
	}
	return ret
}

func IsNil(i interface{}) bool {
	vi := reflect.ValueOf(i)
	if vi.Kind() == reflect.Ptr {
		return vi.IsNil()
	}
	return false
}
