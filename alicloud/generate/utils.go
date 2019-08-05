package generate

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"reflect"
	"strconv"
	"strings"
)

func getRealValueType(value reflect.Value) reflect.Value {
	switch value.Kind() {
	case reflect.Interface:
		return getRealValueType(reflect.ValueOf(value.Interface()))
	default:
		return value
	}
}

func isSchemaResource(sch interface{})bool{
	elemVal := getRealValueType(reflect.ValueOf(sch))
	return elemVal.Type().String() == "*schema.Resource"
}

func isForceNew(s *schema.Schema) bool {
	if s.ForceNew {
		return true
	}
	if s.Type == schema.TypeList || s.Type == schema.TypeSet {
		elemVal := getRealValueType(reflect.ValueOf(s.Elem))
		if elemVal.Type().String() == "*schema.Resource" {
			resourceElem := elemVal.Interface().(*schema.Resource)
			for _, subSchema := range resourceElem.Schema {
				if isForceNew(subSchema) {
					return true
				}
			}
			return false
		} else {
			subSchema := elemVal.Interface().(*schema.Schema)
			return subSchema.ForceNew
		}
	}
	return false
}

func isRequired(s *schema.Schema) bool {
	return s.Required
}

func isOptional(s *schema.Schema) bool {
	return s.Optional
}

func isComputed(s *schema.Schema) bool {
	return s.Computed
}

func mapInterfaceValueCopy(dst map[string]interface{},src map[string]interface{})map[string]interface{}{
	newMap := make(map[string]interface{},0)
	if dst != nil{
		for key,val := range dst{
			newMap[key]=val
		}
	}
	if src != nil{
		for key,srcVal :=range src{
			if _,ok:=newMap[key];ok{
				continue
			}
			newMap[key]= srcVal
		}
	}
	return newMap
}

func mapStringValueCopy(dst map[string]string,src map[string]string)map[string]string{
	newMap := make(map[string]string,0)
	if dst != nil{
		for key,dstVal :=range dst{
			newMap[key]= dstVal
		}
	}

	if src != nil{
		for key,srcVal :=range src{
			if _,ok:=newMap[key];ok{
				continue
			}
			newMap[key]= srcVal
		}
	}

	return newMap
}

func convertToString(val interface{})string{
	value:=getRealValueType(reflect.ValueOf(val))
	switch value.Kind() {
	case reflect.Int:
		return strconv.FormatInt(value.Int(),10)
	case reflect.Float64:
		return strconv.FormatFloat(value.Float(), 'g', -1, 64)
	case reflect.Bool:
		return strconv.FormatBool(value.Bool())
	case reflect.String:
		return strings.TrimSpace(value.String())
	default:
	}
	panic(fmt.Sprintf("unsupport type %s",value.Type().String()))
}