// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/manageddb"
	db "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/database"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	sqlDb     *sql.DB
	txDb      *sql.Tx
	testDb    *manageddb.TestDb
	managedDb *manageddb.ManagedDb
)

func TestQuery(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Query Suite")
}

var _ = BeforeSuite(func() {
	ctx := context.Background()
	By("Starting database")
	testDb = &manageddb.TestDb{}
	var err error
	managedDb, err = testDb.Start(ctx)
	Expect(err).Should(Succeed())
	Expect(managedDb).ShouldNot(BeNil())
	Expect(managedDb.Migrate(ctx, db.Fsys, "migrations")).Should(Succeed())
	sqlDb, err = managedDb.Open(ctx)
	Expect(err).Should(Succeed())
	Expect(sqlDb).ShouldNot(BeNil())
})

var _ = AfterSuite(func() {
	sqlDb.Close()
})
