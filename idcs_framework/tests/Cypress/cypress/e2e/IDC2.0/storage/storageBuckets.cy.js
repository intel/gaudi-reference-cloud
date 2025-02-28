const homePage = require("../../../pages/IDC2.0/homePage.js");
const storageBucketsPage = require("../../../pages/IDC2.0/storageBucketsPage.js");
import storagePrincipalsPage from "../../../pages/IDC2.0/storagePrincipalsPage.js";
const storageLifeCyclePage = require("../../../pages/IDC2.0/storageLifeCyclePage.js");
import TestFilter from "../../../support/testFilter.js";
var regionNum = Cypress.env("region");

TestFilter(["IntelAll"], () => {
  describe("2 Storage - Verify Buckets", () => {
    beforeEach(() => {
      cy.PrepareSession();
      cy.GetSession();
      cy.selectRegion(regionNum);
    });

    afterEach(() => {
      cy.TestClean();
    });

    after(() => {
      cy.TestClean();
    });

    it("5000 | Create Storage Bucket", function () {
      homePage.clickBucketStorageButton();
      // Create first bucket - enabled versioning
      storageBucketsPage.createFirstBucket();
      storageBucketsPage.enterBucketName("firstbucket");
      storageBucketsPage.enterBucketDescription("first bucket descrip");
      storageBucketsPage.enableVersioning();
      if (Cypress.env('accountType') === "Intel") {
        cy.get('[intc-id="Object-Storage-Cost-InputLabel"]').should("contain", "$0.03");
      } else {
        cy.get('[intc-id="Object-Storage-Cost-InputLabel"]').should("contain", "$0.10");
      }
      storageBucketsPage.submitBucket();
      cy.wait(4000);
      storageBucketsPage.checkBucketTableVisible();

      cy.contains("firstbucket").should("be.visible");

      // Create second bucket - no versioning
      storageBucketsPage.createBucket();
      storageBucketsPage.enterBucketName("secondbucket");
      storageBucketsPage.enterBucketDescription("test description");
      storageBucketsPage.submitBucket();
      cy.wait(4000);
      storageBucketsPage.checkBucketTableVisible();
      cy.contains("secondbucket").should("be.visible");
    });

    it("5001 | Create Storage Bucket with invalid name", function () {
      homePage.clickBucketStorageButton();
      storageBucketsPage.createBucket();
      storageBucketsPage.enterBucketName("@@!!");
      storageBucketsPage.checkInvalidBucketNameMessage();
    });

    it("5002| Create Storage Bucket with duplicate name", function () {
      homePage.clickBucketStorageButton();
      storageBucketsPage.createBucket();
      storageBucketsPage.enterBucketName("firstbucket");
      storageBucketsPage.enterBucketDescription("first bucket descrip");
      storageBucketsPage.enableVersioning();
      storageBucketsPage.submitBucket();
      cy.wait(6000);
      cy.get(".modal-header")
        .contains("Could not create your bucket")
        .within(($form) => {
          cy.wrap($form).should("be.visible");
          cy.wait(2000);
        });
      cy.get(".btn.btn-primary").contains("Go back").click({ force: true });
    });

    it("5003 | Create storage Bucket with long name", function () {
      homePage.clickBucketStorageButton();
      storageBucketsPage.createBucket();
      storageBucketsPage.enterBucketName("longstoragebucketnamepreventsnomorethan50character1234");
      storageBucketsPage.getBucketNameInput().should("have.value", "longstoragebucketnamepreventsnomorethan50character");
    });

    it("5004 | Verify bucket's details", function () {
      homePage.clickBucketStorageButton();
      storageBucketsPage.checkBucketTableVisible();
      storageBucketsPage.findBucket("firstbucket");
      cy.wait(4000);
      cy.get(".fw-semibold").should("be.visible").contains("Versioning:")
      cy.contains("first bucket descrip").should("be.visible");
    })

    it("5005 | Search bucket successfully", function () {
      homePage.clickBucketStorageButton();
      storageBucketsPage.checkBucketTableVisible();
      storageBucketsPage.searchBucket("firstbucket");
      cy.wait(2000);
      cy.contains("firstbucket").should("be.visible");
    });

    it("5006 | Search a non-existing bucket", function () {
      homePage.clickBucketStorageButton();
      storageBucketsPage.checkBucketTableVisible();
      storageBucketsPage.searchBucket("fakeBucket");
      cy.contains("The applied filter criteria did not match any items.").should("be.visible");
      storageBucketsPage.clearFilter();
    });

    it("5007 | Verify Cancel Create Bucket", function () {
      homePage.clickBucketStorageButton();
      storageBucketsPage.createBucket();
      storageBucketsPage.enterBucketName("mythirdbucket");
      storageBucketsPage.enterBucketDescription("third bucket descrip");
      storageBucketsPage.enableVersioning();
      storageBucketsPage.cancelBucket();
    });

    it("5008 | Delete a bucket", function () {
      homePage.clickBucketStorageButton();
      // The second bucket will be deleted
      storageBucketsPage.checkBucketTableVisible();
      storageBucketsPage.searchBucket("firstbucket");
      cy.contains("firstbucket").should("be.visible");
      storageBucketsPage.deleteBucket();
      storageBucketsPage.checkConfirmDeleteModal();
      cy.getValue('[intc-id="deleteConfirmationName"]').then(text => {
        cy.get('[intc-id="NameInput"]').should("be.visible").clear().type(text);
      })
      storageBucketsPage.confirmDeleteBucket();
      storageBucketsPage.deleteBucketPrincipalConfirm();
      cy.wait(3000);
    });

    it("5009 | Delete a bucket through Actions", function () {
      homePage.clickBucketStorageButton();
      storageBucketsPage.checkBucketTableVisible();
      storageBucketsPage.findBucket("secondbucket");
      storageBucketsPage.clickActionsButton();
      storageBucketsPage.deleteFromActions();
      storageBucketsPage.checkConfirmDeleteModal();
      cy.getValue('[intc-id="deleteConfirmationName"]').then(text => {
        cy.get('[intc-id="NameInput"]').should("be.visible").clear().type(text);
      })
      storageBucketsPage.confirmDeleteBucket();
      storageBucketsPage.deleteBucketPrincipalConfirm();
      cy.wait(5000);
      storageBucketsPage.checkEmptyBucketTable();
    });
  });
})

