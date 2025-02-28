import { fetch_cloudaccounts } from "../libs/cloudaccounts.js";
import { getAdminToken } from "../../../../utils/HttpClientJF.js";
import { configData } from "../../../../utils/readJsonFileConfig.js";
import { textSummary } from "https://jslib.k6.io/k6-summary/0.0.2/index.js";
import { htmlReport } from "https://raw.githubusercontent.com/benc-uk/k6-reporter/main/dist/bundle.js";

// This will span 10 concurrent users and it will try to execute requests as many as time as possible in 60 seconds
let saturationpoint_scenario = {
    executor: "constant-vus",
    vus: configData.VUs,
    duration: '60s',
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
    "summaryCloudAccountsConstant.json": JSON.stringify(data),
    "summaryCloudAccountsConstant.html": htmlReport(data),
  };
}
