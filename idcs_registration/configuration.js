export default {
    local: {
      name: "LOCAL",
      cdnHost: "http://localhost:8000",
      tenantID: "idcb2cdev.onmicrosoft.com",
      signupURL: "https://auth-dev.tiberaicloud.com/IDCB2CDEV.onmicrosoft.com/oauth2/v2.0/authorize?p=B2C_1A_LOCAL_NEW_SIGNUP&client_id=e4e51a15-c4c5-4102-89a9-918ce8297112&nonce=defaultNonce&redirect_uri=https%3A%2F%2Fjwt.ms%2F&scope=openid&response_type=id_token",
      consoleName: "Intel® Tiber™ AI Cloud"

    },
    development: {
      name: "DEVELOPMENT",
      cdnHost: "https://render-auth-dev.cloud.intel.com",
      tenantID: "idcb2cdev.onmicrosoft.com",
      signupURL: "https://auth-dev.tiberaicloud.com/IDCB2CDEV.onmicrosoft.com/oauth2/v2.0/authorize?p=B2C_1A_DEVELOPMENT_NEW_SIGNUP&client_id=e4e51a15-c4c5-4102-89a9-918ce8297112&nonce=defaultNonce&redirect_uri=https%3A%2F%2Fjwt.ms%2F&scope=openid&response_type=id_token&prompt=login",
      consoleName: "Intel® Tiber™ AI Cloud"
    },
    production: {
      name: "PRODUCTION",
      cdnHost: "https://render-auth.cloud.intel.com",
      tenantID: "idcb2cdev.onmicrosoft.com",
      signupURL: "https://goto/IDCB2CDEV",
      consoleName: "Intel® Tiber™ AI Cloud"
    }
  }
