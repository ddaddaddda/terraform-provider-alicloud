package alicloud

import (
	"bytes"
	"encoding/json"
	"github.com/hashicorp/terraform/helper/schema"
	"log"
	"reflect"
	"testing"
)

type treeNode struct {
	key          string
	childNodeMap map[string]*treeNode
}

func newTreeNode(key string) *treeNode {
	return &treeNode{key: key, childNodeMap: make(map[string]*treeNode, 0)}
}

func (t *treeNode) addChildNode(childNode *treeNode) {
	t.childNodeMap[childNode.key] = childNode
}

func (t *treeNode) getAllBranches() []string {
	var branches []string
	if len(t.childNodeMap) == 0 {
		branches = append(branches, t.key)
	} else {
		for _, child := range t.childNodeMap {
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
		keyNode := newTreeNode("key")
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
		"value":       "",
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
					"value":       "",
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheck(map[string]string{
						"name":  fmt.Sprintf("", defaultRegionToTest, rand),
						"value": "",
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
					"value": "",
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheck(map[string]string{"value": ""}),
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
					"value":       "",
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
