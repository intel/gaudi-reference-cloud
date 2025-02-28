const { defineConfig } = require("cypress");

module.exports = defineConfig({
  // These settings apply everywhere unless overridden
  experimentalModifyObstructiveThirdPartyCode: true,
  chromeWebSecurity: false,
  defaultCommandTimeout: 80000,
  requestTimeout: 70000,
  responseTimeout: 70000,
  pageLoadTimeout: 480000,
  video: false,
  videoCompression: false,
  viewportHeight: 1080,
  viewportWidth: 1920,

  // Viewport settings overridden for component tests
  component: {
    viewportWidth: 500,
    viewportHeight: 500,
  },

  retries: {
    runMode: 1,
    openMode: 2,
  },

  reporter: "mocha-allure-reporter",

  env: {
    baseUrl: "https://internal-placeholder.com/",
    adminUrl: "https://internal-placeholder.com/",
    globalEndpoint: "https://internal-placeholder.com/",
    accountType: "",
    // premium user
    puser: "",
    pupass: "",
    // standard user
    suser: "",
    supass: "",
    // Intel user
    iuser: "",
    iupass: "",
    // Second Intel user
    iuser2: "",
    iupass2: "",
    // Enterprise pending user
    epuser: "",
    eppass: "",
    // Enterprise user
    euser: "",
    epass: "",
    // Admin user
    adminuser: "",
    adminpass: "",
    // coupons and keys
    newCoupon: "",
    alreadyUsedCoupon: "",
    expiredCoupon: "",
    clientId: "",
    token: "",
    publicKey: "",
    // Mailslurp configuration parameters
    MAILSLURP_API_KEY: "",
    premiumAdminInboxId: "",
    enterpriseAdminInboxId: "",
    standardMemberInboxId: "",
    enterpriseMemberInboxId: "",
    standardMemberEmail: "",
    premiumMemberEmail: "",
    enterpriseMemberEmail: "",
    premium2MemberEmail: "",
    premMemberCloudAccount: ""
  },

  // Command timeout overridden for E2E tests
  e2e: {
    experimentalSkipDomainInjection: [
      "*.intel.com",
      "*.microsoftonline.com",
      "*.idcservice.net",
      "*.windows.net",
      "*.microsoft.com",
      "*.microsoftonline-p.com",
      "*.msftauth.net",
      "*.recaptcha.net",
      "*.gstatic.com",
      "*.google.com",
      "*.jsdelivr.net",
      "*.applicationinsights.azure.com",
      "*.googlesyndication.com",
      "ftp.dfp.microsoft.com",
      "geolocation.onetrust.com",
    ],
    experimentalRunAllSpecs: true,
    // We've imported your old cypress plugins here.
    // You may want to clean this up later by importing these.
    setupNodeEvents(on, config) {
      // return require('./cypress/plugins/index.js')(on, config)
      // require("cypress-mochawesome-reporter/plugin")(on);
      // require("cypress-grep/src/plugin")(config);
      on("task", {
        log(message) {
          console.log(message);
          return null;
        },
      });
    },
  },
});
