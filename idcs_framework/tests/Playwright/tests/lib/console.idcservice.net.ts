import { test, Page, expect, Browser, Locator, BrowserContext } from "@playwright/test";
import * as utils from './utils';
import { idcProductMatrix } from './const';

import * as util from 'util';
import * as childProcess from 'child_process';
import exp from "constants";
const exec = util.promisify(childProcess.exec).bind(childProcess);


const SMALL_SCREEN_SIZE = { width: 1280, height: 720 }

export class ConsoleIDCService {
    browser: Browser;
    context!: BrowserContext;
    page!: Page;
    userlogin: string;
    password: string;
    usertype: string;
    env: string;
    envInfo: object;
    url: string;
    region: string;
    testId: string;
    sshPubKey: string;
    sshConnectLine: string;
    sshConnectInitialized: boolean;
    staasFsUser: string;
    staasFsPass: string;
    staasOsAccessKey: string;
    staasOsSecretKey: string;
    staasOsFullName: string;

    locSiteNavBar: Locator
    locSiteToolbar: Locator
    locSideNavPanel: Locator
    locNavigationExpandButton: Locator
    locNavigationCollapseButton: Locator
    
    constructor (browser: Browser, env: string, envInfo: object, region: string, userlogin: string, password: string, usertype: string, testId: string) {
        this.browser = browser
        this.userlogin = userlogin
        this.password = password
        this.usertype = usertype
        this.env = env
        this.envInfo = envInfo      // { "url": "https://console.idcservice.net", "doSshValidation": true, "sshSocksProxy": "internal-placeholder.com:1080" }
        this.region = region
        this.testId = testId
    }

    private log(msg: string) {
        console.log(new Date().toISOString() + ` ${this.testId}: ${msg}`)
    }

    // can't run async functions in constructor
    // don't forget to call this method after instantiation
    async createPage() {
        this.context = await this.browser.newContext(
            {
                recordHar: { path: `har/${this.testId}.har`, urlFilter: /consumer.intel.com|api.idcservice.net|eglb.intel.com/ },
                // recordVideo: { dir: 'videos/', size: { width: 1920, height: 1080 } }
            }
        );
        this.page = await this.context.newPage();
        if (process.env['SCREEN'] == 'small') {
            this.page.setViewportSize(SMALL_SCREEN_SIZE);
        }

        // some common locators
        this.locSiteNavBar       = this.page.locator('nav.siteNavbar')
        this.locSiteToolbar      = this.page.locator('nav.siteToolbar')
        this.locSideNavPanel     = this.page.locator('div[intc-id=SideBarNavigationMain]')
        this.locNavigationExpandButton   = this.locSiteToolbar.getByLabel('Expand side menu')
        this.locNavigationCollapseButton = this.locSiteToolbar.getByLabel('Collapse side menu')
    }

    // * * * * * * * * * * * * * * * * * * * * * * * * * *

    // IDC00-UI
    public async fullUITest() {
        try {
            await this.login()
            await this.selectRegion()
  
            await this.verifyHomePage()
            await this.verifyHelpMenu()

            // Billing
            await this.verifyUserMenu()
            await this.verifyCloudCreditsPage()
            await this.verifyCurrentMonthUsagePage()
            await this.verifyAccountSettingsPage()
            await this.verifyUpgradeAccountPage()
            
            // Catalog
            await this.verifyHardwareCatalogPage()
            await this.verifySoftwareCatalogPage()

            // Compute
            await this.verifyComputeInstancesPage()
            await this.verifyInstanceGroupsPage()
            // await this.verifyLoadBalancersPage()     - TODO
            await this.verifyKeysPage()

            await this.verifyK8sServicePage()

            // await this.verifySupercomputingPage()    - TODO

            // IDC Preview
            await this.verifyPreviewCatalog()
            // Rest of Preview tabs are not available to external users
            // They will be verified in separate IDC02-Preview provisioning test

            // Storage
            // await this.verifyFileStoragePage()
            // await this.verifyObjectStoragePage()

            // await this.verifyAIPlaygroundPage() - not exposed yet
  
            await this.verifyLearningPage()

            // TODO: verify Documentation ext. link
  
            await this.signOut()
        }
        finally {
            await this.context?.close()     // required to save HAR
        }
    }

    // IDC01-VMaas
    public async VMaasTest() {
        try {
            await this.login()
            await this.selectRegion()
  
            await this.VMaasPreCleanup()
            await this.VMaasCreate()
            await this.VMaasVerify()
            await this.VMaasDelete()
            await this.VMaasDeleteKey()
  
            await this.signOut()
        }
        finally {
            await this.context?.close()     // required to save HAR
        }
    }

    // IDC02-Preview
    public async PreviewTest() {
        try {
            await this.login()
            // await this.selectRegion()
  
            await this.PreviewPreCleanup()
            await this.PreviewCreate()
            await this.PreviewVerify()
            await this.PreviewDelete()
            await this.PreviewDeleteKey()
  
            await this.signOut()
        }
        finally {
            await this.context?.close()     // required to save HAR
        }
    }

    // IDC03-STaaS-FS
    public async STaasFSTest() {
        try {
            await this.login()
            await this.selectRegion()

            // do it first, to settle before fs creation
            await this.StaasFsPreCleanup()
  
            // dedicated compute instance
            await this.VMaasPreCleanup('idc03-staas')
            await this.VMaasCreate('idc03-staas')

            await this.StaasFsCreate()
            await this.StaasFsVerify()
            await this.StaasFsDelete()

            // compute cleanup
            await this.VMaasDelete('idc03-staas')
            await this.VMaasDeleteKey('idc03-staas')
  
            await this.signOut()
        }
        finally {
            await this.context?.close()     // required to save HAR
        }
    }

    // IDC04-STaaS-OS
    public async STaasOSTest() {
        try {
            await this.login()
            await this.selectRegion()

            // do it first, because principal deletion sometimes takes time
            await this.StaasOsPreCleanup()
  
            // dedicated compute instance
            await this.VMaasPreCleanup('idc04-staas')
            await this.VMaasCreate('idc04-staas')
            await this.VMaasVerify('idc04-staas')   // required to setup ssh parameters

            await this.StaasOsCreate()
            await this.StaasOsVerify()
            await this.StaasOsDelete()

            // compute cleanup
            await this.VMaasDelete('idc04-staas')
            await this.VMaasDeleteKey('idc04-staas')
  
            await this.signOut()
        }
        finally {
            await this.context?.close()     // required to save HAR
        }
    }

    // ===================================================================================================
    
    public async login() {
        await test.step('Login to IDC Console', async () => {
            this.log(`user ${this.userlogin} tries to login..`)
            const loginStart = Date.now()
            
            await this.createPage()
            const res = await this.page.goto(this.envInfo['url'], { waitUntil: 'domcontentloaded' })
            expect(res?.status(), `Login: trying to receive HTTP 200OK (received ${res?.status()})`).toBe(200)
            
            // EMail input page
            await expect(this.page.locator('#signInName'), 'Login: trying to open Email input page in 20s').toBeVisible({timeout: 20000})
            await this.page.locator('#signInName').fill(this.userlogin)
            await this.page.locator('#continue').click()

            await this.dismissCookieDialog()    // Optional cookie dialog
            
            // Password input page
            const signInPasswInput = this.page.locator('#password')
            await expect(signInPasswInput, 'Login: trying to open Password input page in 20s').toBeVisible({timeout: 20000})
            await signInPasswInput.fill(this.password)
            await this.page.locator('#continue').click()
            
            // Redirect back to IDC Console
            await utils.waitForObjectDisappearance(signInPasswInput, 20000, 'SSO login dialog')
            await this.page.waitForLoadState('domcontentloaded')

            await utils.waitForObject(this.page.locator('#root > nav.navbar > span[intc-id=toolbarNavBrand]'),
                20000, 'redirect from SSO to Console UI')
                
            this.log(`logged-in successfully, user login delay was ${Date.now() - loginStart} msec`)

            await this.dismissCookieDialog()    // Optional cookie dialog
        })
    }

    private async dismissCookieDialog() {
        // Optional cookie dialog
        const cookieDismiss = this.page.locator('button#onetrust-reject-all-handler')
        await utils.waitForObject(cookieDismiss, 3000, '', true)  // tested with 5 sec initially, but it's not observed anymore
        if (await cookieDismiss.isVisible())
            await cookieDismiss.click()
    }

    // Select region (default is this.region)
    public async selectRegion(region: string = '') {
        await test.step('Select Region', async () => {
            if (region)
                this.region = region
            this.log(`selecting ${this.region} region..`)

            await expect(this.page.locator('#dropdown-header-region-toggle')).toBeVisible()
            
            if (await this.page.locator(`#dropdown-header-region-toggle:has-text("${this.region}")`).count()) {
                this.log(`region ${this.region} was already selected active`)
            } 
            else {
                if (! await this.page.locator('#dropdown-header-menu-region').isVisible())
                    await this.page.locator('#dropdown-header-region-toggle').click()
                const linkToRegion = this.page.getByRole('link', { name: `Connect to ${this.region} Region` })
                await expect(linkToRegion, `Region ${this.region} is not available in drop-down`).toBeVisible()
                await linkToRegion.click()
                await expect(this.page.locator('#dropdown-header-region-toggle')).toContainText(this.region)

                this.log(`region ${this.region} was selected successfully`)
            }
        })
    }

    public async verifyUserMenu() {
        await test.step('Verify User Menu', async () => {
            this.log('checking User Menu drop-down')
            if (! await this.page.locator('#dropdown-header-user-menu').isVisible())
                await this.page.getByLabel('User Menu').click()

            await expect(this.page.locator('#dropdown-header-user-menu').locator('span.badge')).toHaveText(this.usertype)
            await expect(this.page.locator('#dropdown-header-user-menu').getByText(this.userlogin)).toBeVisible()
            await expect(this.page.locator('#dropdown-header-user-menu').getByRole('link', { name: 'Account Settings ID:' })).toBeVisible()
            
            if (this.usertype === 'Premium') {
                // Invoices
                this.log('checking User Menu / Invoices')
                await this.page.locator('#dropdown-header-user-menu').getByRole('link', { name: 'Invoices' }).click()
                await expect(this.page.getByRole('main').getByRole('heading', { name: 'Invoices', exact: true })).toBeVisible()
                await expect(this.page.getByRole('main').getByRole('heading', { name: 'Invoice History' })).toBeVisible()
                // there should be some past invoices for Premium user
                const tbl = this.page.getByRole('main').locator('table.table > tbody').locator('tr')
                await expect(tbl.first()).toBeVisible({timeout: 8000})  // at least one invoice should exist
                // TODO: analyze invoices table?

                // back to User Menu
                if (! await this.page.locator('#dropdown-header-user-menu').isVisible())
                    await this.page.getByLabel('User Menu').click()

                // Payment Methods
                this.log('checking User Menu / Payment Methods')
                await this.page.locator('#dropdown-header-user-menu').getByRole('link', { name: 'Payment Methods' }).click()
                await expect(this.page.getByRole('main').getByRole('heading', { name: 'Manage Payment Methods' })).toBeVisible()
                await expect(this.page.getByRole('main').locator('div.card').nth(0).getByRole('heading', { name: 'Credit Card', exact: true })).toBeVisible()
                await expect(this.page.getByRole('main').locator('div.card').nth(1).getByRole('heading', { name: 'Cloud Credits', exact: true })).toBeVisible()
                await expect(this.page.getByRole('main').locator('div.card').nth(1).getByLabel('View Credit Details')).toBeVisible()
                await expect(this.page.getByRole('main').locator('div.card').nth(1).getByLabel('Redeem Coupon')).toBeVisible()
                
                /////// Credit card adding button is removed in UI 2.0
                // await expect(this.page.getByRole('button', { name: 'Add card' })).toBeVisible({timeout: 10000})
                // await this.page.getByRole('button', { name: 'Add card' }).click()
                
                // // check Credit Card input form
                // this.log('checking Add a credit card form')
                // await expect(this.page.getByRole('main').getByRole('heading', { name: 'Add a credit card' })).toBeVisible()
                // await this.page.getByPlaceholder('Card number').fill('4305 7911 0567 7593')
                // await this.page.getByPlaceholder('MM').fill('01')
                // await this.page.getByPlaceholder('YY').fill('30')
                // await this.page.getByLabel('CVC *').fill('123')
                // await this.page.getByLabel('First Name *').fill('John')
                // await this.page.getByLabel('Last Name *').fill('Smith')
                // await this.page.getByLabel('Email *').fill('john.smith@somemail.com')
                // await this.page.getByLabel('Company name').fill('Best AI in Universe')
                // await this.page.getByLabel('Phone').fill('9251112222')
                // await this.page.getByLabel('Country *').selectOption('US')
                // await this.page.getByLabel('Address line 1 *').fill('123 Main Str.')
                // await this.page.getByLabel('City *').fill('San Jose')
                // await this.page.getByLabel('State *').selectOption('CA')
                // await this.page.getByLabel('ZIP code *').fill('94500')
                // await expect(this.page.getByText('Add CardCancel')).toBeVisible()
                // await expect(this.page.getByRole('button', { name: 'Cancel' })).toBeVisible()
                // this.log('form Add a credit card is verified')
            }

            // back to User Menu
            if (! await this.page.locator('#dropdown-header-user-menu').isVisible())
                await this.page.getByLabel('User Menu').click()
            
            // Current month usage link (page will be verified in separate function)
            this.log('checking User Menu / Current month usage')
            let link = this.page.locator('#dropdown-header-user-menu').getByRole('link', { name: 'Current month usage' })
            await expect(link).toBeVisible()
            expect(await link.getAttribute('href')).toEqual('/billing/usages')
            
            if (this.usertype === 'Standard') {
                // Upgrade Account link (page will be verified in separate function)
                this.log('checking User Menu / Upgrade Account')
                link = this.page.locator('#dropdown-header-user-menu').getByRole('link', { name: 'Upgrade Account' })
                await expect(link).toBeVisible()
                expect(await link.getAttribute('href')).toEqual('/upgradeaccount')
                
                // Cloud Credits link (page will be verified in separate function)
                this.log('checking User Menu / Cloud Credits')
                link = this.page.locator('#dropdown-header-user-menu').getByRole('link', { name: 'Cloud Credits' })
                await expect(link).toBeVisible()
                expect(await link.getAttribute('href')).toEqual('/billing/credits')
            }

            await expect(this.page.locator('#dropdown-header-user-menu').getByRole('button', { name: 'Sign-out' })).toBeVisible()

            this.log('User Menu drop-down is verified')
        })
    }

    public async verifyCurrentMonthUsagePage() {
        await test.step('Verify Current Month Usage page', async () => {
            this.log('checking User Menu / Current month usage')

            await this.clickUserMenuLink('Current month usage')
            if (this.usertype == 'Standard') {
                await this.verifyTopHeader('Billing', 
                        [{name: 'Usage', link: '/billing/usages'}, {name: 'Cloud Credits', link: '/billing/credits'}], 
                        'Usage'
                )
            } 
            else if (this.usertype == 'Premium') {
                await this.verifyTopHeader('Billing', 
                        [{name: 'Usage', link: '/billing/usages'}, {name: 'Cloud Credits', link: '/billing/credits'},
                         {name: 'Invoices', link: '/billing/invoices'}, {name: 'Payment Methods', link: '/billing/managePaymentMethods'}], 
                        'Usage'
                )
            }

            // consider APi delays
            const noUsage = this.page.getByRole('main').locator('div.sheet').getByRole('heading', { name: 'No usages found', exact: true })
            const hasUsage = this.page.getByRole('main').locator('div.sheet').getByRole('heading', { name: 'Current month usage', exact: true })
            await expect(noUsage.or(hasUsage)).toBeVisible({timeout: 10000})
            if (await hasUsage.isVisible()) {
                await expect(hasUsage).toBeVisible({timeout: 10000})
                await expect(this.page.getByRole('main').locator('table.table')).toBeVisible({timeout: 10000})   
            }

            // Check Learning bar (top-right)
            this.log('checking Home Page / Learning bar')
            const siteToolbDocumentationLink = this.page.locator('nav.siteToolbar').getByRole('button', { name: 'Open learning bar' })
            await expect(siteToolbDocumentationLink, 'top right Documentation link must be visible').toBeVisible()
            await siteToolbDocumentationLink.click()
            const learnNavBar = this.page.locator('div[intc-id="LearningBarNavigationMain"]')
            await expect(learnNavBar).toBeVisible()
            await expect(learnNavBar.getByRole('heading', { name: 'Documentation' }).first()).toBeVisible()
            await expect(learnNavBar.getByLabel('Learn Billing and account').first()).toBeVisible()
            await expect(learnNavBar.getByLabel('Learn Invoices and Usage').first()).toBeVisible()
            await expect(learnNavBar.getByLabel('Learn Payment Methods').first()).toBeVisible()
            await expect(learnNavBar.getByLabel('Learn Account Type Payments').first()).toBeVisible()
            await expect(learnNavBar.getByLabel('Learn Multi-user accounts').first()).toBeVisible()
            await learnNavBar.getByRole('button').click()

            this.log('Current month usage page is verified')
        })
    }

    public async verifyCloudCreditsPage() {
        await test.step('Verify Cloud Credits page', async () => {
            this.log('checking User Menu / Cloud Credits')

            // for Standard user this page is available via User menu -> Cloud Credits
            if (this.usertype === 'Standard') {
                await this.clickUserMenuLink('Cloud Credits')
                await this.verifyTopHeader('Billing', 
                    [{name: 'Usage', link: '/billing/usages'}, {name: 'Cloud Credits', link: '/billing/credits'}], 
                    'Cloud Credits'
                )
            }
            else // for Premium user this page is available via User menu -> Payment Methods -> button View Credit Details
            if (this.usertype === 'Premium') {
                await this.clickUserMenuLink('Payment Methods')
                let btn = this.page.getByRole('main').locator('div.card').nth(1).getByLabel('View Credit Details')
                await expect(btn).toBeVisible({timeout: 10000})
                await btn.click()
                await this.verifyTopHeader('Billing', 
                    [{name: 'Usage', link: '/billing/usages'}, {name: 'Cloud Credits', link: '/billing/credits'},
                     {name: 'Invoices', link: '/billing/invoices'}, {name: 'Payment Methods', link: '/billing/managePaymentMethods'}], 
                    'Cloud Credits'
                )
            }

            await expect(this.page.getByRole('main').getByRole('heading', { name: 'Cloud Credits', exact: true })).toBeVisible()
            // API call delay
            const tt = this.page.locator('table.table > tbody')
            await expect(tt).toBeVisible({timeout: 10000})
            
            // check balance (warning if <= $40)
            const ccBalance = await this.page.getByRole('main').locator('span[intc-id=remainingCreditsLabel]').textContent()
            const ccb = parseFloat((ccBalance ? ccBalance : '0').replace('$','').replace(',',''))
            // this will go to prom metrics
            this.log(`estimated remaining credit: ${ccb}`)
            if (ccb <= 30)
                this.log(`Warning: estimated remaining credit is low: ${ccb}`)
            
            // check expiration (warning if < 30 days left, consider only non-expired coupons)
            let activeCouponDaysRemaining = 0
            const rwc = await tt.locator('tr').count()
            for (let i=0; i<rwc; i++) {
                let exps = await tt.locator('tr').nth(i).locator('td').nth(2).textContent()
                exps  = exps ? exps : ''
                exps  = exps.substring(0, exps.indexOf(' '))

                let dp = Date.parse(exps)
                let td = Date.now()
                let dr = Math.floor((dp - td)/(24*3600*1000))
                
                if (dr <= 0) {
                    this.log(`found one expired coupon: ${exps} - skipping it..`)
                    continue
                }

                this.log(`found non-expired coupon: ${exps} (expiring in ${dr} days)`)
                if (dr < 30)  // 30 days
                    this.log(`Warning: coupon is expiring in less than 30 days: ${exps}`)
                activeCouponDaysRemaining = (activeCouponDaysRemaining > dr) ? activeCouponDaysRemaining : dr
            }
            // this will go to prom metrics
            this.log(`best available coupon: expiring in ${activeCouponDaysRemaining} days`)
            if (activeCouponDaysRemaining <= 0)
                this.log(`Warning: could not find non-expired coupons!`)

            this.log('checking page Redeem Coupon')
            await expect(this.page.getByRole('link', { name: 'Redeem coupon' })).toBeVisible()
            await this.page.getByRole('link', { name: 'Redeem coupon' }).click();
            await expect(this.page.getByRole('main').getByRole('heading', { name: 'Redeem coupon' })).toBeVisible()
            await this.page.getByPlaceholder('Coupon Code').fill('blabla')
            await expect(this.page.getByRole('button', { name: 'Redeem' })).toBeVisible()
            this.log('Redeem Coupon page is verified')

            this.log('Cloud Credits page is verified')
        })
    }

