package generate

import (
	"github.com/hashicorp/terraform/helper/schema"
	"reflect"
	"strconv"
)

type checkNode interface {
	getCheckMap() map[string]string
}

type leafCheckNode struct {
	checkKey   string
	checkValue string
}

func (l *leafCheckNode) getCheckMap() map[string]string {
	return map[string]string{l.checkKey:l.checkValue}
}

type branchCheckNode struct {
	checkKey           string
	childCheckTreeList []checkNode
}

func (b *branchCheckNode) addChildCheckNode(node checkNode){
	b.childCheckTreeList = append(b.childCheckTreeList,node)
}

func (b *branchCheckNode) getCheckMap() map[string]string {
	checkMap := make(map[string]string,0)
	for _, tree := range b.childCheckTreeList {
		for key, val := range tree.getCheckMap() {
			checkMap[b.checkKey+ "." + key] = val
		}
	}
	return checkMap
}


func getCheckNode(key string,sch *schema.Schema,val interface{})checkNode{
	return getBasicCheckNode(key,sch,reflect.ValueOf(val))
}

func getBasicCheckNode(key string,sch *schema.Schema,val reflect.Value) checkNode{
	val = getRealValueType(val)
	switch sch.Type {
	case schema.TypeBool,schema.TypeInt,schema.TypeFloat,schema.TypeString:
		return &leafCheckNode{checkKey:key,checkValue:val.String()}
	case schema.TypeMap:
		return getMapCheckNode(key,val)
	case schema.TypeList:
		return getListCheckNode(key,sch,val)
	case schema.TypeSet:
		return getSetCheckNode(key,sch,val)
	}
	panic("invalid set element type")
}

func getMapCheckNode(key string,val reflect.Value)checkNode{
	bNode := &branchCheckNode{checkKey:key}

	lenNode := &leafCheckNode{checkKey:"%",checkValue:strconv.FormatInt(int64(val.Len()),10)}
	bNode.addChildCheckNode(lenNode)

	for _,keyReflectVal:= range val.MapKeys(){
		childValReflectVal := getRealValueType(val.MapIndex(keyReflectVal))
		childKey := keyReflectVal.String()
		childVal := childValReflectVal.String()
		childNode:= &leafCheckNode{checkKey:childKey,checkValue:childVal}
		bNode.addChildCheckNode(childNode)
	}

	return bNode
}

func getListCheckNode(key string,sch *schema.Schema,val reflect.Value)checkNode{
	bNode := &branchCheckNode{checkKey:key}

	lenNode := &leafCheckNode{checkKey:"#",checkValue:strconv.FormatInt(int64(val.Len()),10)}
	bNode.addChildCheckNode(lenNode)

	if isSchemaResource(sch.Elem){
		for i:=0;i<val.Len();i++{
			indexCheckNode := &branchCheckNode{checkKey:strconv.FormatInt(int64(i),10)}
			childReflectVal := getRealValueType(val.Index(i))
			for _,childMapKeyReflectVal := range childReflectVal.MapKeys(){
				childMapKey := getRealValueType(childMapKeyReflectVal).String()
				childMapValReflectVal := childReflectVal.MapIndex(childMapKeyReflectVal)
				childSch := sch.Elem.(*schema.Resource).Schema[childMapKey]
				childCheckNode := getBasicCheckNode(childMapKey,childSch,childMapValReflectVal)
				indexCheckNode.addChildCheckNode(childCheckNode)
			}
			bNode.addChildCheckNode(indexCheckNode)
		}
	} else {
		for i:=0;i<val.Len();i++{
			indexCheckNode := &leafCheckNode{
				checkKey:strconv.FormatInt(int64(i),10),
				checkValue:getRealValueType(val.Index(i)).String(),
			}
			bNode.addChildCheckNode(indexCheckNode)
		}
	}

	return bNode
}

func getSetCheckNode(key string,sch *schema.Schema,val reflect.Value)checkNode{
	bNode := &branchCheckNode{checkKey:key}

	lenNode := &leafCheckNode{checkKey:"#",checkValue:strconv.FormatInt(int64(val.Len()),10)}
	bNode.addChildCheckNode(lenNode)

	if isSchemaResource(sch.Elem){
		for i:=0;i<val.Len();i++{
			childReflectVal := getRealValueType(val.Index(i))

			hashcode := getHashFunc(sch)(childReflectVal.Interface())
			hashCheckNode := &branchCheckNode{checkKey: strconv.FormatInt(int64(hashcode),10)}

			for _,childMapKeyReflectVal := range childReflectVal.MapKeys(){
				childMapKey := getRealValueType(childMapKeyReflectVal).String()
				childMapValReflectVal := childReflectVal.MapIndex(childMapKeyReflectVal)
				childSch := sch.Elem.(*schema.Resource).Schema[childMapKey]
				childCheckNode := getBasicCheckNode(childMapKey,childSch,childMapValReflectVal)
				hashCheckNode.addChildCheckNode(childCheckNode)
			}
			bNode.addChildCheckNode(hashCheckNode)
		}
	} else {
		for i:=0;i<val.Len();i++{
			childReflectVal := getRealValueType(val.Index(i))

			hashcode := getHashFunc(sch)(childReflectVal.Interface())

			hashCheckNode := &leafCheckNode{
				checkKey:strconv.FormatInt(int64(hashcode),10),
				checkValue:getRealValueType(val.Index(i)).String(),
			}
			bNode.addChildCheckNode(hashCheckNode)
		}
	}

	return bNode
}

func getHashFunc(sch *schema.Schema)schema.SchemaSetFunc{
	if sch.Set == nil {
		// Default set function uses the schema to hash the whole value
		elem := sch.Elem
		switch t := elem.(type) {
		case *schema.Schema:
			return HashSchema(t)
		case *schema.Resource:
			return HashResource(t)
		default:
			panic("invalid set element type")
		}
	} else {
		return sch.Set
	}
}