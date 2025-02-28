// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package usage

import (
	"errors"
	"regexp"
	"strconv"
	"strings"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
)

const clusterTypeKey = "clusterType"
const instanceGroupKey = "instanceGroupSize"

// Trying to keep it as simple as possible to avoid mismatched selections.
func CheckIfProductMapsToProperties(product *pb.Product, properties map[string]string) (bool, error) {
	// these are all the ands
	var andExpressions []string
	// these are all the ors
	var orExpressions []string
	matchExpression := product.GetMatchExpr()

	if strings.Contains(matchExpression, clusterTypeKey) {
		if _, foundClusterTypeKey := properties[clusterTypeKey]; !foundClusterTypeKey {
			return false, nil
		}
	}

	if _, foundClusterTypeKey := properties[clusterTypeKey]; foundClusterTypeKey {
		if !strings.Contains(matchExpression, clusterTypeKey) {
			return false, nil
		}
	}

	if strings.Contains(matchExpression, instanceGroupKey) {
		if _, foundInstanceGroupKey := properties[instanceGroupKey]; !foundInstanceGroupKey {
			return false, nil
		}
	}

	if _, foundInstanceGroupKey := properties[instanceGroupKey]; foundInstanceGroupKey {
		if !strings.Contains(matchExpression, instanceGroupKey) {
			return false, nil
		}
	}

	// Check if there are && expressions
	// If there are || expressions, they would either be stand alone or as a part of "and" expression.
	// There will be no && expression within a parenthesis
	if strings.Contains(matchExpression, "&&") {
		notTrimmedAndExpressions := strings.Split(product.GetMatchExpr(), "&&")
		var trimmedExpressions []string
		for _, notTrimmedAndExpression := range notTrimmedAndExpressions {
			// If we have open parenthesis as a part of expression and not the closing
			// or the reverse, then we have a "and" expression within parenthesis and that is not valid - because we are splitting using &&
			if ((strings.Contains(notTrimmedAndExpression, "(")) && !(strings.Contains(notTrimmedAndExpression, ")"))) ||
				((strings.Contains(notTrimmedAndExpression, ")")) && !(strings.Contains(notTrimmedAndExpression, "("))) {
				return false, errors.New("invalid match expression")
			}
			trimmedExpression := strings.TrimSpace(notTrimmedAndExpression)
			if len(trimmedExpression) == 0 {
				return false, errors.New("invalid match expression")
			}
			trimmedExpressions = append(trimmedExpressions, trimmedExpression)
		}
		for _, trimmedExpression := range trimmedExpressions {
			// Check if it contains a ||
			// Nested parenthesis is not required and not supported.
			if strings.Contains(trimmedExpression, "(") {
				if !strings.Contains(trimmedExpression, ")") {
					return false, errors.New("invalid match expression")
				}
				// remove the parenthesis
				removeParenthesis := strings.Trim(trimmedExpression, "()")
				notTrimmedOrExpressions := strings.Split(removeParenthesis, "||")
				for _, notTrimmedOrExpression := range notTrimmedOrExpressions {
					orExpressions = append(orExpressions, strings.TrimSpace(notTrimmedOrExpression))
				}
				// Else has to be a &&
			} else {
				andExpressions = append(andExpressions, trimmedExpression)
			}
		}
	} else {
		// trim the spaces and we only have one expression.
		andExpressions = append(andExpressions, strings.TrimSpace(matchExpression))
	}
	var foundMatching = 0
	for _, andExpression := range andExpressions {

		// a expression cannot contain both equals and not equals.
		if (strings.Contains(andExpression, "==")) && (strings.Contains(andExpression, "!=")) {
			return false, errors.New("invalid match expression: " + andExpression)
		}
		expressionMatches, err := checkExpressionMatches(andExpression, properties)
		if err != nil {
			return false, err
		}
		if !expressionMatches {
			return false, nil
		} else {
			foundMatching += 1
		}
	}

	if foundMatching != len(andExpressions) {
		return false, nil
	}

	// All the &&s need to be true
	if (foundMatching == len(andExpressions)) && (len(orExpressions) == 0) {
		return true, nil
	}

	for _, orExpression := range orExpressions {

		// a expression cannot contain both equals and not equals.
		if (strings.Contains(orExpression, "==")) && (strings.Contains(orExpression, "!=")) {
			return false, errors.New("invalid match expression: " + orExpression)
		}
		expressionMatches, err := checkExpressionMatches(orExpression, properties)
		if err != nil {
			return false, err
		}
		// Just one or needs to match
		if expressionMatches {
			return true, nil
		}
	}
	return false, nil
}

