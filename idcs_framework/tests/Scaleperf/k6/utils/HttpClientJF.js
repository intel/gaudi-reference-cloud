import { check } from "k6";
import http from "k6/http";
import { configData } from "./readJsonFileConfig.js";

const baseUrl = configData.base_url;

export function postMethod(post_endpoint, token, payload, randomString) {
  const headers_value = {
    Authorization: String(token),
    "Content-Type": "application/json",
  };

  console.log("payload", payload);

  const response = http.post(baseUrl + post_endpoint, JSON.stringify(payload), {
    headers: headers_value,
  });
  console.log("CODE....." + response.status);
  console.log("RESPONSE....." + response.body);
  check(response, {
    "status code is 200": (res) => res.status === 200,
  });
}

export function getMethod(get_endpoint, token) {
  const headers_value = {
    Authorization: String(token),
    "Content-Type": "application/json",
  };
  const response = http.get(baseUrl + get_endpoint, {
    headers: headers_value,
  });
  console.log("endpoint response code: " + response.status);
  console.log("endpoint response body: " + response.body);
  check(response, {
    "status code is 200": (res) => res.status === 200,
  });
}

export function enroll_user(token, payload) {
  const enroll_endpoint = "/v1/cloudaccounts/enroll";

  let response = postMethod(enroll_endpoint, token, payload);
  return response;
}

export function get_cloudaccount_id(token, email) {
  const cloudaccount_endpoint = "/v1/cloudaccounts/name/";
  let url = baseUrl + cloudaccount_endpoint;

  const headers_value = {
    Authorization: String(token),
    "Content-Type": "application/json",
  };
  const response = http.get(url + email, {
    headers: headers_value,
  });

  console.log("CLOUDACCOUNT ID", response.body);

  return response.json();
}

export function getAzureToken() {
  let token = __ENV.AZURE_TOKEN;
  console.log("token...", token);
  return token;
}

export function getAdminToken() {
  let token = __ENV.AZURE_ADMIN_TOKEN;
  console.log("Admin token...", token);
  return token;
}

export function getCloudaccountMetadata(token, cloudAccountId) {
  const cloudaccount_endpoint = "/v1/cloudaccounts/id/" + cloudAccountId;
  let url = baseUrl + cloudaccount_endpoint;

  const headers_value = {
    Authorization: String(token),
    "Content-Type": "application/json",
  };
  const response = http.get(url + email, {
    headers: headers_value,
  });

  console.log("CLOUDACCOUNT data: ", response.body);
  return response.json();
}

export function getUsages(token, cloudAccountId) {
  const cloudaccount_endpoint = "/v1/billing/usages?cloudAccountId" + cloudAccountId;
  let url = baseUrl + cloudaccount_endpoint;

  const headers_value = {
    Authorization: String(token),
    "Content-Type": "application/json",
  };
  const response = http.get(url + email, {
    headers: headers_value,
  });

  console.log("Usages data: ", response.body);
  return response.json();
}

export function getCredits(token, cloudAccountId) {
  const cloudaccount_endpoint = "/v1/cloudcredits/credit?cloudAccountId" + cloudAccountId;
  let url = baseUrl + cloudaccount_endpoint;

  const headers_value = {
    Authorization: String(token),
    "Content-Type": "application/json",
  };
  const response = http.get(url + email, {
    headers: headers_value,
  });

  console.log("Usages data: ", response.body);
  return response.json();
}

export function getAdminProducts(token) {
  const cloudaccount_endpoint = "/v1/products/admin";
  let url = baseUrl + cloudaccount_endpoint;

  const headers_value = {
    Authorization: String(token),
    "Content-Type": "application/json",
  };
  const response = http.get(url + email, {
    headers: headers_value,
  });

  console.log("products data: ", response.body);
  return response.json();
}

export function getMembers(token, email) {
  const cloudaccount_endpoint = "/v1/cloudaccounts/name/" + email + "/members";
  let url = baseUrl + cloudaccount_endpoint;

  const headers_value = {
    Authorization: String(token),
    "Content-Type": "application/json",
  };
  const response = http.get(url + email, {
    headers: headers_value,
  });

  console.log("members data: ", response.body);
  return response.json();
}

export function getInvoices(token, cloudAccountId) {
  const cloudaccount_endpoint = "/v1/billing/invoices?cloudAccountId=" + cloudAccountId;
  let url = baseUrl + cloudaccount_endpoint;

  const headers_value = {
    Authorization: String(token),
    "Content-Type": "application/json",
  };
  const response = http.get(url + email, {
    headers: headers_value,
  });

  console.log("invoices data: ", response.body);
  return response.json();
}

export function getCoupons(token){
  const coupons_endpoint = "/v1/billing/coupons";
  let url = baseUrl + coupons_endpoint;

  const headers_value = {
   Authorization: String(token),
    "Content-Type": "application/json",
  };
  const response = http.get(url, {
    headers: headers_value,
  });

  console.log("coupons data: ", response.body);
  return response.json();
}

export function createCoupons(token, payload){
  const cloudaccount_endpoint = "/v1/billing/invoices?cloudAccountId=" + cloudAccountId;
  let url = baseUrl + cloudaccount_endpoint;

  let response = postMethod(url, token, payload);
  return response.json();
}

export function createMeteringRecord(token, payload){

}

export function getMeteringRecord(token, payload){

}

export function getCloudaccounts(token){
  const cloudaccounts_endpoint = "/v1/cloudaccounts";
  let url = baseUrl + cloudaccounts_endpoint;

  const headers_value = {
    Authorization: String(token),
    "Content-Type": "application/json",
  };
  const response = http.get(url, {
    headers: headers_value,
  });

  console.log("Cloudaccounts data: ", response.body);
  return response.json();
}