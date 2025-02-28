import { enroll_user, get_cloudaccount_id } from "../../utils/HttpClientJF.js";

export function enroll(token, premium) {
  payload = { premium: premium };
  let result = enroll_user(token, payload);
  return result;
}

export function getCloudAccountId(token, adminToken, email, premium) {
  let cloudaccount_response = get_cloudaccount_id(adminToken, email);
  if (get_cloudaccount_id.status == 404) {
    enroll(token, premium);
    let cloudaccount_response = get_cloudaccount_id(adminToken, email);
    return cloudaccount_response.id;
  }
  return cloudaccount_response.id;
}
