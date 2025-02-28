import TestFilter from "../../../support/testFilter.js";
const homePage = require("../../../pages/IDC2.0/homePage.js");
const invoicesPage = require("../../../pages/IDC2.0/invoicesPage.js");
const currentUsage = require("../../../pages/IDC2.0/currentUsagePage.js");
const cloudCredits = require("../../../pages/IDC2.0/cloudCreditsPage.js");
const instancePage = require("../../../pages/IDC2.0/instancePage.js");
const computePage = require("../../../pages/IDC2.0/computePage.js");
const accountKeyPage = require("../../../pages/IDC2.0/accountKeyPage.js");
const paymentMethods = require("../../../pages/IDC2.0/managePaymentMethodsPage.js");
var publickey = Cypress.env("publicKey");

TestFilter(["IntelAll", "StandardAll", "PremiumAll", "EnterpriseAll"], () => {
  describe("Cloud Credits", () => {
    beforeEach(() => {
      cy.PrepareSession();
      cy.GetSession();
    });

    afterEach(() => {
      cy.TestClean();
    });

    after(() => {
      cy.TestClean();
    });

    it("1096 | Verify redeem invalid coupon", function () {
      homePage.viewCouponPage();
      cy.wait(2000);
      cloudCredits.redeemCoupon();
      cloudCredits.typeCouponCode("123");
      cloudCredits.clickRedeemButton();
      cy.contains("Invalid coupon code format").should("be.visible");
      cy.wait(2000);
    });

    /*
      it.skip("1097 | Verify already used coupon", function () {
        cy.loginAdmin();
        cy.getCoupon();
        const couponCode = Cypress.env("newCoupon");
        homePage.viewCouponPage();
        cy.wrap(couponCode).then(($code) => {
          cloudCredits.typeCouponCode($code);
          cloudCredits.clickRedeemButton();
        });
        cy.wait(1000);
        cy.get(".btn.btn-secondary.mx-0").contains("Redeem Coupon").click();
        cloudCredits.redeemCoupon();
        cy.wait(1000);
        cy.wrap(couponCode).then(($code) => {
          cloudCredits.typeCouponCode($code);
        });
        cloudCredits.clickRedeemButton();
        cy.wait(1000);
        cy.contains(
          `Error the coupon code (${couponCode}) has already been redeemed`
        ).should("be.visible");
      });
    
      it("1098 | Verify expired coupon", function () {
        homePage.clickuserAccountTab();
        homePage.clickCloudCredits();
        cloudCredits.redeemCoupon();
        const couponCode = Cypress.env("expiredCoupon");
        cloudCredits.typeCouponCode(couponCode);
        cloudCredits.clickRedeemButton();
        cy.contains(
          `Error cannot redeem coupon ${couponCode} as it has expired`
        ).should("be.visible");
      });
  */
    it("1099 | Verify cancel redeem coupon", function () {
      homePage.viewCouponPage();
      cloudCredits.redeemCoupon();
      cloudCredits.typeCouponCode("12-CS");
      cloudCredits.clickCancelCouponCodeButton();
      cy.get('[intc-id="cloudCreditsTitle"]').should("be.visible");
    });

    it("1100 | Verify filter credits in cloud history", function () {
      homePage.clickuserAccountTab();
      homePage.clickCloudCredits();
      cloudCredits.filterCredits("10");
    });

    it("1101 | Verify obtained on sort button", function () {
      homePage.clickuserAccountTab();
      homePage.clickCloudCredits();
      cloudCredits.obtainedOnsort();
    });

    it("1102 | Verify expiration date sort button", function () {
      homePage.clickuserAccountTab();
      homePage.clickCloudCredits();
      cloudCredits.expirationDatesort();
    });

    it("1103 | Verify total credit amount sort button", function () {
      homePage.clickuserAccountTab();
      homePage.clickCloudCredits();
      cloudCredits.totalCreditAmountsort();
    });

    it("1104 | Verify amount used sort button", function () {
      homePage.clickuserAccountTab();
      homePage.clickCloudCredits();
      cloudCredits.amountUsedsort();
    });

    it("1105 | Verify amount remaining sort button", function () {
      homePage.clickuserAccountTab();
      homePage.clickCloudCredits();
      cloudCredits.amountRemainingsort();
    });

    it("1107 | Sign out account", function () {
      cy.signout();
    });
  });
});

TestFilter(["Premium"], () => {
  describe("Cloud Credits - Specific to user", function () {
    beforeEach(() => {
      cy.PrepareSession();
      cy.GetSession();
    });

    afterEach(() => {
      cy.TestClean();
    });

    after(() => {
      cy.TestClean();
    });

    it("1109 | Verify cancel button for redeem coupon - Premium user", function () {
      homePage.clickuserAccountTab();
      homePage.clickPaymentMethods();
      paymentMethods.redeemCoupon();
      cloudCredits.clickCancelCouponCodeButton();
      cy.get('[intc-id="cloudCreditsTitle"]').should("be.visible");
    });
  });
});

