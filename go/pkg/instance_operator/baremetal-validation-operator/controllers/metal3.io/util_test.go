// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package validation

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	. "github.com/onsi/gomega"
)

func Test_sortBySubstringPtr(t *testing.T) {
	type args struct {
		arr    []*string
		substr string
	}
	oldImage := "ubuntu-22.04-pvc-metal-cloudimg-amd64-v20240104"
	newImage := "ubuntu-22.04-pvc-1550-metal-cloudimg-amd64-v20240129"
	invalidImage := "ubuntu-22.04-server-oneapi-amd64-latest"
	imageWithoutVersion1 := "ubuntu-22.04-pvc-metal-cloudimg-amd64-1"
	imageWithoutVersion2 := "ubuntu-22.04-pvc-metal-cloudimg-amd64-2"
	oldimageWithMultipleV := "ubuntu-v22.04-pvc-metal-cloudimg-amd64-v20240104"
	newimageWithMultipleV := "ubuntu-v22.04-pvc-1550-metal-cloudimg-amd64-v20240129"
	a := "z-key01"
	b := "a-key02"
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "test normal scenario",
			args: args{
				arr: []*string{
					&oldImage,
					&newImage,
					&invalidImage,
				},
				substr: "-v",
			},
			want: newImage,
		},
		{
			name: "test with no versions",
			args: args{
				arr: []*string{
					// &invalidImage,
					&imageWithoutVersion2,
					&imageWithoutVersion1,
				},
				substr: "-v",
			},
			want: imageWithoutVersion2,
		},
		{
			name: "test multiple substrings",
			args: args{
				arr: []*string{
					&oldimageWithMultipleV,
					&newimageWithMultipleV,
				},
				substr: "-v",
			},
			want: newimageWithMultipleV,
		},
		{
			name: "test multiple substrings",
			args: args{
				arr: []*string{
					&a,
					&b,
				},
				substr: "-key",
			},
			want: b,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := reverseSuffixSort(tt.args.arr, tt.args.substr); *got[0] != tt.want {
				t.Errorf("sortBySubstringPtr() = %v, want %v", *got[0], tt.want)
			}
		})
	}
}

func Test_readMetaFile(t *testing.T) {
	g := NewWithT(t)
	byteArrr, err := os.ReadFile("../../testdata/validation.meta")
	g.Expect(err).ShouldNot(HaveOccurred())
	inputString := strings.TrimSuffix(string(byteArrr), "\n")
	resultMap := convertToMap(inputString)
	g.Expect(len(resultMap)).To(Equal(5))
	g.Expect(resultMap["PYENV_SHELL"]).To(Equal("bash"))
	g.Expect(resultMap["TERM_PROGRAM_VERSION"]).To(Equal(""))

	//Verify it with an empty string
	resultMap = convertToMap("")
	g.Expect(len(resultMap)).To(Equal(0))
	g.Expect(resultMap["PYENV_SHELL"]).To(Equal(""))
}

func Test_isTimedOut(t *testing.T) {
	g := NewWithT(t)
	t0 := time.Now().UTC()
	t1 := t0.Add(-30 * time.Minute)
	// It is not timed out since the Timeout value is 31 mins
	g.Expect(isTimedOut(t1, 31)).To(Equal(false))
	g.Expect(isTimedOut(t1, 30)).To(Equal(true))
	g.Expect(isTimedOut(t1, 29)).To(Equal(true))
}

func Test_error(t *testing.T) {
	g := NewWithT(t)
	e1 := fmt.Errorf("error1")

	er := RetryableError(e1.Error())
	enr := NonRetryableError(e1.Error())
	g.Expect(IsRetryable(er)).To(Equal(true))
	g.Expect(IsNonRetryable(er)).To(Equal(false))
	g.Expect(IsRetryable(enr)).To(Equal(false))
	g.Expect(IsNonRetryable(enr)).To(Equal(true))

	erWrapped := fmt.Errorf("wrapping a retryable error %w", er)
	g.Expect(IsRetryable(erWrapped)).To(Equal(true))
	g.Expect(IsNonRetryable(erWrapped)).To(Equal(false))

	enrWrapped := fmt.Errorf("wrapping a non-retryable error %w", enr)
	g.Expect(IsNonRetryable(enrWrapped)).To(Equal(true))
	g.Expect(IsRetryable(enrWrapped)).To(Equal(false))
}
