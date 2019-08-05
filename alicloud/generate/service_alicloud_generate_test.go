package generate

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/terraform-providers/terraform-provider-alicloud/alicloud"
	"reflect"
)

const REMOVEKEY = "#REMOVEKEY"

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

func (g *generate)getStepNConfig(key string,changeConfigMap map[string]interface{})(map[string]interface{},map[string]interface{}){
	configMap := make(map[string]interface{})
	sch := g.getSchema(key)
	defaultValue, ok := changeConfigMap[key]
	if ok {
		configMap[key] = defaultValue
	}else if sch.Required {
		configMap[key] = getSchemaDefaultValue(key,sch,false)
	} else {
		configMap[key] = g.getSchemaValue(key,sch,false)
	}
	configMap = mapInterfaceValueCopy(configMap,changeConfigMap)
	reverseConfigMap := make(map[string]interface{})
	for key,newVal:=range configMap{
		if oldVal,ok := g.configMap[key];ok{
			if !reflect.DeepEqual(newVal,oldVal){
				reverseConfigMap[key] = oldVal
			}
		} else {
			reverseConfigMap[key] = REMOVEKEY
		}
	}
	return configMap,reverseConfigMap
}

func (g *generate)getStepNCheck(changeCheckMap map[string]string )(map[string]string,map[string]string){
	checkAllMap := make(map[string]string,0)
	for key,val:=range g.configMap{
		checkNode := getCheckNode(key,g.getSchema(key),val)
		checkAllMap = mapStringValueCopy(checkAllMap,checkNode.getCheckMap())
	}
	checkAllMap = mapStringValueCopy(changeCheckMap, checkAllMap)
	checkMap := make(map[string]string,0)
	reverseCheckMap := make(map[string]string,0)
	for key,newVal:=range checkAllMap {
		oldVal,ok := g.checkMap[key]
		if ok {
			if newVal != oldVal{
				reverseCheckMap[key] = oldVal
				checkMap[key] = newVal
			}
		}else {
			checkMap[key] = newVal
			reverseCheckMap[key] = REMOVEKEY
		}

	}

	for key :=range g.checkMap {
		if oldVal,ok := checkAllMap[key];!ok{
			reverseCheckMap[key] = oldVal
			checkMap[key] = REMOVEKEY
		}
	}
	return checkMap,reverseCheckMap
}


type StepChange struct{
	configMap map[string]interface{}
	checkMap map[string]string
}

func (g *generate)getStepN(changeMap map[string]StepChange)[]Step{
	var steps []Step
	for key,change:=range changeMap{
		configMap ,reverseConfigMap := g.getStepNConfig(key,change.configMap)
		g.updateConfig(configMap)
		checkMap,reverseCheckMap := g.getStepNCheck(change.checkMap)
		g.updateCheck(checkMap)
		step :=Step{ConfigMap:configMap,ReverseConfigMap:reverseConfigMap,CheckMap:checkMap,ReverseCheckMap:reverseCheckMap}
		steps = append(steps,step)
	}
	return steps
}

func (g *generate)updateConfig(configMap map[string]interface{}){
	for key,newVal := range configMap{
		_ ,ok := g.configMap[key]
		if ok {
			if strVal,ok := getRealValueType(reflect.ValueOf(newVal)).Interface().(string);
				ok && strVal == REMOVEKEY{
				delete(g.configMap,key)
			} else {
				delete(g.configMap,key)
				g.configMap[key] = newVal
			}
		} else {
			g.configMap[key] = newVal
		}
	}
}

func (g *generate)updateCheck(checkMap map[string]string){
	for key,newVal := range checkMap{
		_ ,ok := g.checkMap[key]
		if ok {
			if ok && newVal == REMOVEKEY{
				delete(g.configMap,key)
			} else {
				delete(g.checkMap,key)
				g.checkMap[key] = newVal
			}
		} else {
			g.checkMap[key] = newVal
		}
	}
}


func(g *generate)getStepAll(stepNs []Step)Step{
	configMap := make(map[string]interface{})
	checkMap := make(map[string]string)
	for i:=len(stepNs)-1;i>=0;i--{
		step := stepNs[i]
		g.updateConfig(step.ReverseConfigMap)
		g.updateCheck(step.ReverseCheckMap)
		for key,newVal := range step.ReverseConfigMap{
			_ ,ok := configMap[key]
			if ok {
				delete(g.configMap,key)
				configMap[key] = newVal
			} else {
				configMap[key] = newVal
			}
		}

		for key,newVal := range step.ReverseCheckMap{
			_ ,ok := checkMap[key]
			if ok {
				delete(g.checkMap,key)
				g.checkMap[key] = newVal
			} else {
				g.checkMap[key] = newVal
			}
		}
	}
	return Step{ConfigMap:configMap,CheckMap:checkMap}
}