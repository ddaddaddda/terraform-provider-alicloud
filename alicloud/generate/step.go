package generate

import (
	"fmt"
	"io"
	"reflect"
	"strings"
)

const (
	REMOVEKEY = "REMOVEKEY"

	CHECKSET = "CHECKSET"
	NOSET = "NOSET"
	REGEXMATCH = "REMOVEKEY"
)

const (
	// indentation symbol
	INDENTATIONSYMBOL = " "

	// child field indend number
	CHILDINDEND = 4
)

type Step struct {
	ConfigMap map[string]interface{}

	CheckMap map[string]string

	ReverseConfigMap map[string]interface{}

	ReverseCheckMap map[string]string
}

func(s *Step)outPut(indentation int,w io.Writer){
	io.WriteString(w,addIndentation(indentation))
	io.WriteString(w,"{\n")
	io.WriteString(w,addIndentation(indentation + CHILDINDEND))
	io.WriteString(w,"Config: testAccConfig(")
	commonString(w,indentation + CHILDINDEND,reflect.ValueOf(s.ConfigMap))
	io.WriteString(w,"),\n")
	io.WriteString(w,addIndentation(indentation + CHILDINDEND))
	io.WriteString(w,"Check: resource.ComposeTestCheckFunc(\n")
	io.WriteString(w,addIndentation(indentation + 2 * CHILDINDEND))
	io.WriteString(w,"testAccCheck(")
	checkMapString(w,indentation + 2 * CHILDINDEND,s.CheckMap)
	io.WriteString(w,"),\n")
	io.WriteString(w,addIndentation(indentation + CHILDINDEND))
	io.WriteString(w,"),\n")
	io.WriteString(w,addIndentation(indentation))
	io.WriteString(w,"}")
}

func commonString(w io.Writer,indentation int,reflectVal reflect.Value){
	reflectVal = getRealValueType(reflectVal)
	switch reflectVal.Kind() {
	case reflect.Map:
		 mapString(w,indentation,reflectVal)
	case reflect.Slice:
		listString(w,indentation,reflectVal)
	case reflect.String:
		str := reflectVal.String()
		if str == REMOVEKEY || str == CHECKSET || str == NOSET || strings.HasPrefix(str,REGEXMATCH){
			io.WriteString(w,fmt.Sprintf("%s",reflectVal.String()))
		} else {
			io.WriteString(w,fmt.Sprintf("\"%s\"",reflectVal.String()))
		}

	default:
		panic(fmt.Sprintf("unsuported type : %s",reflectVal.Type().String()))
	}
}

func mapString(w io.Writer,indentation int, reflectVal reflect.Value){
	io.WriteString(w,"map[string]interface{}{\n")

	for _,childKeyReflectVal:=range reflectVal.MapKeys(){
		childKey := childKeyReflectVal.String()
		childValueReflectVal := reflectVal.MapIndex(childKeyReflectVal)
		io.WriteString(w,addIndentation(indentation + CHILDINDEND))
		io.WriteString(w,childKey)
		io.WriteString(w," : ")
		commonString(w,indentation + CHILDINDEND,childValueReflectVal)
		io.WriteString(w,",\n")
	}
	io.WriteString(w,addIndentation(indentation))
	io.WriteString(w,"}")

}

func listString(w io.Writer,indentation int, reflectVal reflect.Value){
	io.WriteString(w,"[]interface{}{")
	for i:=0;i < reflectVal.Len();i++{
		childReflectVal := reflectVal.Index(i)
		if i == 0 {
			if childReflectVal.Kind() == reflect.String{
				io.WriteString(w," ")
			} else {
				io.WriteString(w,"\n")
				io.WriteString(w,addIndentation(indentation + CHILDINDEND))
			}
		}
		commonString(w,indentation + CHILDINDEND,childReflectVal)

		if childReflectVal.Kind() == reflect.String{
			if i < reflectVal.Len() -1 {
				io.WriteString(w,",")
			}
			io.WriteString(w," ")
		} else {
			io.WriteString(w,",\n")
		}
	}
	if reflectVal.Len() > 0 && reflectVal.Index(0).Kind() != reflect.String{
		io.WriteString(w,addIndentation(indentation))
	}
	io.WriteString(w,"}")
}

func addIndentation(indentation int) string {
	return strings.Repeat(INDENTATIONSYMBOL, indentation)
}

func checkMapString(w io.Writer,indentation int, checkMap map[string]string){
	io.WriteString(w,"map[string]string{\n")

	for key, val :=range checkMap{
		io.WriteString(w,addIndentation(indentation + CHILDINDEND))
		io.WriteString(w,key)
		io.WriteString(w," : ")
		if val == REMOVEKEY || val == CHECKSET || val == NOSET || strings.HasPrefix(val,REGEXMATCH){
			io.WriteString(w,fmt.Sprintf("%s",val))
		} else {
			io.WriteString(w,fmt.Sprintf("\"%s\"",val))
		}
		io.WriteString(w,",\n")
	}
	io.WriteString(w,addIndentation(indentation))
	io.WriteString(w,"}")

}