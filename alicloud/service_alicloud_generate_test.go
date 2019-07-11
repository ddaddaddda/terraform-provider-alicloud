package alicloud

import (
	"bytes"
	"encoding/json"
	"github.com/hashicorp/terraform/helper/schema"
	"log"
	"reflect"
	"sort"
	"testing"
)

type treeNode struct {
	key           string
	childNodeList []*treeNode
}

func newTreeNode(key string) *treeNode {
	return &treeNode{key: key, childNodeList: make([]*treeNode, 0)}
}

func (t *treeNode) addChildNode(childNode *treeNode) {
	t.childNodeList = append(t.childNodeList,childNode)
}

func (t *treeNode) getAllBranches() []string {
	var branches []string
	if len(t.childNodeList) == 0 {
		branches = append(branches, t.key)
	} else {
		for _, child := range t.childNodeList {
			childBranches := child.getAllBranches()
			for _, childName := range childBranches {
				if t.key != "" {
					branches = append(branches, t.key+"."+childName)
				} else {
					branches = append(branches, childName)
				}
			}
		}
	}
	return branches
}

func generateNodeBySchema(key string, s *schema.Schema) *treeNode {
	node := newTreeNode(key)
	switch s.Type {
	case schema.TypeInt, schema.TypeString, schema.TypeFloat, schema.TypeBool:

	case schema.TypeList:
		lenNode := newTreeNode("#")
		node.addChildNode(lenNode)
		indexNode := newTreeNode("index")
		node.addChildNode(indexNode)
		addSubNodes(s.Elem, indexNode)
	case schema.TypeSet:
		lenNode := newTreeNode("#")
		node.addChildNode(lenNode)
		hashcodeNode := newTreeNode("hashcode")
		node.addChildNode(hashcodeNode)
		addSubNodes(s.Elem, hashcodeNode)
	case schema.TypeMap:
		lenNode := newTreeNode("%")
		node.addChildNode(lenNode)
		keyNode := newTreeNode("checkKey")
		node.addChildNode(keyNode)
	}
	return node
}

func addSubNodes(elem interface{}, parentNode *treeNode) {
	elemVal := getRealValueType(reflect.ValueOf(elem))
	switch elemVal.Type().String() {
	case "*schema.Schema":
	case "*schema.Resource":
		resourceElem := elemVal.Interface().(*schema.Resource)
		for _, subNode := range generateNodeByResource(resourceElem) {
			parentNode.addChildNode(subNode)
		}
	default:
		log.Panicf("unsupported type %s", reflect.TypeOf(elem).String())
	}
}

func generateNodeByResource(resource *schema.Resource) []*treeNode {
	var nodeList []*treeNode
	for childKey, childScheme := range resource.Schema {
		node := generateNodeBySchema(childKey, childScheme)
		nodeList = append(nodeList, node)
	}
	return nodeList
}

func TestGenerateNodeByResource(t *testing.T) {
	resource := resourceAliyunInstance()
	treeList := generateNodeByResource(resource)
	for _, tree := range treeList {
		bs, _ := json.Marshal(tree.getAllBranches())
		println(string(bs))
	}
}

type schemaClassify struct {
	resourceName string
	forceNewSchema map[string]*schema.Schema
	requiredSchema map[string]*schema.Schema
	optionalSchema map[string]*schema.Schema
	computedSchema map[string]*schema.Schema
}

func initSchemaClassify(resourceName string)*schemaClassify{
	provider := Provider().(*schema.Provider)
	resource := provider.ResourcesMap[resourceName]
	sc := &schemaClassify{}
	sc.resourceName = resourceName
	sc.forceNewSchema = map[string]*schema.Schema{}
	sc.requiredSchema = map[string]*schema.Schema{}
	sc.optionalSchema = map[string]*schema.Schema{}
	sc.computedSchema = map[string]*schema.Schema{}

	for key,sch := range resource.Schema{
		if isForceNew(sch){
			sc.forceNewSchema[key] = sch
		} else if isRequired(sch){
			sc.requiredSchema[key] = sch
		} else if isOptional(sch){
			sc.optionalSchema[key] = sch
		}

		if isComputed(sch){
			sc.computedSchema[key] = sch
		}
	}
	return sc
}