    public async verifyAccountSettingsPage() {
        await test.step('Verify Account Settings page', async () => {
            this.log('checking User Menu / Account Settings')
            
            await this.clickUserMenuLink('Account Settings ID:')
            await this.verifyTopHeader('Account Settings')          // this is old. TODO: switch to new
            // await this.verifyTopHeader('Account Settings', 
            //     [{name: 'My Information', link: '/profile/accountsettings'}], 'My Information'
            // )

            // TODO: restore after migration
            // await expect(this.page.getByRole('heading', { name: 'Your AI Cloud Account' })).toBeVisible()
            await expect(this.page.locator('#planTypeInfo').getByText(this.usertype)).toBeVisible()

            await expect(this.page.getByRole('heading', { name: 'Your intel.com Account' })).toBeVisible()
            await expect(this.page.getByRole('main').getByText(this.userlogin)).toBeVisible()

            if (this.usertype === 'Standard') {
                this.log('checking Account Settings / Upgrade to premium button')
                await this.page.getByLabel('Upgrade to premium').click()
                // await this.page.waitForURL(`${this.envInfo['url']}/upgradeaccount`, {timeout: 3000});
                await expect(this.page.getByRole('main').getByRole('heading', { name: 'Upgrade your account' })).toBeVisible()
                // Upgrade Account page will be verified in separate function
            }

            if (this.usertype === 'Premium') {
                // this is only in Staging now! Two tabs and "Grant access" in Members tab
                if (await this.page.getByRole('link', { name: 'Members', exact: true }).isVisible()) {  // TODO: remove condition
                    await this.verifyTopHeader('Account Settings', 
                        [{name: 'My Information', link: '/profile/accountsettings'}, {name: 'Members', link: '/profile/accountAccessManagement'}], 
                        'My Information')

                    // switch to Members tab
                    await this.page.getByRole('link', { name: 'Members', exact: true }).click()
                    await this.verifyTopHeader('Account Settings', 
                        [{name: 'My Information', link: '/profile/accountsettings'}, {name: 'Members', link: '/profile/accountAccessManagement'}], 
                        'Members')
                }

                await expect(this.page.getByRole('heading', { name: 'Account Access Management' })).toBeVisible()
                await expect(this.page.getByRole('button', { name: 'Grant access' })).toBeVisible({timeout: 10000})

                this.log('checking Account Settings / Grant access form')
                await this.page.getByRole('button', { name: 'Grant access' }).click()
                await expect(this.page.getByRole('dialog').getByText('Grant access to cloud account ID:')).toBeVisible()
                await this.page.getByRole('dialog').getByPlaceholder('example@domain.com').fill('test@gmail.com')
                await this.page.getByRole('dialog').locator('input[type="date"]').fill('2030-01-01')
                await this.page.getByRole('dialog').getByPlaceholder('Note for user').fill('test')
                await expect(this.page.getByRole('dialog').getByRole('button', { name: 'Cancel' })).toBeVisible()
                await expect(this.page.getByRole('dialog').getByRole('button', { name: 'Grant', exact: true })).toBeVisible()
                await this.page.getByRole('dialog').getByLabel('Close').click()
            }

            this.log('Account Settings page is verified')
        })

        // TODO: verify Documentation links
        // ...
    }

    public async verifyUpgradeAccountPage() {
        await test.step('Verify Upgrade Account page', async () => {
            this.log('checking User Menu / Upgrade Account')

            // Only showing for Standard user
            if (this.usertype === 'Standard') {

                await this.clickUserMenuLink('Upgrade Account')
                // await this.verifyTopHeader('AI Cloud')   // TODO: restore and verify after migration
                await expect(this.page.getByRole('main').getByRole('heading', { name: 'Upgrade your account' })).toBeVisible()

                // Coupon code form
                this.log('checking Upgrade Account / Coupon code form')
                await expect(this.page.getByRole('button', { name: 'Coupon code' })).toBeVisible()
                await this.page.getByRole('button', { name: 'Coupon code' }).click()
                await expect(this.page.getByRole('main').getByPlaceholder('Coupon Code')).toBeVisible()
                await this.page.getByRole('main').getByPlaceholder('Coupon Code').fill('123')
                await expect(this.page.getByRole('main').getByRole('button', { name: 'Redeem' })).toBeVisible()
                await expect(this.page.getByRole('main').getByRole('button', { name: 'Cancel' })).toBeVisible()

                // // Credit card form - button is removed?
                // this.log('checking Upgrade Account / Credit Card form')
                // await expect(this.page.getByRole('button', { name: 'Credit Card' })).toBeVisible()
                // await this.page.getByRole('button', { name: 'Credit Card' }).click()
                // await expect(this.page.locator('input[intc-id=CardnumberInput]')).toBeVisible()
                // await expect(this.page.locator('input[intc-id=MonthInput]')).toBeVisible()
                // await expect(this.page.locator('input[intc-id=YearInput]')).toBeVisible()
                // await expect(this.page.locator('input[intc-id=CVCInput]')).toBeVisible()
                // await expect(this.page.locator('input[intc-id=FirstNameInput]')).toBeVisible()
                // await expect(this.page.locator('input[intc-id=LastNameInput]')).toBeVisible()
                // await expect(this.page.locator('select[intc-id=CountrySelect]')).toBeVisible()
                // await expect(this.page.locator('input[intc-id=Addressline1Input]')).toBeVisible()
                // await expect(this.page.locator('input[intc-id=CityInput]')).toBeVisible()
                // await expect(this.page.locator('select[intc-id=StateSelect]')).toBeVisible()
                // await expect(this.page.locator('input[intc-id=ZIPcodeInput]')).toBeVisible()
                // await expect(this.page.getByPlaceholder('Card number')).toBeVisible()
                // await expect(this.page.getByText('Add CardCancel')).toBeVisible()

                this.log('Upgrade Account page is verified')
            }
            else
                this.log('- SKIPPED. Only available for Standard users')
        })
    }

    // Main Home page
    public async verifyHomePage() {
        await test.step('Verify Home Page', async () => {
            this.log('checking Navigation / Home')

            await this.clickNavigationLink('Go to Home Page')
            await this.verifyTopHeader('Home')

            // // Check Onboarding/Deploy diagram (depends on API)
            // let onboardingDiag = this.page.locator('#root >> div.card').nth(0).locator('div.stepContainer')
            // await expect(onboardingDiag.first(), 'Onboarding diagram should be visible').toBeVisible({timeout: 10000})
            // await utils.sleep(500)
            // expect(await onboardingDiag.count(), 'Onboarding diagram should show at least 4 steps').toBeGreaterThan(4)

            // Dashboard cards
            this.log('checking Home Page / Main cards')
            const dcards = this.page.locator('#root >> div.card')
            // Delayed (spinner) due to API queries
            // await expect(dcards.nth(4).locator('.d-flex > div > div.flex-even:has(h2:has-text("Compute"))').first().locator('div.dashboard-item:nth-child(2) > div:has(span:has-text("Instances"))')).toBeVisible({ timeout: 10000 })
            await expect(dcards.nth(4).getByText('Instance groups').nth(2)).toBeVisible({ timeout: 10000 })
            // await expect(dcards.nth(5).getByText('Current Month').nth(0)).toBeVisible({ timeout: 10000 })
            await expect(this.page.getByText('Current Month Usage').nth(3)).toBeVisible({ timeout: 10000 })
            expect(await dcards.count(), 'Expecting at least 6 segments on dashboar').toBeGreaterThanOrEqual(6)
            
            // #1 card
            await expect(dcards.nth(0).getByRole('heading', { name: 'Learn Free' })).toBeVisible()
            await expect(dcards.nth(0).getByLabel('Get started to Learn tab')).toBeVisible()
            await dcards.nth(0).getByLabel('Get started to Learn tab').click()
            // Learn page
            await expect(this.page.getByRole('main').getByRole('heading', { name: /^Learn/  })).toBeVisible({ timeout: 10000 })
            let checkBtn
            // // old:
            // checkBtn = this.page.getByRole('main').getByLabel('Connect to a GPU')
            // await expect(checkBtn).toBeVisible()
            // new:
            checkBtn = this.page.getByRole('main').getByRole('button', { name: 'Connect now' })
            await expect(checkBtn).toBeVisible()
            await checkBtn.click()
            await expect(this.page.getByRole('main').getByRole('button', { name: 'AI Accelerator' })).toBeVisible()
            await expect(this.page.getByRole('main').getByRole('button', { name: 'GPU' })).toBeVisible()

            checkBtn = this.page.getByRole('main').getByLabel('Learning', { exact: true })
            await expect(checkBtn).toBeVisible()
            expect(await checkBtn.getAttribute('href')).toEqual('/learning/notebooks')
            checkBtn = this.page.getByRole('main').getByLabel('Documentation', { exact: true })
            await expect(checkBtn).toBeVisible()
            expect(await checkBtn.getAttribute('href')).toEqual('/docs/index.html')
            const backHome = this.page.getByRole('main').getByRole('link', { name: 'Home', exact: true })
            await backHome.click()
            
            // #2 card
            await expect(dcards.nth(1).getByRole('heading', { name: 'Evaluate' })).toBeVisible()
            await expect(dcards.nth(1).getByLabel('Get started to Evaluate tab')).toBeVisible()
            await dcards.nth(1).getByLabel('Get started to Evaluate tab').click()
            await expect(this.page.getByRole('main').getByRole('heading', { name: 'Evaluate', exact: true })).toBeVisible({ timeout: 10000 })
            checkBtn = this.page.getByRole('main').getByLabel('Request Instance').first()
            await expect(checkBtn).toBeVisible()
            expect(await checkBtn.getAttribute('href')).toEqual('/preview/hardware?fctg=CPU')
            checkBtn = this.page.getByRole('main').getByLabel('Request Instance').nth(1)
            await expect(checkBtn).toBeVisible()
            expect(await checkBtn.getAttribute('href')).toEqual('/preview/hardware?fctg=GPU%2CAI+Accelerator')
            await backHome.click()
            
            // #3 card
            await expect(dcards.nth(2).getByRole('heading', { name: 'Deploy' })).toBeVisible()
            await expect(dcards.nth(2).getByLabel('Get started to Deploy tab')).toBeVisible()
            await dcards.nth(2).getByLabel('Get started to Deploy tab').click()
            await expect(this.page.getByRole('main').getByRole('heading', { name: 'Deploy', exact: true })).toBeVisible({ timeout: 10000 })
            checkBtn = this.page.getByRole('main').getByLabel('Select Instance').first()
            await expect(checkBtn).toBeVisible()
            expect(await checkBtn.getAttribute('href')).toEqual('/hardware?fctg=GPU%2CAI')
            checkBtn = this.page.getByRole('main').getByLabel('Launch Cluster')
            await expect(checkBtn).toBeVisible()
            expect(await checkBtn.getAttribute('href')).toEqual('/cluster/reserve')
            await backHome.click()

            // #4 card
            await expect(dcards.nth(3).getByRole('heading', { name: 'Recently Visited' })).toBeVisible()
            
            /// Following cards depend on API (compute, billing, etc) and show up with delay
            // #5 card
            await expect(dcards.nth(4).getByRole('heading', { name: 'Compute', exact: true })).toBeVisible({ timeout: 10000 })
            // await expect(dcards.nth(4).locator('.d-flex > div > div > div:nth-child(2)').first()).toBeVisible({ timeout: 10000 })
            await expect(dcards.nth(4).getByText('Instance groups').nth(2)).toBeVisible({ timeout: 10000 })
            await expect(dcards.nth(4).getByRole('heading', { name: 'Storage' })).toBeVisible()
            await expect(dcards.nth(4).getByRole('heading', { name: 'Intel Kubernetes' })).toBeVisible()
            await expect(dcards.nth(4).getByRole('heading', { name: 'Preview' })).toBeVisible()
            // await expect(dcards.nth(4).getByRole('heading', { name: 'Supercomputing' })).toBeVisible()
            
            // #6 card
            await expect(dcards.nth(6).getByRole('heading', { name: 'Usage' })).toBeVisible({ timeout: 10000 })
            await expect(dcards.nth(6).getByText('Current Month').first()).toBeVisible({ timeout: 10000 })
            await expect(dcards.nth(6).getByText('Remaining Credits').first()).toBeVisible()
            
            // Check Learning bar (top-right) // intc-id=btn-toolbar-learning
            this.log('checking Home Page / Learning bar')
            const siteToolbDocumentationLink = this.locSiteToolbar.getByRole('button', { name: 'Open learning bar' })
            await expect(siteToolbDocumentationLink, 'top right Documentation link must be visible').toBeVisible()
            await siteToolbDocumentationLink.click()
            const learnNavBar = this.page.locator('div[intc-id="LearningBarNavigationMain"]')
            await expect(learnNavBar).toBeVisible()
            await expect(learnNavBar.getByRole('heading', { name: 'Documentation' }).first()).toBeVisible()
            await expect(learnNavBar.getByLabel('Learn Get started with the Intel').first()).toBeVisible()
            await expect(learnNavBar.getByLabel('Learn Customer support and community').first()).toBeVisible()
            // await expect(learnNavBar.getByLabel('Learn Overview').first()).toBeVisible()
            // await expect(learnNavBar.getByLabel('Learn Launch Learning').first()).toBeVisible()
            await expect(learnNavBar.getByLabel('Learn Launch Instance').first()).toBeVisible()
            await learnNavBar.getByRole('button').click()

            // Footer links (check URLs?)
            this.log('checking Home Page / Footer links')
            await expect(this.page.locator('div.footer').getByRole('link', { name: '© Intel Corporation' })).toBeVisible()
            await expect(this.page.locator('div.footer').getByRole('link', { name: 'Send Feedback' })).toBeVisible()
            await expect(this.page.locator('div.footer').getByRole('link', { name: 'Terms of Use' })).toBeVisible()
            await expect(this.page.locator('div.footer').getByRole('link', { name: '*Trademarks' })).toBeVisible()
            await expect(this.page.locator('div.footer').getByRole('link', { name: 'Cookies' })).toBeVisible()
            await expect(this.page.locator('div.footer').getByRole('link', { name: 'Privacy', exact: true })).toBeVisible()
            await expect(this.page.locator('div.footer').getByRole('link', { name: 'Supply Chain Transparency' })).toBeVisible()
            await expect(this.page.locator('div.footer').getByRole('link', { name: 'Site Map' })).toBeVisible()
            await expect(this.page.locator('div.footer').getByRole('link', { name: 'Your Privacy Choices' })).toBeVisible()
            await expect(this.page.locator('div.footer').getByRole('link', { name: 'Notice at Collection' })).toBeVisible()

            // Check left side navigation menu
            this.log('checking Home Page / Navigation menu')
            if (! await this.locSideNavPanel.isVisible())
                await this.locNavigationExpandButton.click()
            await expect(this.locSideNavPanel.getByLabel('Go to Home Page')).toBeVisible()
            await expect(this.locSideNavPanel.getByLabel('Go to Catalog Page')).toBeVisible()
            await expect(this.locSideNavPanel.getByLabel('Go to Compute Page')).toBeVisible()
            await expect(this.locSideNavPanel.getByLabel('Go to Preview Page')).toBeVisible()
            await expect(this.locSideNavPanel.getByLabel('Go to Storage Page')).toBeVisible()
            // await expect(this.locSideNavPanel.getByLabel('Go to AI Playground Page')).toBeVisible()
            await expect(this.locSideNavPanel.getByLabel('Go to Learning Page')).toBeVisible()
            await expect(this.locSideNavPanel.getByLabel('Go to Documentation Page')).toBeVisible()
            await this.locNavigationCollapseButton.click()

            this.log('Home Page is verified')
        })
    }

    public async verifyHardwareCatalogPage() {
        await test.step('Verify Hardware Catalog page', async () => {
            this.log('checking Navigation / Catalog / Hardware')

            await this.clickNavigationLink('Go to Catalog Page', 'Go to Hardware Page')
            await this.verifyTopHeader('Catalog', 
                [{name: 'Hardware', link: '/hardware'}, {name: 'Software', link: '/software'}], 
                'Hardware')
            await expect(this.page.getByRole('main').getByRole('heading', { name: 'Available hardware' })).toBeVisible()

            // Instead of exact match to matrix values per region, let's check minimum availability (no less than)
            const reg = (this.env == 'Staging') ? 'us-staging-minimum' : 'us-prod-minimum'

            // API delay
            const noInst = this.page.getByRole('main').locator('div.text-center').getByRole('heading', { name: 'No available services' }) // TODO: verify
            const instTbl = this.page.getByRole('main').locator('div.card-body').first()
            await expect(noInst.or(instTbl)).toBeVisible({timeout: 12000})
            await utils.sleep(500)

            // Total (no filter)
            // await expect(this.page.getByRole('main').getByLabel('Toggle filter Core compute')).toBeVisible()
            await expect(this.page.getByRole('main').getByLabel('Toggle filter all products')).toBeVisible()
            let pcount = await this.page.getByRole('main').locator('div.card-body').count()
            expect (pcount, 'Number of Total available products is wrong')
                .toBeGreaterThanOrEqual(idcProductMatrix[this.usertype][reg]['total'])
            this.log(`available products - hardware - Total: ${pcount}`)
            
            // Non-existent filter
            await this.page.getByRole('main').getByPlaceholder('Type to search...').fill('blabla')
            pcount = await this.page.getByRole('main').locator('div.card-body').count()
            expect (pcount, 'With non-existent filter - available products should be zero').toEqual(0)

            await this.page.getByRole('main').getByPlaceholder('Type to search...').fill('')
            
            // Filter = type_core_compute
            pcount = 0
            if (await this.page.getByLabel('Toggle filter Core compute').isVisible()) {
                await this.page.getByLabel('Toggle filter Core compute').click()
                pcount = await this.page.getByRole('main').locator('div.card-body').count()
                if (idcProductMatrix[this.usertype][reg]['type_core_compute'])
                    expect (pcount, 'Number of Core Compute products is wrong')
                        .toBeGreaterThanOrEqual(idcProductMatrix[this.usertype][reg]['type_core_compute'])
                await this.page.getByLabel('Toggle filter Core compute').click()
            }
            this.log(`available products - hardware - Core Compute: ${pcount}`)

            // Filter = type_gpu
            pcount = 0
            if (await this.page.getByLabel('Toggle filter GPU products').isVisible()) {
                await this.page.getByLabel('Toggle filter GPU products').click()
                pcount = await this.page.getByRole('main').locator('div.card-body').count()
                if (idcProductMatrix[this.usertype][reg]['type_gpu'])
                    expect (pcount, 'Number of Core Compute products is wrong')
                        .toBeGreaterThanOrEqual(idcProductMatrix[this.usertype][reg]['type_gpu'])
                await this.page.getByLabel('Toggle filter GPU products').click()
            }
            this.log(`available products - hardware - GPU: ${pcount}`)

            // Filter = type_hpc
            pcount = 0
            if (await this.page.getByLabel('Toggle filter HPC products').isVisible()) {
                await this.page.getByLabel('Toggle filter HPC products').click()
                pcount = await this.page.getByRole('main').locator('div.card-body').count()
                if (idcProductMatrix[this.usertype][reg]['type_hpc'])
                    expect (pcount, 'Number of Core Compute products is wrong')
                        .toBeGreaterThanOrEqual(idcProductMatrix[this.usertype][reg]['type_hpc'])
                await this.page.getByLabel('Toggle filter HPC products').click()
            }
            this.log(`available products - hardware - HPC: ${pcount}`)

            // Filter = type_ai
            pcount = 0
            if (await this.page.getByLabel('Toggle filter AI products').isVisible()) {
                await this.page.getByLabel('Toggle filter AI products').click()
                pcount = await this.page.getByRole('main').locator('div.card-body').count()
                if (idcProductMatrix[this.usertype][reg]['type_ai'])
                    expect (pcount, 'Number of Core Compute products is wrong')
                        .toBeGreaterThanOrEqual(idcProductMatrix[this.usertype][reg]['type_ai'])
                await this.page.getByLabel('Toggle filter AI products').click()
            }
            this.log(`available products - hardware - AI: ${pcount}`)

            this.log('Hardware Catalog page is verified')
        })
    }

