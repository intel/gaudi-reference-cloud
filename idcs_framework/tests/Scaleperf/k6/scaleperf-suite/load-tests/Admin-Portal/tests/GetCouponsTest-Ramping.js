import { fetch_coupons } from "../libs/coupons.js";
import { getAdminToken } from "../../../../utils/HttpClientJF.js";
import { configData } from "../../../../utils/readJsonFileConfig.js";
import { textSummary } from "https://jslib.k6.io/k6-summary/0.0.2/index.js";
import { htmlReport } from "https://raw.githubusercontent.com/benc-uk/k6-reporter/main/dist/bundle.js";

// 0 to 20 in 30 seconds, down to 10 in 20 seconds
let saturationpoint_scenario = {
    executor: "ramping-vus",
    startVUs: 0,
    stages: [
      { duration: configData.rampMaxDuration, target: configData.maxVU },
      { duration: configData.rampMinDuration, target: configData.minVU },
    ],
    gracefulRampDown: "10s",
    gracefulStop: "30s"
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
    fetch_coupons(admin_token);
}

export function teardown() {
  const d = new Date();
  console.log("Test execution ends at :" + d.getTime().toLocaleString());
}

export function handleSummary(data) {
  return {
    stdout: textSummary(data, { indent: " ", enableColors: true }),
    "summaryCouponsRamping.json": JSON.stringify(data),
    "summaryCouponsRamping.html": htmlReport(data),
  };
}