TestFilter(["IntelAll", "PremiumAll", "EnterpriseAll"], () => {
  describe("3 Storage - Verify LifeCycle Rules", () => {
    beforeEach(() => {
      cy.PrepareSession();
      cy.GetSession();
      cy.selectRegion(regionNum);
    });

    afterEach(() => {
      cy.TestClean();
    });

    after(() => {
      cy.TestClean();
    });

    it("7001 | Create LifeCycle Rule", function () {
      homePage.clickBucketStorageButton();
      // Create first bucket - enabled versioning
      storageBucketsPage.createFirstBucket();
      storageBucketsPage.enterBucketName("rulebucket");
      storageBucketsPage.enableVersioning();
      if (Cypress.env('accountType') === "Intel") {
        cy.get('[intc-id="Object-Storage-Cost-InputLabel"]').should("contain", "$0.03");
      } else {
        cy.get('[intc-id="Object-Storage-Cost-InputLabel"]').should("contain", "$0.10");
      }
      storageBucketsPage.submitBucket();
      cy.wait(2000);
      storageBucketsPage.findBucket("rulebucket");
      storageLifeCyclePage.rulesTab();
      // Create LifeCycle Rule
      storageLifeCyclePage.createRule();
      storageLifeCyclePage.ruleNameInput("rule01");
      storageLifeCyclePage.inputPrefix("prefixtest");
      storageLifeCyclePage.addRule();
      cy.wait(2000);
      storageLifeCyclePage.checkRulesTableVisible();
      cy.contains("rule01").should("be.visible");
    });

    it("7002 | Create  Rule with an invalid name", function () {
      homePage.clickBucketStorageButton();
      storageBucketsPage.checkBucketTableVisible();
      cy.contains("rulebucket").should("be.visible");
      storageBucketsPage.findBucket("rulebucket");
      storageLifeCyclePage.rulesTab();
      // Create LifeCycle Rule
      storageLifeCyclePage.createRule();
      storageLifeCyclePage.ruleNameInput("rule01");
      storageLifeCyclePage.ruleNameInput("@@!!");
      storageLifeCyclePage.checkInvalidRuleNameMessage();
      storageLifeCyclePage.addRuleIsEnabled();
    });

    it("7003 | Create a Rule with long name", function () {
      homePage.clickBucketStorageButton();
      storageBucketsPage.checkBucketTableVisible();
      cy.contains("rulebucket").should("be.visible");
      storageBucketsPage.findBucket("rulebucket");
      storageLifeCyclePage.rulesTab();
      // Create LifeCycle Rule
      storageLifeCyclePage.createRule();
      storageLifeCyclePage.ruleNameInput("longstoragelifecyclerulenamepreventsnomorethan63character1234567");
      storageLifeCyclePage.getRuleName().should("have.value", "longstoragelifecyclerulenamepreventsnomorethan63character123456");
    });

    it("7004 | Verify LifeCyle Rule Expiry Days Max value", function () {
      homePage.clickBucketStorageButton();
      storageBucketsPage.checkBucketTableVisible();
      storageBucketsPage.findBucket("rulebucket");
      storageLifeCyclePage.rulesTab();
      storageLifeCyclePage.createRule();
      storageLifeCyclePage.ruleNameInput("rule02");
      storageLifeCyclePage.inputPrefix("prefixtest");
      storageLifeCyclePage.expiryDaysCheck();
      storageLifeCyclePage.clearExpiryDays();
      storageLifeCyclePage.expiryDaysInput('333333');
      cy.contains('Value more than 2557 is not allowed.');
      storageLifeCyclePage.addRuleIsEnabled();
    })

    it("7005| Verify LifeCyle Rule Non-Expiry Days Max value", function () {
      homePage.clickBucketStorageButton();
      storageBucketsPage.checkBucketTableVisible();
      storageBucketsPage.findBucket("rulebucket");
      storageLifeCyclePage.rulesTab();
      storageLifeCyclePage.createRule();
      storageLifeCyclePage.ruleNameInput("rule02");
      storageLifeCyclePage.inputPrefix("prefixtest");
      storageLifeCyclePage.expiryDaysCheck();
      storageLifeCyclePage.nonExpiryDays('333333');
      cy.contains('Value more than 2557 is not allowed.');
      storageLifeCyclePage.addRuleIsEnabled();
    })

    it("7006 | Verify Edit LifeCyle Rule Expiry Days", function () {
      homePage.clickBucketStorageButton();
      storageBucketsPage.checkBucketTableVisible();
      storageBucketsPage.findBucket("rulebucket");
      storageLifeCyclePage.rulesTab();
      storageLifeCyclePage.editRule();
      cy.wait(2000);
      storageLifeCyclePage.expiryDaysCheck();
      storageLifeCyclePage.clearExpiryDays();
      storageLifeCyclePage.expiryDaysInput('33');
      storageLifeCyclePage.saveEditRule();
      cy.contains('33');
    })

    it("7007 | Verify Edit LifeCyle Rule Non-Expiry Days", function () {
      homePage.clickBucketStorageButton();
      storageBucketsPage.checkBucketTableVisible();
      storageBucketsPage.findBucket("rulebucket");
      storageLifeCyclePage.rulesTab();
      storageLifeCyclePage.editRule();
      cy.wait(2000);
      storageLifeCyclePage.nonExpiryDays('999');
      storageLifeCyclePage.saveEditRule();
      cy.contains('999');
    })

    it("7008 | Verify Edit LifeCyle Rule Non-Expiry Days - max allowed value", function () {
      homePage.clickBucketStorageButton();
      storageBucketsPage.checkBucketTableVisible();
      storageBucketsPage.findBucket("rulebucket");
      storageLifeCyclePage.rulesTab();
      storageLifeCyclePage.editRule();
      cy.wait(2000);
      storageLifeCyclePage.nonExpiryDays('2899');
      cy.contains('Value more than 2557 is not allowed.');
    })

    it("7009 | Verify Edit LifeCyle Rule Prefix name and set Delete Marker", function () {
      homePage.clickBucketStorageButton();
      storageBucketsPage.checkBucketTableVisible();
      storageBucketsPage.findBucket("rulebucket");
      cy.get('[intc-id="Lifecycle RulesTab"]').as('rulesTab')
      cy.get('@rulesTab').trigger('mouseover').wait(1000).click().click({ force: true });
      storageLifeCyclePage.editRule();
      cy.wait(2000);
      storageLifeCyclePage.deleteMarkerCheck();
      storageLifeCyclePage.inputPrefix('newprefix');
      storageLifeCyclePage.saveEditRule();
      cy.contains('newprefix');
    })

    it("7010 | Create Rule with duplicate name", function () {
      homePage.clickBucketStorageButton();
      storageBucketsPage.checkBucketTableVisible();
      storageBucketsPage.findBucket("rulebucket");
      storageLifeCyclePage.rulesTab();
      storageLifeCyclePage.createRule();
      storageLifeCyclePage.ruleNameInput("rule01");
      storageLifeCyclePage.inputPrefix("prefixtest");
      storageLifeCyclePage.addRule();
      cy.get(".modal-header")
        .contains("Could not create your Lifecycle Rule")
        .within(($form) => {
          cy.wrap($form).should("be.visible");
          cy.wait(2000);
        });
    })

    it("7011 | Delete a LifeCycle Rule", function () {
      homePage.clickBucketStorageButton();
      storageBucketsPage.checkBucketTableVisible();
      storageBucketsPage.findBucket("rulebucket");
      storageLifeCyclePage.rulesTab();
      storageLifeCyclePage.deleteRule();
      storageLifeCyclePage.checkConfirmDeleteModal();
      cy.getValue('[intc-id="deleteConfirmationName"]').then(text => {
        cy.get('[intc-id="NameInput"]').should("be.visible").clear().type(text);
      })
      storageLifeCyclePage.confirmDelete();
    });
  })
})