    public async verifySoftwareCatalogPage() {
        await test.step('Verify Software Catalog page', async () => {
            this.log('checking Navigation / Catalog / Software')

            await this.clickNavigationLink('Go to Catalog Page', 'Go to Software Page')
            await this.verifyTopHeader('Catalog', 
                [{name: 'Hardware', link: '/hardware'}, {name: 'Software', link: '/software'}], 
                'Software'
            )

            await expect(this.page.locator('div.filter').getByRole('heading', { name: 'Available software' })).toBeVisible()
            await expect(this.page.locator('div.filter').getByLabel('Toggle filter all products')).toBeVisible()

            // API delay
            const noInst = this.page.getByRole('main').locator('div.text-center').getByRole('heading', { name: 'No available services' }) // TODO: verify
            const instTbl = this.page.getByRole('main').locator('div.card-body').first()
            await expect(noInst.or(instTbl)).toBeVisible({timeout: 12000})
            await utils.sleep(500)

            // Total (no filter)
            await expect(this.page.getByRole('main').getByLabel('Toggle filter all products')).toBeVisible()
            let pcount = await this.page.getByRole('main').locator('div.card-body').count()
            expect (pcount, 'Number of Total available software is < 1').toBeGreaterThanOrEqual(1)
            this.log(`available products - software - Total: ${pcount}`)
            
            // Non-existent filter
            await this.page.getByRole('main').getByPlaceholder('Type to search...').fill('blabla')
            pcount = await this.page.getByRole('main').locator('div.card-body').count()
            expect (pcount, 'With non-existent filter - available products should be zero').toEqual(0)
            await this.page.getByRole('main').getByPlaceholder('Type to search...').fill('')
            
            // Filter = Intel optimized images
            if (this.env == 'Staging') {
                pcount = 0
                if (await this.page.getByLabel('Toggle filter Intel optimized').isVisible()) {
                    await this.page.getByLabel('Toggle filter Intel optimized').click()
                    pcount = await this.page.getByRole('main').locator('div.card-body').count()
                    expect (pcount, 'Number of Intel optimized software is < 2').toBeGreaterThanOrEqual(2)
                    await this.page.getByLabel('Toggle filter Intel optimized').click()
                }
                this.log(`available products - software - Intel optimized: ${pcount}`)
            }

            // Seekr flow and Geti ?

            this.log('Hardware Catalog page is verified')
        })
    }

    public async verifyAIPlaygroundPage() {
        await test.step('Verify AI Playground page', async () => {
            this.log('checking Navigation / AI Playground')

            await this.clickNavigationLink('Go to AI Playground Page')
            await this.verifyTopHeader('AI Playground')
            
            await expect(this.page.locator('div.filter').getByRole('heading', { name: 'Demos' })).toBeVisible()
            await expect(this.page.locator('div.filter').getByLabel('Toggle filter all products')).toBeVisible()
            await expect(this.page.locator('div.filter').getByPlaceholder('Type to search...')).toBeVisible()

            // API delay
            const noInst = this.page.getByRole('main').locator('div.text-center').getByRole('heading', { name: 'No available' }) // TODO: verify
            const instTbl = this.page.getByRole('main').locator('div.card-body').first()
            await expect(noInst.or(instTbl)).toBeVisible({timeout: 12000})
            await utils.sleep(500)

            let link = this.page.getByRole('main').getByLabel('Disclaimer for using models')
            await expect(link).toBeVisible()
            expect(await link.getAttribute('href')).toEqual('/docs/reference/model_disclaimers.html')
            await expect(this.page.getByRole('main').getByRole('heading', { name: 'AI Playground' })).toBeVisible()

            // Total (no filter)
            await expect(this.page.getByRole('main').getByLabel('Toggle filter all products')).toBeVisible()
            let pcount = await this.page.getByRole('main').locator('div.card-body').count()
            expect (pcount, 'Number of AI Playground demos is < 1').toBeGreaterThanOrEqual(1)
            this.log(`available AI Playground demos - Total: ${pcount}`)
            
            // Non-existent filter
            await this.page.getByRole('main').getByPlaceholder('Type to search...').fill('blabla')
            pcount = await this.page.getByRole('main').locator('div.card-body').count()
            expect (pcount, 'With non-existent filter - available products should be zero').toEqual(0)
            await this.page.getByRole('main').getByPlaceholder('Type to search...').fill('')
            
            // Filter = Intel optimized images
            pcount = 0
            if (await this.page.getByLabel('Toggle filter AI Playground').isVisible()) {
                await this.page.getByLabel('Toggle filter AI Playground').click()
                pcount = await this.page.getByRole('main').locator('div.card-body').count()
                expect (pcount, 'Number of AI Playground demos is < 1').toBeGreaterThanOrEqual(1)
                await this.page.getByLabel('Toggle filter AI Playground').click()
            }
            this.log(`available AI Playgrounds - AI Playground: ${pcount}`)

            this.log('AI Playground page is verified')
        })
    }

    public async verifyComputeInstancesPage() {
        await test.step('Verify Compute Instances page', async () => {
            this.log('checking Navigation / Compute / Instances')
            await this.waitComputeConsoleIsLoaded()

            await this.verifyTopHeader('Compute', 
                [{name: 'Instances', link: '/compute'}, {name: 'Instance Groups', link: '/compute-groups'},
                 {name: 'Keys', link: '/security/publickeys'}], 
                'Instances'
            )
            // postponed: {name: 'Load Balancers', link: '/load-balancer'},

            // if we have some instances already:
            const nonEmptyLaunchBtn = this.page.getByRole('main').locator('div.filter').getByText('Launch Instance')
            const nonEmptySearchFld = this.page.getByRole('main').locator('div.filter').getByPlaceholder('Search instances...')
            const nonEmptyTbl = this.page.getByRole('main').locator('table.table')
            // if no instances:
            const emptyMsg = this.page.getByRole('main').getByText('Your account currently has no ')
            const emptyLaunchBtn = this.page.getByRole('main').locator('a[intc-id="LaunchinstanceEmptyViewButton"]')

            if (await nonEmptyTbl.isVisible()) {
                // Search input
                await expect(nonEmptySearchFld).toBeVisible()
                // Launch instance button
                await expect(nonEmptyLaunchBtn).toBeVisible()
                expect(await nonEmptyLaunchBtn.getAttribute('href')).toEqual('/compute/reserve')

                const instExisting = await nonEmptyTbl.locator('tbody > tr').count()
                this.log(`showing ${instExisting} existing instances in table`)
                await nonEmptyLaunchBtn.click()
            } else {
                this.log('showing No instances found currently')
                await emptyLaunchBtn.click()
            }

            this.log('checking Launch Instance dialog')

            await expect(this.page.getByRole('main').getByRole('heading', { name: 'Launch a compute instance' })).toBeVisible()

            await expect(this.page.getByRole('heading', { name: 'Instance configuration' })).toBeVisible()
            await expect(this.page.getByRole('radio', { name: 'Select CPU products' })).toBeVisible()
            await expect(this.page.getByRole('radio', { name: 'Select GPU products' })).toBeVisible()
            if (this.usertype == 'Premium')
                await expect(this.page.getByRole('radio', { name: 'Select AI products' })).toBeVisible()
            // if (this.region.startsWith('us-region-'))   // removed 12/6/2024
            //     await expect(this.page.getByRole('radio', { name: 'Select Core compute products' })).toBeVisible()

            await expect(this.page.locator('div.cost-estimate-card').getByRole('heading', { name: 'Cost estimate', exact: true })).toBeVisible()

            await expect(this.page.getByText('Instance Type', { exact: true })).toBeVisible()
            await expect(this.page.getByText('Machine image: *')).toBeVisible()
            await expect(this.page.getByPlaceholder('Instance name')).toBeVisible()
            
            await expect(this.page.getByRole('heading', { name: 'Public Keys' })).toBeVisible()
            await expect(this.page.getByText('Select keys: *')).toBeVisible()
            await expect(this.page.getByRole('button', { name: '+ Upload Key' })).toBeVisible()
            // This one is optional:
            // await expect(this.page.getByRole('button', { name: 'Select/Deselect All' })).toBeVisible()
            
            if (this.region != 'us-staging-3') {
                await expect(this.page.getByRole('heading', { name: 'One-Click connection' })).toBeVisible()
                await expect(this.page.getByLabel('Enable JupyterLab in my instance')).toBeVisible()
            }
            
            // Upload Key dialog
            this.log('checking Upload Key dialog')
            await this.page.getByRole('button', { name: '+ Upload Key' }).click()
            await expect(this.page.getByRole('dialog').getByText('Upload a public key')).toBeVisible()
            await expect(this.page.getByRole('dialog').getByPlaceholder('Key Name')).toBeVisible()
            await expect(this.page.getByRole('dialog').getByPlaceholder('Paste your key contents')).toBeVisible()
            await expect(this.page.getByLabel('Upload public key modal').getByText('Cancel')).toBeVisible()
            await expect(this.page.getByRole('dialog').getByRole('button', { name: 'Upload key', exact: true })).toBeVisible()
            await this.page.getByLabel('Upload public key modal').getByLabel('Close', { exact: true }).click()
            
            await expect(this.page.getByLabel('Launch instance', { exact: true })).toBeVisible()
            await expect(this.page.getByLabel('Cancel', { exact: true })).toBeVisible()

            // Check Learning bar (top-right) // intc-id=btn-toolbar-learning
            this.log('checking Navigation / Compute / Instances / Learning bar')
            const siteToolbDocumentationLink = this.page.locator('nav.siteToolbar').getByRole('button', { name: 'Open learning bar' })
            await expect(siteToolbDocumentationLink, 'top right Documentation link must be visible').toBeVisible()
            await siteToolbDocumentationLink.click()
            const learnNavBar = this.page.locator('div[intc-id="LearningBarNavigationMain"]')
            await expect(learnNavBar).toBeVisible()
            await expect(learnNavBar.getByRole('heading', { name: 'Documentation' }).first()).toBeVisible()
            await expect(learnNavBar.getByLabel('Learn Billing and account').first()).toBeVisible()
            await expect(learnNavBar.getByLabel('Learn Invoices and Usage').first()).toBeVisible()
            await expect(learnNavBar.getByLabel('Learn Manage a compute').first()).toBeVisible()
            // await expect(learnNavBar.getByLabel('Learn Launch an Instance').first()).toBeVisible()
            // await expect(learnNavBar.getByLabel('Learn Instance States').first()).toBeVisible()
            await expect(learnNavBar.getByLabel('Learn Filter processor by').first()).toBeVisible()
            await expect(learnNavBar.getByLabel('Learn Filter by AI type to').first()).toBeVisible()
            await expect(learnNavBar.getByLabel('Learn Manage SSH Keys').first()).toBeVisible()
            await learnNavBar.getByRole('button').click()

            this.log('Compute Instances page is verified')
        })
    }

    private async waitComputeConsoleIsLoaded(subItem: string = 'Instances', timeout = 30000) {
        // Compute / subItem page
        const itab = this.page.locator(`nav.siteToolbar > div.nav-tabs > div.nav-item > a.nav-link.active:has-text("${subItem}")`)
        if (! await itab.count())
            await this.clickNavigationLink('Go to Compute Page', `Go to ${subItem} Page`)
        else
            await itab.click()

        // wait for either table or button at the bottom (API delay)
        const instTbl = this.page.getByRole('main').locator('div.section >> table.table')
        const launchBtn = this.page.getByRole('main').getByText('Your account currently has no ')
        await expect(instTbl.or(launchBtn)).toBeVisible({timeout: timeout})
        this.log(`verified that Compute Console page (sub-item ${subItem}) is fully loaded`)
    }

    public async verifyInstanceGroupsPage() {
        await test.step('Verify Compute Instance Groups page', async () => {
            this.log('checking Navigation / Compute / Instance Groups')
            await this.waitComputeConsoleIsLoaded('Instance Groups')

            await this.verifyTopHeader('Compute', 
                [{name: 'Instances', link: '/compute'}, {name: 'Instance Groups', link: '/compute-groups'},
                 {name: 'Keys', link: '/security/publickeys'}], 
                'Instance Groups'
            )
            // postponed: {name: 'Load Balancers', link: '/load-balancer'},

            // if we have some instance groups already:
            const nonEmptyLaunchBtn = this.page.getByRole('main').locator('div.filter').getByText('Launch Instance group')
            const nonEmptySearchFld = this.page.getByRole('main').locator('div.filter').getByPlaceholder('Search instances...')
            const nonEmptyTbl = this.page.getByRole('main').locator('table.table')
            // if no instance groups:
            const emptyMsg = this.page.getByRole('main').getByText('Your account currently has no ')
            const emptyLaunchBtn = this.page.getByRole('main').locator('a[intc-id="LaunchinstancegroupEmptyViewButton"]')

            // Wait for API query
            await expect(nonEmptyTbl.or(emptyMsg)).toBeVisible({timeout: 15000})

            if (await nonEmptyTbl.isVisible()) {
                // Search input
                await expect(nonEmptySearchFld).toBeVisible()
                // Launch instance button
                await expect(nonEmptyLaunchBtn).toBeVisible()
                expect(await nonEmptyLaunchBtn.getAttribute('href')).toEqual('/compute-groups/reserve')

                const instExisting = await nonEmptyTbl.locator('tbody > tr').count()
                this.log(`showing ${instExisting} existing instance groups in table`)
                await nonEmptyLaunchBtn.click()
            } else {
                expect(emptyMsg).toBeVisible()
                this.log('showing No instance groups found currently')
                await emptyLaunchBtn.click()
            }

            this.log('checking Launch Instance group dialog')
            
            await expect(this.page.locator('div.modal-header').getByText('Exclusive Feature Access')).toBeVisible()
            await this.page.getByRole('link', { name: 'Go Back' }).click()

            this.log('Compute Instance Groups page is verified')
        })
    }

    public async verifyLoadBalancersPage() {
        await test.step('Verify Compute Load Balancers page', async () => {
            this.log('checking Navigation / Compute / Load Balancers')
            await this.waitComputeConsoleIsLoaded('Load Balancers')

            await this.verifyTopHeader('Compute', 
                [{name: 'Instances', link: '/compute'}, {name: 'Instance Groups', link: '/compute-groups'},
                 {name: 'Load Balancers', link: '/load-balancer'}, {name: 'Keys', link: '/security/publickeys'}], 
                'Load Balancers'
            )

            // if we have some load balancers already:
            const nonEmptyLaunchBtn = this.page.getByRole('main').locator('div.filter').getByText('Launch Load Balancer')
            const nonEmptySearchFld = this.page.getByRole('main').locator('div.filter').getByPlaceholder('Search load balancers...')
            const nonEmptyTbl = this.page.getByRole('main').locator('table.table')
            // if no load balancers:
            const emptyMsg = this.page.getByRole('main').getByText('Your account currently has no ')
            const emptyLaunchBtn = this.page.getByRole('main').locator('a[intc-id="LaunchLoadBalancerEmptyViewButton"]')

            // Wait for API query
            await expect(nonEmptyTbl.or(emptyMsg)).toBeVisible({timeout: 15000})

            if (await nonEmptyTbl.isVisible()) {
                // Search input
                await expect(nonEmptySearchFld).toBeVisible()
                // Launch instance button
                await expect(nonEmptyLaunchBtn).toBeVisible()
                expect(await nonEmptyLaunchBtn.getAttribute('href')).toEqual('/load-balancer/reserve')

                const instExisting = await nonEmptyTbl.locator('tbody > tr').count()
                this.log(`showing ${instExisting} existing load balancers in table`)
                await nonEmptyLaunchBtn.click()
            } else {
                expect(emptyMsg).toBeVisible()
                this.log('showing No load balancers found currently')
                await emptyLaunchBtn.click()
            }

            this.log('checking Launch Load Balancer dialog')

            await expect(this.page.locator('div.modal-header').getByText('Load Balancer')).toBeVisible()
            await this.page.getByRole('button', { name: 'Go Back' }).click()

            this.log('Compute Load Balancers page is verified')
        })
    }

    public async verifyKeysPage() {
        await test.step('Verify Keys page', async () => {
            this.log('checking Navigation / Compute / Keys')
            await this.waitComputeConsoleIsLoaded('Keys', 15000)

            await this.verifyTopHeader('Compute', 
                [{name: 'Instances', link: '/compute'}, {name: 'Instance Groups', link: '/compute-groups'},
                 {name: 'Keys', link: '/security/publickeys'}], 
                'Keys'
            )
            // postponed: {name: 'Load Balancers', link: '/load-balancer'},

            // if we have some keys already:
            const nonEmptyLaunchBtn = this.page.getByRole('main').locator('div.filter').getByText('Upload key')
            const nonEmptySearchFld = this.page.getByRole('main').locator('div.filter').getByPlaceholder('Search keys...')
            const nonEmptyTbl = this.page.getByRole('main').locator('table.table')
            // if no keys:
            const emptyMsg = this.page.getByRole('main').getByText('Your account currently has no ')
            const emptyLaunchBtn = this.page.getByRole('main').locator('a[intc-id="UploadkeyEmptyViewButton"]')

            // Wait for API query
            await expect(nonEmptyTbl.or(emptyMsg)).toBeVisible({timeout: 15000})

            if (await nonEmptyTbl.isVisible()) {
                // Search input
                await expect(nonEmptySearchFld).toBeVisible()
                // Launch instance button
                await expect(nonEmptyLaunchBtn).toBeVisible()
                expect(await nonEmptyLaunchBtn.getAttribute('href')).toEqual('/security/publickeys/import')

                const instExisting = await nonEmptyTbl.locator('tbody > tr').count()
                this.log(`showing ${instExisting} existing keys in table`)
                await nonEmptyLaunchBtn.click()
            } else {
                expect(emptyMsg).toBeVisible()
                this.log('showing No keys found currently')
                await emptyLaunchBtn.click()
            }

            this.log('checking Upload key dialog')
            
            await expect(this.page.getByRole('heading', { name: 'Upload key' })).toBeVisible()
            await expect(this.page.getByPlaceholder('Key Name')).toBeVisible()
            await expect(this.page.getByPlaceholder('Paste your key contents')).toBeVisible()
            await expect(this.page.getByLabel('Upload key')).toBeVisible()
            await expect(this.page.getByRole('link', { name: 'Cancel' })).toBeVisible()

            this.log('Keys page is verified')
        })
    }

