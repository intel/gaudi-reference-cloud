export const typeDef = `
type testSummary {
  pass: Int
  fail: Int
  total: Int
}

type testsSummary {
  id: ID
  test: String
  job: String
  env: String
  buildNumber: Int
  build_url: String
  branch: String
  commit_sha: String
  commit_log: String
  enterpriseUsageE2E: testSummary
  intelUsageE2E: testSummary
  standardUsageE2E: testSummary
  premiumUsageE2E: testSummary
  total: testSummary
}

extend type Query {
  UsagesE2ETestSummaries: [testsSummary!]
  E2ETestUsagesPremiumUserLogs: [ginkgoLog!]
  E2ETestUsagesStandardUserLogs: [ginkgoLog!]
  E2ETestUsagesEnterpriseUserLogs: [ginkgoLog!]
  E2ETestUsagesIntelUserLogs: [ginkgoLog!]
}
`;