package generate

import (
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/terraform-providers/terraform-provider-alicloud/alicloud"
	"testing"
)


type DependResource struct {
	resourceName string
	configs      []string
	dependOn     []string
}

type attributeValue struct {
	sch   *schema.Schema
	value interface{}
}

type Step struct{

	ConfigMap map[string]interface{}

	CheckMap map[string]string

	ReverseConfigMap map[string]interface{}

	ReverseCheckMap map[string]string
}


func initGenerator(resourceName string) *generate {
	provider := alicloud.Provider().(*schema.Provider)
	resource := provider.ResourcesMap[resourceName]
	sc := &generate{}
	sc.resourceName = resourceName
	sc.forceNewSchema = map[string]*schema.Schema{}
	sc.requiredSchema = map[string]*schema.Schema{}
	sc.optionalSchema = map[string]*schema.Schema{}
	sc.computedSchema = map[string]*schema.Schema{}
	sc.configMap = map[string]interface{}{}
	sc.checkMap = map[string]string{}

	for key, sch := range resource.Schema {
		if isForceNew(sch) {
			sc.forceNewSchema[key] = sch
		} else if isRequired(sch) {
			sc.requiredSchema[key] = sch
		} else if isOptional(sch) {
			sc.optionalSchema[key] = sch
		}

		if isComputed(sch) {
			sc.computedSchema[key] = sch
		}
	}
	return sc
}

type generate struct {
	resourceName       string
	forceNewSchema     map[string]*schema.Schema
	requiredSchema     map[string]*schema.Schema
	optionalSchema     map[string]*schema.Schema
	computedSchema     map[string]*schema.Schema
	dependResourceList []DependResource
	configMap          map[string]interface{}
	checkMap          map[string]string
}

func (g *generate)getSchema(key string)*schema.Schema{
	if sch,ok := g.forceNewSchema[key];ok{
		return sch
	} else if sch,ok = g.requiredSchema[key];ok{
		return sch
	} else if sch,ok = g.optionalSchema[key];ok{
		return sch
	} else if sch,ok = g.computedSchema[key];ok{
		return sch
	} else {
		return nil
	}
}

func (g *generate) addDependResource(dependResource ...DependResource) {
	g.dependResourceList = append(g.dependResourceList, dependResource...)
}

func (g *generate) getSchemaValue(key string, sch *schema.Schema, isChange bool)interface{} {
	if val, ok := bridgeMap[key]; ok {
		if hasDependResource(val.resourceName){
			dependResource := getDependFromResourceMap(val.resourceName)
			g.addDependResource(dependResource)
		}
		return val.resourceName
	}
	return getSchemaDefaultValue(key,sch,isChange)
}

func (g *generate)getStep0(changeConfigMap map[string]interface{},changeCheckMap map[string]string) Step {
	configMap := g.getStep0Config(changeConfigMap)
	g.configMap = configMap
	checkMap := g.getStep0Check(changeCheckMap)
	return Step{ConfigMap: configMap, CheckMap:checkMap}
}


func TestGetStep0(t *testing.T){
	name := fmt.Sprintf("tf-testAccEcsInstanceDataDisks%d",  13344)
	gt:=initGenerator("alicloud_instance")
	step := gt.getStep0(map[string]interface{}{
		"image_id":        "${data.alicloud_images.default.images.0.id}",
		"security_groups": []string{"${alicloud_security_group.default.0.id}"},
		"instance_type":   "${data.alicloud_instance_types.default.instance_types.0.id}",

		"availability_zone":             "${data.alicloud_zones.default.zones.0.id}",
		"system_disk_category":          "cloud_efficiency",
		"instance_name":                 "${var.name}",
		"key_name":                      "${alicloud_key_pair.default.key_name}",
		"spot_strategy":                 "NoSpot",
		"spot_price_limit":              "0",
		"security_enhancement_strategy": "Active",
		"user_data":                     "I_am_user_data",

		"instance_charge_type": "PrePaid",
		"vswitch_id":           "${alicloud_vswitch.default.id}",
		"role_name":            "${alicloud_ram_role.default.name}",
		"data_disks": []map[string]string{
			{
				"name":        "disk1",
				"size":        "20",
				"category":    "cloud_efficiency",
				"description": "disk1",
			},
			{
				"name":        "disk2",
				"size":        "20",
				"category":    "cloud_efficiency",
				"description": "disk2",
			},
		},
		"force_delete": "true",
	},map[string]string{
		"instance_name": name,
		"key_name":      name,
		"role_name":     name,
		"user_data":     "I_am_user_data",

		"data_disks.#":             "2",
		"data_disks.0.name":        "disk1",
		"data_disks.0.size":        "20",
		"data_disks.0.category":    "cloud_efficiency",
		"data_disks.0.description": "disk1",
		"data_disks.1.name":        "disk2",
		"data_disks.1.size":        "20",
		"data_disks.1.category":    "cloud_efficiency",
		"data_disks.1.description": "disk2",

		"force_delete":         "true",
		"instance_charge_type": "PrePaid",
		"period":               "1",
		"period_unit":          "Month",
		"renewal_status":       "Normal",
		"auto_renew_period":    "0",
		"include_data_disks":   "true",
		"dry_run":              "false",
	})
	bs,_:=json.Marshal(step)
	println(string(bs))
}



func (g *generate)getStep0Config(changeConfigMap map[string]interface{})map[string]interface{}{
	configMap := make(map[string]interface{})
	for key, sch := range g.requiredSchema{
		if changeConfigMap != nil{
			val ,ok:=changeConfigMap[key]
			if ok {
				configMap[key]=val
				continue
			}
		}
		configMap[key]=g.getSchemaValue(key,sch,false)
	}

	for key, sch := range g.forceNewSchema{
		if changeConfigMap != nil{
			val ,ok:=changeConfigMap[key]
			if ok {
				configMap[key]=val
				continue
			}
		}
		if sch.Default != nil {
			configMap[key] = convertToString(sch.Default)
			continue
		}
		configMap[key]=g.getSchemaValue(key,sch,false)
	}
	return configMap
}

func (g *generate)getStep0Check(changeCheckMap map[string]string)map[string]string{
	checkMap := make(map[string]string)
	for key, val := range g.configMap{
		childCheckMap := getCheckNode(key,g.getSchema(key),val).getCheckMap()
		checkMap = mapStringValueCopy(checkMap,childCheckMap)
	}
	for key,sch := range g.optionalSchema{
		if sch.Default != nil{
			childCheckMap := getCheckNode(key,sch,convertToString(sch.Default)).getCheckMap()
			checkMap = mapStringValueCopy(checkMap,childCheckMap)
		}
	}
	checkMap = mapStringValueCopy(changeCheckMap, checkMap)
	return checkMap
}