    public async verifySupercomputingPage() {
        await test.step('Verify Supercomputing page', async () => {
            this.log('checking Navigation / Supercomputing')

            if (this.usertype != 'Standard') {
                if (this.region == 'us-staging-3') {
                
                    ///////////////////////////////////
                    // Checking Supercomputing Overview
                    await this.clickNavigationLink('Go to Supercomputing Page', 'Go to Overview Page')

                    await this.verifyTopHeader('Supercomputing', 
                        [{name: 'Overview', link: '/supercomputer/overview'}, {name: 'Clusters', link: '/supercomputer'}], 
                        'Overview'
                    )

                    // Launch button
                    const launchLink = this.page.locator('div.filter').getByRole('link', { name: 'Launch Supercomputing Cluster' })
                    await expect(launchLink).toBeVisible()
                    expect(await launchLink.getAttribute('href')).toEqual('/supercomputer/launch')

                    await expect(this.page.getByRole('main').getByRole('heading', { name: 'Train and deploy your AI workloads' })).toBeVisible()
                    expect(await this.page.locator('div.ClusterStepItem').count()).toBeGreaterThanOrEqual(4)
                    
                    // Resources
                    await expect(this.page.getByRole('main').getByRole('heading', { name: 'Resources' })).toBeVisible()
                    let link = this.page.getByRole('main').getByRole('link', { name: 'Documentation', exact: true })
                    await expect(link).toBeVisible()
                    expect(await link.getAttribute('href')).toEqual('/docs/guides/k8s_guide.html')

                    this.log('Supercomputing Overview page is verified')

                    ///////////////////////////////////
                    // Checking Supercomputing Clusters
                    await this.clickNavigationLink('Go to Supercomputing Page', 'Go to Clusters Page')

                    await this.verifyTopHeader('Supercomputing', 
                        [{name: 'Overview', link: '/supercomputer/overview'}, {name: 'Clusters', link: '/supercomputer'}], 
                        'Clusters'
                    )

                    // if we have some clusters already:
                    const nonEmptyLaunchBtn = this.page.getByRole('main').locator('div.filter').getByText('Launch Supercomputing Cluster')
                    const nonEmptySearchFld = this.page.getByRole('main').locator('div.filter').getByPlaceholder('Search clusters...')
                    const nonEmptyTbl = this.page.getByRole('main').locator('table.table')
                    // if no clusters:
                    const emptyMsg = this.page.getByRole('main').getByText('Your account currently has no ')
                    const emptyLaunchBtn = this.page.getByRole('main').locator('a[intc-id="LaunchSupercomputingClusterEmptyViewButton"]')

                    // Wait for API query
                    await expect(nonEmptyTbl.or(emptyMsg)).toBeVisible({timeout: 15000})

                    if (await nonEmptyTbl.isVisible()) {
                        // Search input
                        await expect(nonEmptySearchFld).toBeVisible()
                        // Launch button
                        await expect(nonEmptyLaunchBtn).toBeVisible()
                        expect(await nonEmptyLaunchBtn.getAttribute('href')).toEqual('/cluster/reserve')

                        const instExisting = await nonEmptyTbl.locator('tbody > tr').count()
                        this.log(`showing ${instExisting} existing supercomputer clusters in table`)
                        await nonEmptyLaunchBtn.click()
                    } else {
                        expect(emptyMsg).toBeVisible()
                        this.log('showing No supercomputer clusters found currently')
                        await emptyLaunchBtn.click()
                    }

                    this.log('checking Launch Supercomputing Cluster dialog')
                    
                    // This service requires account whitelisting
                    await expect(this.page.locator('div.modal-header').getByText('Exclusive Feature Access', {exact: true})).toBeVisible()
                    await this.page.getByRole('link', { name: 'Go Back' }).click()

                    this.log('Supercomputing Clusters page is verified')
                }
                else
                    this.log('- SKIPPED (only available now in us-staging-3 region)')   // TODO: check prod
            }
            else 
                this.log('- SKIPPED (not available for Standard user)')
        })
    }

    public async verifyK8sServicePage() {
        await test.step('Verify Intel Kubernetes Services page', async () => {
            this.log('checking Navigation / Intel Kubernetes Services')

            // Not available for Standard users
            if (this.usertype != 'Standard') {
                // Not available yet in us-staging-2
                if (this.region != 'us-staging-2') {
                    
                    ////////////////////////
                    // Checking K8s Overview
                    await this.clickNavigationLink('Go to Kubernetes Page', 'Go to Overview Page')
                    
                    await this.verifyTopHeader('Kubernetes',
                        [{name: 'Overview', link: '/cluster/overview'}, {name: 'Clusters', link: '/cluster'}],
                        'Overview'
                    )

                    // Launch Cluster button
                    const launchLink = this.page.locator('div.filter').getByRole('link', { name: 'Launch Cluster', exact: true })
                    await expect(launchLink).toBeVisible()
                    expect(await launchLink.getAttribute('href')).toEqual('/cluster/reserve')

                    await expect(this.page.getByRole('heading', { name: 'Run your Kubernetes workloads' })).toBeVisible()
                    expect(await this.page.locator('div.ClusterStepItem').count()).toBeGreaterThanOrEqual(4)
                    
                    //// Removed permanently?
                    // await expect(this.page.getByRole('heading', { name: 'Resources', exact: true })).toBeVisible()
                    // this.log('checking Resources links')
                    // let link = this.page.getByRole('main').getByRole('link', { name: 'Documentation' })
                    // await expect(link).toBeVisible()
                    // expect(await link.getAttribute('href')).toEqual('/docs/guides/k8s_guide.html')
                    // // await this.checkLinkToPage(link, 
                    // //     'Intel K8s Service / Resources / Documentation',
                    // //     `${this.envInfo['url']}/docs/guides/k8s_guide.html`)

                    // Check Learning bar (top-right)
                    this.log('checking Navigation / Intel Kubernetes Services / Learning bar')
                    const siteToolbDocumentationLink = this.page.locator('nav.siteToolbar').getByRole('button', { name: 'Open learning bar' })
                    await expect(siteToolbDocumentationLink, 'top right Documentation link must be visible').toBeVisible()
                    await siteToolbDocumentationLink.click()
                    const learnNavBar = this.page.locator('div[intc-id="LearningBarNavigationMain"]')
                    await expect(learnNavBar).toBeVisible()
                    await expect(learnNavBar.getByRole('heading', { name: 'Documentation' }).first()).toBeVisible()
                    await expect(learnNavBar.getByLabel('Learn Intel Kubernetes Guide').first()).toBeVisible()
                    await expect(learnNavBar.getByLabel('Learn Provision Kubernetes').first()).toBeVisible()
                    await expect(learnNavBar.getByLabel('Learn Manage Kubernetes').first()).toBeVisible()
                    await expect(learnNavBar.getByLabel('Learn Manage SSH Keys').first()).toBeVisible()
                    await learnNavBar.getByRole('button').click()

                    this.log('Intel Kubernetes Overview page is verified')

                    ////////////////////////
                    // Checking K8s Clusters
                    await this.clickNavigationLink('Go to Kubernetes Page', 'Go to Clusters Page')
                    
                    await this.verifyTopHeader('Kubernetes',
                        [{name: 'Overview', link: '/cluster/overview'}, {name: 'Clusters', link: '/cluster'}],
                        'Clusters'
                    )

                    // if we have some clusters already:
                    const nonEmptyLaunchBtn = this.page.getByRole('main').locator('div.filter').getByText('Launch Cluster')
                    const nonEmptySearchFld = this.page.getByRole('main').locator('div.filter').getByPlaceholder('Search clusters...')
                    const nonEmptyTbl = this.page.getByRole('main').locator('table.table')
                    // if no clusters:
                    const emptyMsg = this.page.getByRole('main').getByText('Your account currently has no ')
                    const emptyLaunchBtn = this.page.getByRole('main').locator('a[intc-id="LaunchclusterEmptyViewButton"]')

                    // Wait for API query
                    await expect(nonEmptyTbl.or(emptyMsg)).toBeVisible({timeout: 15000})

                    if (await nonEmptyTbl.isVisible()) {
                        // Search input
                        await expect(nonEmptySearchFld).toBeVisible()
                        // Launch button
                        await expect(nonEmptyLaunchBtn).toBeVisible()
                        expect(await nonEmptyLaunchBtn.getAttribute('href')).toEqual('/cluster/reserve')

                        const instExisting = await nonEmptyTbl.locator('tbody > tr').count()
                        this.log(`showing ${instExisting} existing k8s clusters in table`)
                        await nonEmptyLaunchBtn.click()
                    } else {
                        expect(emptyMsg).toBeVisible()
                        this.log('showing No k8s clusters found currently')
                        await emptyLaunchBtn.click()
                    }

                    await expect(this.page.getByRole('main').getByRole('heading', { name: 'Launch Kubernetes cluster' })).toBeVisible()
                    await expect(this.page.getByRole('main').getByRole('heading', { name: 'Cluster details and configuration' })).toBeVisible()
                    await this.page.getByRole('main').getByPlaceholder('Cluster name').fill('test')
                    await this.page.getByRole('main').getByLabel('Select cluster kubernetes').click()
                    await this.page.getByRole('main').getByLabel('Select option 1.28').click()
                    await expect(this.page.getByRole('main').getByRole('button', { name: 'Launch' })).toBeVisible()
                    await expect(this.page.getByRole('main').getByRole('button', { name: 'Cancel' })).toBeVisible()

                    this.log('Intel Kubernetes Clusters page is verified')
                }
                else
                    this.log('- SKIPPED (not available yet in us-staging-2 region)')
            }
            else 
                this.log('- SKIPPED (not available for Standard user)')
        })
    }

    public async verifyLearningPage() {
        await test.step('Verify Learning page', async () => {
            this.log('checking Navigation / Learning')
            
            await this.clickNavigationLink('Go to Learning Page', 'Go to Notebooks Page')
            await this.verifyTopHeader('Learning',
                [{name: 'Notebooks', link: '/learning/notebooks'}], // {name: 'Labs', link: '/learning/labs'}],
                'Notebooks'
            )

            // if (this.region === 'us-region-1' || this.region === 'us-staging-1') {
                
                await expect(this.page.getByRole('main').getByRole('heading', { name: 'Available notebooks' })).toBeVisible()
                await expect(this.page.locator('div.filter').getByPlaceholder('Type to search...')).toBeVisible()
                const btnTrain = this.page.locator('div.filter').getByRole('button', { name: 'Connect now' })
                await expect(btnTrain).toBeVisible({timeout: 10000})
                // TODO: click() and check AI Accelerator and GPU buttons

                // API delay
                const noInst = this.page.getByRole('main').locator('div.text-center').getByRole('heading', { name: 'No available services' })
                const instTbl = this.page.getByRole('main').locator('div.card-body').first()
                await expect(noInst.or(instTbl)).toBeVisible({timeout: 12000})
                await utils.sleep(500)

                // Total notebooks (no filter)
                await expect(this.page.getByRole('main').getByLabel('Toggle filter all products')).toBeVisible()
                let pcount = await this.page.getByRole('main').locator('div.card-body').count()
                expect (pcount, 'Number of Total available notebooks is < 6').toBeGreaterThanOrEqual(6)
                this.log(`available notebooks - Total: ${pcount}`)

                // Non-existent filter
                await this.page.getByRole('main').getByPlaceholder('Type to search...').fill('blabla')
                pcount = await this.page.getByRole('main').locator('div.card-body').count()
                expect (pcount, 'With non-existent filter - available notebooks should be zero').toEqual(0)
                await this.page.getByRole('main').getByPlaceholder('Type to search...').fill('')

                // AI notebooks - Intel
                await this.page.getByRole('main').getByLabel('Toggle filter AI ').first().click()
                pcount = await this.page.getByRole('main').locator('div.card-body').count()
                expect (pcount, 'Number of AI notebooks is < 2').toBeGreaterThanOrEqual(2)
                this.log(`available notebooks - AI: ${pcount}`)
                await this.page.getByRole('main').getByLabel('Toggle filter AI ').first().click()

                // AI notebooks - GPU
                // TODO in staging

                // C++ SYCL notebooks
                await this.page.getByRole('main').getByLabel('Toggle filter C++ SYCL').click()
                pcount = await this.page.getByRole('main').locator('div.card-body').count()
                expect (pcount, 'Number of C++ SYCL notebooks is < 2').toBeGreaterThanOrEqual(2)
                this.log(`available notebooks - C++ SYCL: ${pcount}`)
                await this.page.getByRole('main').getByLabel('Toggle filter C++ SYCL').click()

                // // Gen AI Essentials notebooks - removed
                // await this.page.getByRole('main').getByLabel('Toggle filter Gen AI Essentials').click()
                // pcount = await this.page.getByRole('main').locator('div.card-body').count()
                // expect (pcount, 'Number of Gen AI Essentials notebooks is < 2').toBeGreaterThanOrEqual(2)
                // this.log(`available notebooks - Gen AI Essentials: ${pcount}`)
                // await this.page.getByRole('main').getByLabel('Toggle filter Gen AI Essentials').click()

                // Rendering Toolkit notebooks
                await this.page.getByRole('main').getByLabel('Toggle filter Rendering Toolkit').click()
                pcount = await this.page.getByRole('main').locator('div.card-body').count()
                expect (pcount, 'Number of Rendering Toolkit notebooks is < 1').toBeGreaterThanOrEqual(1)
                this.log(`available notebooks - Rendering Toolkit: ${pcount}`)
                await this.page.getByRole('main').getByLabel('Toggle filter Rendering Toolkit').click()

                // Toggle filter AI with Intel Gaudi2 Accelerator
                await this.page.getByRole('main').getByLabel('Toggle filter AI with Intel Gaudi 2 Accelerator').click()
                pcount = await this.page.getByRole('main').locator('div.card-body').count()
                expect (pcount, 'Number of AI with Intel Gaudi 2 Accelerator notebooks is < 3').toBeGreaterThanOrEqual(3)
                this.log(`available notebooks - AI with Intel Gaudi 2 Accelerator: ${pcount}`)
                await this.page.getByRole('main').getByLabel('Toggle filter AI with Intel Gaudi 2 Accelerator').click()

                // Quantum Computing notebooks
                await this.page.getByRole('main').getByLabel('Toggle filter Quantum Computing').click()
                pcount = await this.page.getByRole('main').locator('div.card-body').count()
                expect (pcount, 'Number of Quantum Computing notebooks is < 1').toBeGreaterThanOrEqual(1)
                this.log(`available notebooks - Quantum Computing: ${pcount}`)
                await this.page.getByRole('main').getByLabel('Toggle filter Quantum Computing').click()

                // // Check Learning bar (top-right) - grayed out so far
                // this.log('checking Navigation / Learning / Learning bar')
                // const siteToolbDocumentationLink = this.page.locator('nav.siteToolbar').getByRole('button', { name: 'Open learning bar' })
                // await expect(siteToolbDocumentationLink, 'top right Documentation link must be visible').toBeVisible()
                // await siteToolbDocumentationLink.click()
                // const learnNavBar = this.page.locator('div[intc-id="LearningBarNavigationMain"]')
                // await expect(learnNavBar).toBeVisible()
                // await expect(learnNavBar.getByRole('heading', { name: 'Documentation' }).first()).toBeVisible()
                // await expect(learnNavBar.getByLabel('Learn Filter processor by').first()).toBeVisible()
                // await expect(learnNavBar.getByLabel('Learn Filter by AI type to').first()).toBeVisible()
                // // await expect(learnNavBar.getByLabel('Learn Gaudi2 Deep Learning')).toBeVisible()
                // await expect(learnNavBar.getByLabel('Learn XPU verify tool for').first()).toBeVisible()
                // await expect(learnNavBar.getByLabel('Learn Jupyter Notebook').first()).toBeVisible()
                // // await expect(learnNavBar.getByLabel('Learn Public AI Tutorials')).toBeVisible()
                // await expect(learnNavBar.getByLabel('Learn Develop in the Cloud').first()).toBeVisible()
                // await learnNavBar.getByRole('button').click()

            // } else
            //     this.log(`Jupiterlab is not available in region ${this.region}, nothing more to verify`)
            
            this.log('Learning / Notebooks page is verified')
            
            // ///////////////////
            // // Learning / Labs

            // await this.clickNavigationLink('Go to Learning Page', 'Go to Labs Page')
            // await this.verifyTopHeader('Learning',
            //     [{name: 'Notebooks', link: '/learning/notebooks'}, {name: 'Labs', link: '/learning/labs'}],
            //     'Labs'
            // )

            // await expect(this.page.getByRole('main').getByRole('heading', { name: 'Available Labs' })).toBeVisible()
            // await expect(this.page.locator('div.filter').getByPlaceholder('Type to search...')).toBeVisible()
            
            // // API delay
            // const cards = this.page.getByRole('main').locator('div.card-body')
            // await expect(cards.first()).toBeVisible({timeout: 10000})
            // await utils.sleep(500)

            // // Total labs (no filter)
            // await expect(this.page.getByRole('main').getByLabel('Toggle filter all products')).toBeVisible()
            // pcount = await cards.count()
            // expect (pcount, 'Number of Total available labs is < 1').toBeGreaterThanOrEqual(1)
            // this.log(`available labs - Total: ${pcount}`)

            // // Non-existent filter
            // await this.page.getByRole('main').getByPlaceholder('Type to search...').fill('blabla')
            // pcount = await cards.count()
            // expect (pcount, 'With non-existent filter - available labs should be zero').toEqual(0)
            // await this.page.getByRole('main').getByPlaceholder('Type to search...').fill('')

            // // AI with Max Series GPU
            // await this.page.getByRole('main').getByLabel('Toggle filter AI with Max').first().click()
            // pcount = await cards.count()
            // expect (pcount, 'Number of AI labs is < 1').toBeGreaterThanOrEqual(1)
            // this.log(`available notebooks - AI: ${pcount}`)
            // await this.page.getByRole('main').getByLabel('Toggle filter AI with Max').first().click()

            // this.log('Learning / Labs page is verified')
            
            this.log('Learning page is verified')
        })
    }

    public async verifyHelpMenu() {
        await test.step('Verify Help Menu', async () => {
            this.log('checking Help drop-down menu')
            // TODO: check if dropdown-menu is already shown
            //if (! await this.page.locator('div.dropdown-menu:has-text("Knowledge Base")').isVisible())
            await this.page.getByLabel('Support Menu').click()

            // Check Help links by URL or by navigating
            this.log('checking Help / Documentation')
            const linkDoc = this.page.locator('#dropdown-header-menu-support').getByLabel('Go to Documentation')
            const linkDoc2 = this.page.locator('#dropdown-header-menu-support').getByLabel('Go to Browse documentation')
            await expect(linkDoc.or(linkDoc2)).toBeVisible()
            // expect(await linkDoc.getAttribute('href')).toContain('/docs/index.html') // TODO: restore
            // await this.checkLinkToPage(linkDoc, 
            //     'Help / Documentation', 
            //     `${this.envInfo['url']}/docs/index.html`)

            this.log('checking Help / Community')
            // await this.page.getByLabel('Support Menu').click()
            const linkComm = this.page.locator('#dropdown-header-menu-support').getByLabel('Go to Community')
            await expect(linkComm).toBeVisible()
            expect(await linkComm.getAttribute('href')).toContain('https://community.intel.com/t5/Intel-Developer-Cloud/bd-p/developer-cloud')
            // await this.checkLinkToPage(linkComm, 
            //     'Help / Community',
            //     'https://community.intel.com/t5/Intel-Developer-Cloud/bd-p/developer-cloud')

            // Knowledge Base external page - may take 20-30s to open
            this.log('checking Help / Knowledge Base')
            // await this.page.getByLabel('Support Menu').click()
            let link = this.page.locator('#dropdown-header-menu-support').getByLabel('Go to Knowledge Base')
            await expect(link).toBeVisible()
            expect(await link.getAttribute('href')).toContain('https://www.intel.com/content/www/us/en/support/')
            // await this.checkLinkToPage(link_kb, 
            //     'Help / Knowledge Base',
            //     'https://www.intel.com/content/www/us/en/support/')

            this.log('checking Help / Submit a ticket')
            // await this.page.getByLabel('Support Menu').click()
            link = this.page.locator('#dropdown-header-menu-support').getByLabel('Go to Submit a ticket')
            await expect(link).toBeVisible()
            expect(await link.getAttribute('href')).toContain('https://supporttickets.intel.com/supportrequest')
            // await this.checkLinkToPage(link_kb, 
            //     'Help / Submit a ticket',
            //     'https://supporttickets.intel.com/s/supportrequest')

            if (this.usertype === 'Premium') {
                this.log('checking Help / Contact support')
                // await this.page.getByLabel('Support Menu').click()
                link = this.page.locator('#dropdown-header-menu-support').getByLabel('Go to Contact support')
                await expect(link).toBeVisible()
                expect(await link.getAttribute('href')).toContain('https://www.intel.com/content/www/us/en/support/contact-intel.html')
            }

            this.log('checking Help / Send Feedback')
            // await this.page.getByLabel('Support Menu').click()
            link = this.page.locator('#dropdown-header-menu-support').getByLabel('Go to Send Feedback')
            await expect(link).toBeVisible()
            expect(await link.getAttribute('href')).toContain('https://intel.az1.qualtrics.com/jfe/form/SV_8cEjBMShr8n3FgW')

            this.log('Help Menu is verified')
        })
    }

