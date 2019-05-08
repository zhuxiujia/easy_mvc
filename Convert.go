package easy_mvc

import (
	"net/http"
	"reflect"
	"strconv"
	"time"
)

func convert(value string, tItemTypeFieldType reflect.Type, w http.ResponseWriter, r *http.Request) (reflect.Value, error) {
	if tItemTypeFieldType.Kind() == reflect.String {
		return reflect.ValueOf(value), nil
	} else {
		if tItemTypeFieldType.Kind() == reflect.Bool {
			newValue, e := strconv.ParseBool(value)
			if e != nil {
				return reflect.Value{}, e
			}
			return reflect.ValueOf(newValue), nil
		} else if tItemTypeFieldType.Kind() == reflect.Int || tItemTypeFieldType.Kind() == reflect.Int16 || tItemTypeFieldType.Kind() == reflect.Int32 || tItemTypeFieldType.Kind() == reflect.Int64 {
			newValue, e := strconv.ParseInt(value, 0, 64)
			if e != nil {
				return reflect.Value{}, e
			}
			if tItemTypeFieldType.Kind() == reflect.Int {
				return reflect.ValueOf(int(newValue)), nil
			}
			if tItemTypeFieldType.Kind() == reflect.Int16 {
				return reflect.ValueOf(int16(newValue)), nil
			}
			if tItemTypeFieldType.Kind() == reflect.Int32 {
				return reflect.ValueOf(int32(newValue)), nil
			}
			return reflect.ValueOf(newValue), nil
		} else if tItemTypeFieldType.Kind() == reflect.Uint || tItemTypeFieldType.Kind() == reflect.Uint8 || tItemTypeFieldType.Kind() == reflect.Uint16 || tItemTypeFieldType.Kind() == reflect.Uint32 || tItemTypeFieldType.Kind() == reflect.Uint64 {
			newValue, e := strconv.ParseUint(value, 0, 64)
			if e != nil {
				return reflect.Value{}, e
			}
			if tItemTypeFieldType.Kind() == reflect.Uint {
				return reflect.ValueOf(uint(newValue)), nil
			}
			if tItemTypeFieldType.Kind() == reflect.Uint16 {
				return reflect.ValueOf(uint16(newValue)), nil
			}
			if tItemTypeFieldType.Kind() == reflect.Uint32 {
				return reflect.ValueOf(uint32(newValue)), nil
			}
			return reflect.ValueOf(newValue), nil
		} else if tItemTypeFieldType.Kind() == reflect.Float32 || tItemTypeFieldType.Kind() == reflect.Float64 {
			newValue, e := strconv.ParseFloat(value, 64)
			if e != nil {
				return reflect.Value{}, e
			}
			if tItemTypeFieldType.Kind() == reflect.Float32 {
				return reflect.ValueOf(float32(newValue)), nil
			}
			return reflect.ValueOf(newValue), nil
		} else if tItemTypeFieldType.Kind() == reflect.Struct {
			if tItemTypeFieldType.String() == "time.Time" {
				newValue, e := time.Parse(string(time.RFC3339), value)
				if e != nil {
					return reflect.Value{}, e
				}
				return reflect.ValueOf(newValue), nil
			} else {

			}
		} else if tItemTypeFieldType.Kind() == reflect.Interface {
			if tItemTypeFieldType.String() == "http.ResponseWriter" {
				return reflect.ValueOf(w), nil
			}
		} else if tItemTypeFieldType.Kind() == reflect.Ptr {
			if tItemTypeFieldType.String() == "*http.Request" {
				return reflect.ValueOf(r), nil
			}
			if value == "" {
				return reflect.Zero(tItemTypeFieldType), nil
			}
			var v, e = convert(value, tItemTypeFieldType.Elem(), w, r)
			var newPtrV = reflect.New(tItemTypeFieldType.Elem())
			if v.IsValid() {
				newPtrV.Elem().Set(v)
			}
			return newPtrV, e
		}
	}

	return reflect.Value{}, nil
}
