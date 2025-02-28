import TestFilter from "../../../support/testFilter";
const previewKeyPage = require("../../../pages/IDC2.0/previewKeyPage");
const previewComputePage = require("../../../pages/IDC2.0/previewComputePage.js");
const previewStoragePage = require("../../../pages/IDC2.0/previewStoragePage.js");
const homePage = require("../../../pages/IDC2.0/homePage.js");
const instancePage = require("../../../pages/IDC2.0/instancePage.js");
var publickey = Cypress.env("publicKey");

TestFilter(["IntelAll", "PremiumAll", "StandardAll", "ependingAll", "EnterpriseAll"], () => {
  describe("1 - Preview Keys", () => {
    beforeEach(() => {
      cy.PrepareSession()
      cy.GetSession();
    });

    afterEach(() => {
      cy.TestClean()
    })

    after(() => {
      cy.TestClean()
    })

    it("3000 | Create preview key", function () {
      homePage.clickpreviewKeyPageButton();
      cy.wait(2000);
      previewKeyPage.uploadFirstKey();
      previewKeyPage.addKeyName("test1");
      previewKeyPage.setAssociatedEmail();
      previewKeyPage.addKeyContent(publickey);
      previewKeyPage.createKey();
    });

    it("3001 | Create preview key with invalid public key", function () {
      homePage.clickpreviewKeyPageButton();
      previewKeyPage.uploadKey();
      previewKeyPage.addKeyName("test-key");
      previewKeyPage.setAssociatedEmail();
      previewKeyPage.addKeyContent("invalidtestkey");
      previewKeyPage.createKey();
      cy.contains(
        "found unsupported key type"
      ).should("be.visible");
    });

    it("3002 | Create preview key with duplicate name", function () {
      homePage.clickpreviewKeyPageButton();
      previewKeyPage.uploadKey();
      previewKeyPage.addKeyName("test1");
      previewKeyPage.setAssociatedEmail();
      previewKeyPage.addKeyContent(publickey);
      previewKeyPage.createKey();
      cy.contains(
        "already exists"
      ).should("be.visible");
    });

    it("3003 | Cancel creation of preview key", function () {
      homePage.clickpreviewKeyPageButton();
      previewKeyPage.uploadKey();
      previewKeyPage.addKeyName("test-cancel");
      previewKeyPage.setAssociatedEmail();
      previewKeyPage.addKeyContent(publickey);
      previewKeyPage.cancelKeyCreate();
    });

    it("3004 | Create preview key with special chars in name", function () {
      homePage.clickpreviewKeyPageButton();
      previewKeyPage.uploadKey();
      previewKeyPage.addKeyName("test@$^%*");
      previewKeyPage.setAssociatedEmail();
      previewKeyPage.addKeyContent(publickey);
      cy.contains(
        "Only lower case alphanumeric and hypen(-) allowed for Key Name: *."
      ).should("be.visible");
    });

    it("3005 | Validate copy button for generating SSH key", function () {
      homePage.clickpreviewKeyPageButton();
      previewKeyPage.uploadKey();
      previewKeyPage.clickCreateSSHDetailBar();
      const selector = '.section.code-line.rounded-3.mt-s4';
      const expectedContent = 'ssh-keygen -t rsa -b 4096 -f $env:UserProfile\.ssh\id_rsa'
      cy.get(selector).eq(1).click();
      cy.copyToClipboard(selector, expectedContent);
    });

    it("3006 | Validate copy button for open public key step - Linux/Mac", function () {
      homePage.clickpreviewKeyPageButton();
      previewKeyPage.uploadKey();
      previewKeyPage.clickCreateSSHDetailBar();
      previewKeyPage.clickLinuxOption();
      const selector = '.row.mt-0.align-items-center';
      const expectedContent = 'cat ~/.ssh/id_rsa.pub'
      cy.get(selector).eq(1).click();
      cy.copyToClipboard(selector, expectedContent);
    });

    it("3007 | Validate warning in upload key page", function () {
      homePage.clickpreviewKeyPageButton();
      previewKeyPage.uploadKey();
      cy.contains(
        "Never share your private keys with anyone. Never create a SSH Private key without a passphrase"
      ).should("be.visible");
    });

    it("3008 | Create and validate preview key with empty name", function () {
      homePage.clickpreviewKeyPageButton();
      previewKeyPage.uploadKey();
      previewKeyPage.addKeyName(" ");
      previewKeyPage.clearKeyName();
      previewKeyPage.setAssociatedEmail();
      previewKeyPage.addKeyContent(publickey);
      cy.contains("Key Name is required").should("be.visible");
    });

    it("3009 | SSH key Documentation", function () {
      homePage.clickpreviewKeyPageButton();
      previewKeyPage.uploadKey();
      previewKeyPage.clickCreateSSHDetailBar();
      cy.get('.valid-feedback').contains("SSH key documentation").should(
        "have.attr",
        "href",
        "/docs/guides/ssh_keys.html"
      );
    });

    it("3010 | Select Linux OS for SSH key creation", function () {
      homePage.clickpreviewKeyPageButton();
      previewKeyPage.uploadKey();
      previewKeyPage.clickCreateSSHDetailBar();
      previewKeyPage.clickLinuxOption();
    });

    it("3011 | Select Windows OS for SSH key creation", function () {
      homePage.clickpreviewKeyPageButton();
      previewKeyPage.uploadKey();
      previewKeyPage.clickCreateSSHDetailBar();
      previewKeyPage.clickLinuxOption();
      previewKeyPage.clickWindowsOption();
    });

    it("3012 | Validate key content error message for preview keys", function () {
      homePage.clickpreviewKeyPageButton();
      previewKeyPage.uploadKey();
      previewKeyPage.addKeyName("test1");
      previewKeyPage.setAssociatedEmail();
      previewKeyPage.addKeyContent(publickey);
      previewKeyPage.clearKeyContent();
      cy.contains("key contents is required").should("be.visible");
    });

    it("3013 | Search preview keys success", function () {
      homePage.clickpreviewKeyPageButton();
      previewKeyPage.uploadKey();
      previewKeyPage.addKeyName("test2");
      previewKeyPage.setAssociatedEmail();
      previewKeyPage.addKeyContent(publickey);
      previewKeyPage.createKey();
      cy.wait(4000);
      previewKeyPage.clicksearchFilter("test2");
      cy.wait(2000);
      cy.contains("test2").should("be.visible");
    });

    it("3014 | Search preview keys with no items found", function () {
      homePage.clickpreviewKeyPageButton();
      previewKeyPage.checkVisibleKeyTable();
      previewKeyPage.clicksearchFilter("test3");
      cy.wait(2000);
      previewKeyPage.checkEmptyKeyTable();
      previewKeyPage.clickClearFilterButton();
      previewKeyPage.checkEmptySearchFilter();
      previewKeyPage.checkVisibleKeyTable();
    });

    it("3015 | Cannot delete preview keys in use", function () {
      //Create compute instance
      homePage.clickpreviewComputePageButton();
      cy.get(".btn.btn-primary").contains("Request instance").click({ force: true });
      previewComputePage.instanceName("testinstance");
      instancePage.elements.CPU().should('be.checked');
      instancePage.checkSierraForest();
      previewComputePage.intendedUsed("test deployment");
      cy.get('input[type="checkbox"][value="test2"]').check();
      cy.get('[intc-id="btn-computelaunch-navigationBottom Request instance - singlenode"]').should('be.enabled').then(() => {
        instancePage.submitPreviewInstance();
      });
      cy.wait(5000);
      // Verify that key cannot be deleted
      cy.get('.tap-inactive.nav-link').contains("Preview Keys").click({ force: true });
      cy.wait(2000);
      cy.contains("Item in use").should("be.visible");
      // Clean up created instance
      cy.get('.tap-inactive.nav-link').contains("Preview Instances").click({ force: true });
      cy.contains("testinstance");
      cy.get('[intc-id="ButtonTable Delete instance"]').first().click({ force: true });
      cy.handleDeleteModal();
      previewComputePage.checkEmptyInstanceTable();
    });

    it("3017 | Verify preview key with an invalid associated email", function () {
      homePage.clickpreviewKeyPageButton();
      previewKeyPage.uploadKey();
      previewKeyPage.addKeyName("mykey");
      previewKeyPage.addEmail("erroremail");
      previewKeyPage.addKeyContent(publickey);
      cy.contains("Invalid email address.").should("be.visible");
      previewKeyPage.clearAssociatedEmail();
      cy.contains("Associated Email is required").should("be.visible");
    });

    it("3018 | Verify preview key with an invalid associated email Domain", function () {
      homePage.clickpreviewKeyPageButton();
      previewKeyPage.uploadKey();
      previewKeyPage.addKeyName("mykey");
      previewKeyPage.addEmail("invalid@error.com");
      previewKeyPage.addKeyContent(publickey);
      previewKeyPage.createKey();
      cy.contains("email domain must match your email domain").should("be.visible");
    });

    it("3019 | Delete preview keys", function () {
      homePage.clickpreviewKeyPageButton();
      cy.deleteFirstKey();
    });
  });
});