TestFilter(["IntelAll", "PremiumAll", "EnterpriseAll"], () => {
  describe("4 Storage - Verify User principals", () => {
    beforeEach(() => {
      cy.PrepareSession();
      cy.GetSession();
      cy.selectRegion(regionNum);
    });

    afterEach(() => {
      cy.TestClean();
    });

    after(() => {
      cy.TestClean();
    });

    it("6000 | Setup Storage Bucket", function () {
      homePage.clickBucketStorageButton();
      // Create first bucket - enabled versioning
      storageBucketsPage.createBucket();
      storageBucketsPage.enterBucketName("secondbucket");
      storageBucketsPage.enterBucketDescription("second bucket descrip");
      storageBucketsPage.enableVersioning();
      if (Cypress.env('accountType') === "Intel") {
        cy.get('[intc-id="Object-Storage-Cost-InputLabel"]').should("contain", "$0.03");
      } else {
        cy.get('[intc-id="Object-Storage-Cost-InputLabel"]').should("contain", "$0.10");
      }
      storageBucketsPage.submitBucket();
      cy.wait(4000);
      storageBucketsPage.checkBucketTableVisible();
      cy.contains("secondbucket").should("be.visible");
    });

    it("6001 | Create Principal for all buckets with all actions and all policies", function () {
      homePage.clickBucketStorageButton();
      storageBucketsPage.clickManagePrincipalButton();
      storagePrincipalsPage.checkEmptyPrincipalTable();
      storagePrincipalsPage.firstPrincipal();
      storagePrincipalsPage.enterPrincipalName("principal1");
      storagePrincipalsPage.applyPermissionsForAllBucketsByDefault();
      storagePrincipalsPage.applySelectAllActions();
      storagePrincipalsPage.applySelectAllPolicies();
      storagePrincipalsPage.submitCreatePrincipal();
      cy.wait(6000);
      cy.get('[intc-id="principal1HyperLinkTable"]').should("be.visible").click();
      // Check the set permissions are correct
      storagePrincipalsPage.clickPermissionsTab();
      cy.contains("Buckets permissions").should("be.visible");
      cy.get('[aria-label]').filter((_, el) => Cypress.$(el).attr('aria-label').endsWith('rulebucket')).check();
      const allowedPolicies = [
        "GetBucketLocation",
        "GetBucketPolicy",
        "ListBucket",
        "ListBucketMultipartUploads",
        "ListMultipartUploadParts",
        "GetBucketTagging"
      ];
      const allowedActions = ["Read", "Write", "Delete"];
      for (let i = 0; i < allowedPolicies.length; i++) {
        cy.contains(allowedPolicies[i]).should("be.visible");
      }
      for (let i = 0; i < allowedActions.length; i++) {
        cy.contains(allowedActions[i]).should("be.visible");
      }

      cy.get('[aria-label]').filter((_, el) => Cypress.$(el).attr('aria-label').endsWith('secondbucket')).check();
      for (let i = 0; i < allowedPolicies.length; i++) {
        cy.contains(allowedPolicies[i]).should("be.visible");
      }
      for (let i = 0; i < allowedActions.length; i++) {
        cy.contains(allowedActions[i]).should("be.visible");
      }
    });

    it("6002 | Create Principal for one bucket with select permisions", function () {
      homePage.clickBucketStorageButton();
      storageBucketsPage.clickManagePrincipalButton();
      storagePrincipalsPage.createPrincipal();
      storagePrincipalsPage.enterPrincipalName("principal2");
      storagePrincipalsPage.applyPermissionsPerBucket();
      cy.get('[aria-label]').filter((_, el) => Cypress.$(el).attr('aria-label').endsWith('rulebucket')).check();
      storagePrincipalsPage.applyGetBucketLocation();
      storagePrincipalsPage.applyListMultipartUploadParts();
      storagePrincipalsPage.applyGetBucketPolicy();
      storagePrincipalsPage.applyReadPolicy();
      storagePrincipalsPage.submitCreatePrincipal();
      cy.wait(4000);
      cy.get('[intc-id="principal2HyperLinkTable"]').as('principal')
      cy.get('@principal').click({ force: true })
      cy.get('[intc-id="PermissionsTab"]').as('permissionsTab')
      cy.get('@permissionsTab').trigger('mouseover').wait(1000).click().click({ force: true });
      cy.wait(4000);
      cy.get('[aria-label]').filter((_, el) => Cypress.$(el).attr('aria-label').endsWith('rulebucket')).should("be.checked");
      const allowedPolicies = [
        "GetBucketLocation",
        "GetBucketPolicy",
        "ListMultipartUploadParts",
      ];
      for (let i = 0; i < allowedPolicies.length; i++) {
        cy.contains(allowedPolicies[i]).should("be.visible");
      }
      cy.contains("ReadBucket").should("be.visible");
    });

    it("6003 | Edit Principal from Actions", function () {
      homePage.clickBucketStorageButton();
      storageBucketsPage.clickManagePrincipalButton();
      cy.wait(4000);
      storagePrincipalsPage.checkPrincipalTableVisible();
      cy.get('[intc-id="principal2HyperLinkTable"]').should("be.visible").click();
      storagePrincipalsPage.clickActionsDropdown();
      storagePrincipalsPage.clickEditActionsButton();
      cy.get('[aria-label]').filter((_, el) => Cypress.$(el).attr('aria-label').endsWith('rulebucket')).check();
      storagePrincipalsPage.uncheckListMultipartUploadParts();
      storagePrincipalsPage.uncheckGetBucketLocation();
      storagePrincipalsPage.applyGetBucketTagging();
      storagePrincipalsPage.uncheckReadPolicy();
      storagePrincipalsPage.applyWritePolicy();
      storagePrincipalsPage.saveEdit();
      cy.wait(4000);
      // Check the set permissions are correct
      storagePrincipalsPage.clickPermissionsTab();
      cy.contains("Buckets permissions").should("be.visible");
      cy.get('[aria-label]').filter((_, el) => Cypress.$(el).attr('aria-label').endsWith('rulebucket')).should('be.checked');
      cy.contains("GetBucketPolicy").should("be.visible");
      cy.contains("GetBucketTagging").should("be.visible");
      cy.contains("WriteBucket").should("be.visible");
      cy.contains("ReadBucket").should("not.exist");
    });

    it("6004 | Check principal allow removing all permissions", function () {
      homePage.clickBucketStorageButton();
      storageBucketsPage.clickManagePrincipalButton();
      storagePrincipalsPage.checkPrincipalTableVisible();
      // Edit principal2
      storagePrincipalsPage.editPrincipalFromTable(0);
      cy.get('[aria-label]').filter((_, el) => Cypress.$(el).attr('aria-label').endsWith('rulebucket')).check();
      // Uncheck existing policies & actions
      storagePrincipalsPage.uncheckGetBucketPolicy();
      storagePrincipalsPage.uncheckGetBucketTagging();
      storagePrincipalsPage.uncheckWritePolicy();
      storagePrincipalsPage.saveEdit();
    });

    it("6005 | Cancel Edit principal", function () {
      homePage.clickBucketStorageButton();
      storageBucketsPage.clickManagePrincipalButton();
      storagePrincipalsPage.checkPrincipalTableVisible();
      // Edit principal1
      storagePrincipalsPage.editPrincipalFromTable(0);
      cy.get('[aria-label]').filter((_, el) => Cypress.$(el).attr('aria-label').endsWith('rulebucket')).check();
      storagePrincipalsPage.applyGetBucketLocation();
      storagePrincipalsPage.applyWritePolicy();
      storagePrincipalsPage.inputAllowedPolicyPath("/test");
      storagePrincipalsPage.cancelEditButton();
      storagePrincipalsPage.checkPrincipalTableVisible();
    });

    it("6006 | Create principal with invalid name", function () {
      homePage.clickBucketStorageButton();
      storageBucketsPage.clickManagePrincipalButton();
      storagePrincipalsPage.createPrincipal();
      storagePrincipalsPage.enterPrincipalName("$#@%");
      cy.contains("Only lower case alphanumeric and hypen(-) allowed for Name:.")
    });

    it("6007 | Create principal with long name", function () {
      homePage.clickBucketStorageButton();
      storageBucketsPage.clickManagePrincipalButton();
      storagePrincipalsPage.createPrincipal();
      storagePrincipalsPage.enterPrincipalName("maxlength-12abc");
      storagePrincipalsPage.getPrincipalNameInput().should("have.value", "maxlength-12");
    });

    it("6008 | Edit Principal for all buckets", function () {
      homePage.clickBucketStorageButton();
      storageBucketsPage.clickManagePrincipalButton();
      storagePrincipalsPage.checkPrincipalTableVisible();
      // Edit principal2
      storagePrincipalsPage.editPrincipalFromTable(0);
      storagePrincipalsPage.applyPermissionsForAllBuckets();
      storagePrincipalsPage.applyListBucket();
      storagePrincipalsPage.applyListBucketMultipartUploads();
      storagePrincipalsPage.applyDeletePolicy();
      storagePrincipalsPage.inputAllowedPolicyPath("/images");
      storagePrincipalsPage.saveEdit();
      cy.wait(4000);
      cy.get('[intc-id="principal2HyperLinkTable"]').should("be.visible").click({ force: true });
      // Check the set permissions are correct
      cy.get('[intc-id="PermissionsTab"]').as('permissionTab')
      cy.get('@permissionTab').trigger('mouseover').wait(1000).click().click({ force: true });
      cy.wait(3000);
      for (let i = 0; i < 2; i++) {
        const labelPart = i === 0 ? "rulebucket" : "secondbucket";
        cy.get('[aria-label]').filter((_, el) => Cypress.$(el).attr('aria-label').endsWith(labelPart)).check();
        const allowedActions = [
          "ListBucket",
          "ListBucketMultipartUploads",
        ];
        for (let i = 0; i < allowedActions.length; i++) {
          cy.contains(allowedActions[i]).should("be.visible");
        }
        cy.contains("Delete").should("be.visible");
      }
    });

    it("6009 | View Buckets from Principals Page", function () {
      homePage.clickBucketStorageButton();
      storageBucketsPage.clickManagePrincipalButton();
      storagePrincipalsPage.viewBuckets();
      storageBucketsPage.checkBucketTableVisible();
    });

    it("6010 | Verify cancel create principal", function () {
      homePage.clickBucketStorageButton();
      storageBucketsPage.clickManagePrincipalButton();
      storagePrincipalsPage.createPrincipal();
      storagePrincipalsPage.enterPrincipalName("tobecancel");
      storagePrincipalsPage.cancelCreate();
      cy.contains("Manage Principals and Permissions");
    });

    it("6011 | Delete all principals", function () {
      homePage.clickBucketStorageButton();
      storageBucketsPage.clickManagePrincipalButton();
      storagePrincipalsPage.deletePrincipals();
    });

    it("6012 | Delete all buckets", function () {
      homePage.clickBucketStorageButton();
      storageBucketsPage.deleteAllBuckets();
    });
  })
})