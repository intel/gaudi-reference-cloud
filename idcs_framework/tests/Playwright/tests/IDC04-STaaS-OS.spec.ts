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

function IDC03_TEST_PLAN (env: string, reg: string, utype: string) {
   console.log(`
      Scenario IDC04-STaaS-OS:
          Login user in to ${env} IDC Console, ${reg}
          Perform pre-cleanup (instances and ssh keys)
          Open Compute Instances page
          Initiate creating small VM, with default options
          Create new SSH key, use it for new VM
          Verify that new VM is created with 3-min timeout
          
          Perform STaas pre-cleanup
          Create Object Storage (90-sec timeout)
          Create Principal, generate credentials
          Verify SSH connectivity
          Make remote AWS CLI install (6-min timeout)
          Make remote bucket verifications (via aws-cli)

          Delete Object Storage in UI
          Delete VM and SSH key in UI
          Sign out
      `)
}

// Test Matrix (for parameterized tests)
// if requested 'all-regions' or 'all-types' - then try all combinations
const testedRegions   = (REGION == 'all-regions') ? ((ENVIRONMENT == 'Production') ? ['us-region-1','us-region-2','us-region-3'] : ['us-staging-1','us-staging-3','us-staging-4','us-qa1-1']) : [REGION]
const testedUserTypes = (USER_TYPE == 'all-types') ? ['Premium'] : [USER_TYPE]

test.describe('IDC04-STaas-OS', () => {

  // test.describe.configure({ retries: 0 })   // default is 1 retry

  // this is parameterized test, name will always contain short region name, and short user type (eg, IDC01-reg1-std)
  for (const testedRegion of testedRegions) {
    for (const testedUserType of testedUserTypes) {

      test(`IDC04-${idcRegionsShorts[testedRegion]}-${idcUsersShorts[testedUserType]}`, async ({ browser }, testInfo) => {
        
        test.setTimeout(720000)  // 12-min timeout for this test
  
        utils.log(`${testInfo.title} - test started`)
        const startTime = new Date().getTime()
    
        IDC03_TEST_PLAN(ENVIRONMENT, testedRegion, testedUserType)

        const ENV_INFO = testedRegion.startsWith('us-qa') ? JSON.parse(String(process.env['IDC_URL']))['QA'] : JSON.parse(String(process.env['IDC_URL']))[ENVIRONMENT]

        const USER = JSON.parse(String(process.env['IDC_user']))[testedUserType]
    
        const page1 = new ConsoleIDCService(browser, ENVIRONMENT, ENV_INFO, testedRegion,
            USER.login, USER.password, testedUserType, `${testInfo.title}-${testInfo.retry}`)
        
        await page1.STaasOSTest()
  
        utils.log(`${testInfo.title} - test completed`)
        utils.log(`Execution time: ${Math.ceil((new Date().getTime() - startTime)/1000)} sec`)
      })

    }
  }
  
})
