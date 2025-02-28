## CYPRESS IDC UI AUTOMATION

## HOW TO RUN CYPRESS LOCALLY

1. cd to idcs_framework\tests\Cypress    
2. Install required packages by running:  npm --proxy http://internal-placeholder.com:912 install
3. Make sure to update [cypress.config.js](./cypress.config.js) and [fixtures\IDC1.0\login.json](./cypress/fixtures/IDC1.0/login.json) with the relevant users credentials, env variables and environment URLs
4. Make sure proxy settings are present in [.npmrc file](/.npmrc)  - see ".npmrc configurations" at the end of this file.
5. Start Cypress runner with command:  npx cypress open --env testTags='["Standard","StandardAll"]'
6. Choose desired browser and spec file for e2e testing. 

* NOTE: 
(Running with other user account types)
npx cypress open --env testTags='["Intel","IntelAll"]'
npx cypress open --env testTags='["Standard","StandardlAll"]'
npx cypress open --env testTags='["Premium","PremiumAll"]'
npx cypress open --env testTags='["Enterprise","EnterpriseAll"]'

testTags are used for filtering and to avoid test cases duplication, as same TCs may apply for different user account types.

## Running cypress in CLI mode using: npx cypress run

* Running an individual spec in CLI:

npx cypress run --env testTags='["Intel","IntelAll"]' --spec cypress/e2e/IDC1.0/compute/computeFlow.cy.js

* Running all specs in a folder CLI:

npx cypress run --env testTags='["Intel","IntelAll"]' --spec cypress/e2e/IDC1.0/compute/**
npx cypress run --env testTags='["Intel","IntelAll"]' --spec cypress/e2e/IDC1.0/**/**

* Running specific browser (chrome, firefox, edge) and/or headed mode in CLI (cypress run command uses headless mode as default):

npx cypress run --env testTags='["Intel","IntelAll"]' --spec cypress/e2e/IDC1.0/compute/computeFlow.cy.js --browser chrome --headed

For more command line options refer to Cypress official docs [https://docs.cypress.io/guides/guides/command-line]

## General Instructions for new spec files

Use the following syntax for your specs

```js
import TestFilter from '../../support/testFilter';
TestFilter(["IntelAll", "PremiumAll", "StandardAll"], () => {
  describe('Your suite descrition', () => {

        beforeEach(() => {
          cy.PrepareSession()
          cy.GetSession()
        });

        afterEach(() => {
          cy.TestClean()
        })

        after(() => {
          cy.TestClean()
        })

        it('1001 | Test 1 description', function() {        
           Your test code
        });
        
        it('1002 | Test 2 description', function() {        
           Your test code
        });

        it('1003 | Test 3 description', function() {        
           Your test code
        });
  })
})

```

* This is the minimum you need to call in order to authenticate and run the tests.
* The order of methods presented in the code example is very important.
* The its are just example you need to remove it and change by yours.

## Proxy settings

For Windows, you must set this system variable and restart vscode:

NO_PROXY=intel.com,.intel.com,localhost,127.0.0.1,msftauth.net

## .npmrc configurations

A common cypress login failure reaching Intel sites like "consumer.intel.com", are caused by using a wrong proxy configuration.
The following [.npmrc file](/.npmrc) sample configuration will help to avoid proxy issues.

* Use the following setting for AWS hosted environments i.e dev3, Staging and Prod
registry=https://registry.npmjs.org
http_proxy=http://internal-placeholder.com:912
https_proxy=http://internal-placeholder.com:912
NO_PROXY=localhost,127.0.0.1,msftauth.net

* Use the following settings for other test environments i.e dev1, dev7, dev8, dev10
registry=https://registry.npmjs.org
http_proxy=http://internal-placeholder.com:912
https_proxy=http://internal-placeholder.com:912
NO_PROXY=localhost,127.0.0.1,msftauth.net

## Troubleshoot

* To avoid errors on runner when working locally, do not change the code while the login process is in progress on a spec. If you do it and get an "URL is null" something error, just exit the runner and start again from shell.

