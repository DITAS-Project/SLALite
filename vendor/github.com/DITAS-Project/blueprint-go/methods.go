package blueprint

import (
	"github.com/go-openapi/spec"
)

type ExtendedMethods struct {
	Properties AbstractPropertiesMethodType
	Method     DataManagementMethodType
	Path       string
	HTTPMethod string
	Tags       []string
}

type ExtendedOps struct {
	Ops    spec.Operation
	Path   string
	Method string
}

func (b BlueprintType) GetMethodMap() map[string]ExtendedMethods {
	//get all Swagger Operations form Blueprint
	ops := AssembleOperationsMap(b)

	//get all AbstractProperties form Blueprint
	properties := assemblePropertiesMap(b)

	//get all Methods form Blueprint
	methods := assembleMethodMap(b)

	//get all method tags
	tags := assembleTagMap(b)

	//create aggergatged map
	results := make(map[string]ExtendedMethods)

	for k, v := range properties {
		exOps := ops[k]
		result := ExtendedMethods{
			Properties: v,
			Path:       exOps.Path,
			HTTPMethod: exOps.Method,
		}

		if val, ok := methods[k]; ok {
			result.Method = val
		}

		if val, ok := tags[k]; ok {
			result.Tags = val
		}

		results[k] = result
	}

	return results
}

func AssembleOperationsMap(b BlueprintType) map[string]ExtendedOps {
	ops := make(map[string]ExtendedOps)

	addToOps := func(method string, path string, ops *spec.Operation, data map[string]ExtendedOps) {
		data[ops.ID] = ExtendedOps{Ops: *ops, Path: path, Method: method}
	}

	//Thats some ugly code :P but thats what worked
	if b.API.Paths != nil {
		for k, v := range b.API.Paths.Paths {
			if v.Get != nil {
				addToOps("GET", k, v.Get, ops)
			}
			if v.Post != nil {
				addToOps("POST", k, v.Post, ops)
			}
			if v.Put != nil {
				addToOps("PUT", k, v.Put, ops)
			}
			if v.Delete != nil {
				addToOps("DELETE", k, v.Delete, ops)
			}
			if v.Head != nil {
				addToOps("HEAD", k, v.Head, ops)
			}
			if v.Options != nil {
				addToOps("OPTIONS", k, v.Options, ops)
			}
			if v.Patch != nil {
				addToOps("PATCH", k, v.Patch, ops)
			}
		}
	}

	return ops
}

func assemblePropertiesMap(b BlueprintType) map[string]AbstractPropertiesMethodType {
	properties := make(map[string]AbstractPropertiesMethodType)

	for _, v := range b.AbstractProperties {
		if v.MethodId == nil {
			continue
		}

		properties[*v.MethodId] = v
	}
	return properties
}

func assembleMethodMap(b BlueprintType) map[string]DataManagementMethodType {
	methods := make(map[string]DataManagementMethodType)

	for _, v := range b.DataManagement {
		if v.MethodId == nil {
			continue
		}

		methods[*v.MethodId] = v
	}

	return methods
}

func assembleTagMap(b BlueprintType) map[string][]string {
	tags := make(map[string][]string)

	for _, t := range b.InternalStructure.Overview.Tags {
		tags[t.ID] = t.Tags
	}

	return tags
}
