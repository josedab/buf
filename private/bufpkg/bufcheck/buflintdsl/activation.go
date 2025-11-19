// Copyright 2020-2025 Buf Technologies, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package buflintdsl

import (
	"fmt"

	"github.com/bufbuild/buf/private/bufpkg/bufprotosource"
	"github.com/google/cel-go/interpreter"
	"google.golang.org/protobuf/types/descriptorpb"
)

// itemsOfKind extracts all items of the given kind from a file.
func itemsOfKind(file bufprotosource.File, kind MatchKind) ([]any, error) {
	var items []any

	switch kind {
	case MatchKindFile:
		items = append(items, file)

	case MatchKindMessage:
		if err := bufprotosource.ForEachMessage(func(message bufprotosource.Message) error {
			if !message.IsMapEntry() {
				items = append(items, message)
			}
			return nil
		}, file); err != nil {
			return nil, err
		}

	case MatchKindField:
		if err := bufprotosource.ForEachMessage(func(message bufprotosource.Message) error {
			for _, field := range message.Fields() {
				items = append(items, field)
			}
			return nil
		}, file); err != nil {
			return nil, err
		}

	case MatchKindEnum:
		if err := bufprotosource.ForEachEnum(func(enum bufprotosource.Enum) error {
			items = append(items, enum)
			return nil
		}, file); err != nil {
			return nil, err
		}

	case MatchKindEnumValue:
		if err := bufprotosource.ForEachEnum(func(enum bufprotosource.Enum) error {
			for _, value := range enum.Values() {
				items = append(items, value)
			}
			return nil
		}, file); err != nil {
			return nil, err
		}

	case MatchKindService:
		for _, service := range file.Services() {
			items = append(items, service)
		}

	case MatchKindRPC:
		for _, service := range file.Services() {
			for _, method := range service.Methods() {
				items = append(items, method)
			}
		}

	case MatchKindOneof:
		if err := bufprotosource.ForEachMessage(func(message bufprotosource.Message) error {
			for _, oneof := range message.Oneofs() {
				items = append(items, oneof)
			}
			return nil
		}, file); err != nil {
			return nil, err
		}

	default:
		return nil, fmt.Errorf("unknown match kind: %s", kind)
	}

	return items, nil
}

// activationForItem creates a CEL activation (variable binding) for the given item.
func activationForItem(item any, kind MatchKind) interpreter.Activation {
	vars := make(map[string]any)

	switch v := item.(type) {
	case bufprotosource.Message:
		vars["name"] = v.Name()
		vars["full_name"] = v.FullName()
		vars["nested_name"] = v.NestedName()
		vars["comment"] = getComment(v)
		vars["deprecated"] = v.Deprecated()
		vars["is_map_entry"] = v.IsMapEntry()
		vars["field_count"] = int64(len(v.Fields()))
		vars["oneof_count"] = int64(len(v.Oneofs()))
		vars["nested_message_count"] = int64(len(v.Messages()))
		vars["nested_enum_count"] = int64(len(v.Enums()))
		vars["has_comment"] = hasComment(v)

	case bufprotosource.Field:
		vars["name"] = v.Name()
		vars["full_name"] = v.FullName()
		vars["number"] = int64(v.Number())
		vars["field_type"] = fieldTypeToString(v.Type())
		vars["type_name"] = v.TypeName()
		vars["label"] = fieldLabelToString(v.Label())
		vars["comment"] = getComment(v)
		vars["deprecated"] = v.Deprecated()
		vars["is_repeated"] = v.Label() == descriptorpb.FieldDescriptorProto_LABEL_REPEATED
		vars["is_map"] = isMapField(v)
		vars["is_optional"] = v.Label() == descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL
		vars["is_required"] = v.Label() == descriptorpb.FieldDescriptorProto_LABEL_REQUIRED
		vars["has_comment"] = hasComment(v)
		vars["json_name"] = v.JSONName()
		if v.ParentMessage() != nil {
			vars["message_name"] = v.ParentMessage().Name()
		} else {
			vars["message_name"] = ""
		}
		vars["has_default"] = v.Default() != ""
		vars["in_oneof"] = v.Oneof() != nil

	case bufprotosource.Enum:
		vars["name"] = v.Name()
		vars["full_name"] = v.FullName()
		vars["nested_name"] = v.NestedName()
		vars["comment"] = getComment(v)
		vars["deprecated"] = v.Deprecated()
		vars["value_count"] = int64(len(v.Values()))
		vars["has_comment"] = hasComment(v)
		vars["allow_alias"] = v.AllowAlias()

	case bufprotosource.EnumValue:
		vars["name"] = v.Name()
		vars["full_name"] = v.FullName()
		vars["number"] = int64(v.Number())
		vars["comment"] = getComment(v)
		vars["deprecated"] = v.Deprecated()
		vars["has_comment"] = hasComment(v)
		vars["enum_name"] = v.Enum().Name()

	case bufprotosource.Service:
		vars["name"] = v.Name()
		vars["full_name"] = v.FullName()
		vars["comment"] = getComment(v)
		vars["deprecated"] = v.Deprecated()
		vars["method_count"] = int64(len(v.Methods()))
		vars["has_comment"] = hasComment(v)

	case bufprotosource.Method:
		vars["name"] = v.Name()
		vars["full_name"] = v.FullName()
		vars["comment"] = getComment(v)
		vars["deprecated"] = v.Deprecated()
		vars["input_type"] = v.InputTypeName()
		vars["output_type"] = v.OutputTypeName()
		vars["client_streaming"] = v.ClientStreaming()
		vars["server_streaming"] = v.ServerStreaming()
		vars["has_comment"] = hasComment(v)
		vars["service_name"] = v.Service().Name()

	case bufprotosource.File:
		vars["path"] = v.Path()
		vars["package"] = v.Package()
		vars["syntax"] = v.Syntax().String()
		vars["message_count"] = int64(len(v.Messages()))
		vars["enum_count"] = int64(len(v.Enums()))
		vars["service_count"] = int64(len(v.Services()))
		vars["deprecated"] = v.Deprecated()

	case bufprotosource.Oneof:
		vars["name"] = v.Name()
		vars["full_name"] = v.FullName()
		vars["comment"] = getComment(v)
		vars["has_comment"] = hasComment(v)
		vars["field_count"] = int64(len(v.Fields()))
		vars["message_name"] = v.Message().Name()
	}

	return &mapActivation{vars: vars}
}

