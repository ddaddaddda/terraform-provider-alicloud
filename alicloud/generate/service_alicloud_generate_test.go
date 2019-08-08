package generate

import (
	"bytes"
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/terraform-providers/terraform-provider-alicloud/alicloud"
	"io"
	"reflect"
	"sort"
	"strings"
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

func initGenerator(resourceName string) *generate {
	provider := alicloud.Provider().(*schema.Provider)
	resource := provider.ResourcesMap[resourceName]
	sc := &generate{}
	sc.resourceName = resourceName
	sc.preCheckList = []string{"testAccPreCheck(t)"}
	sc.providers = "testAccProviders"
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
	name               string
	preCheckList       []string
	providers          string
	describeMethod     string
	forceNewSchema     map[string]*schema.Schema
	requiredSchema     map[string]*schema.Schema
	optionalSchema     map[string]*schema.Schema
	computedSchema     map[string]*schema.Schema
	dependResourceList []DependResource
	configMap          map[string]interface{}
	checkMap           map[string]string
	step0              Step
	stepN              []Step
	stepAll            Step
}

func(g *generate)SetGlobalName(name string){
	g.name = name
}

func(g *generate)SetPreCheckList(preCheckList []string){
	g.preCheckList = preCheckList
}

func(g *generate)SetProviders(providers string){
	g.providers = providers
}

func (g *generate) getSchema(key string) *schema.Schema {
	if sch, ok := g.forceNewSchema[key]; ok {
		return sch
	} else if sch, ok = g.requiredSchema[key]; ok {
		return sch
	} else if sch, ok = g.optionalSchema[key]; ok {
		return sch
	} else if sch, ok = g.computedSchema[key]; ok {
		return sch
	} else {
		return nil
	}
}

func (g *generate) addDependResource(dependResource ...DependResource) {
	g.dependResourceList = append(g.dependResourceList, dependResource...)
}

func (g *generate) getSchemaValue(key string, sch *schema.Schema, isChange bool) interface{} {
	if val, ok := bridgeMap[key]; ok {
		if hasDependResource(val.resourceName) {
			dependResource := getDependFromResourceMap(val.resourceName)
			g.addDependResource(dependResource)
		}
		return val.resourceName
	}
	return getSchemaDefaultValue(key, sch, isChange)
}

func (g *generate) Step0(changeConfigMap map[string]interface{}, changeCheckMap map[string]string){
	configMap := g.step0Config(changeConfigMap)
	g.configMap = mapInterfaceValueCopy(configMap,nil)
	checkMap := g.step0Check(changeCheckMap)
	g.checkMap = mapStringValueCopy(checkMap,nil)
	g.step0 = Step{ConfigMap: configMap, CheckMap: checkMap}
}

func (g *generate) step0Config(changeConfigMap map[string]interface{}) map[string]interface{} {
	configMap := make(map[string]interface{})
	for key, sch := range g.requiredSchema {
		if changeConfigMap != nil {
			val, ok := changeConfigMap[key]
			if ok {
				configMap[key] = val
				continue
			}
		}
		configMap[key] = g.getSchemaValue(key, sch, false)
	}

	for key, sch := range g.forceNewSchema {
		if changeConfigMap != nil {
			val, ok := changeConfigMap[key]
			if ok {
				configMap[key] = val
				continue
			}
		}
		if sch.Default != nil {
			configMap[key] = convertToString(sch.Default)
			continue
		}
		configMap[key] = g.getSchemaValue(key, sch, false)
	}
	configMap = mapInterfaceValueCopy(configMap,changeConfigMap)

	return configMap
}

func (g *generate) step0Check(changeCheckMap map[string]string) map[string]string {
	checkMap := make(map[string]string)
	for key, val := range g.configMap {
		childCheckMap := getCheckNode(key, g.getSchema(key), val).getCheckMap()
		checkMap = mapStringValueCopy(checkMap, childCheckMap)
	}
	for key, sch := range g.optionalSchema {
		if sch.Default != nil {
			childCheckMap := getCheckNode(key, sch, convertToString(sch.Default)).getCheckMap()
			checkMap = mapStringValueCopy(checkMap, childCheckMap)
		}
	}
	checkMap = mapStringValueCopy(changeCheckMap, checkMap)
	return checkMap
}

func (g *generate) stepNConfig(key string, changeConfigMap map[string]interface{}) (map[string]interface{}, map[string]interface{}) {
	configMap := make(map[string]interface{})
	sch := g.getSchema(key)
	defaultValue, ok := changeConfigMap[key]
	if ok {
		configMap[key] = defaultValue
	} else if sch.Required {
		configMap[key] = getSchemaDefaultValue(key, sch, true)
	} else {
		configMap[key] = g.getSchemaValue(key, sch, true)
	}
	configMap = mapInterfaceValueCopy(changeConfigMap,configMap)
	reverseConfigMap := make(map[string]interface{})
	for key, newVal := range configMap {
		if oldVal, ok := g.configMap[key]; ok {
			if !reflect.DeepEqual(newVal, oldVal) {
				reverseConfigMap[key] = oldVal
			}
		} else {
			reverseConfigMap[key] = REMOVEKEY
		}
	}
	return configMap, reverseConfigMap
}

func (g *generate) stepNCheck(configMap map[string]interface{},changeCheckMap map[string]string) (map[string]string, map[string]string) {

	var newCheckMap map[string]string
	var oldCheckMap map[string]string

	for key ,newVal :=range configMap{
		newChildCheckMap := getCheckNode(key,g.getSchema(key),newVal).getCheckMap()
		newCheckMap = mapStringValueCopy(newCheckMap,newChildCheckMap)
		if oldVal ,ok:=g.configMap[key];ok{
			oldChildCheckMap := getCheckNode(key,g.getSchema(key),oldVal).getCheckMap()
			oldCheckMap = mapStringValueCopy(oldCheckMap,oldChildCheckMap)
		}
	}
	newCheckMap = mapStringValueCopy(changeCheckMap,newCheckMap)

	checkMap := make(map[string]string, 0)
	reverseCheckMap := make(map[string]string, 0)

	for key,oldVal :=range oldCheckMap{
		if _,ok := newCheckMap[key];!ok{
			checkMap[key] = REMOVEKEY
			reverseCheckMap[key] = oldVal
		}
	}

	for key, newVal := range newCheckMap {
		oldVal, ok := g.checkMap[key]
		if ok {
			if newVal != oldVal {
				reverseCheckMap[key] = oldVal
				checkMap[key] = newVal
			}
		} else {
			checkMap[key] = newVal
			reverseCheckMap[key] = REMOVEKEY
		}

	}
	return checkMap, reverseCheckMap
}

type StepChange struct {
	configMap map[string]interface{}
	checkMap  map[string]string
}

func(g *generate)getStepN(key string,changeConfigMap map[string]interface{},changeCheckMap map[string]string)Step{
	configMap, reverseConfigMap := g.stepNConfig(key, changeConfigMap)
	checkMap, reverseCheckMap := g.stepNCheck(configMap,changeCheckMap)
	g.updateConfig(configMap)
	g.updateCheck(checkMap)
	return Step{ConfigMap: configMap, ReverseConfigMap: reverseConfigMap, CheckMap: checkMap, ReverseCheckMap: reverseCheckMap}
}

func (g *generate) StepN(changeMap map[string]StepChange) {
	var steps []Step

	for key:=range g.requiredSchema{
		changeConfigMap := make(map[string]interface{})
		changeCheckMap := make(map[string]string)
		if change ,ok :=changeMap[key];ok{
			changeConfigMap = change.configMap
			changeCheckMap = change.checkMap
		}
		step := g.getStepN(key,changeConfigMap,changeCheckMap)
		steps = append(steps, step)
	}

	for key:=range g.optionalSchema{
		changeConfigMap := make(map[string]interface{})
		changeCheckMap := make(map[string]string)
		if change ,ok :=changeMap[key];ok{
			changeConfigMap = change.configMap
			changeCheckMap = change.checkMap
		}
		step := g.getStepN(key,changeConfigMap,changeCheckMap)
		steps = append(steps, step)
	}

	g.stepN = steps
}

func (g *generate) StepAll(changeConfigMap map[string]interface{},changeCheckMap map[string]string) {
	configMap := make(map[string]interface{})
	checkMap := make(map[string]string)
	for i := len(g.stepN) - 1; i >= 0; i-- {
		step := g.stepN[i]
		g.updateConfig(step.ReverseConfigMap)
		g.updateCheck(step.ReverseCheckMap)
		for key, newVal := range step.ReverseConfigMap {
			_, ok := configMap[key]
			if ok {
				delete(g.configMap, key)
				configMap[key] = newVal
			} else {
				configMap[key] = newVal
			}
		}

		for key, newVal := range step.ReverseCheckMap {
			_, ok := checkMap[key]
			if ok {
				delete(g.checkMap, key)
				checkMap[key] = newVal
			} else {
				checkMap[key] = newVal
			}
		}
	}

	g.updateConfig(changeConfigMap)
	g.updateCheck(changeCheckMap)
	for key, newVal := range changeConfigMap {
		_, ok := configMap[key]
		if ok {
			delete(g.configMap, key)
			configMap[key] = newVal
		} else {
			configMap[key] = newVal
		}
	}

	for key, newVal := range changeCheckMap {
		_, ok := checkMap[key]
		if ok {
			delete(g.checkMap, key)
			checkMap[key] = newVal
		} else {
			checkMap[key] = newVal
		}
	}


	g.stepAll = Step{ConfigMap: configMap, CheckMap: checkMap}
}


func (g *generate) updateConfig(configMap map[string]interface{}) {
	for key, newVal := range configMap {
		_, ok := g.configMap[key]
		if ok {
			if strVal, ok := getRealValueType(reflect.ValueOf(newVal)).Interface().(string);
				ok && strVal == REMOVEKEY {
				delete(g.configMap, key)
			} else {
				delete(g.configMap, key)
				g.configMap[key] = newVal
			}
		} else {
			g.configMap[key] = newVal
		}
	}
}

func (g *generate) updateCheck(checkMap map[string]string) {
	for key, newVal := range checkMap {
		_, ok := g.checkMap[key]
		if ok {
			if ok && newVal == REMOVEKEY {
				delete(g.configMap, key)
			} else {
				delete(g.checkMap, key)
				g.checkMap[key] = newVal
			}
		} else {
			g.checkMap[key] = newVal
		}
	}
}

func (g *generate) getConfigDependence() string {
	var dependOnAll []string
	var configList []string
	resourceSet := make(map[string]struct{}, 0)
	for _, depend := range g.dependResourceList {
		dependOnAll = append(dependOnAll, depend.dependOn...)
		configList = append(configList, depend.configs...)
		resourceSet[depend.resourceName] = struct{}{}
	}
	for _, resourceName := range dependOnAll {
		_, ok := resourceSet[resourceName]
		if !ok {
			dependResource := getDependFromResourceMap(resourceName)
			for _, depend := range dependResource.dependOn {
				if count := sort.SearchStrings(dependOnAll, depend); count == 0 {
					dependOnAll = append(dependOnAll, depend)
				}
			}
			configList = append(configList, dependResource.configs...)
			resourceSet[resourceName] = struct{}{}
		}
	}


	return fmt.Sprintf(`func (name string) string {
		return fmt.Sprintf(` + "`\n%s\n" +"	  `, name)\n}",strings.Join(configList, "\n'"))
}

func(g *generate)outPutTestCode(w io.Writer){
	if g.describeMethod == ""{
		g.describeMethod = getResourceDescribeMethod(g.resourceName)
	}

	resourceId := g.resourceName + ".default"
	serviceName, responseType := getStructNameAboutChectExist(g.describeMethod)

	resourceConfigDependence := g.getConfigDependence()


	io.WriteString(w,"func TestAccAlicloud" + strings.Replace(g.describeMethod,"Describe","",0) + "Basic(t *testing.T) {\n")
	io.WriteString(w,fmt.Sprintf("    var v %s\n\n",responseType))

	io.WriteString(w,fmt.Sprintf("    resourceId := \"%s\"\n",resourceId))
	io.WriteString(w,"    ra := resourceAttrInit(resourceId, nil)\n")
	io.WriteString(w,"    serviceFunc := func() interface{} {\n")
	io.WriteString(w,fmt.Sprintf("        return &%s{testAccProvider.Meta().(*connectivity.AliyunClient)}\n",serviceName))
	io.WriteString(w,"    }\n")
	io.WriteString(w,"    rc := resourceCheckInit(resourceId, &v, serviceFunc)\n")
	io.WriteString(w,"    rac := resourceAttrCheckInit(rc, ra)\n\n")

	io.WriteString(w,"    testAccCheck := rac.resourceAttrMapUpdateSet()\n")
	io.WriteString(w,fmt.Sprintf("    name := \"%s\"\n\n",g.name))

	io.WriteString(w,fmt.Sprintf("    resourceConfigDependence = %s\n\n",resourceConfigDependence))

	io.WriteString(w,"    testAccConfig := resourceTestAccConfigFunc(resourceId, name, resourceConfigDependence)\n\n")

	io.WriteString(w,"    resource.Test(t, resource.TestCase{\n")
	io.WriteString(w,"        PreCheck: func() {\n")
	for _,preCheck:=range g.preCheckList{
		io.WriteString(w,fmt.Sprintf("            %s\n",preCheck))
	}
	io.WriteString(w,"        },\n")
	io.WriteString(w,fmt.Sprintf("        IDRefreshName: %s,\n",resourceId))
	io.WriteString(w,fmt.Sprintf("        Providers:     %s,\n",g.providers))
	io.WriteString(w,"        CheckDestroy:  rac.checkResourceDestroy(),\n")
	io.WriteString(w,"        Steps: []resource.TestStep{\n")

	g.step0.outPut(12,w)
	io.WriteString(w,"\n")
	for _,step :=range g.stepN{
		step.outPut(12,w)
		io.WriteString(w,"\n")
	}
	g.stepAll.outPut(12,w)
	io.WriteString(w,"\n")
	io.WriteString(w,"        },\n")
	io.WriteString(w,"    })\n")
	io.WriteString(w,"}\n")
}


func TestSearchMethod(t *testing.T){
	g := initGenerator("alicloud_instance")
	g.Step0(map[string]interface{}{
			/*"image_id":        "${data.alicloud_images.default.images.0.id}",
			"security_groups": []string{"${alicloud_security_group.default.0.id}","123"},
			"instance_type":   "${data.alicloud_instance_types.default.instance_types.0.id}",

			"availability_zone":             "${data.alicloud_zones.default.zones.0.id}",
			"system_disk_category":          "cloud_efficiency",
			"instance_name":                 "${var.name}",
			"key_name":                      "${alicloud_key_pair.default.key_name}",
			"spot_strategy":                 "NoSpot",
			"spot_price_limit":              "0",
			"security_enhancement_strategy": "Active",
			"user_data":                     "I_am_user_data",

			"vswitch_id": "${alicloud_vswitch.default.id}",
			"role_name":  "${alicloud_ram_role.default.name}",*/
		},
		map[string]string{
			/*"image_id":          CHECKSET,
			"instance_type":     CHECKSET,
			"security_groups.#": "1",

			"availability_zone":             CHECKSET,
			"system_disk_category":          "cloud_efficiency",
			"spot_strategy":                 "NoSpot",
			"spot_price_limit":              "0",
			"security_enhancement_strategy": "Active",
			"vswitch_id":                    CHECKSET,
			"user_data":                     "I_am_user_data",

			"description":      "",
			"host_name":        CHECKSET,
			"password":         "",
			"is_outdated":      NOSET,
			"system_disk_size": "40",

			"data_disks.#":  NOSET,
			"volume_tags.%": "0",
			"tags.%":        NOSET,

			"private_ip": CHECKSET,
			"public_ip":  "",
			"status":     "Running",

			"internet_charge_type":       "PayByTraffic",
			"internet_max_bandwidth_in":  "-1",
			"internet_max_bandwidth_out": "0",

			"instance_charge_type": "PostPaid",
			// the attributes of below are suppressed  when the value of instance_charge_type is `PostPaid`
			"period":             NOSET,
			"period_unit":        NOSET,
			"renewal_status":     NOSET,
			"auto_renew_period":  NOSET,
			"force_delete":       NOSET,
			"include_data_disks": NOSET,
			"dry_run":            NOSET,*/
		},
		)
	g.StepN(nil)
	g.StepAll(nil,nil)
	buf:=bytes.NewBufferString("")
	g.outPutTestCode(buf)
	println(buf.String())

	//getStructNameAboutChectExist("DescribeInstance")
}


