// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package mock

import (
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"strconv"
)

// In development :- works for slices only if we know the length, but more nested structs still store empty values
// Using fmt purposefully
func parseFormToStruct(formData url.Values, target interface{}, tagName string) error {
	formFieldMap := make(map[string]reflect.Value)
	// Step 2: Loop through form values and populate the map
	for formName := range formData {
		formFieldMap[formName] = reflect.ValueOf(formData[formName][0])
	}
	//fmt.Println(formFieldMap)
	targetValue := reflect.ValueOf(target)
	if targetValue.Kind() != reflect.Ptr || targetValue.IsNil() {
		return errors.New("pointer type accepted")
	}
	targetElem := targetValue.Elem()
	fmt.Println(targetElem.Kind())
	if targetElem.Kind() != reflect.Struct {
		return errors.New("target must be a pointer to a struct")
	}

	targetType := targetElem.Type()
	fmt.Println("Target Type is:- ", targetType)
	for i := 0; i < targetType.NumField(); i++ {
		fieldValue := targetElem.Field(i)
		//fmt.Printf("%v:%v\n", i, fieldValue)
		fieldType := targetType.Field(i)
		fmt.Printf("%v Type:%#v\n", i, fieldType)
		formName := fieldType.Tag.Get("json")
		fmt.Printf("%v jsonName:%v\n", i, formName)
		if fieldValue.Kind() == reflect.Slice {
			slicefieldType := fieldValue.Type()
			sliceElemType := slicefieldType.Elem()
			slice := reflect.MakeSlice(slicefieldType, 0, 0)

			for j := 0; j < 3; j++ {
				elementFormName := tagName + fmt.Sprintf("%s[%d]", formName, j)
				element := reflect.New(sliceElemType).Elem()
				if _, ok := formFieldMap[elementFormName]; !ok {
					fmt.Println("Going to recursion!")
					err := parseFormToStruct(formData, element.Addr().Interface(), elementFormName)
					if err != nil {
						return err
					}
					// if err != nil {
					// 	return errors.New("recursion error")
					// }
				}
				fmt.Printf("%d:in if for:%+v\n", i, element)
				slice = reflect.Append(slice, element)
				fmt.Println(slice)
				fieldValue.Set(slice)

			}
		} else {
			var elementFormName string
			if tagName == "" {
				elementFormName = formName
			} else {
				elementFormName = fmt.Sprintf("%s[%s]", tagName, formName)
			}

			//fmt.Printf("%v-in if string:%v\n", i, elementFormName)
			if value, ok := formData[elementFormName]; ok {
				fmt.Println(value[0])
				setFieldValue(fieldValue, value[0])
			}
		}
		//else if fieldValue.Kind() == reflect.String {
		// elementFormName := fmt.Sprintf("%s[%s]", tagName, formName)
		// fmt.Printf("%v-in if string:%v\n", i, elementFormName)
		// str := fmt.Sprint(formData[elementFormName])
		// res := strings.Trim(str, "[]")
		// fieldValue.SetString(res)
		// }

	}
	fmt.Printf("End Value is:- %+v\n", target)
	// fmt.Println(reflect.TypeOf(target))
	return nil
}

// This converter works fine
func setFieldValue(fieldValue reflect.Value, value string) {
	switch fieldValue.Kind() {
	case reflect.String:
		fieldValue.SetString(value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		intValue, err := strconv.ParseInt(value, 10, 64)
		if err == nil {
			fieldValue.SetInt(intValue)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		uintValue, err := strconv.ParseUint(value, 10, 64)
		if err == nil {
			fieldValue.SetUint(uintValue)
		}
	case reflect.Float32, reflect.Float64:
		floatValue, err := strconv.ParseFloat(value, 64)
		if err == nil {
			fieldValue.SetFloat(floatValue)
		}
	case reflect.Bool:
		boolValue, err := strconv.ParseBool(value)
		if err == nil {
			fieldValue.SetBool(boolValue)
		}
	}
}
