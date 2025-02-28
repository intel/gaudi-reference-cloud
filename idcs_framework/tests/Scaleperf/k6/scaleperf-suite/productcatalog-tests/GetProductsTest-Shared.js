import { fetch_products } from "../productcatalog-libs/FetchProducts.js";
import { getCloudAccountId } from "../productcatalog-libs/FetchCloudAccount.js";
import { getAzureToken, getAdminToken } from "../../utils/HttpClientJF.js";
import { data, configData } from "../../utils/readJsonFileConfig.js";
import { textSummary } from "https://jslib.k6.io/k6-summary/0.0.2/index.js";
import { htmlReport } from "https://raw.githubusercontent.com/benc-uk/k6-reporter/main/dist/bundle.js";

// load the required payload
let payload = data[2];

// 10 concurrent users, execute 200 requests dividing amongst each other unequally
let saturationpoint_scenario = {
  executor: "shared-iterations",
  vus: configData.VUs,
  iterations: configData.iterations,
  maxDuration: configData.maxDuration,
};

export const options = {
  scenarios: { saturationpoint_scenario },
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
      productFilter: payload.productFilter,
    },
  };
  return custom_payload;
}

export default function (custom_payload) {
  fetch_products(custom_payload.token, custom_payload.payload);
}

export function teardown() {
  const d = new Date();
  console.log("Test execution ends at :" + d.getTime().toLocaleString());
}

export function handleSummary(data) {
  return {
    stdout: textSummary(data, { indent: " ", enableColors: true }),
    "summaryProductsShared.json": JSON.stringify(data),
    "summaryProductsShared.html": htmlReport(data),
  };
}
