import { Page, Locator } from "@playwright/test";

export function log(msg: string) {
    //msg = this.constructor.name + ': ' + msg;
    console.log(new Date().toISOString() + ' ' + msg);
}

// By default will generate exception
export function error(msg: string, skipException = false) {
    //msg = this.constructor.name + ': ' + msg;
    console.error(new Date().toISOString() + ' ' + msg);
    if (!skipException)
        throw new Error(msg);
}

export function sleep (ms: number) {
    return new Promise((r) => setTimeout(r, ms));
}

// if not exists - throw exception (end test) or return "false" (further handling)
export async function waitForObject(locator: Locator, timeout = 5000, errMsg = '', skipException = false){ 
    for (let v=0,t=0; t < timeout; t+=300){
        try {
            v = await locator.count();
        } catch (e) {
            v = 0;
        }
        if (v)
            return true;
        await new Promise(r => setTimeout(r, 300));
    }
    
    if (!skipException)
        throw new Error(`Timed out waiting for object (${errMsg})`);
    else 
        return false;
}

export async function waitForObjectDisappearance(locator: Locator, timeout = 5000, errMsg = '', skipException = false){ 
    for (let v=0,t=0; t < timeout; t+=300){
        try {
            v = await locator.count();
        } catch (e) {
            v = 0;
            return true;
        }
        if (v == 0)
            return true;
        await sleep(300);
    }
    
    if (!skipException)
        throw new Error(`Timed out waiting for object disappearance (${errMsg})`);
    else 
        return false;
}

// workaround for deleting compute instance in us-region-2
// also waits for doc to be downloaded
export async function waitForObjectDisappearance1(pg: Page, locator: Locator, timeout = 5000, errMsg = '', skipException = false){ 
    for (let v=0,t=0; t < timeout; t+=1000){
        try {
            await pg.waitForLoadState('domcontentloaded');
            await pg.waitForLoadState('networkidle');
            v = await locator.count();
        } catch (e) {
            v = 0;
            return true;
        }
        if (v == 0)
            return true;
        await sleep(1000);
    }
    
    if (!skipException)
        throw new Error(`Timed out waiting for object disappearance (${errMsg})`);
    else 
        return false;
}