TestFilter(["enterprisePending"], () => {
  describe("Validate products - No Credits", function () {
    beforeEach(() => {
      cy.loginWithSessionIntel2("IntelSecond");
      cy.visit(Cypress.env("baseUrl"));
      cy.wait(4000);
    });

    afterEach(() => {
      cy.TestClean();
    });

    after(() => {
      cy.TestClean();
    });

    it("1159 | Verify launching small VM with no credits ", function () {
      // launch paid instance
      homePage.hardwareCatalog();
      homePage.clickfourthGenIntelProd();
      cy.wait(1000);
      instancePage.instanceType("small");
      cy.wait(1000);
      instancePage.instanceName("test-small");
      cy.wait(1000);
      instancePage.clickCreateKeyButton();
      accountKeyPage.addKeyName("test-nocreds");
      accountKeyPage.addKeyContent(publickey);
      instancePage.uploadKey();
      cy.wait(5000);
      computePage.launchInstance();
      cy.get(".modal-dialog.modal-md.modal-dialog-centered").should("be.visible")
      cy.contains("Insufficient cloud credits");
      cy.wait(3000);
      cy.get('[class="text-decoration-none btn btn-link"]')
        .contains("Cancel")
        .click();
    });

    it("1160 | Verify launching medium VM with no credits ", function () {
      // launch paid instance
      homePage.hardwareCatalog();
      homePage.clickfourthGenIntelProd();
      cy.wait(1000);
      instancePage.instanceType("medium");
      instancePage.instanceName("test-medium");
      computePage.launchInstance();
      cy.get(".modal-dialog.modal-md.modal-dialog-centered").should("be.visible")
      cy.contains("Insufficient cloud credits");
      cy.wait(3000);
      cy.get('[class="text-decoration-none btn btn-link"]')
        .contains("Cancel")
        .click();
    });

    it("1061 | Delete all account keys used", function () {
      homePage.clickKeysPageButton();
      cy.deleteAllAccountKeys();
    });
  });
});

TestFilter(["IntelAll", "PremiumAll", "StandardAll"], () => {
  describe("Current month usage", () => {
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

    it("Session creation for usage", function () {
      cy.clearAllCookies();
      cy.clearAllLocalStorage();
      cy.clearAllSessionStorage();
    });

    it("1109 | Filter by Compute as a Service", function () {
      homePage.clickuserAccountTab();
      homePage.clickCurrentUsage();
      currentUsage.typeFilterproducts("Compute as a Service: Bare Metal and Virtual Machine");
      cy.get('[intc-id="fillTable"]').contains("Storage as a Service").should("not.exist");
    });

    it("1110 | Filter by Storage as a Service", function () {
      homePage.clickuserAccountTab();
      homePage.clickCurrentUsage();
      currentUsage.typeFilterproducts("Storage as a Service");
      cy.get('[intc-id="fillTable"]').contains("Compute as a Service: Bare Metal and Virtual Machine").should("not.exist");
    });

    it("1111 | Verify clear filter button", function () {
      homePage.clickuserAccountTab();
      homePage.clickCurrentUsage();
      currentUsage.typeFilterproducts("-1");
      currentUsage.clearFilters();
    });
  });
});


