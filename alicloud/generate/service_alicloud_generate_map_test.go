package generate

import (
	"fmt"
)

var resourceMap = map[string]DependResource{
	"alicloud_zones": {
		resourceName: "alicloud_zones",
		configs: []string{`
			data "alicloud_zones" "default" {
				available_disk_category     = "cloud_efficiency"
				available_resource_creation = "VSwitch"
			}
           `},
		dependOn: []string{},
	},
	"alicloud_instance_types": {
		resourceName: "alicloud_instance_types",
		configs: []string{`
			data "alicloud_instance_types" "default" {
				availability_zone = "${data.alicloud_zones.default.zones.0.id}"
			}
           `},
		dependOn: []string{ "availability_zone" },
	},

	"alicloud_images": {
		resourceName: "alicloud_images",
		configs: []string{`
			data "alicloud_images" "default" {
				name_regex  = "^ubuntu*"
				owners      = "system"
			}
           `},
		dependOn: []string{ },
	},
	"alicloud_vpc": {
		resourceName: "alicloud_vpc",
		configs: []string{`
			resource "alicloud_vpc" "default" {
				name       = "${var.name}"
				cidr_block = "172.16.0.0/16"
			}
           `},
		dependOn: []string{ "name" },
	},

	"alicloud_vswitch": {
		resourceName: "alicloud_vswitch",
		configs: []string{`
			resource "alicloud_vswitch" "default" {
				vpc_id            = "${alicloud_vpc.default.id}"
				cidr_block        = "172.16.0.0/24"
				availability_zone = "${data.alicloud_zones.default.zones.0.id}"
				name              = "${var.name}"
			}
           `},
		dependOn: []string{ "alicloud_vpc", "alicloud_zones"},
	},
	"alicloud_security_group": {
		resourceName: "alicloud_security_group",
		configs: []string{`
			resource "alicloud_security_group" "default" {
				count = "2"
				name   = "${var.name}"
				vpc_id = "${alicloud_vpc.default.id}"
			}
           `},
		dependOn: []string{ "alicloud_vpc"},
	},
	"alicloud_security_group_rule": {
		resourceName: "alicloud_security_group_rule",
		configs: []string{`
			resource "alicloud_security_group_rule" "default" {
				count = 2
				type = "ingress"
				ip_protocol = "tcp"
				nic_type = "intranet"
				policy = "accept"
				port_range = "22/22"
				priority = 1
				security_group_id = "${element(alicloud_security_group.default.*.id,count.index)}"
				cidr_ip = "172.16.0.0/24"
			}
           `},
		dependOn: []string{ "alicloud_vpc"},
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





/*variable "name" {
default = "%s"
}

resource "alicloud_ram_role" "default" {
name = "${var.name}"
services = ["ecs.aliyuncs.com"]
force = "true"
}

resource "alicloud_key_pair" "default" {
key_name = "${var.name}"
}
*/