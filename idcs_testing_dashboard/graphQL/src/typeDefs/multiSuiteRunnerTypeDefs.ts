export const typeDef = `

type testSummary {
  pass: Int
  fail: Int
  total: Int
}

type regressionTestSummary {
  id: ID
  test: String
  job: String
  env: String
  date: String
  buildNumber: Int
  build_url: String
  branch: String
  commit_sha: String
  commit_log: String
  metering: testSummary
  cloudAccount: testSummary
  productCatalog: testSummary
  billing: testSummary
  authZ: testSummary
  productCatalogE2E: testSummary
  total: testSummary
}

type RegressionTestProductCatalogLog {
  id: ID
  Time: String
  Action: String
  Package: String
  Output: String
  buildNumber: Int
}

type RegressionTestProductCatalogE2ELog {
  id: ID
  Time: String
  Action: String
  Package: String
  Output: String
  buildNumber: Int
}

type RegressionTestCloudAccountLog {
  id: ID
  Time: String
  Action: String
  Package: String
  Output: String
  buildNumber: Int
}

type RegressionTestBillingLog {
  id: ID
  Time: String
  Action: String
  Package: String
  Output: String
  buildNumber: Int
}

type RegressionTestAuthzLog {
  id: ID
  Time: String
  Action: String
  Package: String
  Output: String
  buildNumber: Int
}

type Query {
  RegressionTestSummaries: [regressionTestSummary!]
  RegressionTestProductCatalogLogs(buildNumber: Int): [RegressionTestProductCatalogLog!]
  RegressionTestProductCatalogE2ELogs(buildNumber: Int): [RegressionTestProductCatalogE2ELog!]
  RegressionTestCloudAccountLogs(buildNumber: Int): [RegressionTestCloudAccountLog!]
  RegressionTestBillingLogs(buildNumber: Int): [RegressionTestBillingLog!]
  RegressionTestAuthzLogs(buildNumber: Int): [RegressionTestAuthzLog!]
}
`;