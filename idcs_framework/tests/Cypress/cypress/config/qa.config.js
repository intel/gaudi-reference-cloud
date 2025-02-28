const { defineConfig } = require("cypress");

module.exports = defineConfig({
  extends: "../../cypress.json",
  experimentalModifyObstructiveThirdPartyCode: true,
  chromeWebSecurity: false,
  defaultCommandTimeout: 80000,
  requestTimeout: 80000,
  responseTimeout: 200000,
  pageLoadTimeout: 50000,
  videoCompression: false,
  videoUploadOnPasses: false,
  viewportHeight: 660,
  viewportWidth: 1000,

  retries: {
    runMode: 2,
    openMode: 2,
  },

  env: {
    baseUrl: "",
    username: "",
    password: "",
  },
  /*
  e2e: {
    // We've imported your old cypress plugins here.
    // You may want to clean this up later by importing these.
    setupNodeEvents(on, config) {
      return require('/cypress/plugins/index.js')(on, config)
    },
  },
*/ e2e: {
    setupNodeEvents(on, config) {
      // implement node event listeners here
    },
  },
});