    public async verifyPreviewCatalog() {
        await test.step('Verify Preview Catalog page', async () => {
            this.log('checking Navigation / Preview / Preview Catalog')
            
            await this.clickNavigationLink('Go to Preview Page', 'Go to Preview Catalog Page')
                    
            await this.verifyTopHeader('Preview',
                [{name: 'Preview Catalog', link: '/preview/hardware'}, {name: 'Preview Instances', link: '/preview/compute'},
                 {name: 'Preview Keys', link: '/preview/security/publickeys'}],
                'Preview Catalog'
            )
            // {name: 'Preview Storage', link: '/preview/storage'},  - hidden in Prod now

            await expect(this.page.getByRole('main').getByRole('heading', { name: 'Available hardware' })).toBeVisible()
            await expect(this.page.locator('div.filter').getByPlaceholder('Type to search...')).toBeVisible()

            // API delay
            const noCards = this.page.getByRole('main').locator('div.text-center').getByRole('heading', { name: 'No available ' })
            const cards = this.page.getByRole('main').locator('div.card-body').first()
            await expect(noCards.or(cards)).toBeVisible({timeout: 12000})
            await utils.sleep(500)

            // Total products (no filter)
            await expect(this.page.getByRole('main').getByLabel('Toggle filter all products')).toBeVisible()
            let pcount = await this.page.getByRole('main').locator('div.card-body').count()
            expect (pcount, 'Number of Total available products is < 6').toBeGreaterThanOrEqual(6)
            this.log(`available products - Total: ${pcount}`)

            // Non-existent filter
            await this.page.getByRole('main').getByPlaceholder('Type to search...').fill('blabla')
            pcount = await this.page.getByRole('main').locator('div.card-body').count()
            expect (pcount, 'With non-existent filter - available products should be zero').toEqual(0)
            await this.page.getByRole('main').getByPlaceholder('Type to search...').fill('')

            // CPU products
            await this.page.getByRole('main').getByLabel('Toggle filter CPU products').click()
            pcount = await this.page.getByRole('main').locator('div.card-body').count()
            expect (pcount, 'Number of CPU products is < 4').toBeGreaterThanOrEqual(4)
            this.log(`available products - CPU: ${pcount}`)
            await this.page.getByRole('main').getByLabel('Toggle filter CPU products').click()

            // GPU products
            await this.page.getByRole('main').getByLabel('Toggle filter GPU products').click()
            pcount = await this.page.getByRole('main').locator('div.card-body').count()
            expect (pcount, 'Number of GPU products is < 2').toBeGreaterThanOrEqual(2)
            this.log(`available products - GPU: ${pcount}`)
            await this.page.getByRole('main').getByLabel('Toggle filter GPU products').click()

            if (this.usertype === 'Premium') {
                // // AI products  - not available in Stage
                // await this.page.getByRole('main').getByLabel('Toggle filter AI products').click()
                // pcount = await this.page.getByRole('main').locator('div.card-body').count()
                // expect (pcount, 'Number of AI products is < 1').toBeGreaterThanOrEqual(1)
                // this.log(`available products - AI: ${pcount}`)
                // await this.page.getByRole('main').getByLabel('Toggle filter AI products').click()

                // AI PC products
                await this.page.getByRole('main').getByLabel('Toggle filter AI PC products').click()
                pcount = await this.page.getByRole('main').locator('div.card-body').count()
                expect (pcount, 'Number of AI PC products is < 1').toBeGreaterThanOrEqual(1)
                this.log(`available products - AI PC: ${pcount}`)
                await this.page.getByRole('main').getByLabel('Toggle filter AI PC products').click()
            }

            // Check Learning bar (top-right) - grayed out so far
            this.log('checking Navigation / Preview / Learning bar')
            const siteToolbDocumentationLink = this.page.locator('nav.siteToolbar').getByRole('button', { name: 'Open learning bar' })
            await expect(siteToolbDocumentationLink, 'top right Documentation link must be visible').toBeVisible()
            await siteToolbDocumentationLink.click()
            const learnNavBar = this.page.locator('div[intc-id="LearningBarNavigationMain"]')
            await expect(learnNavBar).toBeVisible()
            await expect(learnNavBar.getByRole('heading', { name: 'Documentation' }).first()).toBeVisible()
            await expect(learnNavBar.getByLabel('Learn Preview Catalog').first()).toBeVisible()
            await expect(learnNavBar.getByLabel('Learn Manage SSH Keys').first()).toBeVisible()
            await expect(learnNavBar.getByLabel('Learn Prerequisites').first()).toBeVisible()  // resolves to two links
            await expect(learnNavBar.getByLabel('Learn Request a Preview').first()).toBeVisible()
            await expect(learnNavBar.getByLabel('Learn Generate an SSH Key').first()).toBeVisible()
            await expect(learnNavBar.getByLabel('Learn Upload an SSH Key').first()).toBeVisible()
            await expect(learnNavBar.getByLabel('Learn Update an SSH Key').first()).toBeVisible()
            await learnNavBar.getByRole('button').click()

            this.log('Preview Catalog page is verified')
        })
    }


    //////////////////////////////////////////////////
    // Check link - it must be opening in separate tab
    private async checkLinkToPage(loc: Locator, name: string, expectedUrl: string = '') {
        this.log(`checking link ${name}`);

        const [newPage] = await Promise.all([
            this.context.waitForEvent('page', {timeout: 10000}),
            loc.click()
        ])

        await newPage.waitForResponse(response => response.status() === 200, {timeout: 5000})

        // Wait for Page Load
        // Advance once content is loaded from DOM
        await newPage.waitForLoadState('networkidle', {timeout: 30000})  //domcontentloaded
        await newPage.waitForTimeout(200)

        const newPageUrl: string = newPage.url();
        this.log('new page is opened on URL: ' + newPageUrl);
        
        // if need to validate URL
        if (expectedUrl) {
            if (newPageUrl.startsWith(expectedUrl))
                this.log('new page matches expected URL: ' + expectedUrl)
            else 
                throw new Error('new page does not match expected URL: ' + expectedUrl)
        }

        // title of new tab page
        const pt = await newPage.title()
        this.log(`new page title is: ${pt}`)

        // error pages
        if (pt.match( /50[0-4] / ))
            throw new Error('landed on error page')

        await newPage.close()
        this.log(`link ${name} is verified successfully`)
    }

    private async verifyTopHeader (header: string, tabs: Object[] = [], activeTab: string = '') {
        await expect(this.page.locator('nav.siteToolbar')
            .getByRole('heading', { name: header, exact: true })).toBeVisible()
        
        // optionally, verify tabs, eg. [{name: 'Invoices', link: '/billing/invoices'}]
        for (let tb, i = 0; i < tabs.length; i++)
            if (tabs[i].hasOwnProperty("name")) {
                tb = this.page.locator('nav.siteToolbar').getByRole('link', { name: tabs[i]["name"], exact: true })
                await expect(tb).toBeVisible()
                if (tabs[i].hasOwnProperty("link") && tabs[i]["link"] != '')
                    expect(await tb.getAttribute('href')).toEqual(tabs[i]["link"])
            }
        
        // optionally, verify active tab
        if (activeTab)
            await expect(this.page.locator(`nav.siteToolbar > div.nav-tabs > div.nav-item > a.nav-link.active:text-is("${activeTab}")`)).toBeVisible()
    }

    private async clickNavigationLink (item: string, subItem: string = '') {
        await utils.waitForObjectDisappearance(this.locSideNavPanel, 500, '', true)
        if (! await this.locSideNavPanel.isVisible())
            await this.locNavigationExpandButton.click()
        
        // new 10/7/2024 - top menu item can be already expanded
        const mItem = this.locSideNavPanel.getByLabel(item, {exact: true})
        if (await mItem.getAttribute('data-bs-toggle') !== 'collapse' || await mItem.getAttribute('aria-expanded') === 'false')
            await mItem.click()

        // sub-menu items can be non-unique, eg. for K8s and Supercomputing
        if (subItem)
            await this.locSideNavPanel.locator(`div:has(> a[aria-label="${item}"])`).getByLabel(subItem, {exact: true} ).click()

        // // if navigation panel is still open, close it - not auto-closing since 10/17/2024
        // await utils.waitForObjectDisappearance(this.locSideNavPanel, 500, '', true)
        // if (await this.locSideNavPanel.isVisible())
        //     await this.locNavigationCollapseButton.click()
    }

    private async clickUserMenuLink (item: string) {
        let userMenu = this.page.locator('#dropdown-header-user-menu')
        await utils.waitForObjectDisappearance(userMenu, 500, '', true)
        if (! await userMenu.isVisible())
            await this.page.getByLabel('User Menu').click()
        await this.page.locator('#dropdown-header-user-menu').getByRole('link', { name: item }).click()
    }

    public async signOut() {
        await test.step('Sign out', async () => {
            this.log('initiating Sign out..')
            const userMenu = this.page.getByLabel('User Menu')
            if (! await this.page.locator('#dropdown-header-user-menu').isVisible())
                await userMenu.click()
            await this.page.locator('#dropdown-header-user-menu').getByRole('button', { name: 'Sign-out' }).click()
            // Redirect back to intel.com
            await utils.waitForObjectDisappearance(userMenu, 10000, 'signout error - Console page is still open')
            // await this.page.waitForLoadState('domcontentloaded')
            // await utils.waitForObject..

            this.log('signed out successfully')
        })
    }
    
    // ***************************************************************
    // **************************** VMaaS ****************************

    public async VMaasPreCleanup(instname = 'idc01-vmaas') {
        await test.step(`VMaaS pre-cleanup for ${instname}`, async () => {
            
            this.log('checking pre-existing instances and ssh keys..')

            // delete ${instname} ssh-key if exists
            await this.waitComputeConsoleIsLoaded('Keys')
            const tbl     = this.page.getByRole('main').locator('table.table')
            const tblr    = tbl.locator('tr', {hasText: instname})          // our key only, do not touch other keys
            const tblrtrm = tblr.locator('td', {hasText: 'Terminating'})    // if it's terminating status
            
            // // IDCSRE-4748 - workaround for UI bug, when it immediately shows "no keys"
            // await utils.sleep(1000)  // seems like fixed
            
            if (await tblr.count()) {
                this.log(`found pre-existing ssh-key ${instname}, trying to delete it..`)
                await this.VMaasDeleteKey(instname)
            }

            // delete ${instname} instance if exists
            await this.waitComputeConsoleIsLoaded('Instances')
            // wait if still in "Terminating" status
            if (await tblrtrm.count()) {
                this.log(`found pre-existing instance ${instname} in terminating state, waiting..`)
                await utils.waitForObjectDisappearance(tblr, 90000, `Can not pre-cleanup instance ${instname}, it is stuck in terminating state`);
            }
            if (await tblr.count()) {
                this.log(`found pre-existing instance ${instname}, trying to delete it..`)
                await this.VMaasDelete(instname)
            }

            this.log(`IDC01-VMaaS pre-cleanup ${instname} is done`)
        })
    }

    // Create Compute instance by name
    public async VMaasCreate(instname = 'idc01-vmaas', timeout = 180000) {
        await test.step(`Create VM instance ${instname}`, async () => {
            this.log('Navigation / Compute / Compute Instances VM - create instance')
            await this.waitComputeConsoleIsLoaded()

            // Create SSH keypair locally, and get private key
            if (this.envInfo.hasOwnProperty('doSshValidation') && this.envInfo['doSshValidation'])
                await this.VMaasPrepareSsh()

            // if we have some instances already:
            const nonEmptyLaunchBtn = this.page.getByRole('main').locator('div.filter').getByText('Launch Instance')
            // if no instances:
            const emptyLaunchBtn = this.page.getByRole('main').locator('a[intc-id="LaunchinstanceEmptyViewButton"]')
            
            if (await emptyLaunchBtn.isVisible())
                await emptyLaunchBtn.click()
            else 
                await nonEmptyLaunchBtn.click()

            await expect(this.page.getByRole('main').getByRole('heading', { name: 'Launch a compute instance' })).toBeVisible()

            // Verify all form fields
            await expect(this.page.getByRole('heading', { name: 'Instance configuration' })).toBeVisible()
            // await this.page.getByRole('radio', { name: 'Select Core compute products' }).check()
            // await this.page.getByRole('radio', { name: 'Select CPU products' }).check()
            // await this.page.getByRole('radio', { name: 'Select GPU products' }).check()
            // await this.page.getByRole('radio', { name: 'Select AI products' }).check()

            await expect(this.page.getByText('Instance Type', { exact: true })).toBeVisible()
            // await this.page.locator('input[name="checkBoxTable-Instancetype-grid-vm-spr-sml"]').check()
            // await this.page.locator('input[name="checkBoxTable-Instancetype-grid-vm-spr-med"]').check()
            // await this.page.locator('input[name="checkBoxTable-Instancetype-grid-vm-spr-lrg"]').check()

            await expect(this.page.getByText('Machine image: *')).toBeVisible()
            // await this.page.getByLabel('Machine image: *').click()
            // await this.page.getByLabel('Select option ubuntu-2204').first().click()
            await expect(this.page.getByText('Instance name: *')).toBeVisible()

            // Use default options (small VM) - TODO: select non-default parameters if required
            // (instance family, instance type, VM image)

            // // Compare instance types dialog - removed 11/27/2024
            // await this.page.getByLabel('Compare instance types').click()
            // await expect(this.page.locator('div.modal-header')).toHaveText('Compare instance types')
            // let dlg_tbl = this.page.locator('div.modal-body').getByRole('table')
            // // await expect(dlg_tbl.getByRole('cell', { name: 'Core compute' })).toBeVisible()
            // utils.sleep(500)
            // // expect(await dlg_tbl.locator('tr').count()).toBeGreaterThan(2)
            // await expect(this.page.locator('div.modal-content').getByRole('button', { name: 'Select', exact: true })).toBeVisible()
            // await this.page.locator('div.modal-content').getByRole('button', { name: 'Cancel', exact: true }).click()

            await this.page.getByRole('main').getByPlaceholder('Instance name').fill(instname)

            // Upload Key dialog
            this.log('uploading SSH Key')
            await this.page.getByRole('button', { name: '+ Upload Key' }).click()
            await this.page.getByRole('dialog').getByPlaceholder('Key Name').fill(instname)
            await this.page.getByRole('dialog').getByPlaceholder('Paste your key contents').fill(this.sshPubKey)
            await expect(this.page.getByRole('dialog').locator('div.modal-body').getByRole('button', { name: 'Cancel' })).toBeVisible()
            await this.page.getByRole('dialog').locator('div.modal-body').getByLabel('Upload key').click()
            await this.page.getByRole('main').locator(`input.form-check-input[value="${instname}"]`).check()
            await this.page.getByRole('main').getByLabel('Launch instance', { exact: true }).click()
            this.log('launching new VM..')
            // optional progress dialog?
            // await expect(this.page.getByText('Working on your reservation')).toBeVisible()

            const tblr = this.page.getByRole('main').locator(`table.table > tbody > tr:has-text("${instname}")`)
            await expect(tblr).toBeVisible({timeout: 30000})
            // await expect(tblr.locator('td').nth(2)).toContainText('Provisioning', {timeout: 10000})

            // third column will become Ready status
            this.log('launch is in progress..')
            await expect(tblr.locator('td').nth(2),
                'Timed out (3 min) waiting for Ready status of created instance').toHaveText('Ready', {timeout: timeout})
            this.log('VM instance is Ready!')

            this.log(`VM instance ${instname} is created`)
        })
    }

    // Verify Compute instance by name
    public async VMaasVerify(instname = 'idc01-vmaas') {
        await test.step(`Verify VM instance ${instname}`, async () => {
            this.log('Navigation / Compute / Instances verify VM instance')
            await this.waitComputeConsoleIsLoaded()

            const tbl = this.page.getByRole('main').locator('table.table')
            const tblr = tbl.locator('tr', {hasText: instname})          // checking our VM row only

            await expect(tblr.locator('td').nth(2)).toHaveText('Ready')
            
            // Check VM info
            await tblr.locator('td').nth(0).getByRole('button', { name: instname }).click()
            // ++ tab "Details"
            let tdetr = this.page.locator('div.section:has(> h3:has-text("Instance type information"))').locator('div.row')
            // Instance type info
            await expect(tdetr.nth(0).locator('div.col-md-3').nth(2).locator('span').nth(0)).toHaveText('Status:')
            await expect(tdetr.nth(0).locator('div.col-md-3').nth(2).locator('span').nth(1)).toHaveText('Ready')
            // VM image info
            await expect(tdetr.nth(1).getByRole('heading', { name: 'Machine image information' })).toBeVisible()
            await expect(tdetr.nth(1).locator('div.col-md-3').nth(1)).toContainText('Ubuntu')
            
            // ++ tab "Network"
            await this.page.getByRole('main').getByRole('button', { name: 'Networking' }).click()
            await expect(this.page.getByRole('main').getByRole('heading', { name: 'Networking interfaces' })).toBeVisible()
            tdetr = this.page.locator('div.section:has(> h3:has-text("Networking interfaces"))')
            await expect(tdetr.locator('div.flex-column > span').nth(0)).toHaveText('IP')
            const ip = await tdetr.locator('div.flex-column > span').nth(1).textContent()
            this.log(`VM has IP = ${ip}`)

            // ++ tab "Security"
            await this.page.getByRole('main').getByRole('button', { name: 'Security' }).click()
            await expect(this.page.getByRole('main').getByRole('heading', { name: 'Instance Public Keys' })).toBeVisible()
            tdetr = this.page.locator('div.section:has(> h3:has-text("Instance Public Keys"))')
            await expect(tdetr.locator('div.row').nth(0)).toContainText(instname)
            // How to connect dialog
            await this.page.getByRole('main').getByRole('button', { name: 'How to Connect' }).click()
            await expect(this.page.locator('div.modal-header').getByText('How to connect to your instance')).toBeVisible()
            await expect(this.page.locator('div.modal-body').locator('pre > span').nth(1)).toContainText('ssh -J guest@')
            this.sshConnectLine = await this.page.locator('div.modal-body').locator('pre > span').nth(1).textContent() || ''
            this.log(`ssh connect line: ${this.sshConnectLine}`)
            await this.page.locator('div.modal-footer').getByText('Close').click()

            // Actions / Edit instance
            await this.page.getByRole('main').getByRole('button', { name: 'Actions' }).click()
            await this.page.getByRole('main').getByRole('button', { name: 'Edit' }).click()
            await expect(this.page.getByRole('main').getByRole('heading', { name: 'Edit instance' })).toBeVisible()
            await expect(this.page.getByRole('main').getByPlaceholder('Instance name')).toBeDisabled()
            await expect(this.page.getByRole('main').locator('div.input-group').locator(`input.form-check-input[value="${instname}"]`)).toBeChecked()
            await this.page.getByRole('main').getByLabel('Cancel').click()

            // Test by connecting via SSH and running command
            if (this.envInfo.hasOwnProperty('doSshValidation') && this.envInfo['doSshValidation'])
                expect(await this.RunSshCommand('uname', /^Linux/), 'checking by Linux uname').toEqual(true)
            else
                this.log('skipping SSH test')

            this.log(`VM instance ${instname} is verified`)
        })
    }

