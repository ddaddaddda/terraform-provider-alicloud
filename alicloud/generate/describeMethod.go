package generate

import (
	"fmt"
	"github.com/terraform-providers/terraform-provider-alicloud/alicloud"
	"reflect"
	"strings"
)

var serviceTypeMap = map[string]reflect.Type{
	"ActionTrailService" : reflect.TypeOf(new(alicloud.ActionTrailService)),
	"CloudApiService" : reflect.TypeOf(new(alicloud.CloudApiService)),
	"CasService":reflect.TypeOf(new(alicloud.CasService)),
	"CdnService":reflect.TypeOf(new(alicloud.CdnService)),
	"CenService":reflect.TypeOf(new(alicloud.CenService)),
	"CmsService":reflect.TypeOf(new(alicloud.CmsService)),
	"CrService":reflect.TypeOf(new(alicloud.CrService)),
	"CsService":reflect.TypeOf(new(alicloud.CsService)),
	"DdoscooService":reflect.TypeOf(new(alicloud.DdoscooService)),
	"DnsService":reflect.TypeOf(new(alicloud.DnsService)),
	"DrdsService":reflect.TypeOf(new(alicloud.DrdsService)),
	"EcsService":reflect.TypeOf(new(alicloud.EcsService)),
	"ElasticsearchService":reflect.TypeOf(new(alicloud.ElasticsearchService)),
	"EssService":reflect.TypeOf(new(alicloud.EssService)),
	"FcService":reflect.TypeOf(new(alicloud.FcService)),
	"GpdbService":reflect.TypeOf(new(alicloud.GpdbService)),
	"HaVipService":reflect.TypeOf(new(alicloud.HaVipService)),
	"KmsService":reflect.TypeOf(new(alicloud.KmsService)),
	"KvstoreService":reflect.TypeOf(new(alicloud.KvstoreService)),
	"LogService":reflect.TypeOf(new(alicloud.LogService)),
	"MnsService":reflect.TypeOf(new(alicloud.MnsService)),
	"MongoDBService":reflect.TypeOf(new(alicloud.MongoDBService)),
	"NasService":reflect.TypeOf(new(alicloud.NasService)),
	"OnsService":reflect.TypeOf(new(alicloud.OnsService)),
	"OssService":reflect.TypeOf(new(alicloud.OssService)),
	"OtsService":reflect.TypeOf(new(alicloud.OtsService)),
	"PvtzService":reflect.TypeOf(new(alicloud.PvtzService)),
	"RamService":reflect.TypeOf(new(alicloud.RamService)),
	"RdsService":reflect.TypeOf(new(alicloud.RdsService)),
	"SlbService":reflect.TypeOf(new(alicloud.SlbService)),
	"VpcService":reflect.TypeOf(new(alicloud.VpcService)),
	"VpnGatewayService":reflect.TypeOf(new(alicloud.VpnGatewayService)),
}

func getStructNameAboutChectExist(describeMethod string)(string,string){
	for serviceName,serviceType :=range serviceTypeMap{
		method ,ok:=serviceType.MethodByName(describeMethod)
		if ok {
			funcStr := method.Type.String()
			start:=strings.LastIndex(funcStr,"(")+len("(")
			end :=strings.LastIndex(funcStr,",")
			return serviceName,funcStr[start:end]
		}
	}
	panic("not fount response type !")
}

func getResourceDescribeMethod(resourceName string) string {
	start := strings.Index(resourceName, "alicloud_")
	if start < 0 {
		panic(fmt.Errorf("the parameter \"name\" don't contain string \"alicloud_\""))
	}
	start += len("alicloud_")
	strs := strings.Split(resourceName[start:], "_")
	describeName := "Describe"
	for _, str := range strs {
		describeName = describeName + strings.Title(str)
	}
	return describeName
}