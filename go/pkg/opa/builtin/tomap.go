// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package builtin

import (
	"google.golang.org/protobuf/reflect/protoreflect"
)

func ProtoMessageToMap(obj protoreflect.ProtoMessage) map[string]any {
	result := map[string]any{}
	obj.ProtoReflect().Range(
		func(fd protoreflect.FieldDescriptor, val protoreflect.Value) bool {
			if fd.Kind() == protoreflect.MessageKind {
				if !fd.IsList() {
					// Skip embedded non-list messages for now since we don't need
					// anything from a non-list embedded message yet
					return true
				}
				vals := []any{}
				list := val.List()
				for ii := 0; ii < list.Len(); ii++ {
					lval := list.Get(ii)
					vals = append(vals, ProtoMessageToMap(lval.Message().Interface()))
				}
				result[string(fd.Name())] = vals
				return true
			}
			result[string(fd.Name())] = val.Interface()
			return true
		},
	)
	return result
}
