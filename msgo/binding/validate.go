package binding

import (
	"github.com/go-playground/validator/v10"
	"reflect"
	"sync"
)

type StructValidator interface {
	ValidateExtStruct(v any) error // 结构体验证，如果错误返回对应的错误信息
}

// DefaultValidator 对外暴露使用的验证器
var DefaultValidator StructValidator = &defaultStructValidator{}

type defaultStructValidator struct {
	one      sync.Once
	validate *validator.Validate
}

func (d *defaultStructValidator) ValidateExtStruct(v any) error {
	if v == nil {
		return nil
	}

	of := reflect.ValueOf(v)
	switch of.Kind() {
	case reflect.Pointer:
		return d.ValidateExtStruct(of.Elem().Interface())
	case reflect.Slice, reflect.Array:
		count := of.Len()
		validateRet := make(SliceValidationError, 0)
		for i := 0; i < count; i++ {
			if err := d.validateStruct(of.Index(i).Interface()); err != nil {
				validateRet = append(validateRet, err)
			}
		}
		if len(validateRet) == 0 {
			return nil
		}
		return validateRet
	case reflect.Struct:
		d.validateStruct(v)
	}
	return nil
}

func (d *defaultStructValidator) validateStruct(obj any) error {
	d.lazyInit()
	return d.validate.Struct(obj)
}

func (d *defaultStructValidator) lazyInit() {
	d.one.Do(func() {
		d.validate = validator.New()
	})
}
