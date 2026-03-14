package runtime

import (
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"strings"
)

func validateJSONObject(value map[string]any, field string) error {
	return validateJSONValue(value, field)
}

func cloneJSONObject(value map[string]any) map[string]any {
	if value == nil {
		return nil
	}
	raw, err := json.Marshal(value)
	if err != nil {
		return nil
	}
	var out map[string]any
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil
	}
	return out
}

func validateJSONValue(value any, path string) error {
	if path == "" {
		path = "value"
	}
	if value == nil {
		return nil
	}

	switch typed := value.(type) {
	case string, bool, json.Number:
		return nil
	case float32:
		if math.IsNaN(float64(typed)) || math.IsInf(float64(typed), 0) {
			return fmt.Errorf("%s contains non-finite number", path)
		}
		return nil
	case float64:
		if math.IsNaN(typed) || math.IsInf(typed, 0) {
			return fmt.Errorf("%s contains non-finite number", path)
		}
		return nil
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return nil
	case map[string]any:
		for key, item := range typed {
			if strings.TrimSpace(key) == "" {
				return fmt.Errorf("%s contains empty object key", path)
			}
			if err := validateJSONValue(item, fmt.Sprintf("%s.%s", path, key)); err != nil {
				return err
			}
		}
		return nil
	case []any:
		for index, item := range typed {
			if err := validateJSONValue(item, fmt.Sprintf("%s[%d]", path, index)); err != nil {
				return err
			}
		}
		return nil
	}

	reflected := reflect.ValueOf(value)
	switch reflected.Kind() {
	case reflect.Interface, reflect.Pointer:
		if reflected.IsNil() {
			return nil
		}
		return validateJSONValue(reflected.Elem().Interface(), path)
	case reflect.Map:
		if reflected.Type().Key().Kind() != reflect.String {
			return fmt.Errorf("%s contains non-string map key", path)
		}
		iter := reflected.MapRange()
		for iter.Next() {
			key := iter.Key().String()
			if strings.TrimSpace(key) == "" {
				return fmt.Errorf("%s contains empty object key", path)
			}
			if err := validateJSONValue(iter.Value().Interface(), fmt.Sprintf("%s.%s", path, key)); err != nil {
				return err
			}
		}
		return nil
	case reflect.Slice, reflect.Array:
		for index := 0; index < reflected.Len(); index++ {
			if err := validateJSONValue(reflected.Index(index).Interface(), fmt.Sprintf("%s[%d]", path, index)); err != nil {
				return err
			}
		}
		return nil
	case reflect.String, reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return nil
	case reflect.Float32, reflect.Float64:
		value := reflected.Float()
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return fmt.Errorf("%s contains non-finite number", path)
		}
		return nil
	default:
		return fmt.Errorf("%s contains unsupported JSON value %T", path, value)
	}
}
