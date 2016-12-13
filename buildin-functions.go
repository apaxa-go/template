package template

import (
	"github.com/apaxa-go/helper/strconvh"
	"reflect"
)

var buildinFuncs map[string]interface{} = map[string]interface{}{
	"len": func(a interface{}) int { return reflect.ValueOf(a).Len() },
	"cap": func(a interface{}) int { return reflect.ValueOf(a).Cap() },
	"not": func(a bool) bool { return !a },
	"isNil": func(a interface{}) bool {
		if a == nil {
			return true
		}
		switch v := reflect.ValueOf(a); v.Kind() {
		case reflect.Chan:
			fallthrough
		case reflect.Func:
			fallthrough
		case reflect.Map:
			fallthrough
		case reflect.Ptr:
			fallthrough
		case reflect.Interface:
			fallthrough
		case reflect.Slice:
			return v.IsNil()
		default:
			return false
		}
	},
	"isZero":        func(a int) bool { return a == 0 },
	"isPositive":    func(a int) bool { return a > 0 },
	"isNegative":    func(a int) bool { return a < 0 },
	"isNotNegative": func(a int) bool { return a >= 0 },
	"isNotPositive": func(a int) bool { return a <= 0 },
	"intToString": func(a interface{}) string {
		switch v := a.(type) {
		case int:
			return strconvh.FormatInt(v)
		case int8:
			return strconvh.FormatInt8(v)
		case int16:
			return strconvh.FormatInt16(v)
		case int32:
			return strconvh.FormatInt32(v)
		case int64:
			return strconvh.FormatInt64(v)
		default:
			panic(reflect.TypeOf(a).String() + " is not signed integer")
		}
	},
	"uintToString": func(a interface{}) string {
		switch v := a.(type) {
		case uint:
			return strconvh.FormatUint(v)
		case uint8:
			return strconvh.FormatUint8(v)
		case uint16:
			return strconvh.FormatUint16(v)
		case uint32:
			return strconvh.FormatUint32(v)
		case uint64:
			return strconvh.FormatUint64(v)
		default:
			panic(reflect.TypeOf(a).String() + " is not unsigned integer")
		}
	},
	"htmlChecked": func(a bool)string{
		if a{
			return " checked=\"\""
		}
		return ""
	},
}

func NewFuncs() map[string]interface{} {
	f := make(map[string]interface{})
	for k, v := range buildinFuncs {
		f[k] = v
	}
	return f
}
