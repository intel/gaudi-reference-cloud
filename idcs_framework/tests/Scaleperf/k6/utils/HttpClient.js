import { check } from "k6";
import http from "k6/http";

const baseUrl = "https://internal-placeholder.com";

export function postMethod(post_endpoint, token, payload, randomString) {
  const headers_value = {
    Authorization: "Bearer " + String(token),
  };

  const response = http.post(baseUrl + post_endpoint, JSON.stringify(payload), {
    headers: headers_value,
  });
  //console.log(response.status)
  //console.log(response)
  check(response, {
    "status code is 201": (res) => res.status === 201,
    //'response validation': (res) => res.body.includes("load-vm-"+randomString)
  });
}

export function getMethod(get_endpoint, token, getInstance) {
  const headers_value = {
    Authorization: "Bearer " + String(token),
  };

  const response = http.get(baseUrl + get_endpoint + "/" + getInstance, {
    headers: headers_value,
  });
  //console.log(response.body)
  check(response, {
    "status code is 200": (res) => res.status === 200,
    //'response validation': (res) => res.body().includes(getInstance)
  });
  return response;
}

export function putMethod(put_endpoint, token, payload) {}

export function deleteMethod(delete_endpoint, token, deleteInstance) {
  const headers_value = {
    Authorization: "Bearer " + String(token),
  };

  const response = http.del(
    baseUrl + delete_endpoint + "/" + deleteInstance,
    null,
    { headers: headers_value }
  );
  check(response, {
    "status code is 200": (res) => res.status === 200,
  });
}
