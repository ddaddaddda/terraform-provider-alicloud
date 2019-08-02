package generate

import (
	"github.com/hashicorp/terraform/helper/schema"
	"reflect"
)

func getSchemaDefaultValue(key string, sch *schema.Schema, isChange bool) interface{} {
	switch sch.Type {
	case schema.TypeInt:

		if isChange {
			return "2"
		} else {
			return "1"
		}
	case schema.TypeString:
		if isChange {
			return key + "String_change"
		} else {
			return key + "String"
		}
	case schema.TypeFloat:
		if isChange {
			return "2.1"
		} else {
			return "1.1"
		}
	case schema.TypeBool:

		if isChange {
			return "true"
		} else {
			return "false"
		}

	case schema.TypeList, schema.TypeSet:
		return getResourceDefaultValue(key, sch, isChange)
	case schema.TypeMap:
		if isChange{
			return map[string]string{
				"test1Key": "test1Value_change",
			}
		}
		return map[string]string{
			"test1Key": "test1Value",
		}
	}
	return nil
}

func getResourceDefaultValue(key string, sch *schema.Schema, isChange bool) interface{} {
	elemVal := getRealValueType(reflect.ValueOf(sch.Elem))
	var listOrSet []interface{}
	if isSchemaResource(sch.Elem) {
		resourceElem := elemVal.Interface().(*schema.Resource)
		defaultValueMap := map[string]interface{}{}
		for childKey, childSch := range resourceElem.Schema {
			defaultValueMap[childKey] = getSchemaDefaultValue(childKey, childSch, isChange)
		}
		listOrSet = append(listOrSet,defaultValueMap)
	} else {
		listOrSet = append(listOrSet, getSchemaDefaultValue(key, elemVal.Interface().(*schema.Schema), isChange))
	}
	return listOrSet
}