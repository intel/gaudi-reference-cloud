import { fetch_cloudaccounts } from "../libs/cloudaccounts.js";
import { getAdminToken } from "../../../../utils/HttpClientJF.js";
import { configData } from "../../../../utils/readJsonFileConfig.js";
import { textSummary } from "https://jslib.k6.io/k6-summary/0.0.2/index.js";
import { htmlReport } from "https://raw.githubusercontent.com/benc-uk/k6-reporter/main/dist/bundle.js";

let saturationpoint_scenario = {
    executor: "shared-iterations",
    vus: configData.VUs,
    iterations: configData.iterations,
    maxDuration: '60s',
    gracefulStop: "10s"
};
  
export const options = {
    scenarios: { saturationpoint_scenario },
};

export function setup() {
  const d = new Date();
  console.log("Test execution starts at :" + d.getTime().toLocaleString());
  let adminToken = getAdminToken();
  return adminToken
}

export default function (admin_token) {
    fetch_cloudaccounts(admin_token);
}

export function teardown() {
  const d = new Date();
  console.log("Test execution ends at :" + d.getTime().toLocaleString());
}

export function handleSummary(data) {
  return {
    stdout: textSummary(data, { indent: " ", enableColors: true }),
    "summaryCloudAccountsShared.json": JSON.stringify(data),
    "summaryCloudAccountsShared.html": htmlReport(data),
  };
}