func(sc *schemaClassify)getStep0Config()string{
	buf :=bytes.NewBufferString(addIndentation(12))
	buf.WriteString("{\n")
	buf.WriteString(addIndentation(16))
	buf.WriteString("Config: testAccConfig(map[string]interface{}{\n")
	iterateFunc := buildIterateFunc(20,buf)
	sc.iterateRequired(iterateFunc)
	sc.iterateForceNew(iterateFunc)
	buf.WriteString(addIndentation(16))
	buf.WriteString("}\n")
	buf.WriteString(addIndentation(12))
	buf.WriteString("}\n")
	return buf.String()
}

func TestGetStep0Config(t *testing.T){
	config := initSchemaClassify("alicloud_instance").getStep0Config()
	println(config)
}


func(sc *schemaClassify)iterateForceNew(iterateFunc func(key string,sch *schema.Schema)){
	for key,sch :=range sc.forceNewSchema{
		iterateFunc(key,sch)
	}
}

func(sc *schemaClassify)iterateRequired(iterateFunc func(key string,sch *schema.Schema)){
	for key,sch :=range sc.requiredSchema{
		iterateFunc(key,sch)
	}
}

func(sc *schemaClassify)iterateOptional(iterateFunc func(key string,sch *schema.Schema)){
	for key,sch :=range sc.optionalSchema{
		iterateFunc(key,sch)
	}
}

func(sc *schemaClassify)iterateComputed(iterateFunc func(key string,sch *schema.Schema)){
	for key,sch :=range sc.computedSchema{
		iterateFunc(key,sch)
	}
}


func buildIterateFunc(indentation int,buf *bytes.Buffer) func(string,*schema.Schema){
	return func(key string,sch *schema.Schema){
		iterateSingleFunc(buf,indentation,key,sch)
	}
}


func iterateSingleFunc(buf *bytes.Buffer,indentation int,key string,sch *schema.Schema){
	buf.WriteString(addIndentation(indentation))
	buf.WriteString("\"")
	buf.WriteString(key)
	buf.WriteString("\"")
	buf.WriteString(" : ")
	switch sch.Type {
	case schema.TypeInt,schema.TypeString,schema.TypeFloat,schema.TypeBool:
		buf.WriteString("\"")
		buf.WriteString(getValue(key,schema.TypeInt))
		buf.WriteString("\"")
	case schema.TypeSet,schema.TypeList:
		buf.WriteString("[]interface{}{")
		iterateListFunc(buf,indentation + CHILDINDEND,key,sch)
		buf.WriteString("}")
	case schema.TypeMap:
		buf.WriteString("map[string]string{/n")
		buf.WriteString(addIndentation(indentation + CHILDINDEND))
		buf.WriteString(key)
		buf.WriteString(" = \"")
		buf.WriteString(getValue(key,schema.TypeString))
		buf.WriteString("\"\n")
		buf.WriteString(addIndentation(indentation))
		buf.WriteString("}")
	}
	buf.WriteString(",\n")
}

func iterateListFunc(buf *bytes.Buffer,indentation int,parentKey string,sch *schema.Schema){
	elemVal := getRealValueType(reflect.ValueOf(sch.Elem))
	if elemVal.Type().String() == "*schema.Resource" {
		buf.WriteString("\n")
		resourceElem := elemVal.Interface().(*schema.Resource)
		buf.WriteString(addIndentation(indentation))
		buf.WriteString("map[string]interface{}{\n")
		for key,sch := range resourceElem.Schema {
			iterateSingleFunc(buf,indentation + CHILDINDEND,key,sch)
		}
		buf.WriteString(addIndentation(indentation))
		buf.WriteString("},\n")
		buf.WriteString(addIndentation(indentation - CHILDINDEND))
	} else {
		buf.WriteString("\"")
		buf.WriteString(getValue(parentKey,sch.Type))
		buf.WriteString("\"")
	}
}

func getValue(key string,valueType schema.ValueType)string{
	return key + "Value"
}





