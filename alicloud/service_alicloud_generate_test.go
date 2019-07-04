package alicloud

import (
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