// mapActivation implements interpreter.Activation.
type mapActivation struct {
	vars map[string]any
}

func (a *mapActivation) ResolveName(name string) (any, bool) {
	v, ok := a.vars[name]
	return v, ok
}

func (a *mapActivation) Parent() interpreter.Activation {
	return nil
}

// Helper functions

func getComment(item any) string {
	if namedDesc, ok := item.(bufprotosource.NamedDescriptor); ok {
		if loc := namedDesc.NameLocation(); loc != nil {
			return loc.LeadingComments()
		}
	}
	return ""
}

func hasComment(item any) bool {
	return getComment(item) != ""
}

func fieldTypeToString(t descriptorpb.FieldDescriptorProto_Type) string {
	switch t {
	case descriptorpb.FieldDescriptorProto_TYPE_DOUBLE:
		return "double"
	case descriptorpb.FieldDescriptorProto_TYPE_FLOAT:
		return "float"
	case descriptorpb.FieldDescriptorProto_TYPE_INT64:
		return "int64"
	case descriptorpb.FieldDescriptorProto_TYPE_UINT64:
		return "uint64"
	case descriptorpb.FieldDescriptorProto_TYPE_INT32:
		return "int32"
	case descriptorpb.FieldDescriptorProto_TYPE_FIXED64:
		return "fixed64"
	case descriptorpb.FieldDescriptorProto_TYPE_FIXED32:
		return "fixed32"
	case descriptorpb.FieldDescriptorProto_TYPE_BOOL:
		return "bool"
	case descriptorpb.FieldDescriptorProto_TYPE_STRING:
		return "string"
	case descriptorpb.FieldDescriptorProto_TYPE_GROUP:
		return "group"
	case descriptorpb.FieldDescriptorProto_TYPE_MESSAGE:
		return "message"
	case descriptorpb.FieldDescriptorProto_TYPE_BYTES:
		return "bytes"
	case descriptorpb.FieldDescriptorProto_TYPE_UINT32:
		return "uint32"
	case descriptorpb.FieldDescriptorProto_TYPE_ENUM:
		return "enum"
	case descriptorpb.FieldDescriptorProto_TYPE_SFIXED32:
		return "sfixed32"
	case descriptorpb.FieldDescriptorProto_TYPE_SFIXED64:
		return "sfixed64"
	case descriptorpb.FieldDescriptorProto_TYPE_SINT32:
		return "sint32"
	case descriptorpb.FieldDescriptorProto_TYPE_SINT64:
		return "sint64"
	default:
		return "unknown"
	}
}

func fieldLabelToString(l descriptorpb.FieldDescriptorProto_Label) string {
	switch l {
	case descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL:
		return "optional"
	case descriptorpb.FieldDescriptorProto_LABEL_REQUIRED:
		return "required"
	case descriptorpb.FieldDescriptorProto_LABEL_REPEATED:
		return "repeated"
	default:
		return "unknown"
	}
}

func isMapField(field bufprotosource.Field) bool {
	// A field is a map if it's repeated and its type is a message that is a map entry
	if field.Label() != descriptorpb.FieldDescriptorProto_LABEL_REPEATED {
		return false
	}
	if field.Type() != descriptorpb.FieldDescriptorProto_TYPE_MESSAGE {
		return false
	}
	// Check type name ends with "Entry" (convention for map entries)
	typeName := field.TypeName()
	if len(typeName) > 5 && typeName[len(typeName)-5:] == "Entry" {
		return true
	}
	return false
}