/*func TestAccAlicloudDnsRecordBasic(t *testing.T) {

	resourceDnsRecordConfigDependence := func(name string) string {
		return fmt.Sprintf(`
resource "alicloud_dns" "default" {
  name = "%s"
}
`, name)
	}

	var basicMap = map[string]string{
		"host_record": "",
		"type":        "",
		"ttl":         "",
		"priority":    "",
		"checkValue":       "",
		"routing":     "",
		"status":      "",
		"locked":      "",
	}

	var v *alidns.DescribeDomainRecordInfoResponse

	resourceId := ""
	ra := resourceAttrInit(resourceId, basicMap)

	serviceFunc := func() interface{} {
		return &DnsService{testAccProvider.Meta().(*connectivity.AliyunClient)}
	}
	rc := resourceCheckInit(resourceId, &v, serviceFunc)

	rac := resourceAttrCheckInit(rc, ra)

	testAccCheck := rac.resourceAttrMapUpdateSet()
	rand := acctest.RandInt()
	name := "name"
	testAccConfig := resourceTestAccConfigFunc(resourceId, name, resourceDnsRecordConfigDependence)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		// module name
		IDRefreshName: resourceId,
		Providers:     testAccProviders,
		CheckDestroy:  rac.checkResourceDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testAccConfig(map[string]interface{}{
					"name":        "",
					"host_record": "",
					"type":        "",
					"checkValue":       "",
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheck(map[string]string{
						"name":  fmt.Sprintf("", defaultRegionToTest, rand),
						"checkValue": "",
					}),
				),
			},
			{
				ResourceName:      resourceId,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccConfig(map[string]interface{}{
					"host_record": "",
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheck(map[string]string{"host_record": ""}),
				),
			},
			{
				Config: testAccConfig(map[string]interface{}{
					"type":     "",
					"priority": "",
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheck(map[string]string{
						"type":     "",
						"priority": "",
					}),
				),
			},

			{
				Config: testAccConfig(map[string]interface{}{
					"priority": "",
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheck(map[string]string{"priority": ""}),
				),
			},
			{
				Config: testAccConfig(map[string]interface{}{
					"checkValue": "",
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheck(map[string]string{"checkValue": ""}),
				),
			},
			{
				Config: testAccConfig(map[string]interface{}{
					"ttl": "",
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheck(map[string]string{"ttl": ""}),
				),
			},
			{
				Config: testAccConfig(map[string]interface{}{
					"routing": "",
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheck(map[string]string{"routing": ""}),
				),
			},
			{
				Config: testAccConfig(map[string]interface{}{
					"ttl": "",
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheck(map[string]string{"ttl": ""}),
				),
			},
			{
				Config: testAccConfig(map[string]interface{}{
					"name":        "",
					"host_record": "",
					"type":        "",
					"checkValue":       "",
					"ttl":         "",
					"priority":    "",
					"routing":     "",
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheck(basicMap),
				),
			},
		},
	})

}*/

type DependResource struct{
	resourceName string
	configs []string
	dependOn []string
}

type attributeValue struct {
	sch *schema.Schema
	value interface{}
}

func initGenerator(resourceName string)*generator{
	provider := Provider().(*schema.Provider)
	resource := provider.ResourcesMap[resourceName]
	sc := &generator{}
	sc.resourceName = resourceName
	sc.forceNewSchema = map[string]*schema.Schema{}
	sc.requiredSchema = map[string]*schema.Schema{}
	sc.optionalSchema = map[string]*schema.Schema{}
	sc.computedSchema = map[string]*schema.Schema{}
	sc.attributeMap = map[string]*attributeValue{}

	for key,sch := range resource.Schema{
		if isForceNew(sch){
			sc.forceNewSchema[key] = sch
		} else if isRequired(sch){
			sc.requiredSchema[key] = sch
		} else if isOptional(sch){
			sc.optionalSchema[key] = sch
		}

		if isComputed(sch){
			sc.computedSchema[key] = sch
		}
	}
	return sc
}

type generator struct {
	resourceName string
	forceNewSchema map[string]*schema.Schema
	requiredSchema map[string]*schema.Schema
	optionalSchema map[string]*schema.Schema
	computedSchema map[string]*schema.Schema
	dependResourceList []DependResource
	attributeMap map[string]*attributeValue
}

func iterateSchemaMap(schMap map[string]*schema.Schema, config map[string]interface{},isChange bool,
	iterateFunc func(string,*schema.Schema,map[string]interface{},bool)){
	var schKeys []string
	for key:=range schMap{
		schKeys = append(schKeys,key)
	}
	sort.Strings(schKeys)
	for _,key:=range schKeys{
		iterateFunc(key,schMap[key],config,isChange)
	}
}

func(g *generator)getDependResource(resourceName string)(DependResource,bool){
	for _,dependResource:=range g.dependResourceList{
		if dependResource.resourceName == resourceName{
			return dependResource,true
		}
	}
	return DependResource{},false
}

