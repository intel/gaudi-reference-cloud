import { getMethod } from "../../utils/HttpClientJF.js";
import { group } from "k6";

const post_endpoint = "/v1/vendors";

export function fetch_vendors(token, payload) {
  group("Fetch vendors", function () {
    getMethod(post_endpoint, token);
  });
}