TestFilter(["ePending"], () => {
  describe("Current month usage", () => {
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

    it.skip("1113 | Validate total minutes with estimated usage time", function () {
      homePage.clickuserAccountTab();
      homePage.clickCurrentUsage();
      let totalMinutes = 0;
      let sum = 0;
      cy.get('[intc-id="totalUsageLabel"]').invoke('text').then((totalMin) => {
        // Check if text contains days
        const daysMatch = totalMin.match(/(\d+) days?/);
        if (daysMatch) {
          totalMinutes += parseInt(daysMatch[1], 10) * 24 * 60;
        }
        // Check if text contains hours
        const hoursMatch = totalMin.match(/(\d+) hours?/);
        if (hoursMatch) {
          totalMinutes += parseInt(hoursMatch[1], 10) * 60;
        }
        // Check if text contains minutes
        const minutesMatch = totalMin.match(/(\d+) minutes?/);
        if (minutesMatch) {
          totalMinutes += parseInt(minutesMatch[1], 10);
        }
      })
      // Add the minutes for each product
      cy.get('tbody').within(() => {
        cy.get('tr').each(($row) => {
          cy.wrap($row).within(() => {
            cy.get('td').eq(4).invoke('text').then((Minutes) => {
              sum += Math.round(Minutes)
            })
          })
        })
      }).then(() => {
        expect(sum).to.be.equal(totalMinutes)
      })
    });

    it.skip("1114 | Validate usage for VM small instance", function () {

      homePage.hardwareCatalog()
      // create instance
      homePage.clickfourthGenIntelProd();
      instancePage.instanceType("small");
      instancePage.instanceName("test-minutes2");
      instancePage.clickCreateKeyButton();
      accountKeyPage.addKeyName("test-minutes2");
      accountKeyPage.addKeyContent(publickey);
      instancePage.uploadKey();
      cy.wait(2000);
      computePage.launchInstance();
      cy.wait(40000);

      // delete instance
      computePage.searchInstance('test-minutes2')
      computePage.deleteInstance()
      computePage.confirmDeleteInstance()

      // delete used public key
      homePage.clickKeysPageButton();
      cy.deleteAllAccountKeys();

      // check usage

      homePage.clickuserAccountTab();
      homePage.clickCurrentUsage();
      cy.wait(6000);
      cy.get('tbody tr').then($rows => {
        cy.wrap($rows).find('td').eq(1).should('have.text', 'vm-spr-sml').parent("tr").then(($row) => {
          cy.wrap($row).then(() => {
            cy.get($row).find(":nth-child(5) > .mt-1 > span").invoke("text").as("value1");
            cy.get($row).find(":nth-child(6) > .mt-1 > span").invoke("text").as("value2");
            cy.get($row).find(':nth-child(7) > .mt-1 > span').invoke("text").as("value3");
            cy.wait(2000);
            cy.get("@value1").then((value1) => {
              cy.get("@value2").then((value2) => {
                cy.get("@value3").then((value3) => {
                  const cleanValue1 = Number(value1.replace("$", ""));
                  const cleanValue2 = Number(value2.replace("$", ""));
                  const estimateAmount = Number(value3.replace("$", ""));
                  const expectedResult = (cleanValue1 / 60) * cleanValue2;
                  const tolerancelow = 0.4
                  const tolerancehigh = 0.2
                  expect(expectedResult).to.be.within(estimateAmount - tolerancelow, estimateAmount + tolerancehigh)
                })
              })
            })
          })
        })
      })
    });
  });
});

TestFilter(["PremiumAll"], () => {
  describe("Invoices", function () {
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

    it("1120 | Search product invoices", function () {
      homePage.clickuserAccountTab();
      homePage.clickInvoices();
      invoicesPage.filterProducts("June");
    });

    it("1121 | Filter by start date and end date in invoices", function () {
      homePage.clickuserAccountTab();
      homePage.clickInvoices();
      invoicesPage.startDate("2023-05-01");
      invoicesPage.endDate("2023-06-30");
      cy.get('[intc-id="StartDateInput"]')
        .invoke("val")
        .then((startdate) => {
          cy.get('[intc-id="EndDateInput"]')
            .invoke("val")
            .then((enddate) => {
              expect(Date.parse(enddate)).to.be.greaterThan(
                Date.parse(startdate)
              );
            });
        });
    });

    it("1122 | Verify current month usage link", function () {
      homePage.clickuserAccountTab();
      homePage.clickInvoices();
      invoicesPage.ClickcurrentMonthUsageLink();
    });

    it("1123 | Verify invoices id sort button", function () {
      homePage.clickuserAccountTab();
      homePage.clickInvoices();
      invoicesPage.invoicesIdSort();
    });

    it("1124 | Verify billing period sort button", function () {
      homePage.clickuserAccountTab();
      homePage.clickInvoices();
      invoicesPage.billingPeriodSort();
    });

    it("1125 | Verify period start sort button", function () {
      homePage.clickuserAccountTab();
      homePage.clickInvoices();
      invoicesPage.periodStartSort();
    });

    it("1126 | Verify period end sort button", function () {
      homePage.clickuserAccountTab();
      homePage.clickInvoices();
      invoicesPage.periodEndSort();
    });

    it("1127 | Verify total amount sort button", function () {
      homePage.clickuserAccountTab();
      homePage.clickInvoices();
      invoicesPage.totalAmountSort();
    });

    it("1128 | Verify amount paid sort button", function () {
      homePage.clickuserAccountTab();
      homePage.clickInvoices();
      invoicesPage.amountpaidSort();
    });

    it("1129 | Verify amount due sort button", function () {
      homePage.clickuserAccountTab();
      homePage.clickInvoices();
      invoicesPage.amountDueSort();
    });

    it("1130 | Verify clear filter button in invoices", function () {
      homePage.clickuserAccountTab();
      homePage.clickInvoices();
      invoicesPage.filterProducts("April");
      invoicesPage.clearFilter();
    });
  });
});
