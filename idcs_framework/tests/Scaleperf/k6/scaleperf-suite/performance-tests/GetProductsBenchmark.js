import { fetch_products } from "../productcatalog-libs/FetchProducts.js";
import { getCloudAccountId } from "../productcatalog-libs/FetchCloudAccount.js";
import { getAzureToken, getAdminToken, getAdminProducts } from "../../utils/HttpClientJF.js";
import { data, configData } from "../../utils/readJsonFileConfig.js";
import { textSummary } from "https://jslib.k6.io/k6-summary/0.0.2/index.js";
import { htmlReport } from "https://raw.githubusercontent.com/benc-uk/k6-reporter/main/dist/bundle.js";

// load the required payload
let payload = data[3];

export const options = {
  scenarios: {
    constant_request_rate: {
      executor: 'constant-arrival-rate',
      rate: 10,
      timeUnit: '1m', // 10 iterations per minute, i.e. 1000 RPS
      duration: '1m',
      preAllocatedVUs: 15, // how large the initial pool of VUs would be
      maxVUs: 20, // if the preAllocatedVUs are not enough, we can initialize more
    },
  },
};

export function setup() {
  const d = new Date();
  console.log("Test execution starts at :" + d.getTime().toLocaleString());
  let token = getAzureToken();
  let adminToken = getAdminToken();
  let custom_payload = {
    token: token,
    payload: {
      cloudaccountId: getCloudAccountId(
        token,
        adminToken,
        payload.email,
        payload.premium
      ),
    },
  };
  return custom_payload;
}

export default function (custom_payload) {
  getAdminProducts(custom_payload.token);
}

export function teardown() {
  const d = new Date();
  console.log("Test execution ends at :" + d.getTime().toLocaleString());
}

export function handleSummary(data) {
  return {
    stdout: textSummary(data, { indent: " ", enableColors: true }),
    "summaryProductsPerformance.json": JSON.stringify(data),
    "summaryProductsPerformance.html": htmlReport(data),
  };
}
