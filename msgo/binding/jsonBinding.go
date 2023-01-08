package binding

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"
)

type jsonBinding struct {
	DisallowUnknownFields bool
	IsValidate            bool
}

func (b *jsonBinding) Name() string {
	return "json"
}

func (b *jsonBinding) Bind(r *http.Request, v any) error {
	return b.decodeJson(r, v)
}

func (b *jsonBinding) decodeJson(r *http.Request, v any) error {
	body := r.Body

	if body == nil {
		return errors.New(" request body is empty")
	}

	decoder := json.NewDecoder(body)
	if b.DisallowUnknownFields {
		decoder.DisallowUnknownFields()
	}

	if b.IsValidate {
		err := MyValidate(decoder, v)
		if err != nil {
			return err
		}
	} else {
		err := decoder.Decode(v)
		if err != nil {
			return err
		}
	}

	// 抽象为接口，并只做一次实例化
	return validate(v)
}

// MyValidate 自定义的
func MyValidate(decoder *json.Decoder, v any) error {
	if v == nil {
		return errors.New(" v is nil")
	}

	valueOf := reflect.ValueOf(v)
	if valueOf.Kind() != reflect.Pointer {
		return errors.New("is not Pointer type")
	}

	// 获取指针的元素
	elems := valueOf.Elem().Interface()
	of := reflect.ValueOf(elems)

	switch of.Kind() {
	case reflect.Struct:
		return checkJsonPara(decoder, v, of)
	case reflect.Slice, reflect.Array:
		ofElem := of.Type().Elem()
		if ofElem.Kind() == reflect.Struct {
			err := checkJsonParaSlice(decoder, v, ofElem)
			if err != nil {
				return err
			}
		}
	default:
		err := decoder.Decode(v)
		if err != nil {
			return err
		}
	}
	return nil
}

// checkJsonPara 自定义的验证单个josn对象
func checkJsonPara(decoder *json.Decoder, v any, of reflect.Value) error {
	// 先转换成map, 再校验，把map转成json;再把json转换成结构体
	valMap := make(map[string]any)
	if err := decoder.Decode(&valMap); err != nil {
		return err
	}
	for i := 0; i < of.NumField(); i++ {
		field := of.Type().Field(i)
		fieldName := of.Type().Field(i).Name
		jsonTag := field.Tag.Get("json")
		if jsonTag != "" {
			jsonTagArray := strings.Split(jsonTag, ",")
			fieldName = jsonTagArray[0]
		}

		required := field.Tag.Get("msgo")
		fieldVal := valMap[fieldName]
		if fieldVal == nil && required == "required" {
			return errors.New(fieldName + " field is empty ")
		}
	}

	// 如果验证没有问题，再次解析json
	mapJson, _ := json.Marshal(&valMap)
	json.Unmarshal(mapJson, v)
	return nil
}

// checkJsonParaSlice 自定义的验证切片或数组
func checkJsonParaSlice(decoder *json.Decoder, v any, ofElem reflect.Type) error {
	valMap := make([]map[string]any, 0)
	if err := decoder.Decode(&valMap); err != nil {
		return err
	}
	if len(valMap) <= 0 {
		return errors.New("array is empty")
	}

	for i := 0; i < ofElem.NumField(); i++ {
		field := ofElem.Field(i)
		fieldName := field.Name
		jsonTag := field.Tag.Get("json")
		if jsonTag != "" {
			jsonTagArray := strings.Split(jsonTag, ",")
			fieldName = jsonTagArray[0]
		}

		required := field.Tag.Get("msgo")

		for i, val := range valMap {
			fieldVal := val[fieldName]
			if fieldVal == nil && required == "required" {
				return errors.New(fieldName + " field is empty,by indx: " + strconv.Itoa(i))
			}
		}
	}

	// 如果验证没有问题，再次解析json
	mapJson, _ := json.Marshal(&valMap)
	json.Unmarshal(mapJson, v)
	return nil
}

type SliceValidationError []error

func (err SliceValidationError) Error() string {
	n := len(err)
	switch n {
	case 0:
		return ""
	default:
		var b strings.Builder
		if err[0] != nil {
			fmt.Fprintf(&b, "[%d]: %s", 0, err[0].Error())
		}
		if n > 1 {
			for i := 1; i < n; i++ {
				if err[i] != nil {
					b.WriteString("\n")
					fmt.Fprintf(&b, "[%d]: %s", i, err[i].Error())
				}
			}
		}
		return b.String()
	}
}

func validate(v any) error {
	return DefaultValidator.ValidateExtStruct(v)
}
