import { test, chromium } from '@playwright/test';
import { ConsoleIDCService } from './lib/console.idcservice.net';
import * as utils from './lib/utils';
import { idcRegionsShorts, idcUsersShorts } from './lib/const';

import * as dotenv from 'dotenv';
dotenv.config()

// Input variables via .env or env variables (latter take precedence)
const ENVIRONMENT = String(process.env['IDC_environment']) ?? 'Staging'
const REGION = String(process.env['IDC_region']) ?? 'us-region-1'
const USER_TYPE = String(process.env['IDC_user_type']) ?? 'Standard'

function IDC00_TEST_PLAN (env: string, reg: string, utype: string) {
   console.log(`
      Scenario IDC00-UI:
          Login user in to ${env} IDC Console
          Switch to region ${reg}
          Verify Home page, with all links and buttons
          Verify Help menu, with all links
          Verify User Menu (incl. user ID, account type)
          Verify Cloud Credits page
          Verify Current month usage page
          Verify Account Settings page
          Verify Upgrade Account page
          Verify Hardware Catalog, counting products available to ${utype} user
          Verify Software Catalog
          Verify Compute service pages: Instances, Instance Groups, Load Balancers, SSH Keys
          Verify K8s service page
          Verify Supercomputing page
          Verify AI Playground page
          Verify Trainings page, count available trainings
          Sign out
      `)
}

// Test Matrix (for parameterized tests)
// if requested 'all-regions' or 'all-types' - then try all combinations
const testedRegions   = (REGION == 'all-regions') ? ((ENVIRONMENT == 'Production') ? ['us-region-1','us-region-2','us-region-3'] : ['us-staging-1','us-staging-3','us-staging-4','us-qa1-1']) : [REGION]
const testedUserTypes = (USER_TYPE == 'all-types') ? ['Standard', 'Premium'] : [USER_TYPE]

test.describe('IDC00-UI', () => {

  // IDC00 - Baseline UI test
  // this is parameterized test, name will always contain short region name, and short user type (eg, IDC00-reg1-std)
  for (const testedRegion of testedRegions) {
    for (const testedUserType of testedUserTypes) {

      test(`IDC00-${idcRegionsShorts[testedRegion]}-${idcUsersShorts[testedUserType]}`, async ({ browser }, testInfo) => {

        test.setTimeout(240000)  // 4-min timeout for this test
  
        utils.log(`${testInfo.title} - test started`)
        const startTime = new Date().getTime()
    
        IDC00_TEST_PLAN(ENVIRONMENT, testedRegion, testedUserType)

        const ENV_INFO = testedRegion.startsWith('us-qa') ? JSON.parse(String(process.env['IDC_URL']))['QA'] : JSON.parse(String(process.env['IDC_URL']))[ENVIRONMENT]

        const USER = JSON.parse(String(process.env['IDC_user']))[testedUserType]
    
        const page1 = new ConsoleIDCService(browser, ENVIRONMENT, ENV_INFO, testedRegion,
            USER.login, USER.password, testedUserType, `${testInfo.title}-${testInfo.retry}`)
        await page1.fullUITest()
  
        utils.log(`${testInfo.title} - test completed`)
        utils.log(`Execution time: ${Math.ceil((new Date().getTime() - startTime)/1000)} sec`)
      })

    }
  }
  
})
