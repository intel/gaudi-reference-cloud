import TestFilter from "../../../support/testFilter";
import homePage from "../../../pages/IDC2.0/homePage";

TestFilter(["IntelAll", "PremiumAll", "StandardAll", "ependingAll"], () => {
  describe("Home and Product Catalog verification", () => {
    beforeEach(() => {
      cy.PrepareSession();
      cy.GetSession();
      cy.viewport(1280, 720);
    });

    afterEach(() => {
      cy.TestClean();
    });

    after(() => {
      cy.TestClean();
    });

    it("1073 | Search for 4th Gen Intel family using filter", function () {
      homePage.hardwareCatalog();
      cy.wait(2000);
      homePage.checkHardwarePage();
      cy.searchFilter("4th Generation Intel");
      cy.get(".card-body").should("have.length", 2);
    });

    it("1074 | Collapsed Side Menu", function () {
      cy.get('[intc-id="sidebarnavLink/home"]').should("be.visible");
      homePage.collapasedSideMenu();
      cy.wait(1000);
      homePage.expandSideMenu();
    });
    /*
        it.skip("1075 | Filter Products by CPU", function () {
          homePage.hardwareCatalog();
          cy.wait(2000);
          homePage.checkHardwarePage();
          homePage.clickcpuProcessor();
          cy.get(".card-body").should("have.length", 3);
        });
    */
    it("1076 | Validate Documentation URL", function () {
      homePage.clickhelpButton();
      cy.get('[intc-id="help-menu-Browse-documentation"]')
        .contains("Browse documentation")
        .then((link) => {
          cy.request(link.prop("href")).its("status").should("eq", 200);
        });
      cy.wait(3000);
      homePage.clickDocumentationButton();
    });

    it("1077 | Validate KnowledgeBase URL", function () {
      homePage.clickhelpButton();
      cy.get('[intc-id="help-menu-Knowledge-base"]').should("be.visible");
      cy.get('[intc-id="help-menu-Knowledge-base"]').should(
        "have.attr",
        "href",
        "https://www.intel.com/content/www/us/en/support/products/236984/services/intel-developer-cloud/intel-developer-cloud-hardware-services.html"
      );
    });

    it("1078 | Validate Community URL", function () {
      homePage.clickhelpButton();
      cy.get('[intc-id="help-menu-Community"]').should(
        "have.attr",
        "href",
        "https://community.intel.com/t5/Intel-Developer-Cloud/bd-p/developer-cloud"
      );
    });

    it("1079 | Validate Send feedback URL", function () {
      homePage.clickhelpButton();
      cy.get('[intc-id="help-menu-Send-feedback"]').should(
        "have.attr",
        "href",
        "https://intel.az1.qualtrics.com/jfe/form/SV_8cEjBMShr8n3FgW"
      );
    });

    it("1080 | Current month usage page", function () {
      homePage.clickuserAccountTab();
      homePage.clickCurrentUsage();
      cy.contains("Current month usage");
    });

    it("1081 | Verify Account Settings profile", function () {
      homePage.clickuserAccountTab();
      homePage.clickProfileSettings();
      cy.contains("Your AI Cloud Account");
    });
  });
});

TestFilter(["Premium", "Intel", "Enterprise"], () => {
  describe("Home and Product Catalog verification - Specific to user", () => {
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

    // it("1093 | Validate Submit a ticket URL", function () {
    //   homePage.clickhelpButton();
    //   cy.get('[intc-id="help-menu-Submit-a-ticket"]').should(
    //     "have.attr",
    //     "href",
    //     "https://supporttickets.intel.com/supportrequest?lang=en-US&productId=236985:15897"
    //   );
    // });

    it("1094 | Validate Contact Support URL", function () {
      homePage.clickhelpButton();
      cy.get('[intc-id="help-menu-Contact-support"]').should(
        "have.attr",
        "href",
        "https://www.intel.com/content/www/us/en/support/contact-intel.html#support-intel-products_67709:59441:2314824"
      );
    });
  });
});

