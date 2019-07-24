package alicloud

var resourceMap = map[string]resourceMapValue{
	"alicloud_dns": {
		resourceName: "alicloud_dns",
		config: `
			resource "alicloud_dns" "default" {
			    name = "tf-testAccXXXXX"
			    group_id = ${alicloud_dns_group.id}
			}
           `,
		dependOn: []string{"alicloud_dns_group"},
	},
}

var bridgeMap = map[string]bridgeMapValue{}

type resourceMapValue struct {
	resourceName string
	config       string
	dependOn     []string
}

type bridgeMapValue struct {
	keyValue     string
	resourceName string
}
