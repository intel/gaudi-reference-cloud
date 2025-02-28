import { postMethod } from "../../utils/HttpClientJF.js";
import { group } from "k6";

const post_endpoint = "/v1/products";

export function fetch_products(token, payload) {
  let randomString = "";
  group("Fetch products", function () {
    postMethod(post_endpoint, token, payload, randomString);
  });
}
