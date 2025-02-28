// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package transformer

import (
	"strings"
)

// When using FromRow, the SQL SELECT query must select these columns.
func ColumnsForFromRow() string {
	cols := []string{"cloud_account_id", "resource_id", "name", "deleted_timestamp", "resource_version", "value"}
	return strings.Join(cols, ", ")
}