    // Create local default ssh keypair if absent
    // Remember public key as this.sshPubKey
    private async VMaasPrepareSsh() {
        this.log('Prepare SSH key')
        try{
            let stdout = await exec(`if [ ! -f ~/.ssh/id_rsa ]; then timeout 10 ssh-keygen -q -t rsa -b 4096 -N '' -f ~/.ssh/id_rsa -C idc@test.com; fi`)
            let key = await exec(`cat ~/.ssh/id_rsa.pub`)
            this.sshPubKey = key.stdout.replace('\n','')
        } catch (err) {
            this.log(`SSH prepare error: ${err}`)
        }
    }

    // Run <command> over SSH, test <outPattern> Regex
    // Return: true if output matches Regex
    private async RunSshCommand(command = 'uname', outPattern = /^$/, sshpasswd = '', timeout = 0) {
        this.log(`Run remote ssh command: ${command.substring(0,50)}`)
        try{
            // Fail test by timeout if timeout is non zero
            let failByTimeout
            if (timeout > 0)
                failByTimeout = setTimeout(() => { throw new Error(`Timed out waiting for response`) }, timeout)

            // ignore Preview pre-command "ssh-keygen -R 192.168.12.2;"
            const cmdStart = this.sshConnectLine.indexOf('ssh -J')
            const cmdLine = this.sshConnectLine.substring(cmdStart)

            const sshConnOpts = '-o ConnectTimeout=12 -o StrictHostKeyChecking=no' // -o BatchMode=yes
            
            // can be 'ubuntu' or 'devcloud' user
            const usru = cmdLine.indexOf('ubuntu@')
            const usrd = cmdLine.indexOf('devcloud@')
            const usrpos = usru > 0 ? usru : usrd

            if (! this.sshConnectInitialized) {
                // If need to use socks proxy
                if (this.envInfo.hasOwnProperty('sshSocksProxy') && this.envInfo['sshSocksProxy']) {
                    const jhost = cmdLine.substring(cmdLine.indexOf('guest') + 6, usrpos - 1)
                    await exec(`echo 'Host ${jhost}
        ProxyCommand nc -x internal-placeholder.com:1080 %h %p' > ~/.ssh/config`)
                }

                // connect to proxy first
                const cmd1 = cmdLine.substring(0, usrpos).replace(' -J','')
                const outp1 = await exec(`${cmd1} ${sshConnOpts}`)
                if (outp1.err) {
                    this.log(`ERROR: failed to connect to ssh proxy: ${outp1.err}`)   // ex: 'Connection timed out'
                    return false
                }

                this.sshConnectInitialized = true
            }

            // now connect to target host via bastion, and issue command
            // eg, ssh -J guest@146.152.232.8 ubuntu@100.80.195.99 uname
            const sshpass = sshpasswd ? `sshpass -p ${sshpasswd} ` : ''     // sshpass is required for Preview instance

            const outp2 = await exec(`${sshpass}${cmdLine} ${sshConnOpts} '${command}'`)
            clearTimeout(failByTimeout)

            if (outPattern.test(outp2.stdout)) {
                this.log('SSH command ran successfully, output verified')
                return true
            } else {
                // weird situation with Weka mount (successful output in stderr only)
                if (outPattern.test(outp2.stderr)) {
                    this.log('SSH command ran successfully, but output pattern is in stderr')
                    return true
                }
                this.log(`ERROR: ssh command issued but output is not as expected: ${outp2.stdout} (err: ${outp2.stderr})`)
                return false
            }

        } catch (err) {
            this.log(`SSH connection error: ${err}`)
            return false
        }
    }

    // Delete Compute instance by name (must exist!)
    public async VMaasDelete(instname = 'idc01-vmaas', timeout = 180000) {
        await test.step(`Delete VM instance ${instname}`, async () => {
            this.log('Navigation / Compute / Instances - delete VM instance')
            await this.waitComputeConsoleIsLoaded()

            const tbl = this.page.getByRole('main').locator('div.section >> table.table')
            //locator(`table.table > tbody > tr:has(td:nth-of-type(1):has-text("instname"))`)
            const tblr = tbl.locator('tr', {hasText: instname})          // checking our VM row only

            // Delete VM
            this.log('attempting to Delete the VM..')
            await expect(tblr.getByText('Delete')).toBeVisible({timeout: 10000})
            await tblr.getByText('Delete').click()
            // await tblr.getByLabel('Delete instance').click()

            // Deletion confirm dialog window
            await expect(this.page.locator('div.modal-header').getByText(/^Delete /)).toBeVisible()

            // new dialog requires to type-in name of instance
            const delDlgNew = this.page.locator('div.modal-body').getByText('To confirm deletion enter the name of the instance below')
            await expect(delDlgNew, 'Delete confirmation dialog must be on screen').toBeVisible()

            await expect(this.page.locator('div.modal-body').getByPlaceholder('Enter instance name')).toBeVisible()
            await this.page.locator('div.modal-body').getByPlaceholder('Enter instance name').fill(instname)
            await expect(this.page.locator('div.modal-footer').getByLabel('Cancel')).toBeVisible()
            await this.page.locator('div.modal-footer').getByLabel('Delete', { exact: true }).click()
            
            this.log('instance deletion is in progress..')
            
            // Fails on pre-cleanup, when instance is stuck in Provisioning, then we terminate and it deletes w/o Terminating state
            // await expect(tblr, 
            //     'Timed out (20s) waiting for status Terminating').toContainText('Terminating', {timeout: 20000})

            // set timer to reload page (SRE-3341)
            const reloadTimer = setTimeout(async () => {
                // reload only if instance is still in Ready state (IDCCOMP-3028)
                if (await tblr.getByText('Ready', {exact: true}).count()) {
                    await this.page.reload()
                    this.log('page looks to be stale, reloading the page..')
                }
            }, 30000)

            // wait for row or entire table disappearance
            // await expect(tblr, `Timeout ${timeout} ms exceeded on instance deletion`).not.toBeVisible({timeout: timeout})
            await utils.waitForObjectDisappearance1(this.page, tblr, timeout, `Instance deletion with timeout ${timeout}`)

            clearTimeout(reloadTimer)

            // reload page once again to make 100% sure deleted instance is gone
            await this.page.reload()
            await this.page.waitForLoadState('networkidle')
            await utils.waitForObjectDisappearance(tblr, 60000, `Instance deletion with timeout ${timeout}`)

            this.log(`VM instance ${instname} is successfully deleted`)
        })
    }

    // Delete SSH key by name (must exist!)
    public async VMaasDeleteKey(keyname = 'idc01-vmaas', timeout = 10000) {
        await test.step(`Delete SSH key ${keyname}`, async () => {
            await this.waitComputeConsoleIsLoaded('Keys')

            let tblr = this.page.getByRole('main').locator('table.table').locator('tr', {hasText: keyname})
            await expect(tblr.getByText('Delete')).toBeVisible({timeout: timeout})
            await tblr.getByText('Delete').click()

            // Deletion confirm dialog window
            await expect(this.page.locator('div.modal-header').getByText(/^Delete /)).toBeVisible()

            // new dialog requires to type-in name of key
            const delDlgNew = this.page.locator('div.modal-body').getByText('To confirm deletion enter the name of the key below')
            await expect(delDlgNew, 'Delete confirmation dialog must be on screen').toBeVisible()

            await expect(this.page.locator('div.modal-body').getByPlaceholder('Enter key name')).toBeVisible()
            await this.page.locator('div.modal-body').getByPlaceholder('Enter key name').fill(keyname)
            await expect(this.page.locator('div.modal-footer').getByLabel('Cancel')).toBeVisible()
            await this.page.locator('div.modal-footer').getByLabel('Delete', { exact: true }).click()
            
            this.log('key deletion is in progress..')
            
            await expect(tblr, `Timeout ${timeout} exceeded on SSH key deletion`).not.toBeVisible({timeout: timeout})

            this.log(`SSH key ${keyname} is deleted`)
        })
    }

    // *****************************************************************
    // **************************** Preview ****************************

    private async waitPreviewConsoleIsLoaded(subItem: string = 'Preview Instances', timeout = 30000) {
        // Compute / subItem page
        const itab = this.page.locator(`nav.siteToolbar > div.nav-tabs > div.nav-item > a.nav-link.active:has-text("${subItem}")`)
        if (! await itab.count())
            await this.clickNavigationLink('Go to Preview Page', `Go to ${subItem} Page`)
        else
            await itab.click()

        // wait for either table or button at the bottom (API delay)
        const instTbl = this.page.getByRole('main').locator('div.section >> table.table')
        const launchBtn = this.page.getByRole('main').getByText('Your account currently has no ')
        await expect(instTbl.or(launchBtn)).toBeVisible({timeout: timeout})
        this.log(`verified that Preview Console page (sub-item ${subItem}) is fully loaded`)
    }

    public async PreviewPreCleanup(instname = 'idc02-preview') {
        await test.step(`Preview pre-cleanup for ${instname}`, async () => {

            await this.waitPreviewConsoleIsLoaded()

            const tbl     = this.page.getByRole('main').locator('table.table')
            const tblr    = tbl.locator('tr', {hasText: instname})          // our key only, do not touch other keys
            const tblrtrm = tblr.locator('td', {hasText: 'Terminating'})    // if it's terminating status

            // delete ${instname} instance if exists
            // wait if still in "Terminating" status
            if (await tblrtrm.count()) {
                this.log(`found pre-existing instance ${instname} in terminating state, waiting..`)
                await utils.waitForObjectDisappearance(tblr, 90000, `Can not pre-cleanup instance ${instname}, it is stuck in terminating state`)
                this.log('instance has just been terminated, need to wait 1min for backend VM to unprovision')
                await utils.sleep(60000)
            }
            if (await tblr.count()) {
                this.log(`found pre-existing instance ${instname}, trying to delete it..`)
                await this.PreviewDelete(instname)
                this.log('instance has just been deleted, need to wait 1min for backend VM to unprovision')
                await utils.sleep(60000)
            }

            // delete ${instname} ssh-key if exists
            await this.waitPreviewConsoleIsLoaded('Preview Keys')
            if (await tblr.count()) {
                this.log(`found pre-existing ssh-key ${instname}, trying to delete it..`)
                await this.PreviewDeleteKey(instname)
            }

            this.log(`Preview pre-cleanup ${instname} is done`)
        })
    }

    // Create Preview Compute instance by name
    public async PreviewCreate(instname = 'idc02-preview') {
        await test.step(`Create Preview compute instance ${instname}`, async () => {
            this.log('Navigation / Preview / Compute Instances - create instance')
            await this.waitPreviewConsoleIsLoaded()

            // Create SSH keypair locally, and get private key
            if (this.envInfo.hasOwnProperty('doSshValidation') && this.envInfo['doSshValidation'])
                await this.VMaasPrepareSsh()

            // if we have some instances already:
            // const nonEmptyLaunchBtn = this.page.getByRole('main').locator('div.filter').getByText('Request instance')
            // if no instances:
            const emptyLaunchBtn = this.page.getByRole('main').locator('a[intc-id="RequestinstanceEmptyViewButton"]')
            
            // *** for IDC Preview - we may not launch more than one instance
            await expect(this.page.getByRole('main').getByText('Your account currently has no instances'),
                'Must be no active instances/requests at this point').toBeVisible()
            await expect(emptyLaunchBtn, 'Must be no active instances/requests at this point').toBeVisible()
            await emptyLaunchBtn.click()

            // Choose Instance Type to create
            const instanceType = {
                'Production': 'checkBoxTable-Instancetype-grid-pre-bm-icx',
                'Staging': 'checkBoxTable-Instancetype-grid-pre-bm-srf-sp'
            }

            // Verify all form fields
            await expect(this.page.getByRole('heading', { name: 'Instance configuration' })).toBeVisible()
            await expect(this.page.getByRole('radio', { name: 'Select CPU products' })).toBeVisible()
            await expect(this.page.getByRole('radio', { name: 'Select GPU products' })).toBeVisible()
            await expect(this.page.getByRole('radio', { name: 'Select AI products' })).toBeVisible()
            await expect(this.page.getByRole('radio', { name: 'Select AI PC products' })).toBeVisible()
            await this.page.getByRole('radio', { name: 'Select CPU products' }).check()

            await expect(this.page.getByText('Instance Type', { exact: true })).toBeVisible()
            await this.page.locator(`input[name="${instanceType[this.env]}"]`).check()

            await expect(this.page.locator('div.cost-estimate-card').getByRole('heading', { name: 'Cost estimate', exact: true })).toBeVisible()

            await expect(this.page.getByText('Machine image: *')).toBeVisible()
            // await this.page.getByLabel('Machine image: *').click()
            // await this.page.getByLabel('Select option ubuntu-2204').first().click()

            // // Compare instance types dialog
            // await this.page.getByLabel('Compare instance types').click()
            // await expect(this.page.locator('div.modal-header')).toHaveText('Compare instance types')
            // let dlg_tbl = this.page.locator('div.modal-body').getByRole('table')
            // await expect(dlg_tbl.getByRole('cell', { name: 'CPU', exact: true })).toBeVisible()
            // await expect(dlg_tbl.getByRole('cell', { name: 'GPU', exact: true })).toBeVisible()
            // utils.sleep(500)
            // expect(await dlg_tbl.locator('tr').count()).toBeGreaterThan(2)
            // await expect(this.page.locator('div.modal-content').getByRole('button', { name: 'Select', exact: true })).toBeVisible()
            // await this.page.locator('div.modal-content').getByRole('button', { name: 'Cancel', exact: true }).click()

            await expect(this.page.getByText('Use case: *')).toBeVisible()
            await expect(this.page.getByText('Duration: *')).toBeVisible()
            await expect(this.page.getByPlaceholder('e.g., AI model(s) used')).toBeVisible()
            await this.page.getByPlaceholder('e.g., AI model(s) used').fill('this is automated E2E test')

            // Upload Key dialog
            this.log('uploading Preview SSH Key')
            await this.page.getByRole('button', { name: '+ Upload Key' }).click()
            await this.page.getByRole('dialog').getByLabel('Key Name: *').fill(instname)
            await this.page.getByRole('dialog').getByLabel('Associated Email: *').fill(this.userlogin)
            await this.page.getByRole('dialog').getByLabel('Paste your key contents: *').fill(this.sshPubKey)
            await expect(this.page.getByRole('dialog').locator('div.modal-body').getByRole('button', { name: 'Cancel' })).toBeVisible()
            await this.page.getByRole('dialog').locator('div.modal-body').getByLabel('Upload key').click()
            await expect(this.page.getByRole('dialog').locator('div.modal-body').getByLabel('Upload key')).not.toBeVisible({timeout: 10000})

            // due to UI glitch, fill-out instance name now
            // await expect(this.page.getByText('Instance name: *')).toBeVisible()
            // await this.page.getByText('Instance name: *').fill(instname)
            await expect(this.page.getByPlaceholder('Instance name')).toBeVisible()
            await this.page.getByPlaceholder('Instance name').fill(instname)
            
            await this.page.locator(`input.form-check-input[value="${instname}"]`).check()
            await expect(this.page.getByLabel('Cancel', {exact: true})).toBeVisible()
            await expect(this.page.getByRole('main').getByLabel('Request instance', { exact: true })).toBeVisible()
            await this.page.getByRole('main').getByLabel('Request instance', { exact: true }).click()
            this.log('launching new VM..')
            // optional progress dialog?
            // await expect(this.page.getByText('Working on your reservation')).toBeVisible()

            const tblr = this.page.getByRole('main').locator(`table.table > tbody > tr:has-text("${instname}")`)
            await expect(tblr).toBeVisible({timeout: 30000})
            // await expect(tblr.locator('td').nth(2)).toContainText('Provisioning', {timeout: 10000})
            
            // third column will become Ready status
            this.log('launch is in progress..')
            await expect(tblr.locator('td').nth(2),
                'Timed out (3 min) waiting for Ready status of created instance').toHaveText('Ready', {timeout: 180000})
            this.log('Preview compute instance is Ready!')

            this.log(`Preview compute instance ${instname} is created`)
        })
    }

    // Verify Preview Compute instance by name
    public async PreviewVerify(instname = 'idc02-preview') {
        await test.step(`Verify Preview instance ${instname}`, async () => {
            this.log('Navigation / Preview / Compute verify instance')
            await this.waitPreviewConsoleIsLoaded()

            const tbl = this.page.getByRole('main').locator('table.table')
            const tblr = tbl.locator('tr', {hasText: instname})          // checking our instance row only

            await expect(tblr.locator('td').nth(2)).toHaveText('Ready')
            
            // Check instance info
            await tblr.locator('td').nth(0).getByRole('button', { name: instname }).click()
            // ++ tab "Details"
            let tdetr = this.page.locator('div.section:has(> h3:has-text("Instance type information"))').locator('div.row')
            // Instance type info
            await expect(tdetr.nth(0).locator('div.col-md-3').nth(2).locator('span').nth(0)).toHaveText('Status:')
            await expect(tdetr.nth(0).locator('div.col-md-3').nth(2).locator('span').nth(1)).toHaveText('Ready')
            // VM image info
            await expect(tdetr.nth(1).getByRole('heading', { name: 'Machine image information' })).toBeVisible()
            await expect(tdetr.nth(1).locator('div.col-md-3').nth(1)).toContainText(/Ubuntu|CentOS/)
            
            // ++ tab "Network"
            await this.page.getByRole('main').getByRole('button', { name: 'Networking' }).click()
            await expect(this.page.getByRole('main').getByRole('heading', { name: 'Networking interfaces' })).toBeVisible()
            tdetr = this.page.locator('div.section:has(> h3:has-text("Networking interfaces"))')
            await expect(tdetr.locator('div.flex-column > span').nth(0)).toHaveText('IP')
            const ip = await tdetr.locator('div.flex-column > span').nth(1).textContent()
            this.log(`VM has IP = ${ip}`)

            // ++ tab "Security"
            await this.page.getByRole('main').getByRole('button', { name: 'Security' }).click()
            await expect(this.page.getByRole('main').getByRole('heading', { name: 'Instance Public Keys' })).toBeVisible()
            tdetr = this.page.locator('div.section:has(> h3:has-text("Instance Public Keys"))')
            await expect(tdetr.locator('div.row').nth(0)).toContainText(instname)
            // How to connect dialog
            await this.page.getByRole('main').getByRole('button', { name: 'How to Connect' }).click()
            await expect(this.page.locator('div.modal-header').getByText('How to connect to your instance')).toBeVisible()
            await expect(this.page.locator('div.modal-body').locator('pre > span').nth(1)).toContainText('ssh -J guest@')
            this.sshConnectLine = await this.page.locator('div.modal-body').locator('pre > span').nth(1).textContent() || ''
            this.log(`ssh connect line: ${this.sshConnectLine}`)
            await this.page.locator('div.modal-footer').getByText('Close').click()

            // Actions / Edit instance
            await this.page.getByRole('main').getByRole('button', { name: 'Actions' }).click()
            await this.page.getByRole('main').getByRole('button', { name: 'Edit' }).click()
            await expect(this.page.getByRole('main').getByRole('heading', { name: 'Edit instance' })).toBeVisible()
            await expect(this.page.getByRole('main').getByPlaceholder('Instance name')).toBeDisabled()
            await expect(this.page.getByRole('main').locator('div.input-group').locator(`input.form-check-input[value="${instname}"]`)).toBeChecked()
            await this.page.getByRole('main').getByLabel('Cancel').click()

            // may need to cleanup local cache: ssh-keygen -R 192.168.12.2

            // Test by connecting via SSH and running command
            if (this.envInfo.hasOwnProperty('doSshValidation') && this.envInfo['doSshValidation'])
                expect(await this.RunSshCommand('uname', /^Linux/, 'devcloud'), 'checking by Linux uname').toEqual(true)
            else
                this.log('skipping SSH test')

            this.log(`VM instance ${instname} is verified`)
        })
    }

