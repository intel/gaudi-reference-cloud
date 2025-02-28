import { ApolloServer } from '@apollo/server';
import { startStandaloneServer } from '@apollo/server/standalone';
import { readFileSync } from 'fs';
import { MongoClient } from 'mongodb';
import { typeDef as ginkgoTypeDefs } from './typeDefs/ginkgoTypeDefs.js';
import { typeDef as multiSuiteRunnerTypeDefs } from './typeDefs/multiSuiteRunnerTypeDefs.js';
import { typeDef as usagesE2ETypeDefs } from './typeDefs/usagesE2ETypeDefs.js';



let db = null


const resolvers = {
  Query: {
    RegressionTestSummaries: async () => {
      return await db.collection('regressionTestSummary').find().toArray()
    },
    UsagesE2ETestSummaries: async () => {
      return await db.collection('E2ETestUsagesSummary').find().toArray()
    },
    E2ETestUsagesPremiumUserLogs: async () => {
      return await db.collection('E2ETestUsagesPremiumUserLog').find().toArray()
    },
    E2ETestUsagesStandardUserLogs: async () => {
      return await db.collection('E2ETestUsagesStandardUserLog').find().toArray()
    },
    E2ETestUsagesEnterpriseUserLogs: async () => {
      return await db.collection('E2ETestUsagesEnterpriseUserLog').find().toArray()
    },
    E2ETestUsagesIntelUserLogs: async () => {
      return await db.collection('E2ETestUsagesIntelUserLog').find().toArray()
    },
    RegressionTestProductCatalogLogs: async (_, {buildNumber}) => {
      return await db.collection('RegressionTestProductCatalogLog').find({ buildNumber: buildNumber.toString()}).toArray()
    },
    RegressionTestProductCatalogE2ELogs: async (_, {buildNumber}) => {
      return await db.collection('RegressionTestProductCatalogE2ELog').find({ buildNumber: buildNumber.toString()}).toArray()
    },
    RegressionTestCloudAccountLogs: async (_, {buildNumber}) => {
      return await db.collection('RegressionTestCloudAccountLog').find({ buildNumber: buildNumber.toString()}).toArray()
    },
    RegressionTestBillingLogs: async (_, {buildNumber}) => {
      return await db.collection('RegressionTestBillingLog').find({ buildNumber: buildNumber.toString()}).toArray()
    },
    RegressionTestAuthzLogs: async (_, {buildNumber}) => {
      return await db.collection('RegressionTestAuthzLog').find({ buildNumber: buildNumber.toString()}).toArray()
    },
  },
};


interface MyContext {
  authScope?: String;
}

const start = async () => {
  const uri = readFileSync('../../local/secrets/TestingDashboardsMongoDBConnectionString.txt', 'utf-8');
  //linux: /etc/ssl/certs/ca-certificates.crt
  const ca_certificate = '/etc/ssl/certs/ca-certificates.crt'
  const client = new MongoClient(uri, { tls: true, tlsCAFile: ca_certificate})
  await client.connect()
  db = client.db("Testing")
  const server = new ApolloServer<MyContext>({ typeDefs: [ ginkgoTypeDefs, multiSuiteRunnerTypeDefs, usagesE2ETypeDefs], resolvers });
  const { url } = await startStandaloneServer(server, {
    listen: { port: 4000 },
  });
  console.log(`🚀  Server ready at: ${url}`);
}

start()