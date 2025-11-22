package models

import (
	"errors"
	"reflect"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
)

// determineKey returns the DynamoDB key name for a struct field.
func determineKey(sf reflect.StructField) string {
	key := sf.Tag.Get("dynamodbav")
	if key != "" {
		return key
	}
	key = sf.Tag.Get("json")
	if idx := strings.Index(key, ","); idx != -1 {
		key = key[:idx]
	}
	if key == "" || key == "-" {
		return sf.Name
	}
	return key
}

// valueToInterface converts a non-pointer reflect.Value into an interface{} suitable for marshalling.
// It assumes fv is not a pointer. Special-cases time.Time -> RFC3339 string.
func valueToInterface(fv reflect.Value) interface{} {
	if fv.Type() == reflect.TypeOf(time.Time{}) {
		return fv.Interface().(time.Time).Format(time.RFC3339)
	}
	return fv.Interface()
}

// ToDynamoMap converts any struct value into a map[string]types.AttributeValue
// suitable for DynamoDB PutItem/Update operations.
func ToDynamoMap[T any](obj *T) (map[string]types.AttributeValue, error) {
	if obj == nil {
		return nil, errors.New("nil input")
	}
	v := reflect.ValueOf(obj)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return nil, errors.New("input must be a struct or pointer to struct")
	}

	t := v.Type()
	m := make(map[string]interface{})
	for i := 0; i < t.NumField(); i++ {
		sf := t.Field(i)
		// skip unexported fields
		if sf.PkgPath != "" {
			continue
		}
		key := determineKey(sf)
		fv := v.Field(i)
		// handle pointer fields: skip nil pointers, dereference non-nil
		if fv.Kind() == reflect.Ptr {
			if fv.IsNil() {
				continue
			}
			fv = fv.Elem()
		}
		m[key] = valueToInterface(fv)
	}

	return attributevalue.MarshalMap(m)
}

func GenerateID() string {
	return uuid.New().String()
}