    // Delete Compute instance by name (must exist!)
    public async PreviewDelete(instname = 'idc02-preview', timeout = 90000) {
        await test.step(`Delete Preview VM instance ${instname}`, async () => {
            this.log('Navigation / Preview / Instances - delete VM instance')
            await this.waitPreviewConsoleIsLoaded()

            const tbl = this.page.getByRole('main').locator('div.section >> table.table')
            const tblr = tbl.locator('tr', {hasText: instname})          // checking our VM row only

            // Delete VM
            this.log('attempting to Delete the VM..')
            await expect(tblr.getByText('Delete')).toBeVisible({timeout: 10000})
            await tblr.getByText('Delete').click()

            // Deletion confirm dialog window
            await expect(this.page.locator('div.modal-header').getByText(/^Delete /)).toBeVisible()

            // new dialog requires to type-in name of instance
            const delDlgNew = this.page.locator('div.modal-body').getByText('To confirm deletion enter the name of the instance below')
            await expect(delDlgNew, 'Delete confirmation dialog must be on screen').toBeVisible()

            await expect(this.page.locator('div.modal-body').getByPlaceholder('Enter instance name')).toBeVisible()
            await this.page.locator('div.modal-body').getByPlaceholder('Enter instance name').fill(instname)
            await expect(this.page.locator('div.modal-footer').getByLabel('Cancel')).toBeVisible()
            await this.page.locator('div.modal-footer').getByLabel('Delete', { exact: true }).click()
            
            this.log('instance deletion is in progress..')

            // set timer to reload page (SRE-3341)
            // const reloadTimer = setTimeout(() => this.page.reload(), 60000)

            // wait for row or entire table disappearance
            await expect(tblr, `Timeout ${timeout} ms exceeded on instance deletion`).not.toBeVisible({timeout: timeout})

            // clearTimeout(reloadTimer)

            this.log(`Preview VM instance ${instname} is successfully deleted`)
        })
    }

    // Delete SSH key by name (must exist!)
    public async PreviewDeleteKey(keyname = 'idc02-preview', timeout = 10000) {
        await test.step(`Delete Preview SSH key ${keyname}`, async () => {
            await this.waitPreviewConsoleIsLoaded('Preview Keys')

            let tblr = this.page.getByRole('main').locator('table.table').locator('tr', {hasText: keyname})
            await expect(tblr.getByText('Delete')).toBeVisible({timeout: timeout})
            await tblr.getByText('Delete').click()

            // Deletion confirm dialog window
            await expect(this.page.locator('div.modal-header').getByText(/^Delete /)).toBeVisible()

            // new dialog requires to type-in name of key
            const delDlgNew = this.page.locator('div.modal-body').getByText('To confirm deletion enter the name of the key below')
            await expect(delDlgNew, 'Delete confirmation dialog must be on screen').toBeVisible()

            await expect(this.page.locator('div.modal-body').getByPlaceholder('Enter key name')).toBeVisible()
            await this.page.locator('div.modal-body').getByPlaceholder('Enter key name').fill(keyname)
            await expect(this.page.locator('div.modal-footer').getByLabel('Cancel')).toBeVisible()
            await this.page.locator('div.modal-footer').getByLabel('Delete', { exact: true }).click()
            
            this.log('key deletion is in progress..')

            await expect(tblr, `Timeout ${timeout} exceeded on SSH key deletion`).not.toBeVisible({timeout: timeout})

            this.log(`Preview SSH key ${keyname} is deleted`)
        })
    }


    // ******************************************************************
    // **************************** STaaS-FS ****************************

    private async waitStaasIsLoaded(subItem: string = 'File Storage', timeout = 30000) {
        // Compute / subItem page
        const itab = this.page.locator(`nav.siteToolbar > div.nav-tabs > div.nav-item > a.nav-link.active:has-text("${subItem}")`)
        if (! await itab.count())
            await this.clickNavigationLink('Go to Storage Page', `Go to ${subItem} Page`)
        else
            await itab.click()

        // wait for either table or button at the bottom (API delay)
        const instTbl = this.page.getByRole('main').locator('div.section >> table.table')
        const launchBtn = this.page.getByRole('main').getByText('Your account currently has no ')
        await expect(instTbl.or(launchBtn)).toBeVisible({timeout: timeout})
        this.log(`verified that Storage page (sub-item ${subItem}) is fully loaded`)
    }

    public async StaasFsPreCleanup(instname = 'idc03-staas') {
        await test.step(`STaaS-FS pre-cleanup for ${instname}`, async () => {

            await this.waitStaasIsLoaded()

            const tbl     = this.page.getByRole('main').locator('table.table')
            const tblr    = tbl.locator('tr', {hasText: instname})       // our volume only, do not touch other volumes
            const tblrtrm = tblr.locator('td', {hasText: 'Deleting'})    // if it's deleting status

            // delete ${instname} if exists
            // wait if still in "Deleting" status
            if (await tblrtrm.count()) {
                this.log(`found pre-existing volume ${instname} in deleting state, waiting..`)
                await utils.waitForObjectDisappearance(tblr, 90000, `Can not pre-cleanup volume ${instname}, it is stuck in deleting state`);
            }
            if (await tblr.count()) {
                this.log(`found pre-existing volume ${instname}, trying to delete it..`)
                await this.StaasFsDelete(instname)
            }

            this.log(`STaaS-FS pre-cleanup ${instname} is done`)
        })
    }

    // Create STaaS-FS volume by name
    public async StaasFsCreate(instname = 'idc03-staas') {
        await test.step(`Create STaaS-FS volume ${instname}`, async () => {
            this.log('Navigation / Storage / File Storage - create volume')
            await this.waitStaasIsLoaded()

            // if we have some volume already:
            const nonEmptyLaunchBtn = this.page.getByRole('main').locator('div.filter').getByText('Create Volume')
            // if no volumes:
            const emptyLaunchBtn = this.page.getByRole('main').locator('a[intc-id="CreatevolumeEmptyViewButton"]')

            if (await emptyLaunchBtn.isVisible())
                await emptyLaunchBtn.click()
            else 
                await nonEmptyLaunchBtn.click()

            await expect(this.page.getByRole('main').getByRole('heading', { name: 'Create a storage volume' })).toBeVisible()

            // Verify all form fields
            await expect(this.page.getByPlaceholder('Name')).toBeVisible()
            await this.page.getByPlaceholder('Name').fill(instname)
            await expect(this.page.getByLabel('Storage Size (TB): *')).toBeVisible()
            await this.page.getByLabel('Storage Size (TB): *').fill('1')
            await expect(this.page.getByLabel('Cancel')).toBeVisible()
            await expect(this.page.getByLabel('Launch')).toBeVisible()
            await this.page.getByLabel('Launch').click()
            this.log('volume creation is in progress..')

            const tblr = this.page.getByRole('main').locator(`table.table > tbody > tr:has-text("${instname}")`)
            await expect(tblr).toBeVisible({timeout: 30000})
            
            // third column will become Ready status
            await expect(tblr.locator('td').nth(2),
                'Timed out (1.5 min) waiting for Ready status of created volume').toHaveText('Ready', {timeout: 90000})
            this.log('STaaS-FS volume is Ready!')

            this.log(`STaaS-FS volume ${instname} is created`)
        })
    }

    // Delete STaaS-FS volume by name (must exist!)
    public async StaasFsDelete(instname = 'idc03-staas', timeout = 60000) {
        await test.step(`Delete STaaS-FS volume ${instname}`, async () => {
            this.log('Navigation / Storage / File Storage - delete volume')
            await this.waitStaasIsLoaded()

            const tbl = this.page.getByRole('main').locator('div.section >> table.table')
            const tblr = tbl.locator('tr', {hasText: instname})          // checking our volume row only

            // Delete volume
            this.log('attempting to Delete the volume..')
            await expect(tblr.getByText('Delete')).toBeVisible({timeout: 10000})
            await tblr.getByText('Delete').click()

            // Deletion confirm dialog window
            await expect(this.page.locator('div.modal-header').getByText(/^Delete /)).toBeVisible()

            // new dialog requires to type-in name of instance
            const delDlgNew = this.page.locator('div.modal-body').getByText('To confirm deletion enter the name')
            await expect(delDlgNew, 'Delete confirmation dialog must be on screen').toBeVisible()

            await expect(this.page.locator('div.modal-body').getByPlaceholder('Enter storage name')).toBeVisible()
            await this.page.locator('div.modal-body').getByPlaceholder('Enter storage name').fill(instname)
            await expect(this.page.locator('div.modal-footer').getByLabel('Cancel')).toBeVisible()
            await this.page.locator('div.modal-footer').getByLabel('Delete', { exact: true }).click()
            
            this.log('volume deletion is in progress..')

            // wait for row or entire table disappearance
            await expect(tblr, `Timeout exceeded on volume deletion`).not.toBeVisible({timeout: timeout})

            this.log(`STaaS-FS volume ${instname} is successfully deleted`)
        })
    }

    // Verify STaaS-FS volume by name
    public async StaasFsVerify(instname = 'idc03-staas') {
        await test.step(`Verify STaaS-FS volume ${instname}`, async () => {
            this.log('Navigation / Storage / File Storage verify mounted volume')
            await this.waitStaasIsLoaded()

            const tbl = this.page.getByRole('main').locator('table.table')
            const tblr = tbl.locator('tr', {hasText: instname})          // checking our VM row only

            await expect(tblr.locator('td').nth(2)).toHaveText('Ready')
            this.log(`seeing volume ${instname} in Ready state`)
            
            // Check volume info
            this.log('checking volume info..')
            await tblr.locator('td').nth(0).getByRole('button', { name: instname }).click()
            // ++ tab "Details"
            let tdetr = this.page.locator('div.section:has(> h3:has-text("Volume information"))').locator('div.row')
            // Volume info - ID
            await expect(tdetr.nth(0).locator('div.flex-column').nth(0).locator('span').nth(0)).toContainText('Volume ID:')
            const volId = await tdetr.nth(0).locator('div.flex-column').nth(0).locator('span').nth(1).textContent()
            this.log(`seeing volume ID: ${volId}`)
            // Volume info - status
            await expect(tdetr.nth(0).locator('div.flex-column').nth(2).locator('span').nth(0)).toContainText('State:')
            await expect(tdetr.nth(0).locator('div.flex-column').nth(2).locator('span').nth(1)).toHaveText('Ready')
            this.log('volume status is Ready')
            // Volume info - cluster name
            await expect(tdetr.nth(1).locator('div.flex-column').nth(0).locator('span').nth(0)).toContainText('Cluster')
            const clusterName = await tdetr.nth(1).locator('div.flex-column').nth(0).locator('span').nth(1).textContent()
            this.log(`seeing cluster name: ${clusterName}`)
            // Volume info - encryption - disable check for a while. TODO: re-enabled once STaas fixes API (encrypted --> Encrypted)
            await expect(tdetr.nth(1).locator('div.flex-column').nth(1).locator('span').nth(0)).toContainText('Encryption')
            await expect(tdetr.nth(1).locator('div.flex-column').nth(1).locator('span').nth(1)).toHaveText('Enabled')
            this.log('volume encryption is Enabled')
            // Volume info - namespace
            await expect(tdetr.nth(1).locator('div.flex-column').nth(2).locator('span').nth(0)).toContainText('Namespace')
            const nameSpace = await tdetr.nth(1).locator('div.flex-column').nth(2).locator('span').nth(1).textContent()
            this.log(`seeing namespace: ${nameSpace}`)

            // ++ tab "Security"
            this.log('checking security tab..')
            await this.page.getByRole('main').getByRole('button', { name: 'Security' }).click()
            await expect(this.page.getByRole('main').getByRole('heading', { name: 'Volume credentials' })).toBeVisible()

            tdetr = this.page.locator('div.section:has(> h3:has-text("Volume credentials"))')
            await expect(tdetr).toContainText('User:')
            await expect(tdetr.getByLabel('Copy User')).toBeVisible()
            await tdetr.getByLabel('Copy User').click()
            this.staasFsUser = await this.page.evaluate("navigator.clipboard.readText()")
            this.log(`copied user name: ${this.staasFsUser}`)
            this.log('generating password..')
            await tdetr.getByRole('button', { name: 'Generate password' }).click()
            await expect(tdetr).toContainText('Password:', {timeout: 30000})
            await expect(tdetr.getByLabel('Copy Password')).toBeVisible()
            await tdetr.getByLabel('Copy Password').click()
            this.staasFsPass = await this.page.evaluate("navigator.clipboard.readText()")
            this.log(`generated and copied password: ${this.staasFsPass}`)

            // How to mount dialog
            this.log('checking How to mount dialog..')
            await this.page.getByRole('main').getByRole('button', { name: 'How to mount' }).click()
            await expect(this.page.locator('div.modal-header').getByText('How to mount storage volume')).toBeVisible()
            await this.page.locator('div.modal-body').getByRole('button', { name: 'Single instance' }).click()
            await this.page.locator('div.modal-body').getByLabel('Instance: *').click()
            await expect(this.page.locator('div.modal-body').getByLabel(`Select option ${instname} -IP: `)).toBeVisible()
            await this.page.locator('div.modal-body').getByLabel(`Select option ${instname} -IP: `).click()
            // ssh connect line
            await this.page.locator('div.modal-body').getByRole('button', { name: 'Connect to your instance:' }).click()
            await expect(this.page.locator('div.modal-body').locator('pre > span').nth(1)).toContainText('ssh -J guest@')
            this.sshConnectLine = await this.page.locator('div.modal-body').locator('pre > span').nth(1).textContent() || ''
            // eg: ssh -J guest@146.152.227.17 ubuntu@100.83.56.181
            this.log(`ssh connect line: ${this.sshConnectLine}`)
            await this.page.locator('div.modal-body').getByRole('button', { name: 'Connect to your instance:' }).click()
            
            // mount volume commands
            await this.page.locator('div.modal-body').getByRole('button', { name: 'Mount your volume:' }).click()
            await this.page.locator('div.modal-body').locator('button:has-text("Copy")').nth(2).click()
            const cmdInstallWeka = await this.page.evaluate("navigator.clipboard.readText()")
            // eg: curl http://pdx05-dev-2.us-staging-1.cloud.intel.com:14000/dist/v1/install | sudo sh
            
            await this.page.locator('div.modal-body').locator('button:has-text("Copy")').nth(3).click()
            const cmdWekaLogin = await this.page.evaluate("navigator.clipboard.readText()")
            // eg: weka user login -H pdx05-dev-2.us-staging-1.cloud.intel.com --org ns889513527774 u889513527774

            await this.page.locator('div.modal-body').locator('button:has-text("Copy")').nth(5).click()
            const cmdWekaMount = await this.page.evaluate("navigator.clipboard.readText()")
            // eg: sudo weka mount -t wekafs -o net=udp pdx05-dev-2.us-staging-1.cloud.intel.com/test04 /mnt/test
            await this.page.locator('div.modal-footer').getByText('Close').click()

            // How to unmount dialog
            this.log('checking How to unmount dialog..')
            await this.page.getByRole('main').getByRole('button', { name: 'How to unmount' }).click()
            await expect(this.page.locator('div.modal-header').getByText('How to unmount storage volume')).toBeVisible()
            await this.page.locator('div.modal-body').getByRole('button', { name: 'Single instance' }).click()
            await this.page.locator('div.modal-body').getByLabel('Instance: *').click()
            await expect(this.page.locator('div.modal-body').getByLabel(`Select option ${instname} -IP: `)).toBeVisible()
            await this.page.locator('div.modal-body').getByLabel(`Select option ${instname} -IP: `).click()
            // connect and unmount volume commands
            await this.page.locator('div.modal-body').getByRole('button', { name: 'Connect to your instance:' }).click()
            await expect(this.page.locator('div.modal-body').locator('button:has-text("Copy")').nth(0)).toBeVisible()
            await expect(this.page.locator('div.modal-body').locator('button:has-text("Copy")').nth(1)).toBeVisible()
            await this.page.locator('div.modal-body').getByRole('button', { name: 'Connect to your instance:' }).click()
            await this.page.locator('div.modal-body').getByRole('button', { name: 'Unmount your volume:' }).click()
            await expect(this.page.locator('div.modal-body').locator('button:has-text("Copy")').nth(2)).toBeVisible()
            await this.page.locator('div.modal-footer').getByText('Close').click()

            // Actions / Edit instance
            this.log('checking Actions/Edit/Delete..')
            await this.page.getByRole('main').getByRole('button', { name: 'Actions' }).click()
            await expect(this.page.getByRole('main').getByRole('button', { name: 'Delete' })).toBeVisible()
            // ?? Check why in Prod it's absent ??
            // await expect(this.page.getByRole('main').getByRole('button', { name: 'Edit' })).toBeVisible()
            // await this.page.getByRole('main').getByRole('button', { name: 'Edit' }).click()
            // await expect(this.page.getByRole('main').getByRole('heading', { name: 'Edit storage volume' })).toBeVisible()
            // await expect(this.page.getByRole('main').getByLabel('Storage Size (TB): *')).toBeVisible()
            // await this.page.getByRole('main').getByLabel('Storage Size (TB): *').fill('2')
            // await expect(this.page.getByRole('main').getByLabel('Edit')).toBeVisible()
            // await expect(this.page.getByRole('main').getByLabel('Cancel')).toBeVisible()
            // await this.page.getByRole('main').getByLabel('Cancel').click()

            // if (this.envInfo.hasOwnProperty('doSshValidation') && this.envInfo['doSshValidation'])

            // SSH access is mandatory
            this.log('checking basic ssh connectivity..')
            expect(await this.RunSshCommand('uname', /^Linux/), 'checking SSH connectivity').toEqual(true)
            this.log('creating test directory..')
            expect(await this.RunSshCommand('sudo mkdir /mnt/test'), 'creating test dir').toEqual(true)
            this.log('installing weka cli (30-sec timeout)..')
            expect(await this.RunSshCommand(String(cmdInstallWeka), /Installation finished successfully/, '', 30000), 'making Weka CLI istallation').toEqual(true)
            this.log('making weka login (30-sec timeout)..')
            // Some symbols in pass must be escaped (eg, backticks)
            let escapedFsPass = this.staasFsPass
                .replace('`', '\\`')
                .replace('\'', '\\\'')
                .replace('"', '\\"')
                .replace('$', '\\$')
            
            expect(await this.RunSshCommand(`${cmdWekaLogin} "${escapedFsPass}"`, /Login completed successfully/, '', 30000), 'making Weka login').toEqual(true)
            
            this.log('mounting weka volume (7-min timeout)..')
            const mount_result = await this.RunSshCommand(String(cmdWekaMount), /Mount completed successfully/, '', 420000)
            // now if command failed, show weka log file
            if (! mount_result) {
                this.log('*** DEBUG: cat /opt/weka/data/dependencies/build-20*.log')
                await this.RunSshCommand(`sudo cat /opt/weka/data/dependencies/build-20*.log`, /Random string to show the stdout here../, '', 30000)
            }
            expect(mount_result, 'making Weka volume mount').toEqual(true)

            // Verifications
            this.log('making FS mount verification..')
            expect(await this.RunSshCommand('mount| grep wekafs', new RegExp(`${instname} on /mnt/test type wekafs`)), 'checking Weka volume mount').toEqual(true)
            expect(await this.RunSshCommand('df -h /mnt/test/', new RegExp(instname)), 'checking by df command').toEqual(true)

            // Unmount volume
            this.log('making unmount..')
            expect(await this.RunSshCommand('sudo umount /mnt/test', /Umount completed successfully/), 'making Weka volume unmount').toEqual(true)

            this.log(`STaaS-FS volume ${instname} is verified`)
        })
    }

