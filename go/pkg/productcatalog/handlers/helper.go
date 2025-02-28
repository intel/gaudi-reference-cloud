// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"regexp"
	"strings"
	"unicode"

	"github.com/google/uuid"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func IsValidDecimal(s string) bool {
	_, ok := new(big.Rat).SetString(s)
	return ok
}

func IsValidJSON(s string) bool {
	var js json.RawMessage
	return json.Unmarshal([]byte(s), &js) == nil
}

// Checks if values are set, if a value is a string cannot be empty.
func ValidateProtoMessageFieldsNotEmpty(m proto.Message, exceptions []string) error {
	message := m.ProtoReflect()
	fields := message.Descriptor().Fields()

	for i := 0; i < fields.Len(); i++ {
		field := fields.Get(i)

		// Skip field names exceptions

		if containsString(exceptions, string(field.Name())) {
			continue
		}

		// Check if the field is set
		if !message.Has(field) {
			return fmt.Errorf("field '%s' is not set", field.Name())
		}

		// Get the field value
		value := message.Get(field)

		// Check for non-empty strings
		if field.Kind() == protoreflect.StringKind && value.String() == "" {
			return fmt.Errorf("string field '%s' is empty", field.Name())
		}
	}

	return nil
}

// Return the params and values needed for an update query.
// Each param has the following format: column_name = $1
func GetUpdateQueryPartsFromStruct(object interface{}) ([]string, []interface{}, error) {
	v := reflect.ValueOf(object)
	if v.Kind() == reflect.Ptr {
		v = v.Elem() // Dereference the pointer if necessary
	}

	if v.Kind() != reflect.Struct {
		return nil, nil, errors.New("object must be a struct or a pointer to a struct")
	}

	params := []string{}
	vals := []interface{}{}
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i)

		// Skip unexported fields or fields with a zero value
		if !value.CanInterface() || value.IsZero() {
			continue
		}

		// Use the JSON tag as the column name, if available, otherwise use the field name
		tag := field.Tag.Get("json")
		if tag == "" || tag == "-" {
			tag = field.Name
		} else {
			tag = strings.Split(tag, ",")[0] // Use the part before the comma, if any
		}

		// Convert tag name to sql format (snake case)
		column, err := camelToSnake(tag)
		if err != nil {
			return nil, nil, err
		}

		// Add the parameter and value
		params = append(params, fmt.Sprintf("%s = $%d", column, len(vals)+1))
		vals = append(vals, value.Interface())
	}

	return params, vals, nil
}

func camelToSnake(s string) (string, error) {
	var snakeCase strings.Builder
	for i, r := range s {
		if unicode.IsUpper(r) && i > 0 {
			if _, err := snakeCase.WriteRune('_'); err != nil {
				return "", err
			}
		}
		if _, err := snakeCase.WriteRune(unicode.ToLower(r)); err != nil {
			return "", err
		}
	}
	return snakeCase.String(), nil
}

// containsString checks if the slice contains the given string.
func containsString(slice []string, str string) bool {
	for _, item := range slice {
		if item == str {
			return true
		}
	}
	return false
}

func validateProductFilter(filter *pb.ProductFilter) error {
	if filter == nil {
		return fmt.Errorf("filter cannot be nil")
	}

	// Validate id
	if filter.Id != nil {
		idPattern := "^[0-9a-zA-Z]{8}-[0-9a-zA-Z]{4}-[0-9a-zA-Z]{4}-[0-9a-zA-Z]{4}-[0-9a-zA-Z]{12}$"
		matched, err := regexp.MatchString(idPattern, *filter.Id)
		if err != nil || !matched {
			return fmt.Errorf("id must match pattern %s", idPattern)
		}
	}

	// Validate vendorId
	if filter.VendorId != nil {
		if _, err := uuid.Parse(*filter.VendorId); err != nil {
			return fmt.Errorf("vendorId must be a valid uuid")
		}
	}

	// Validate familyId
	if filter.FamilyId != nil {
		if _, err := uuid.Parse(*filter.FamilyId); err != nil {
			return fmt.Errorf("familyid must be a valid uuid")
		}
	}

	// Validate metadata
	if len(filter.Metadata) > 0 {
		for key, value := range filter.Metadata {
			if key == "" || value == "" {
				return fmt.Errorf("metadata keys and values cannot be empty")
			}
		}
	}

	// Validate matchExpr
	if filter.MatchExpr != nil && len(*filter.MatchExpr) < 1 {
		return fmt.Errorf("matchexpr must be at least 1 character long")
	}

	return nil
}

func validateProductUserFilter(filter *pb.ProductUserFilter) error {
	if filter == nil {
		return fmt.Errorf("filter cannot be nil")
	}

	// Validate cloudaccountId
	if filter.CloudaccountId == "" {
		return fmt.Errorf("cloudaccountId cannot be empty")
	}

	idPattern := "^[0-9]{12}$"
	matched, err := regexp.MatchString(idPattern, filter.CloudaccountId)
	if err != nil || !matched {
		return fmt.Errorf("cloudaccountId must match pattern %s", idPattern)
	}

	// Validate nested ProductFilter
	if err := validateProductFilter(filter.ProductFilter); err != nil {
		return fmt.Errorf("invalid product filter: %v", err)
	}

	return nil
}

// incrementPlaceholders increments the placeholders by a given offset
func IncrementPlaceholders(placeholders string, offset int) string {
	parts := strings.Split(placeholders, ",")
	for i, part := range parts {
		var num int
		_, err := fmt.Sscanf(part, "$%d", &num)
		if err != nil {
			fmt.Printf("warning: failed to parse placeholder '%s': %v\n", part, err)
		}
		parts[i] = fmt.Sprintf("$%d", num+offset)
	}
	return strings.Join(parts, ", ")
}