func(g *generator)addDependResource(dependResource ...DependResource)(DependResource,bool){
	g.dependResourceList = append(g.dependResourceList,dependResource...)
	return DependResource{},false
}

func(g *generator)getStep0(config map[string]interface{},check map[string]string){
	iterateSchemaMap(g.requiredSchema,config,false,g.getAttributeValue)
	iterateSchemaMap(g.forceNewSchema,config,false,g.getAttributeValue)
}

func(g *generator)getAttributeValue(key string,sch *schema.Schema,config map[string]interface{},isChange bool){
	if val ,ok := config[key];ok{
		g.attributeMap[key] = &attributeValue{
			sch:sch,
			value:val,
		}
		return
	}
	if val ,ok := bridgeMap[key];ok{
		g.attributeMap[key] = &attributeValue{
			sch:sch,
			value:val,
		}
		resourceValue :=resourceMap[val.resourceName]
		dr :=DependResource{
			resourceName:resourceValue.resourceName,
			configs :[]string{resourceValue.resourceName},
			dependOn :resourceValue.dependOn,
		}
		g.dependResourceList = append(g.dependResourceList,dr)
		return
	}
	g.attributeMap[key] = &attributeValue{
		sch:sch,
		value:getSchemaDefaultValue(key,sch,false),
	}
}

func getSchemaDefaultValue(key string,sch *schema.Schema,isChange bool)interface{}{
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
		if sch.Default != nil{
			if isChange{
				return !sch.Default.(bool)
			}
			return sch.Default.(bool)
		}

		if isChange {
			return "true"
		} else {
			return "false"
		}

	case schema.TypeList,schema.TypeSet:
		return getResourceDefaultValue(key,sch,isChange)
	case schema.TypeMap:
		return map[string]string{
			"test1Key":"test1Value",
		}
	}
	return nil
}


func getResourceDefaultValue(key string,sch *schema.Schema,isChange bool)interface{}{
	elemVal := getRealValueType(reflect.ValueOf(sch.Elem))
	if elemVal.Type().String() == "*schema.Resource" {
		resourceElem := elemVal.Interface().(*schema.Resource)
		defaultValueMap := map[string]interface{}{}
		for childKey,childSch := range resourceElem.Schema {
			defaultValueMap[childKey] = getSchemaDefaultValue(childKey,childSch,isChange)
		}
		return defaultValueMap
	} else {
		var listOrSet []interface{}
		listOrSet = append(listOrSet,getSchemaDefaultValue(key,elemVal.Interface().(*schema.Schema),isChange))
		return listOrSet
	}
}





func isForceNew(s *schema.Schema)bool{
	if s.ForceNew {
		return true
	}
	if s.Type == schema.TypeList || s.Type == schema.TypeSet {
		elemVal := getRealValueType(reflect.ValueOf(s.Elem))
		if elemVal.Type().String() == "*schema.Resource" {
			resourceElem := elemVal.Interface().(*schema.Resource)
			for _,subSchema :=range resourceElem.Schema{
				if isForceNew(subSchema){
					return true
				}
			}
			return false
		}else {
			subSchema := elemVal.Interface().(*schema.Schema)
			return subSchema.ForceNew
		}
	}
	return false
}

func isRequired(s *schema.Schema)bool{
	return s.Required
}

func isOptional(s *schema.Schema)bool{
	return s.Optional
}

func isComputed(s *schema.Schema)bool{
	return s.Computed
}

type checkPair struct {
	checkKey string
	checkValue string
}


type checkTree interface {
	getCheckPairList()[]*checkPair
}

type leafCheckTree struct {
	checkKey   string
	checkValue string
}

func (l *leafCheckTree)getCheckPairList()[]*checkPair{
	return []*checkPair{{checkKey:l.checkKey,checkValue:l.checkValue}}
}

type branchCheckTree struct {
	key string
	childCheckTreeList []checkTree
}

func(b *branchCheckTree)getCheckPairList()[]*checkPair{
	var list []*checkPair
	for _,tree:=range b.childCheckTreeList{
		for _,pair:=range tree.getCheckPairList(){
			newPair := &checkPair{checkKey: b.key + "." + pair.checkKey ,checkValue:pair.checkValue}
			list = append(list,newPair)
		}
	}
}

