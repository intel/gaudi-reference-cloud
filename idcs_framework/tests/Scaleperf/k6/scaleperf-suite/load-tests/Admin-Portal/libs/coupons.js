import { getCoupons, createCoupons } from "../../../../utils/HttpClientJF.js";
import { group } from "k6";

export function fetch_coupons(token) {
  group("Fetch coupons", function () {
    getCoupons(token);
  });
}

export function create_coupons(token, payload) {
    group("Create coupons", function () {
        createCoupons(token, payload);
    });
  }
  