TestFilter(["IntelAll", "PremiumAll", "StandardAll", "ependingAll", "EnterpriseAll"], () => {
  describe("2 - Reserve Preview instance verification", () => {
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

    it("4001 | Reserve Preview instance with public Key", function () {
      homePage.clickpreviewComputePageButton();
      previewComputePage.requestInstance();
      previewComputePage.machineImageLabel();
      previewComputePage.instanceName("testinstance");
      instancePage.elements.CPU().should('be.checked');
      instancePage.checkSierraForest();
      previewComputePage.intendedUsed("test deployment");
      instancePage.clickCreateKeyButton();
      previewKeyPage.addKeyName("testkey");
      previewKeyPage.setAssociatedEmail();
      previewKeyPage.addKeyContent(publickey);
      instancePage.uploadKey();
      cy.wait(3000);
      instancePage.submitPreviewInstance();
      cy.wait(8000);
      cy.contains("PRE-BM-SRF-SP");
    });

    it("4002 | Preview instance Edit name is disabled", function () {
      homePage.clickpreviewComputePageButton();
      cy.wait(1000);
      previewComputePage.searchInstance("testinstance");
      cy.wait(2000);
      previewComputePage.editInstance();
      previewComputePage.editInstanceNameIsDisabled();
      previewComputePage.cancelEdit();
    });

    it("4003 | Cancel Delete VM action", function () {
      homePage.clickpreviewComputePageButton();
      cy.wait(1000);
      previewComputePage.searchInstance("testinstance");
      previewComputePage.deleteInstance();
      cy.wait(1000);
      previewComputePage.cancelDeleteInstance();
    });

    it("4004 | Verify instance details tab", function () {
      homePage.clickpreviewComputePageButton();
      previewComputePage.searchInstance("testinstance");
      cy.get('[intc-id="testinstanceHyperLinkTable"]').contains("testinstance").click();
      cy
        .contains("Instance Type:")
        .scrollIntoView({ duration: 2000 });
      cy.contains("Intel® Xeon® 6700E-series processors (formerly codenamed Sierra Forest Scalable Processors)").should("be.visible");
    });

    it("4005 | Verify instance Security tab", function () {
      homePage.clickpreviewComputePageButton();
      previewComputePage.searchInstance("testinstance");
      cy.get('[intc-id="testinstanceHyperLinkTable"]').contains("testinstance").click();
      previewComputePage.clickSecurityTab();
      cy.contains("Key name").scrollIntoView({ duration: 2000 });
      cy.get('[intc-id="fillTable"]').contains("testkey").should("be.visible");
    });

    it("4006 | Verify instance Networking tab", function () {
      homePage.clickpreviewComputePageButton();
      previewComputePage.searchInstance("testinstance");
      cy.get('[intc-id="testinstanceHyperLinkTable"]').contains("testinstance").click();
      previewComputePage.clickNetworkingTab();
      cy.contains("vNet name").scrollIntoView({ duration: 2000 });
      cy.get('[intc-id="fillTable"]').contains("idc-preview").should("be.visible");
    });

    it("4007 | Verify close icon for how to connect", function () {
      homePage.clickpreviewComputePageButton();
      previewComputePage.searchInstance("testinstance");
      cy.get('[intc-id="testinstanceHyperLinkTable"]').contains("testinstance").click();
      previewComputePage.clickhowToConnectButton();
      previewComputePage.clickcloseHowtoConnectIcon();
    });

    it("4008 | Verify instance edit page", function () {
      homePage.clickpreviewComputePageButton();
      previewComputePage.checkInstanceTableVisible();
      previewComputePage.editInstance();
      previewComputePage.cancelEdit();
    });

    it("4009 | Cancel deleting an instance using actions", function () {
      homePage.clickpreviewComputePageButton();
      cy.get('[intc-id="testinstanceHyperLinkTable"]').contains("testinstance").click();
      previewComputePage.clickactionsButton();
      cy.get('[intc-id="myReservationActionsDropdownItemButton3"]').click();
      previewComputePage.checkConfirmDeleteModal();
      previewComputePage.cancelDeleteInstance();
    });

    it("4010 | Verify Preview Request Extension - invalid Min value", function () {
      homePage.clickpreviewComputePageButton();
      previewComputePage.searchInstance("testinstance");
      cy.wait(2000);
      previewComputePage.clickExtend();
      previewComputePage.inputExtensionDays("0");
      cy.get('[intc-id="Extensionday(s)InvalidMessage"]').contains("Value less than 1 is not allowed.");
    });

    it("4011 | Verify Preview Request Extension - invalid Max value", function () {
      homePage.clickpreviewComputePageButton();
      previewComputePage.searchInstance("testinstance");
      cy.wait(2000);
      previewComputePage.clickExtend();
      previewComputePage.inputExtensionDays("20");
      cy.get('[intc-id="Extensionday(s)InvalidMessage"]').contains("Value more than 14 is not allowed.");
    });

    it("4012 | Search a non-existing instance", function () {
      homePage.clickpreviewComputePageButton();
      previewComputePage.checkInstanceTableVisible();
      previewComputePage.searchInstance("fakeInstance");
      previewComputePage.checkEmptyInstanceTable();
    });

    it("4013 | Edit an instance's key to switch to a new key", function () {
      homePage.clickpreviewComputePageButton();
      cy.wait(2000);
      previewComputePage.checkInstanceTableVisible();
      previewComputePage.editInstance();
      cy.wait(1000);
      previewComputePage.clickUploadKeyButton();
      previewKeyPage.addKeyName("switch-key");
      previewKeyPage.setAssociatedEmail();
      previewKeyPage.addKeyContent(publickey);
      cy.get('[intc-id="btn-ssh-createpublickey"]').should("be.visible").click({ force: true });
      cy.wait(4000);
      cy.get('input[type="checkbox"][value="switch-key"]').should("be.visible");
      cy.get('input[type="checkbox"][value="switch-key"]').check();
      cy.get('input[type="checkbox"][value="testkey"]').uncheck();
      previewComputePage.saveEditInstance();
      cy.get('[intc-id="testinstanceHyperLinkTable"]').should("be.visible");
      cy.get('[intc-id="testinstanceHyperLinkTable"]').click();
      previewComputePage.clickSecurityTab();
      cy.get('[intc-id="KeynameColumn1"]')
        .contains("Key name")
        .scrollIntoView({ duration: 2000 });
      cy.get('[intc-id="fillTable"]').contains("switch-key").should("be.visible");
      cy.get('[intc-id="fillTable"]').should("not.contain", "testkey");
    });

    it("4014 | Edit an instance's key to include another key", function () {
      homePage.clickpreviewComputePageButton();
      cy.wait(2000);
      previewComputePage.checkInstanceTableVisible();
      previewComputePage.editInstance();
      cy.wait(1000);
      cy.get('input[type="checkbox"][value="testkey"]').check();
      previewComputePage.saveEditInstance();
      cy.get('[intc-id="testinstanceHyperLinkTable"]').should("be.visible");
      cy.get('[intc-id="testinstanceHyperLinkTable"]').click();
      previewComputePage.clickSecurityTab();
      cy.get('[intc-id="KeynameColumn1"]')
        .contains("Key name")
        .scrollIntoView({ duration: 2000 });
      cy.get('[intc-id="fillTable"]').contains("switch-key").should("be.visible");
      cy.get('[intc-id="fillTable"]').contains("testkey").should("be.visible");
    });

    it("4015 | Verify Gaudi3 displays UseCase select field", function () {
      homePage.clickpreviewComputePageButton();
      previewComputePage.requestFromTable();
      instancePage.clickAI();
      previewComputePage.selectGaudi3();
      cy.contains("Object Storage access is not available for this instance type.")
      previewComputePage.selectDeployAI_useCase();
      previewComputePage.intendedUsed("test deployment");
      previewComputePage.cancelRequest();
    });

    it("4016 | Verify closing Request Extend modal from details Actions", function () {
      homePage.clickpreviewComputePageButton();
      cy.get('[intc-id="testinstanceHyperLinkTable"]').contains("testinstance").click();
      cy.wait(2000);
      previewComputePage.clickactionsButton();
      previewComputePage.clickExtendActionsButton();
      previewComputePage.closeExtendModal();
    });

    it("4017 | Verify how to connect steps for windows option", function () {
      homePage.clickpreviewComputePageButton();
      cy.get('[intc-id="testinstanceHyperLinkTable"]').click();
      cy.wait(4000);
      previewComputePage.clickhowToConnectButton();
      cy.contains("chmod 400 my-key.ssh").should("be.visible");
    });

    it("4018 | Delete preview instance from table grid", function () {
      // Delete created instance
      homePage.clickpreviewComputePageButton();
      previewComputePage.deleteInstance();
      previewComputePage.checkConfirmDeleteModal();
      cy.getValue('[intc-id="deleteConfirmationName"]').then(text => {
        cy.get('[intc-id="NameInput"]').should("be.visible").clear().type(text);
      })
      previewComputePage.confirmDeleteInstance();
      previewComputePage.checkEmptyInstanceTable();
    });

    it("4019 | Delete test preview keys", function () {
      homePage.clickpreviewKeyPageButton();
      // Delete created keys
      previewKeyPage.checkVisibleKeyTable();
      cy.deleteAllAccountKeys();
    });
  });
});

