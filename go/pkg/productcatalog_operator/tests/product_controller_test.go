// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package tests

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/productcatalog_operator/apis/private.cloud/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func NewNamespace(namespace string) *v1.Namespace {
	return &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}
}

func NewProduct(namespace string, productName string, productId string, vendorId string, familyId string, description string, eccn string, pcq string, matchexpr string) *v1alpha1.Product {
	return &v1alpha1.Product{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "private.cloud.intel.com/v1alpha1",
			Kind:       "Product",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      productName,
			Namespace: namespace,
		},
		Spec: v1alpha1.ProductSpec{
			ID:          productId,
			VendorID:    vendorId,
			FamilyID:    familyId,
			Description: description,
			ECCN:        eccn,
			PCQ:         pcq,
			MatchExpr:   matchexpr,
			Rates:       []v1alpha1.ProductRate{},
			Metadata:    []v1alpha1.ProductMetadata{},
		},
		Status: v1alpha1.ProductStatus{
			State: v1alpha1.ProductStateUndetermined,
		},
	}
}

var _ = Describe("product catalog tests", func() {

	ctx := context.Background()
	Context("Product Catalog tests", func() {
		It("Product catalog operator test", func() {

			namespace := uuid.NewString()
			productName := uuid.NewString()

			productId := "product-id"
			vendorId := "vendorid"
			familyId := "familyid"
			description := "Xeon4 large"
			eccn := "eccn2"
			pcq := "pcq2"
			matchxxpr := "matchexpr"

			product := NewProduct(namespace, productName, productId, vendorId, familyId, description, eccn, pcq, matchxxpr)

			nsObject := NewNamespace(namespace)

			By("Creating namespace successfully")
			Expect(k8sClient.Create(ctx, nsObject)).Should(Succeed())

			By("Creating product")
			Expect(k8sClient.Create(ctx, product)).Should(Succeed())

			By("Fetching the product")
			obj := types.NamespacedName{Name: productName, Namespace: namespace}
			prod := &v1alpha1.Product{}
			Eventually(func(g Gomega) {
				err := k8sClient.Get(ctx, obj, prod)
				g.Expect(err).Should(BeNil())
			}, timeout, time.Millisecond*500).Should(Succeed())

		})
	})
})