    // **************************** STaaS-OS ****************************

    public async StaasOsPreCleanup(instname = 'idc04-staas') {
        await test.step(`STaaS-OS pre-cleanup for ${instname}`, async () => {

            await this.waitStaasIsLoaded('Object Storage')

            const tbl     = this.page.getByRole('main').locator('table.table')
            const tblr    = tbl.locator('tr', {hasText: instname})       // our bucket only, do not touch other buckets
            const tblrtrm = tblr.locator('td', {hasText: 'Deleting'})    // if it's deleting status

            // delete ${instname} if exists
            // wait if still in "Deleting" status
            if (await tblrtrm.count()) {
                this.log(`found pre-existing bucket ${instname} in deleting state, waiting..`)
                await utils.waitForObjectDisappearance(tblr, 90000, `Can not pre-cleanup bucket ${instname}, it is stuck in deleting state`);
            }
            if (await tblr.count()) {
                this.log(`found pre-existing bucket ${instname}, trying to delete it..`)
                await this.StaasOsDelete(instname)
            }

            this.log(`STaaS-OS pre-cleanup ${instname} is done`)
        })
    }

    // Create STaaS-OS bucket by name
    public async StaasOsCreate(instname = 'idc04-staas') {
        await test.step(`Create STaaS-OS bucket ${instname}`, async () => {
            this.log('Navigation / Storage / Object Storage - create bucket')
            await this.waitStaasIsLoaded('Object Storage')

            // if we have some bucket already:
            const nonEmptyLaunchBtn = this.page.getByRole('main').locator('div.filter').getByText('Create bucket')
            // if no buckets:
            const emptyLaunchBtn = this.page.getByRole('main').locator('a[intc-id="CreatebucketEmptyViewButton"]')

            if (await emptyLaunchBtn.isVisible())
                await emptyLaunchBtn.click()
            else 
                await nonEmptyLaunchBtn.click()

            await expect(this.page.getByRole('main').getByRole('heading', { name: 'Create storage bucket' })).toBeVisible()

            // Verify all form fields
            await expect(this.page.getByPlaceholder('Name')).toBeVisible()
            await this.page.getByPlaceholder('Name').fill(instname)
            await expect(this.page.getByPlaceholder('Description')).toBeVisible()
            await this.page.getByPlaceholder('Description').fill('this is automated E2E test')
            await expect(this.page.getByLabel('Enable versioning')).toBeVisible()
            await expect(this.page.getByLabel('Enable versioning')).not.toBeChecked()

            await utils.sleep(1000)
            await expect(this.page.getByLabel('Cancel', { exact: true })).toBeVisible()
            await expect(this.page.getByLabel('Create', { exact: true })).toBeVisible()
            await this.page.getByLabel('Create', { exact: true }).click()
            this.log('bucket creation is in progress..')

            const tblr = this.page.getByRole('main').locator(`table.table > tbody > tr:has-text("${instname}")`)
            await expect(tblr).toBeVisible({timeout: 30000})
            
            // third column will become Ready status
            await expect(tblr.locator('td').nth(2),
                'Timed out (2 min) waiting for Ready status of created bucket').toHaveText('Ready', {timeout: 90000})
            this.log('STaaS-OS bucket is Ready!')

            this.staasOsFullName = String(await tblr.locator('td').nth(0).textContent())

            // -- Principal --
            this.log('creating principal..')
            await expect(this.page.getByRole('link', { name: 'Manage principals and permissions' })).toBeVisible()
            await this.page.getByRole('link', { name: 'Manage principals and permissions' }).click()
            await expect(this.page.getByRole('heading', { name: 'Manage Principals and Permissions' })).toBeVisible()
            // button "create principal"
            const nonEmptyPrincBtn = this.page.getByRole('main').locator('div.filter').getByText('Create principal')
            const emptyPrincBtn = this.page.getByRole('main').locator('a[intc-id="CreateprincipalEmptyViewButton"]')
            await utils.waitForObject(emptyPrincBtn, 10000, '', true)
            // await expect(emptyPrincBtn.or(nonEmptyPrincBtn)).toBeVisible({timeout: 10000})   // does not work, nonEmptyPrincBtn in hidden div
          
            if (await emptyPrincBtn.isVisible())
                await emptyPrincBtn.click()
            else 
                await nonEmptyPrincBtn.click()

            await expect(this.page.getByPlaceholder('Name')).toBeVisible()
            await this.page.getByPlaceholder('Name').fill(instname)
            await expect(this.page.getByText('Select/Deselect All').nth(0)).toBeVisible()
            await expect(this.page.getByText('Select/Deselect All').nth(1)).toBeVisible()
            await this.page.getByText('Select/Deselect All').nth(0).click()
            await this.page.getByText('Select/Deselect All').nth(1).click()
            await expect(this.page.getByLabel('Cancel', { exact: true })).toBeVisible()
            await expect(this.page.getByLabel('Create', { exact: true })).toBeVisible()
            await this.page.getByLabel('Create', { exact: true }).click()

            const tblr1 = this.page.getByRole('main').locator(`table.table > tbody > tr:has-text("${instname}")`)
            await expect(tblr1).toBeVisible({timeout: 30000})
            
            // second column will become Ready status
            await expect(tblr1.locator('td').nth(1),
                'Timed out (1 min) waiting for Ready status of created principal').toHaveText('Ready', {timeout: 60000})
            this.log('principal is Ready!')

            // few checks
            await expect(this.page.locator('div.filter').getByRole('link', { name: 'Create principal' })).toBeVisible()
            await expect(this.page.locator('div.filter').getByRole('link', { name: 'View buckets' })).toBeVisible()
            expect(await this.page.locator('div.filter').getByRole('link', { name: 'View buckets' }).getAttribute('href')).toEqual('/buckets')
            await expect(this.page.locator('div.filter').getByPlaceholder('Search principals...')).toBeVisible()

            // check edit principal menu
            await tblr1.getByLabel('Edit principal').click()
            await expect(this.page.getByRole('heading', { name: `Edit principal - ${instname}` })).toBeVisible()
            await expect(this.page.getByText('Select/Deselect All').nth(0)).toBeVisible()
            await expect(this.page.getByText('Select/Deselect All').nth(1)).toBeVisible()
            await expect(this.page.getByText('Select/Deselect All').nth(0)).toBeChecked()
            await expect(this.page.getByText('Select/Deselect All').nth(1)).toBeChecked()
            await expect(this.page.getByRole('button', { name: 'Save' })).toBeVisible()
            await expect(this.page.getByRole('button', { name: 'Cancel' })).toBeVisible()
            await this.page.getByRole('button', { name: 'Cancel' }).click()
            await expect(tblr1).toBeVisible({timeout: 30000})

            // edit principal
            await tblr1.getByText(instname).click()
            await expect(this.page.getByRole('heading', { name: `Principal: ${instname}` })).toBeVisible()
            await expect(this.page.getByRole('button', { name: 'Actions' })).toBeVisible()
            await this.page.getByRole('button', { name: 'Actions' }).click()
            await expect(this.page.getByRole('button', { name: 'Edit' })).toBeVisible()
            await expect(this.page.getByRole('button', { name: 'Delete' })).toBeVisible()
            await expect(this.page.getByRole('heading', { name: 'Bucket credentials' })).toBeVisible()
            this.log('generating credentials..')
            await expect(this.page.getByLabel('Generate password')).toBeVisible()
            await this.page.getByLabel('Generate password').click()
            await expect(this.page.getByText('Access Key:')).toBeVisible({timeout: 10000})
            await this.page.getByLabel('Copy access key').click()
            this.staasOsAccessKey = await this.page.evaluate("navigator.clipboard.readText()")
            this.log('copied access key')
            await expect(this.page.getByText('Secret Key:')).toBeVisible()
            await this.page.getByLabel('Copy secret key').click()
            this.staasOsSecretKey = await this.page.evaluate("navigator.clipboard.readText()")
            this.log('copied secret key')
            
            await this.page.getByRole('button', { name: 'Permissions' }).click()
            await expect(this.page.getByRole('heading', { name: 'Buckets permissions' })).toBeVisible()
            await expect(this.page.getByText('GetBucketLocation', { exact: true })).toBeVisible()
            await expect(this.page.getByText('GetBucketPolicy', { exact: true })).toBeVisible()
            await expect(this.page.getByText('ListBucket', { exact: true })).toBeVisible()
            await expect(this.page.getByText('ReadBucket', { exact: true })).toBeVisible()
            await expect(this.page.getByText('WriteBucket', { exact: true })).toBeVisible()
            await expect(this.page.getByText('DeleteBucket', { exact: true })).toBeVisible()

            this.log(`STaaS-OS bucket ${instname} is created`)
        })
    }

    // Delete STaaS-OS bucket by name (must exist!)
    public async StaasOsDelete(instname = 'idc04-staas', timeout = 60000) {
        await test.step(`Delete STaaS-OS bucket ${instname}`, async () => {
            this.log('Navigation / Storage / Object Storage - delete bucket')
            await this.waitStaasIsLoaded('Object Storage')

            const tbl = this.page.getByRole('main').locator('div.section >> table.table')
            const tblr = tbl.locator('tr', {hasText: instname})          // checking our row only

            // Delete bucket
            this.log('attempting to Delete the bucket..')
            await expect(tblr.getByText('Delete')).toBeVisible({timeout: 10000})
            await tblr.getByText('Delete').click()

            // Deletion confirm dialog window
            await expect(this.page.locator('div.modal-header').getByText(/^Delete /)).toBeVisible()

            const staasOsFullName = String(await tblr.locator('td').nth(0).textContent())

            // while new dialog requires to type-in name of instance
            const delDlgNew = this.page.locator('div.modal-body').getByText('To confirm deletion enter the name')
            await expect(delDlgNew, 'Delete confirmation dialog must be on screen').toBeVisible()

            await expect(this.page.locator('div.modal-body').getByPlaceholder('Enter bucket name')).toBeVisible()
            await this.page.locator('div.modal-body').getByPlaceholder('Enter bucket name').fill(staasOsFullName)
            await expect(this.page.locator('div.modal-footer').getByLabel('Cancel')).toBeVisible()
            await this.page.locator('div.modal-footer').getByLabel('Delete', { exact: true }).click()
            
            this.log('bucket deletion is in progress..')

            // wait for row or entire table disappearance
            await expect(tblr, `Timeout exceeded on bucket deletion`).not.toBeVisible({timeout: timeout})

            await expect(this.page.locator('div.modal-body').getByText('Your bucket was deleted')).toBeVisible()
            await this.page.locator('div.modal-footer').getByRole('button', { name: 'OK' }).click()

            // delete Principal (if exists)
            this.log('checking if we need to delete principal..')
            await expect(this.page.getByRole('link', { name: 'Manage principals and permissions' })).toBeVisible()
            await this.page.getByRole('link', { name: 'Manage principals and permissions' }).click()

            const instTbl = this.page.getByRole('main').locator('div.section >> table.table')
            const noObjMsg = this.page.getByRole('main').getByText('Your account currently has no ')
            await expect(instTbl.or(noObjMsg)).toBeVisible({timeout: 30000})
            await utils.sleep(500)

            const tblr1 = this.page.getByRole('main').locator(`table.table > tbody > tr:has-text("${instname}")`)
            if (await tblr1.isVisible()) {
                this.log(`seeing principal ${instname}, deleting it..`)
                await tblr1.getByLabel('Delete').click()
                const delPrincDlg = this.page.locator('div.modal-body').getByText('To confirm deletion enter the name')
                await expect(delPrincDlg, 'Delete confirmation dialog must be on screen').toBeVisible()
                await expect(this.page.locator('div.modal-body').getByPlaceholder('Enter principal name')).toBeVisible()
                await this.page.locator('div.modal-body').getByPlaceholder('Enter principal name').fill(instname)
                await expect(this.page.locator('div.modal-footer').getByLabel('Cancel')).toBeVisible()
                await this.page.locator('div.modal-footer').getByLabel('Delete', { exact: true }).click()
                // wait for row or entire table disappearance
                await expect(tblr1, `Timeout exceeded on principal deletion`).not.toBeVisible({timeout: 30000})
                this.log(`principal ${instname} has been deleted`)
            }
            else
                this.log(`seeing that account has no ${instname} principal`)

            this.log(`STaaS-OS bucket ${instname} is successfully deleted`)
        })
    }

    // Verify STaaS-OS volume by name
    public async StaasOsVerify(instname = 'idc04-staas') {
        await test.step(`Verify STaaS-OS volume ${instname}`, async () => {
            this.log('Navigation / Storage / Object Storage verify bucket')
            await this.waitStaasIsLoaded('Object Storage')

            const tbl = this.page.getByRole('main').locator('table.table')
            const tblr = tbl.locator('tr', {hasText: instname})          // checking our row only

            await expect(tblr.locator('td').nth(2)).toHaveText('Ready')
            this.log(`seeing bucket ${instname} in Ready state`)
            
            // Check bucket info
            this.log('checking bucket info..')
            await tblr.locator('td').nth(0).getByText(instname).click()

            await expect(this.page.locator('a[intc-id="btn how-to-connect"]')).toBeVisible()
            expect(await this.page.locator('a[intc-id="btn how-to-connect"]').getAttribute('href')).toEqual('/docs/guides/staas_object.html')
            await expect(this.page.getByRole('button', { name: 'Actions' })).toBeVisible()
            await this.page.getByRole('button', { name: 'Actions' }).click()
            await expect(this.page.getByRole('button', { name: 'Delete' })).toBeVisible()

            // ++ tab "Details"
            let tdetr = this.page.locator('div.section:has(> h3:has-text("Bucket information"))').locator('div.row')
            // Bucket info - ID
            await expect(tdetr.nth(0).locator('div.flex-column').nth(0).locator('span').nth(0)).toContainText('Bucket ID:')
            const bucketId = await tdetr.nth(0).locator('div.flex-column').nth(0).locator('span').nth(1).textContent()
            this.log(`seeing bucket ID: ${bucketId}`)
            // Bucket info - status
            await expect(tdetr.nth(0).locator('div.flex-column').nth(3).locator('span').nth(0)).toContainText('State:')
            await expect(tdetr.nth(0).locator('div.flex-column').nth(3).locator('span').nth(1)).toHaveText('Ready')
            this.log('seeing bucket status: Ready')

            // Bucket info - endpoint URL
            await expect(tdetr.nth(1).locator('div.flex-column').nth(0).locator('span').nth(0)).toContainText('Private Endpoint URL:')
            // const endpointUrl = await tdetr.nth(1).locator('div.flex-column').nth(0).locator('span').nth(1).textContent()
            await this.page.getByLabel('Copy Private Endpoint URL').click()
            const endpointUrl  = await this.page.evaluate("navigator.clipboard.readText()")
            this.log(`seeing endpoint URL: ${endpointUrl}`)
            // Bucket info - security group
            await expect(tdetr.nth(1).locator('div.flex-column').nth(1).locator('span').nth(0)).toContainText('Network Security Group')
            const securityGroup = await tdetr.nth(1).locator('div.flex-column').nth(1).locator('span').nth(1).textContent()
            this.log(`seeing security group: ${securityGroup}`)

            // ++ tab "Principals"
            this.log('checking principals tab..')
            await this.page.getByRole('button', { name: 'Principals' }).click()
            expect(await this.page.getByRole('link', { name: 'Manage buckets principals and' }).getAttribute('href')).toEqual('/buckets/users')

            // ++ tab "Lifecycle Rules"
            this.log('checking lifecycle rules tab..')
            await this.page.getByRole('main').getByRole('button', { name: 'Lifecycle Rules' }).click()
            await this.page.getByRole('link', { name: 'Create rule' }).click()
            await expect(this.page.getByRole('heading', { name: 'Add Lifecycle Rule' })).toBeVisible()
            await expect(this.page.getByLabel('Name: *')).toBeVisible()
            await expect(this.page.getByLabel('Prefix:')).toBeVisible()
            await expect(this.page.getByText('Delete Marker')).toBeVisible()
            await expect(this.page.getByText('Delete Marker')).toBeChecked()
            await expect(this.page.getByText('Expiry Days:', { exact: true })).toBeVisible()
            await expect(this.page.getByRole('textbox').nth(2)).toBeVisible()
            await expect(this.page.getByLabel('Non current expiry days:')).toBeVisible()
            await expect(this.page.getByLabel('Add')).toBeVisible()
            await expect(this.page.getByLabel('Cancel')).toBeVisible()
            await this.page.getByLabel('Cancel').click()

            // Next we will connect to compute instance via SSH and will run remote commands
            // sudo DEBIAN_FRONTEND=noninteractive NEEDRESTART_MODE=a apt-get update && sudo DEBIAN_FRONTEND=noninteractive apt-get install awscli -y
            // aws --version
            // AWS_ACCESS_KEY_ID='xx' AWS_SECRET_ACCESS_KEY='xx' aws s3 ls --endpoint-url=https://s3-pdx11-2.tenantiglb.us-staging-3.cloud.intel.com:9000

            this.log('will run remote shell commands on compute instance now')

            // basic check already done in VMaas verification, skip here
            // this.log('checking basic ssh connectivity..')
            // expect(await this.RunSshCommand(), 'checking SSH connectivity').toEqual(true)

            // await utils.sleep(15000)    // trying to deal with timeouts in us-region-1

            this.log('disable & stop unattended-upgrades (45-sec timeout)..')     // in us-staging-1 it takes 2 min!
            await this.RunSshCommand('sudo systemctl disable --now unattended-upgrades', /^$/, '', 45000)
            await this.RunSshCommand('sudo systemctl stop unattended-upgrades', /^$/, '', 45000)
            
            this.log('running apt-get update (3-min timeout)..')
            expect(await this.RunSshCommand('sudo DEBIAN_FRONTEND=noninteractive NEEDRESTART_MODE=a apt-get update', /^$/, '', 180000), 'run apt-get update').toEqual(true)
            
            this.log('installing awscli (7-min timeout)..')     // in us-region-1 it takes >5 min
            expect(await this.RunSshCommand('sudo DEBIAN_FRONTEND=noninteractive NEEDRESTART_MODE=a apt-get install awscli -y', /^$/, '', 420000), 'install awscli').toEqual(true)
            
            // Verifications
            this.log('checking awscli version (30-sec timeout)..')
            expect(await this.RunSshCommand('aws --version', /aws-cli\/.+ Python\/.+ Linux\/.+ botocore\/.+/, '', 30000), 'check awscli version').toEqual(true)
            
            // Some symbols in pass must be escaped (eg, backticks)
            let escapedPass = this.staasOsSecretKey
                .replace('`', '\\`')
                .replace('\'', '\\\'')
                .replace('"', '\\"')
                .replace('$', '\\$')

            this.log('checking bucket by aws s3 (30-sec timeout)..')
            const commAwsS3 = `AWS_ACCESS_KEY_ID="${this.staasOsAccessKey}" AWS_SECRET_ACCESS_KEY="${escapedPass}" aws s3 ls --endpoint-url=${endpointUrl}`
            expect(await this.RunSshCommand(commAwsS3, new RegExp(this.staasOsFullName), '', 30000), 'check bucket by aws s3').toEqual(true)

            // We also can copy and verify file
            // echo "this is test file" > idc04-testfile
            // aws s3 cp idc04-testfile s3://249241798056-idc04-staas
            // aws s3 ls 249241798056-idc04-staas
            // aws s3 rm s3://249241798056-idc04-staas/idc04-testfile

            this.log(`STaaS-OS bucket ${instname} is verified`)
        })
    }
    
}