TestFilter(["IntelAll", "PremiumAll", "StandardAll", "ependingAll", "EnterpriseAll"], () => {
  describe("3 - Reserve Preview Storage verification", () => {
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

    it("5031 | Reserve Preview instance with No public Key for Storage use", function () {
      homePage.clickpreviewComputePageButton();
      previewComputePage.requestInstance();
      previewComputePage.instanceName("storageinstance");
      instancePage.elements.CPU().should('be.checked');
      instancePage.checkSierraForest();
      previewComputePage.intendedUsed("test deployment");
      instancePage.submitPreviewInstance();
      cy.wait(30000);
      cy.contains("PRE-BM-SRF-SP");
    });

    it("5032 | Request a 10 GB Storage Bucket for Preview instance", function () {
      homePage.clickpreviewStoragePageButton();
      cy.wait(1000);
      previewStoragePage.requestStorage();
      cy.wait(2000);
      previewStoragePage.selectSize10GB();
      previewStoragePage.requestBucket();
      previewStoragePage.checkStorageTableIsVisible();
      cy.contains("10");
    });

    it("5033 | Verify Storage Preview Extend reservation not allowed before 30 days", function () {
      homePage.clickpreviewStoragePageButton();
      cy.wait(1000);
      previewStoragePage.checkStorageTableIsVisible();
      previewStoragePage.requestExtendFromTable();
      cy.contains("To request an extension, please submit your request within the last 30 days of your reservation.");
      previewStoragePage.cancelExtend();
    });

    it("5034| Verify Cancel Edit Size for Preview storage", function () {
      homePage.clickpreviewStoragePageButton();
      cy.wait(1000);
      previewStoragePage.editBucketFromTable();
      previewStoragePage.selectSize20GB();
      previewStoragePage.cancelRequest();
    });

    it("5035 | Verify Edit Storage size from 10GB to 20GB", function () {
      homePage.clickpreviewStoragePageButton();
      cy.wait(1000);
      previewStoragePage.editBucketFromTable();
      previewStoragePage.selectSize20GB();
      previewStoragePage.requestBucket();
      previewStoragePage.checkStorageTableIsVisible();
      cy.contains("20");
    });

    it("5036 | Verify Edit Storage size from 20GB to 50GB", function () {
      homePage.clickpreviewStoragePageButton();
      cy.wait(1000);
      previewStoragePage.editBucketFromTable();
      previewStoragePage.selectSize50GB();
      previewStoragePage.requestBucket();
      previewStoragePage.checkStorageTableIsVisible();
      cy.contains("50");
    });

    it("5037 | Verify Cancel Storage Preview Extend reservation before 20 days of expiration", function () {
      homePage.clickpreviewStoragePageButton();
      cy.wait(1000);
      previewStoragePage.checkStorageTableIsVisible();
      previewStoragePage.requestExtendFromTable();
      previewStoragePage.cancelExtend();
    });

    it("5038 | Verify Storage Preview Extend reservation before 20 days of expiration", function () {
      homePage.clickpreviewStoragePageButton();
      cy.wait(1000);
      previewStoragePage.checkStorageTableIsVisible();
      previewStoragePage.requestExtendFromTable();
      cy.get(".modal-content").as('modal');
      cy.get("@modal").should("be.visible").contains("Extend storage reservation");
      previewStoragePage.confirmExtend();
    });


    it("5039 | Verify How to use - Preview storage", function () {
      homePage.clickpreviewStoragePageButton();
      cy.wait(1000);
      previewStoragePage.checkStorageTableIsVisible();
      previewStoragePage.clickHowToUse();
      cy.contains("To connect to a bucket you need an access key. When creating a key, save it in a secure location outside the instance.");
      previewStoragePage.clickGenerateKey();
    });

    it("5040 | Cancel deleting Preview Storage bucket", function () {
      homePage.clickpreviewStoragePageButton();
      cy.wait(1000);
      previewStoragePage.checkStorageTableIsVisible();
      previewStoragePage.deleteBucketFromTable();
      previewStoragePage.checkConfirmDeleteModal();
      previewStoragePage.cancelDelete();
    });

    it("5041 | Verify delete Preview Storage", function () {
      homePage.clickpreviewStoragePageButton();
      cy.wait(1000);
      previewStoragePage.checkStorageTableIsVisible();
      previewStoragePage.deleteBucketFromTable();
      cy.handleDeleteModal();
    });

    it("5042 | Delete preview test instances", function () {
      // Delete created instance
      homePage.clickpreviewComputePageButton();
      previewComputePage.deleteInstance();
      previewComputePage.checkConfirmDeleteModal();
      cy.getValue('[intc-id="deleteConfirmationName"]').then(text => {
        cy.get('[intc-id="NameInput"]').should("be.visible").clear().type(text);
      })
      previewComputePage.confirmDeleteInstance();
      previewComputePage.checkEmptyInstanceTable();
    });

    it("5043 | Verify Requesting Storage without any Preview instance", function () {
      homePage.clickpreviewStoragePageButton();
      cy.wait(1000);
      cy.contains("Cannot Request Storage").should("be.visible");
    });
  });
});