TestFilter(["Premium"], () => {
  describe("Home and Product Catalog verification - Specific to user", function () {
    beforeEach(() => {
      cy.PrepareSession();
      cy.GetSession();
      cy.viewport(1280, 720);
    });

    afterEach(() => {
      cy.TestClean();
    });

    after(() => {
      cy.TestClean();
    });

    it("1091 | Validate Submit a ticket URL", function () {
      homePage.clickhelpButton();
      cy.get('[intc-id="help-menu-Submit-a-ticket"]').should(
        "have.attr",
        "href",
        "https://supporttickets.intel.com/supportrequest?lang=en-US&productId=236985:15897"
      );
    });

    it("1092 | Invoices page", function () {
      homePage.clickuserAccountTab();
      homePage.clickInvoices();
    });

    it("1093 | Ensure user account is Premium", function () {
      homePage.clickuserAccountTab();
      cy.get(
        '[class="badge header-badge header-badge-bg-premium text-capitalize"]'
      )
        .contains("Premium")
        .should("be.visible");
    });

    it("1094 | Verify Payment methods is available for Premium account", function () {
      homePage.clickuserAccountTab();
      cy.get('[intc-id="paymentMethodsHeaderButton"]').should("exist");
    });

    it("1095 | Verify Chat with an agent is available for Premium", function () {
      homePage.hardwareCatalog();
      homePage.clickhelpButton();
      cy.get('[intc-id="help-menu-Chat-with-an-agent"]').should("exist");
    });

  });
});

TestFilter(["Intel"], () => {
  describe("Home and Product Catalog verification - Specific to user", function () {
    beforeEach(() => {
      cy.PrepareSession();
      cy.GetSession();
      cy.viewport(1280, 720);
    });

    afterEach(() => {
      cy.TestClean();
    });

    after(() => {
      cy.TestClean();
    });

    it("1095 | Ensure user account is Intel", function () {
      homePage.clickuserAccountTab();
      cy.get(
        '[class="badge header-badge header-badge-bg-intel text-capitalize"]'
      )
        .contains("Intel")
        .should("be.visible");
    });

    it("1096 | Validate Submit a ticket URL", function () {
      homePage.clickhelpButton();
      cy.get('[intc-id="help-menu-Submit-a-ticket"]').should(
        "have.attr",
        "href",
        "https://supporttickets.intel.com/supportrequest?lang=en-US&productId=236756:15738"
      );
    });
    /*
        it("1096 | Filter Products by core compute and GPU", function () {
          homePage.hardwareCatalog();
          cy.wait(2000);
          homePage.checkHardwarePage();
          homePage.clickCoreCompute();
          cy.wait(1000);
          homePage.clickgpuProcessor();
          cy.wait(1000);
          cy.get(".card-body").should("have.length", 4);
        });
    */
    it("1097 | Search for GPU based products using filter", function () {
      homePage.hardwareCatalog();
      homePage.checkHardwarePage();
      cy.searchFilter("GPU");
      cy.get(".card-body").should("have.length", 2);
    });

    it("1098 | Cloud credits page", function () {
      homePage.clickuserAccountTab();
      homePage.clickCloudCredits();
      cy.get('[intc-id="cloudCreditsTitle"]').should("be.visible");
    });

    it("1099 | Verify Chat with an agent is available for Intel Account", function () {
      homePage.hardwareCatalog();
      homePage.clickhelpButton();
      cy.get('[intc-id="help-menu-Chat-with-an-agent"]').should("exist");
    });
  });
});

TestFilter(["enpending"], () => {
  describe("Home and Product Catalog verification - Specific to user", function () {
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

    it("1088 | Ensure user account is Enteprise Pending", function () {
      homePage.clickuserAccountTab();
      cy.get(
        '[class="badge header-badge header-badge-bg-enterprise_pending text-capitalize"]'
      )
        .contains("Pending confirmation")
        .should("be.visible");
    });
  });
});

