package generate

import (
	"bytes"
	"fmt"
	"github.com/hashicorp/terraform/helper/hashcode"
	"github.com/hashicorp/terraform/helper/schema"
	"reflect"
	"sort"
)

func HashResource(res *schema.Resource) schema.SchemaSetFunc {
	return func(v interface{}) int {
		var buf bytes.Buffer
		SerializeResourceForHash(&buf, v, res)
		return hashcode.String(buf.String())
	}
}

// HashSchema hashes values that are described using a *Schema. This is the
// default set implementation used when a set's element type is a single
// schema.
func HashSchema(sch *schema.Schema) schema.SchemaSetFunc {
	return func(v interface{}) int {
		var buf bytes.Buffer
		SerializeValueForHash(&buf, v, sch)
		return hashcode.String(buf.String())
	}
}


func SerializeValueForHash(buf *bytes.Buffer, val interface{}, sch *schema.Schema) {

	if val == nil {
		buf.WriteRune(';')
		return
	}
	val = getRealValueType(reflect.ValueOf(val)).Interface()

	switch sch.Type {
	case schema.TypeBool:
		if val.(string) == "true" {
			buf.WriteRune('1')
		} else {
			buf.WriteRune('0')
		}
	case schema.TypeInt,schema.TypeFloat,schema.TypeString:
		buf.WriteString(val.(string))
	case schema.TypeList:
		buf.WriteRune('(')
		l := val.([]interface{})
		for _, innerVal := range l {
			serializeCollectionMemberForHash(buf, innerVal, sch.Elem)
		}
		buf.WriteRune(')')
	case schema.TypeMap:

		m := val.(map[string]interface{})
		var keys []string
		for k := range m {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		buf.WriteRune('[')
		for _, k := range keys {
			innerVal := m[k]
			if innerVal == nil {
				continue
			}
			buf.WriteString(k)
			buf.WriteRune(':')
			buf.WriteString(innerVal.(string))
			buf.WriteRune(';')
		}
		buf.WriteRune(']')
	case schema.TypeSet:
		buf.WriteRune('{')
		s := val.(*schema.Set)
		for _, innerVal := range s.List() {
			serializeCollectionMemberForHash(buf, innerVal, sch.Elem)
		}
		buf.WriteRune('}')
	default:
		panic("unknown schema type to serialize")
	}
	buf.WriteRune(';')
}

// SerializeValueForHash appends a serialization of the given resource config
// to the given buffer, guaranteeing deterministic results given the same value
// and schema.
//
// Its primary purpose is as input into a hashing function in order
// to hash complex substructures when used in sets, and so the serialization
// is not reversible.
func SerializeResourceForHash(buf *bytes.Buffer, val interface{}, res *schema.Resource) {
	if val == nil {
		return
	}
	sm := res.Schema
	m := val.(map[string]interface{})
	var keys []string
	for k := range sm {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		innerSchema := sm[k]
		// Skip attributes that are not user-provided. Computed attributes
		// do not contribute to the hash since their ultimate value cannot
		// be known at plan/diff time.
		if !(innerSchema.Required || innerSchema.Optional) {
			continue
		}

		buf.WriteString(k)
		buf.WriteRune(':')
		innerVal := m[k]
		SerializeValueForHash(buf, innerVal, innerSchema)
	}
}

func serializeCollectionMemberForHash(buf *bytes.Buffer, val interface{}, elem interface{}) {
	switch tElem := elem.(type) {
	case *schema.Schema:
		SerializeValueForHash(buf, val, tElem)
	case *schema.Resource:
		buf.WriteRune('<')
		SerializeResourceForHash(buf, val, tElem)
		buf.WriteString(">;")
	default:
		panic(fmt.Sprintf("invalid element type: %T", tElem))
	}
}