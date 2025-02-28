// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package protodb

import (
	"strconv"
	"strings"
)

// AddArrayArgs adds the elements of addArgs to args while constructing
// a string to refer to those arguments in an SQL query.
//
// For example, to run a query like this:
//
//	UPDATE usage_report SET reported=true WHERE id IN (2,4,6,8)
//
// Use AddArrayArgs like this:
//
//	args := []any{true}
//	args, argStr := protodb.AddArrayArgs(args, []int64{2,4,6,8})
//	query := "UPDATE usage_report SET reported=$1 WHERE id IN (" + argStr + ")"
//	_, err := dbconn.ExecContext(ctx, query, args...)
func AddArrayArgs[T any](args []any, addArgs []T) ([]any, string) {
	argStr := strings.Builder{}
	for _, arg := range addArgs {
		args = append(args, arg)
		if argStr.Len() > 0 {
			argStr.WriteByte(',')
		}
		argStr.WriteByte('$')
		argStr.WriteString(strconv.FormatInt(int64(len(args)), 10))
	}
	return args, argStr.String()
}

// AddArrayArgValues adds a set of arguments to args for each element of addArgs
// while constructing a string to refer to those arguements in an SQL query.
//
// For example, to run a query like this:
//
//	INSERT into credit_usage (usage_id, amount) VALUES (1,5),(2,8),(3,11)
//
// Use AddArrayArgValues like this:
//
//	 type CreditUsage struct {
//	 	id     int64
//	 	amount float64
//	 }
//	 args, argsStr := AddArrayArgValues(
//		[]any{},
//	 	[]*CreditUsage{{id: 1, amount: 5}, {id: 2, amount: 8}, {id: 3, amount: 11}},
//	 	func(creditUsage *CreditUsage) []any {
//			return []any{creditUsage.id, creditUsage.amount}
//		},
//	 )
//	 query := "INSERT into credit_usage (usage_id, amount) VALUES" + argsStr
//	 _, err := dbconn.ExecContext(ctx, query, args...)
func AddArrayArgValues[T any](args []any, addArgs []T, getValueArgs func(val T) []any) ([]any, string) {
	argStr := strings.Builder{}
	for _, arg := range addArgs {
		valArgs := getValueArgs(arg)
		if argStr.Len() > 0 {
			argStr.WriteByte(',')
		}
		argStr.WriteByte('(')
		for ii, valArg := range valArgs {
			args = append(args, valArg)
			if ii > 0 {
				argStr.WriteByte(',')
			}
			argStr.WriteByte('$')
			argStr.WriteString(strconv.FormatInt(int64(len(args)), 10))
		}
		argStr.WriteByte(')')
	}
	return args, argStr.String()
}
