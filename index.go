package template

import (
	"reflect"
	"errors"
	"github.com/apaxa-go/helper/strconvh"
)

type Index int

func (i Index)Apply(a reflect.Value) (reflect.Value, error){
	if kind:=a.Kind(); kind!=reflect.Array && kind!=reflect.Slice && kind!=reflect.String{
		return reflect.Value{}, errors.New("access by index allowed only for arrays, slices and strings, but got "+kind.String())
	}
	l:=a.Len()
	if int(i)>=l {
		return reflect.Value{}, errors.New("index "+strconvh.FormatInt(int(i))+" out of range ("+strconvh.FormatInt(l)+")")
	}
	return a.Index(int(i)), nil
}