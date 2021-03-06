package golang

import (
	"fmt"
	"strings"

	"github.com/Jumpscale/go-raml/raml"
)

// FieldDef defines a field of a struct
type fieldDef struct {
	Name          string // field name
	Type          string // field type
	IsComposition bool   // composition type
	IsOmitted     bool   // omitted empty
	UniqueItems   bool
	Enum          *enum // not nil if this field contains enum

	Validators string
}

func newFieldDef(structName string, prop raml.Property, pkg string) fieldDef {
	fd := fieldDef{
		Name:      formatFieldName(prop.Name),
		Type:      convertToGoType(prop.TypeString(), prop.Items),
		IsOmitted: !prop.Required,
	}
	fd.buildValidators(prop)
	if prop.IsEnum() {
		fd.Enum = newEnum(structName, prop, pkg, false)
		fd.Type = fd.Enum.Name
	}

	return fd
}

func (fd *fieldDef) buildValidators(p raml.Property) {
	validators := []string{}
	addVal := func(s string) {
		validators = append(validators, s)
	}
	// string
	if p.MinLength != nil {
		addVal(fmt.Sprintf("min=%v", *p.MinLength))
	}
	if p.MaxLength != nil {
		addVal(fmt.Sprintf("max=%v", *p.MaxLength))
	}
	if p.Pattern != nil {
		addVal(fmt.Sprintf("regexp=%v", *p.Pattern))
	}

	// Number
	if p.Minimum != nil {
		addVal(fmt.Sprintf("min=%v", *p.Minimum))
	}

	if p.Maximum != nil {
		addVal(fmt.Sprintf("max=%v", *p.Maximum))
	}

	if p.MultipleOf != nil {
		addVal(fmt.Sprintf("multipleOf=%v", *p.MultipleOf))
	}

	//if p.Format != nil {
	//}

	// Array & Map
	if p.MinItems != nil {
		addVal(fmt.Sprintf("min=%v", *p.MinItems))
	}
	if p.MaxItems != nil {
		addVal(fmt.Sprintf("max=%v", *p.MaxItems))
	}
	if p.UniqueItems {
		fd.UniqueItems = true
	}

	// Required
	if !fd.IsOmitted && fd.Type != "bool" {
		addVal("nonzero")
	}

	fd.Validators = strings.Join(validators, ",")
}

// format struct's field name
// - Title it
// - replace '-' with camel case version
func formatFieldName(name string) string {
	var formatted string
	for _, v := range strings.Split(name, "-") {
		formatted += strings.Title(v)
	}
	return formatted
}
