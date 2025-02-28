import { test, chromium } from '@playwright/test';
import { Devcloud } from './lib/devcloud.intel.com';
import * as utils from './lib/utils';

import * as dotenv from 'dotenv';
dotenv.config()

// Input variables via .env or env variables
const URL  = String(process.env['test1_URL'])
const user = JSON.parse(String(process.env['test1_user']))

test.describe('OneAPI-Sign-In', () => {

    test('TEST_01', async () => {
      
      test.setTimeout(60000);

      utils.log('TEST_01 - test started');
      const browser = await chromium.launch()
      const startTime = new Date().getTime();
  
      const TEST_PLAN = `
          Scenario 01:
          Login user in to devcloud.intel.com/oneapi/
          Verify user ID after login
          Log out
      `;
      console.log(TEST_PLAN);
  
      const page1 = new Devcloud(browser, URL, user.login, user.name, user.password);
      await page1.login();
      await page1.verifyLoggedUser();
      await utils.sleep(1000);
      await page1.logout();
  
      utils.log(`TEST_01 - test completed`);
      utils.log(`Execution time: ${Math.ceil((new Date().getTime() - startTime)/1000)} sec`);
  
      //browser.close();
    })
  
  })