func checkExpressionMatches(expression string, properties map[string]string) (bool, error) {
	var expressionMatches = false
	var err error = nil
	// check for the equals.
	if strings.Contains(expression, "==") {
		expressionMatches, err = matchEqualsExpression(expression, properties)
	} else if strings.Contains(expression, "!=") {
		expressionMatches, err = matchNotEqualsExpression(expression, properties)
		// check if not of a attribute
	} else if strings.HasPrefix(expression, "!") {
		expressionMatches, err = matchNotExpression(expression, properties)
		// check if is a boolean attribute
	} else {
		expressionMatches, err = checkExpressionResultsTrue(expression, properties)
	}
	if err != nil {
		return false, err
	}
	if expressionMatches {
		return true, nil
	}
	return false, nil
}

func matchEqualsExpression(expression string, properties map[string]string) (bool, error) {
	if !strings.Contains(expression, "==") {
		return false, errors.New("invalid equals match expression: " + expression)
	}
	equals := strings.Split(expression, "==")
	if len(equals) != 2 {
		return false, errors.New("invalid equals match expression: " + expression)
	}
	equalsLeft := strings.Trim(equals[0], " ")
	equalsRight := strings.Trim(equals[1], " \"")
	isLeftAlphaNumeric := regexp.MustCompile(`^[a-zA-Z0-9]*$`).MatchString(equalsLeft)
	isRightCorrect := regexp.MustCompile(`^[a-zA-Z0-9\.\-]*$`).MatchString(equalsLeft)
	if !isLeftAlphaNumeric || !isRightCorrect {
		return false, errors.New("invalid equals match expression: " + expression)
	}
	if val, ok := properties[equalsLeft]; ok {
		if len(equalsRight) > 0 {
			if val == equalsRight {
				return true, nil
			} else {
				return false, nil
			}
		}
		if len(equalsRight) == 0 {
			// handle numbers
			if numVal, err := strconv.ParseInt(val, 10, 64); err == nil {
				if numVal == 0 {
					return true, nil
				} else {
					return false, nil
				}
			} else {
				if val == equalsRight {
					return true, nil
				} else {
					return false, nil
				}
			}
		}
	} else {
		// Either empty or 0 if not found in the properties.
		if (equalsRight == "") || (equalsRight == "0") {
			return true, nil
		}
	}
	return false, nil
}

func matchNotEqualsExpression(expression string, properties map[string]string) (bool, error) {
	if !strings.Contains(expression, "!=") {
		return false, errors.New("invalid not equals match expression: " + expression)
	}
	notequals := strings.Split(expression, "!=")
	if len(notequals) != 2 {
		return false, errors.New("invalid not equals match expression: " + expression)
	}
	notequalsLeft := strings.Trim(notequals[0], " ")
	notequalsRight := strings.Trim(notequals[1], " \"")
	isLeftAlphaNumeric := regexp.MustCompile(`^[a-zA-Z0-9]*$`).MatchString(notequalsLeft)
	isRightCorrect := regexp.MustCompile(`^[a-zA-Z0-9\.\-]*$`).MatchString(notequalsLeft)
	if !isLeftAlphaNumeric || !isRightCorrect {
		return false, errors.New("invalid not equals match expression: " + expression)
	}
	if val, ok := properties[notequalsLeft]; ok {
		if len(notequalsRight) > 0 {
			if val != notequalsRight {
				return true, nil
			} else {
				return false, nil
			}
		}
		if len(notequalsRight) == 0 {
			if numVal, err := strconv.ParseInt(val, 10, 64); err == nil {
				if numVal != 0 {
					return true, nil
				} else {
					return false, nil
				}
			} else {
				if len(val) != 0 {
					return true, nil
				} else {
					return false, nil
				}
			}
		}
	}
	return false, nil
}

func matchNotExpression(expression string, properties map[string]string) (bool, error) {
	if !strings.Contains(expression, "!") {
		return false, errors.New("invalid not expression: " + expression)
	}
	removeNot := strings.Trim(expression, "!")
	isAlphaNumeric := regexp.MustCompile(`^[a-zA-Z0-9]*$`).MatchString(removeNot)
	// can only be a alpha numeric and cannot be a expression.
	if !isAlphaNumeric {
		return false, errors.New("invalid not expression: " + expression)
	}
	if val, ok := properties[removeNot]; ok {
		booleanValue, err := strconv.ParseBool(val)
		if err != nil {
			return false, errors.New("invalid not expression: " + expression)
		}
		if !booleanValue {
			return true, nil
		} else {
			return false, nil
		}
	}
	return false, nil
}

func checkExpressionResultsTrue(expression string, properties map[string]string) (bool, error) {
	isAlphaNumeric := regexp.MustCompile(`^[a-zA-Z0-9]*$`).MatchString(expression)
	// can only be a alpha numeric and cannot be a expression.
	if !isAlphaNumeric {
		return false, errors.New("invalid match expression: " + expression)
	}
	if val, ok := properties[expression]; ok {
		booleanValue, err := strconv.ParseBool(val)
		if err != nil {
			return false, errors.New("invalid match expression: " + expression)
		}
		if booleanValue {
			return true, nil
		}
	}
	return false, nil
}
