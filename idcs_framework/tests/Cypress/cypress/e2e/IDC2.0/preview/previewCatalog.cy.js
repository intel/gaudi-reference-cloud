import TestFilter from "../../../support/testFilter";
import homePage from "../../../pages/IDC2.0/homePage";

TestFilter(["ependingAll"], () => {
  describe("0 - Preview Catalog verification", () => {
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

    it("2000 | Search for Flex products using filter", function () {
      homePage.clickpreviewHardwareCatalog();
      homePage.searchBoxIsVisible();
      cy.wait(2000);
      cy.searchFilter("Flex");
      cy.get(".bg-surface-gradient product-card.card").should("have.length", 1);
    });

    it("2001 | Search for Non-existing preview product", function () {
      homePage.clickpreviewHardwareCatalog();
      homePage.searchBoxIsVisible();
      cy.wait(2000);
      cy.searchFilter("try");
      cy.get(".bg-surface-gradient product-card.card").should('not.exist');
    });

    it("2002 | Filter Products by AI Processor", function () {
      homePage.clickpreviewHardwareCatalog();
      homePage.searchBoxIsVisible();
      cy.wait(2000);
      homePage.clickAIprocessor();
      cy.get(".bg-surface-gradient product-card.card").should("have.length", 1);
    });

    it("2003 | Filter Products by CPU Processor", function () {
      homePage.clickpreviewHardwareCatalog();
      homePage.searchBoxIsVisible();
      cy.wait(2000);
      homePage.clickcpuProcessor();
      cy.get(".bg-surface-gradient product-card.card").should("have.length", 9);
    });

    it("2004 | Filter Products by GPU Processor", function () {
      homePage.clickpreviewHardwareCatalog();
      homePage.searchBoxIsVisible();
      cy.wait(2000);
      homePage.clickgpuProcessor();
      cy.get(".bg-surface-gradient product-card.card").should("have.length", 2);
    });

    it("2005 | Filter by Sierra Forest products using Search", function () {
      homePage.clickpreviewHardwareCatalog();
      homePage.searchBoxIsVisible();
      cy.wait(2000);
      cy.searchFilter("Sierra");
      cy.get(".bg-surface-gradient product-card.card").should("have.length", 3);
    });

    it("2006 | Search for GPU based preview products", function () {
      homePage.clickpreviewHardwareCatalog();
      homePage.searchBoxIsVisible();
      cy.wait(2000);
      homePage.searchFilter("GPU");
      cy.get(".bg-surface-gradient product-card.card").should("have.length", 2);
    });

    it("2007 | Search for AI PC preview products", function () {
      homePage.clickpreviewHardwareCatalog();
      homePage.searchBoxIsVisible();
      cy.wait(2000);
      homePage.clickAI_PC();
      cy.get(".bg-surface-gradient product-card.card").should("have.length", 2);
    });
  });
});
