import { getCloudaccounts } from "../../../../utils/HttpClientJF.js";
import { group } from "k6";

export function fetch_cloudaccounts(token) {
  group("Fetch coupons", function () {
    getCloudaccounts(token);
  });
}