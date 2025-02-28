package testsetup

import (
	"github.com/tidwall/gjson"
	"strings"
)

func GetRegions() []string {
	data := gjson.Get(ConfigData, "regions").String()
	trimmedStr := strings.Trim(data, "[]")
	return strings.Split(trimmedStr, ",")
}