package alicloud

import "github.com/hashicorp/terraform/helper/schema"


type treeNode struct{
	key        string
	childNodes []*treeNode
}

func(t *treeNode)addChildNode(childNode *treeNode){
	t.childNodes = append(t.childNodes,childNode)
}

func (t *treeNode)getAllBranches()[]string{
	var branches []string
	if len(t.childNodes) == 0{
		branches = append(branches,t.key)
	}else{
		for _,child := range t.childNodes{
			childBranches := child.getAllBranches()
			for _,childName := range childBranches{
				if t.key != ""{
					branches = append(branches,t.key+ "." + childName)
				}else {
					branches = append(branches,childName)
				}
			}
		}
	}
	return branches
}

func generateNodeBySchema(key string, s *schema.Schema,parentNode *treeNode){
	switch s.Type {
	case schema.TypeInt, schema.TypeString, schema.TypeFloat, schema.TypeBool:
		node :=&treeNode{key:key}
		parentNode.addChildNode(node)
	case schema.TypeList:


	case schema.TypeSet:

	case schema.TypeMap:
	}
}

func generateNodeByResource(key string,resource *schema.Resource)*treeNode{
	node :=&treeNode{key:key}
	for childKey,childScheme :=range resource.Schema{
		generateNodeBySchema(childKey,childScheme,node)
	}
	return node
}
