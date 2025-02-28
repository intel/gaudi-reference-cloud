import { Page, expect, Browser, BrowserContext } from "@playwright/test";
import * as utils from './utils';

const SMALL_SCREEN_SIZE = { width: 1280, height: 720 }

export class Devcloud {
    browser: Browser;
    context!: BrowserContext;
    page!: Page;
    url: string;
    userlogin: string;
    userid: string;
    password: string;
    ver: string;
    
    constructor (browser: Browser, url: string, userlogin: string, userid: string, password: string) {
        this.browser = browser;
        this.url = url;
        this.userlogin = userlogin;
        this.userid = userid;
        this.password = password;
    };

    // can't run async functions in constructor
    // don't forget to call this method after instantiation
    async createPage() {
        this.context = await this.browser.newContext(
            {
                recordHar: { path: 'har/requests.har', urlFilter: /api.idcservice.net|consumer.intel.com/ },
                // recordVideo: { dir: 'videos/', size: { width: 1920, height: 1080 } }
            }
        );
        this.page = await this.context.newPage();
        if (process.env['SCREEN'] == 'small') {
            this.page.setViewportSize(SMALL_SCREEN_SIZE);
        }
    }
    
    public async login() {
        await this.createPage();
        const res = await this.page.goto(this.url, { waitUntil: 'domcontentloaded' });  //networkidle
        expect(res?.status(), `Expected 200, received ${res?.status()}`).toBe(200);
        
        await this.page.getByRole('button', { name: 'Sign In' }).click();
        await this.page.getByPlaceholder('Email').click();
        await this.page.getByLabel('Email').fill(this.userlogin);
        await this.page.getByLabel('Sign In').click();
        await this.page.getByPlaceholder('Password').click();
        await this.page.getByLabel('Password').fill(this.password);
        await this.page.getByLabel('Sign In').click();
        await this.page.waitForTimeout(1000);
        utils.log('logged-in successfully');
    }

    public async verifyLoggedUser() {
        await expect(this.page.locator('div.navbar-right').getByText(`Sign out (${this.userid})`)).toBeVisible();
        utils.log('logged-in user ID verified: ' + this.userid);
    }

    public async verifyEnrolled() {
        await this.page.getByRole('link', { name: 'Enroll Now' }).click();
        await expect(this.page).toHaveURL('https://intel.com');
        await expect(this.page.getByRole('heading', { name: 'You are already enrolled in' })).toBeVisible();
        utils.log('verified user is already enrolled');
    }

    public async logout() {
        await this.page.locator('div.navbar-right').getByRole('link', { name: 'Sign out' }).click();
        await expect(this.page.locator('div.navbar-right').getByRole('button', { name: 'Sign In' })).toBeVisible();
        utils.log('logged out successfully');
    }
}
