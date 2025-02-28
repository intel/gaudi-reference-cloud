package common

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	TimestampInfinityStr = "infinity"
	KErrUniqueViolation  = "23505"
)

func ResourceUniqueColumnAndValue(resourceId string, name string) (string, string, error) {
	if resourceId != "" {
		return "resource_id", resourceId, nil
	}
	if name != "" {
		return "name", name, nil
	}
	return "", "", status.Error(codes.InvalidArgument, "either ResourceId or Name must be provided")
}
