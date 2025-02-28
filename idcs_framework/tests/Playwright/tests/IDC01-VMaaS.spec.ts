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

function IDC01_TEST_PLAN (env: string, reg: string, utype: string) {
   console.log(`
      Scenario IDC01-VMaaS:
          Login user in to ${env} IDC Console, ${reg}
          Perform pre-cleanup (instances and ssh keys)
          Open Compute Instances page
          Initiate creating small VM, with default options
          Create new SSH key, use it for new VM
          Verify that new VM is creating
          Verify that new VM is created (3-min timeout)
          Check all VM settings
          Verify connectivity via SSH
          Delete VM and SSH key (2-min timeout)
          Sign out
      `)
}

// Test Matrix (for parameterized tests)
// if requested 'all-regions' or 'all-types' - then try all combinations
// 12/6/2024 - removed us-region-1 from 'all-regions' scheduled tests
const testedRegions   = (REGION == 'all-regions') ? ((ENVIRONMENT == 'Production') ? ['us-region-2','us-region-3'] : ['us-staging-1','us-staging-3','us-staging-4','us-qa1-1']) : [REGION]
const testedUserTypes = (USER_TYPE == 'all-types') ? ['Standard', 'Premium'] : [USER_TYPE]

test.describe('IDC01-VMaaS', () => {

  // test.describe.configure({ retries: 0 })   // reduce load on the VMaaS system
  // Actually, in order to reduce monitoring noise, need to have retries (default is 1 retry)

  // IDC01 - VMaaS test
  // this is parameterized test, name will always contain short region name, and short user type (eg, IDC01-reg1-std)
  for (const testedRegion of testedRegions) {
    for (const testedUserType of testedUserTypes) {

      test(`IDC01-${idcRegionsShorts[testedRegion]}-${idcUsersShorts[testedUserType]}`, async ({ browser }, testInfo) => {
        
        test.setTimeout(360000)  // 6-min timeout for this test
  
        utils.log(`${testInfo.title} - test started`)
        const startTime = new Date().getTime()
    
        IDC01_TEST_PLAN(ENVIRONMENT, testedRegion, testedUserType)

        const ENV_INFO = testedRegion.startsWith('us-qa') ? JSON.parse(String(process.env['IDC_URL']))['QA'] : JSON.parse(String(process.env['IDC_URL']))[ENVIRONMENT]

        const USER = JSON.parse(String(process.env['IDC_user']))[testedUserType]
    
        const page1 = new ConsoleIDCService(browser, ENVIRONMENT, ENV_INFO, testedRegion,
            USER.login, USER.password, testedUserType, `${testInfo.title}-${testInfo.retry}`)
        
        await page1.VMaasTest()
  
        utils.log(`${testInfo.title} - test completed`)
        utils.log(`Execution time: ${Math.ceil((new Date().getTime() - startTime)/1000)} sec`)
      })

    }
  }
  
})
