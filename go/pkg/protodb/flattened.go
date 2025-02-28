// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package protodb

import (
	"fmt"
	"strings"
)

// A Flattened object can be used to construct SQL statements.
type Flattened struct {
	// Column names
	Columns []string
	// Field values in same order as Columns
	Values []any
}

func (f *Flattened) Add(column string, value any) {
	f.Columns = append(f.Columns, column)
	f.Values = append(f.Values, value)
}

// Returns the list of columns separated by commas.
func (f *Flattened) GetColumnsString() string {
	return strings.Join(f.Columns, ", ")
}

// Returns an expression with place holders that can be added to a SQL INSERT VALUES expression.
// For example: $1, $2
func (f *Flattened) GetInsertValuesString(firstPlaceholderNumber int) string {
	var p []string
	for i := 0; i < len(f.Columns); i++ {
		p = append(p, fmt.Sprintf("$%d", i+firstPlaceholderNumber))
	}
	return strings.Join(p, ", ")
}

// Returns an expression that can be added to a SQL UPDATE SET expression.
// For example: name1 = $1, name2 = $2
func (f *Flattened) GetUpdateSetString(firstPlaceholderNumber int) string {
	var p []string
	for i := 0; i < len(f.Columns); i++ {
		p = append(p, fmt.Sprintf("%s = $%d", f.Columns[i], i+firstPlaceholderNumber))
	}
	return strings.Join(p, ", ")
}

// Returns an expression that can be added to a SQL WHERE expression.
// If there are no columns, returns "1 = 1".
// For example: name1 = $1 and name2 = $2
func (f *Flattened) GetWhereString(firstPlaceholderNumber int) string {
	if len(f.Columns) == 0 {
		return "1 = 1"
	}
	var p []string
	for i := 0; i < len(f.Columns); i++ {
		p = append(p, fmt.Sprintf("%s = $%d", f.Columns[i], i+firstPlaceholderNumber))
	}
	return strings.Join(p, " and ")
}

// Returns an expression that can be added to a SQL WHERE expression
// which uses the contains @> jsonb operator.
// If there are no columns, returns "1 = 1".
// For example: jsonb expression1 @> $1 and jsonb expression2 @> $2
func (f *Flattened) GetWhereContains(firstPlaceholderNumber int) string {
	if len(f.Columns) == 0 {
		return "1 = 1"
	}
	var p []string
	for i := 0; i < len(f.Columns); i++ {
		p = append(p, fmt.Sprintf("%s @> $%d", f.Columns[i], i+firstPlaceholderNumber))
	}
	return strings.Join(p, " and ")
}