TestFilter(["Enterprise"], () => {
  describe("Home and Product Catalog verification - Specific to user", function () {
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

    it("1095 | Verify Chat with an agent is available for Enterprise Account", function () {
      homePage.clickhelpButton();
      cy.get('[intc-id="help-menu-Chat-with-an-agent"]').should("exist");
    });

    it("1096 | Validate Submit a ticket URL", function () {
      homePage.clickhelpButton();
      cy.get('[intc-id="help-menu-Submit-a-ticket"]').should(
        "have.attr",
        "href",
        "https://supporttickets.intel.com/supportrequest?lang=en-US&productId=236984:15896"
      );
    });
  });
});

TestFilter(["Standard"], () => {
  describe("Home and Product Catalog verification - Specific to user", function () {
    beforeEach(() => {
      cy.PrepareSession();
      cy.GetSession();
      cy.viewport(1280, 720);
    });

    afterEach(() => {
      cy.TestClean();
    });

    after(() => {
      cy.TestClean();
    });

    it("1081 | Filter Products by core compute and GPU", function () {
      homePage.hardwareCatalog();
      homePage.checkHardwarePage();
      homePage.clickCoreCompute();
      cy.wait(1000);
      homePage.clickgpuProcessor();
      cy.wait(1000);
      cy.get(".card-body").should("have.length", 3);
    });

    it("1082 | Search for GPU based products using filter", function () {
      homePage.hardwareCatalog();
      homePage.checkHardwarePage();
      cy.searchFilter("GPU");
      cy.get(".card-body").should("have.length", 1);
    });

    it("1083 | Verify Gaudi2 product is not available for standard user", function () {
      homePage.hardwareCatalog();
      homePage.checkHardwarePage();
      cy.get(
        '[intc-id="btn-select Habana* Gaudi2 Deep Learning Server"]'
      ).should("not.exist");
    });

    it("1084 | Verify Intel® Xeon® processors - Sapphire Rapids (HBM) not available for standard user", function () {
      homePage.hardwareCatalog();
      homePage.checkHardwarePage();
      cy.get(
        '[intc-id="btn-select Intel® Xeon® processors, codenamed Sapphire Rapids with high bandwidth memory (HBM)"]'
      ).should("not.exist");
    });

    it("1085 | Verify Intel® Data Center GPU product not available for standard user", function () {
      homePage.hardwareCatalog();
      homePage.checkHardwarePage();
      cy.get(
        '[intc-id="btn-select Intel® Data Center GPU Flex Series on latest Intel® Xeon® processors"]'
      ).should("not.exist");
    });

    it("1090 | Verify Submit a ticket is not available", function () {
      homePage.clickhelpButton();
      cy.get('[intc-id="help-menu-Submit-a-ticket"]').should("not.exist");
    });

    it("1091 | Verify Contact support is not available", function () {
      homePage.clickhelpButton();
      cy.get('[intc-id="help-menu-Contact-support"]').should("not.exist");
    });

    it("1092 | Verify Chat with an agent is not available for Standard account", function () {
      homePage.clickhelpButton();
      cy.get('[intc-id="help-menu-Chat-with-an-agent"]').should("not.exist");
    });

    it("1093 | Verify invoices is not available for Standard account", function () {
      homePage.clickuserAccountTab();
      cy.get('[intc-id="invoicesHeaderButton"]').should("not.exist");
    });

    it("1094 | Verify Payment methods is not available for Standard account", function () {
      homePage.clickuserAccountTab();
      cy.get('[intc-id="paymentMethodsHeaderButton"]').should("not.exist");
    });

    it("1095 | Ensure user account is Standard", function () {
      homePage.clickuserAccountTab();
      cy.get(
        '[class="badge header-badge header-badge-bg-standard text-capitalize"]'
      )
        .contains("Standard")
        .should("be.visible");
    });
  });
});
