import TestFilter from "../../support/testFilter";
const computePage = require("../../pages/IDC2.0/computePage.js");
const homePage = require("../../pages/IDC2.0/homePage.js");
const instancePage = require("../../pages/IDC2.0/instancePage.js");
var publickey = Cypress.env("publicKey");
import accountKeyPage from "../../pages/IDC2.0/accountKeyPage";

TestFilter(["IntelAll", "PremiumAll", "StandardAll"], () => {
  describe("Reserve Virtual Machine verification", () => {
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

    it("101 | Reserve VM", function () {
      homePage.hardwareCatalog();
      homePage.clickfourthGenIntelProd();
      instancePage.instanceType("small");
      instancePage.instanceName("test6");
      cy.wait(1000);
      instancePage.clickCreateKeyButton();
      accountKeyPage.clickkeyName("test-key");
      accountKeyPage.clickKeyContents(publickey);
      instancePage.uploadKey()
      instancePage.checkFirstKey()
      cy.wait(2000);
      computePage.launchInstance();
      cy.wait(40000);
      computePage.computeConsolePage();
      cy.wait(5000);
      cy.get(".mt-1").contains("Ready").as('Ready')
      cy.get("@Ready").should("be.visible")
    });

    it("102 | Verify hardware catalog page", function () {
      homePage.hardwareCatalog();
    });

    it("103 | Verify software catalog page", function () {
      homePage.softwareCatalog();
    });

    it("104 | Verify training and workshops page", function () {
      homePage.trainingAndWorkshops();
    });

    it("105 | Current month usage page", function () {
      homePage.clickuserAccountTab();
      homePage.clickCurrentUsage();
      cy.get(".m-3").should("have.text", "Current month usage");
      cy.wait(2000);
    });

    it("106 | Filter Virtual Machine Type Products", function () {
      homePage.hardwareCatalog();
      homePage.clickvirtualMachineType();
      cy.wait(2000);
      cy.get(".catalog-product-title").should("have.length", 1);
    });

    it("107 | Verify instance details tab", function () {
      homePage.clickcomputePageButton();
      computePage.searchInstance("test1");
      cy.get('[intc-id="test1HyperLinkTable"]').contains("test1").click();
      cy.get('[class="fw-bold small mb-auto"]')
        .contains("Instance Type:")
        .scrollIntoView({ duration: 4000 });
      cy.contains("4th Generation Intel® Xeon® Scalable processor").should(
        "be.visible");
      cy.get('[class="small reserve-detail"]').contains("Ready")
    });

    it("108 | Delete an instance using actions", function () {
      homePage.clickcomputePageButton();
      cy.get('[intc-id="test1HyperLinkTable"]').contains("test1").click();
      computePage.clickactionsButton();
      computePage.clickdeleteActionsButton();
      computePage.confirmDeleteInstance();
      cy.wait(5000);
      homePage.clickhomePageButton();
      homePage.clickaccountKeyPageButton();
      cy.deleteAllAccountKeys();
    });
  });
});
