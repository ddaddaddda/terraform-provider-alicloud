package generate

import (
	"fmt"
)

var resourceMap = map[string]DependResource{
	"alicloud_dns": {
		resourceName: "alicloud_dns",
		configs: []string{`
			resource "alicloud_dns" "default" {
			    name = "tf-testAccXXXXX"
			    group_id = ${alicloud_dns_group.id}
			}
           `},
		dependOn: []string{"alicloud_dns_group"},
	},
}

var bridgeMap = map[string]bridgeMapValue{}


type bridgeMapValue struct {
	keyValue     string
	resourceName string
}

func getDependFromResourceMap(resourceName string)DependResource{
	if depend,ok:=resourceMap[resourceName];ok{
		return depend
	}
	panic(fmt.Sprintf("can't get '%s' dependResource!",resourceName))
}

func hasDependResource(resourceName string)bool{
	_,ok:=resourceMap[resourceName]
	return ok
}