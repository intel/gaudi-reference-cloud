import TestFilter from "../../../support/testFilter";
const accountKeyPage = require("../../../pages/IDC2.0/accountKeyPage");
const computePage = require("../../../pages/IDC2.0/computePage.js");
const computeGroupPage = require("../../../pages/IDC2.0/computeGroupPage.js");
const homePage = require("../../../pages/IDC2.0/homePage.js");
const instancePage = require("../../../pages/IDC2.0/instancePage.js");
const VMdata = require("../../../fixtures/IDC2.0/reserveVM_data.json");
var publickey = Cypress.env("publicKey");
var regionNum = Cypress.env("region");

TestFilter(["IntelAll", "PremiumAll", "StandardAll", "ependingAll", "EnterpriseAll"], () => {
  describe("0 - Account Keys", () => {
    beforeEach(() => {
      cy.PrepareSession()
      cy.GetSession()
      cy.selectRegion(regionNum)
    });

    afterEach(() => {
      cy.TestClean()
    })

    after(() => {
      cy.TestClean()
    })

    it("1010 | Create account key", function () {
      homePage.clickKeysPageButton();
      cy.wait(2000);
      accountKeyPage.createFirstKey();
      cy.wait(2000);
      accountKeyPage.addKeyName("test1");
      accountKeyPage.addKeyContent(publickey);
      cy.wait(2000);
      accountKeyPage.createKey();
    });

    it("1011 | Create account key with invalid public key", function () {
      homePage.clickKeysPageButton();
      accountKeyPage.uploadKey();
      accountKeyPage.addKeyName("test2");
      accountKeyPage.addKeyContent("invalidtestkey");
      accountKeyPage.createKey();
      cy.contains(
        "SshPublicKey should have at least algorithm and publickey"
      ).should("be.visible");
    });

    it("1012 | Create account key with duplicate name", function () {
      homePage.clickKeysPageButton();
      accountKeyPage.uploadKey();
      accountKeyPage.addKeyName("test1");
      accountKeyPage.addKeyContent(publickey);
      accountKeyPage.createKey();
      accountKeyPage.duplicateKeyToastMessage();
    });

    it("1013 | Cancel creation of account key", function () {
      homePage.clickKeysPageButton();
      accountKeyPage.uploadKey();
      accountKeyPage.addKeyName("test-cancel");
      accountKeyPage.addKeyContent(publickey);
      accountKeyPage.cancelKeyCreate();
    });

    it("1014 | Create account key with special chars in name", function () {
      homePage.clickKeysPageButton();
      accountKeyPage.uploadKey();
      accountKeyPage.addKeyName("test@$^%*");
      accountKeyPage.addKeyContent(publickey);
      cy.contains(
        "Only lower case alphanumeric and hypen(-) allowed for Key Name: *."
      ).should("be.visible");
    });

    it("1015 | Validate copy button for generating SSH key", function () {
      homePage.clickKeysPageButton();
      accountKeyPage.uploadKey();
      accountKeyPage.clickCreateSSHDetailBar();
      const selector = '.section.code-line.rounded-3.mt-s4';
      const expectedContent = 'ssh-keygen -t rsa -b 4096 -f $env:UserProfile\.ssh\id_rsa'
      cy.get(selector).eq(1).click();
      cy.copyToClipboard(selector, expectedContent);
    });

    it("1016 | Validate copy button for open public key step", function () {
      homePage.clickKeysPageButton();
      accountKeyPage.uploadKey();
      accountKeyPage.clickCreateSSHDetailBar();
      accountKeyPage.clickLinuxOption();
      const selector = '.row.mt-0.align-items-center';
      const expectedContent = 'ssh-keygen -t rsa -b 4096 -f ~/.ssh/id_rsa'
      cy.get(selector).eq(0).click();
      cy.copyToClipboard(selector, expectedContent);
    });

    it("1017 | Validate warning in upload key page", function () {
      homePage.clickKeysPageButton();
      accountKeyPage.uploadKey();
      cy.contains(
        "Never share your private keys with anyone. Never create a SSH Private key without a passphrase"
      ).should("be.visible");
    });

    it("1018 | Create and validate account key with empty name", function () {
      homePage.clickKeysPageButton();
      accountKeyPage.uploadKey();
      accountKeyPage.addKeyName(" ");
      accountKeyPage.clearKeyName();
      accountKeyPage.addKeyContent(publickey);
      cy.contains("Key Name is required").should("be.visible");
    });

    it("1019 | SSH key Documentation", function () {
      homePage.clickKeysPageButton();
      accountKeyPage.uploadKey();
      accountKeyPage.clickCreateSSHDetailBar();
      cy.get('.valid-feedback').contains("SSH key documentation").should(
        "have.attr",
        "href",
        "/docs/guides/ssh_keys.html"
      );
    });

    it("1020 | Select Linux OS for SSH key creation", function () {
      homePage.clickKeysPageButton();
      accountKeyPage.uploadKey();
      accountKeyPage.clickCreateSSHDetailBar();
      accountKeyPage.clickLinuxOption();
    });

    it("1021 | Select Windows OS for SSH key creation", function () {
      homePage.clickKeysPageButton();
      accountKeyPage.uploadKey();
      accountKeyPage.clickCreateSSHDetailBar();
      accountKeyPage.clickLinuxOption();
      accountKeyPage.clickWindowsOption();
    });

    it("1022 | Validate key content error message for account keys", function () {
      homePage.clickKeysPageButton();
      accountKeyPage.uploadKey();
      accountKeyPage.addKeyName("test1");
      accountKeyPage.addKeyContent(publickey);
      accountKeyPage.clearKeyContent();
      cy.contains("key contents is required").should("be.visible");
    });

    it("1023 | Validate warning in upload key page", function () {
      homePage.clickKeysPageButton();
      accountKeyPage.uploadKey();
      cy.contains(
        "Never share your private keys with anyone. Never create a SSH Private key without a passphrase"
      ).should("be.visible");
    });

    it("1024 | Delete account keys", function () {
      homePage.clickKeysPageButton();
      cy.deleteFirstKey();
    });
  });
});

TestFilter(["IntelAll", "PremiumAll", "StandardAll", "ependingAll", "EnterpriseAll"], () => {
  describe("1 - Reserve Virtual Machine verification", () => {
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
    /*  Disabling dynamic coupon creation
        it("1025 | Redeem new coupon", function () {
          homePage.checkDashboard();
          cy.checkRemainingCredit().then(isRemainingCreditZero => {
            if (isRemainingCreditZero) {
              cy.loginAdmin();
              cy.wait(5000);
              cy.getCoupon();
              homePage.viewCouponPage();
              cloudCredits.redeemCoupon();
              cy.wrap(Cypress.env("newCoupon")).then(() => {
                cloudCredits.typeCouponCode(Cypress.env("newCoupon"));
              });
              cloudCredits.clickRedeemButton();
            } else {
              assert.isOk("OK", "Credit is available.");
            }
          });
        });
    */
    it("1026 | Reserve VM", function () {
      homePage.hardwareCatalog();
      homePage.clickfourthGenIntelProd();
      cy.wait(1000);
      instancePage.instanceType("small");
      cy.wait(1000);
      instancePage.instanceName("test2");
      cy.wait(1000);
      instancePage.clickCreateKeyButton();
      accountKeyPage.addKeyName("test-key");
      accountKeyPage.addKeyContent(publickey);
      instancePage.uploadKey()
      cy.wait(2000);
      computePage.launchInstance();
      cy.wait(40000);
    });

    it("1027 | Edit VM action", function () {
      homePage.clickcomputePageButton();
      computePage.searchInstance("test2");
      cy.wait(2000);
      computePage.editInstance();
      cy.wait(1000);
      computePage.editInstanceName();
    });

    it("1028 | Cancel Delete VM action", function () {
      homePage.clickcomputePageButton();
      cy.wait(1000);
      computePage.searchInstance("test2");
      computePage.deleteInstance();
      computePage.checkConfirmDeleteModal();
      computePage.cancelDeleteInstance();
    });

    it("1029 | Delete VM action", function () {
      homePage.clickcomputePageButton();
      cy.wait(1000);
      computePage.searchInstance("test2");
      computePage.deleteInstance();
      computePage.checkConfirmDeleteModal();
      cy.getValue('[intc-id="deleteConfirmationName"]').then(text => {
        cy.get('[intc-id="NameInput"]').should("be.visible").clear().type(text);
      })
      computePage.confirmDeleteInstance();
    });

    it("1030 | Reserve VM with different OS", function () {
      homePage.hardwareCatalog();
      homePage.clickfourthGenIntelProd();
      cy.wait(1000);
      instancePage.instanceType("small");
      cy.wait(1000);
      instancePage.machineImage();
      cy.get('[intc-id="Machineimage-form-select-option-ubuntu-2204-jammy-v20230122"]').click();
      instancePage.instanceName("test1");
      cy.wait(1000);
      computePage.launchInstance();
      cy.wait(40000);
      computePage.instanceTableIsVisible();
      computePage.searchInstance("test1");
      computePage.checkInstanceIsReady();
    });

    it("1031 | Reserve VM with Duplicate Name", function () {
      homePage.hardwareCatalog();
      homePage.clickfourthGenIntelProd();
      cy.wait(1000);
      instancePage.instanceType("small");
      instancePage.instanceName("test1");
      computePage.launchInstance();
      cy.wait(8000);
      cy.get(".modal-header").should("be.visible").as('modal')
      cy.get("@modal").should("be.visible").contains("Could not launch your instance");
      cy.get(".btn.btn-primary").contains("Go back").click({ force: true });
      instancePage.instanceName("test5");
      computePage.launchInstance();
      cy.wait(4000);
    });

    it("1032 | Reserve VM with Long Name", function () {
      homePage.hardwareCatalog();
      homePage.clickfourthGenIntelProd();
      cy.wait(1000);
      instancePage.instanceType("small");
      instancePage.instanceName("test-sample-with-more-than-sixty-three-characters-1234567891011121314152");
      cy.wait(1000);
      cy.get('[intc-id="InstancenameInvalidMessage"]').contains(
        "Max length 63 characters."
      );
      cy.wait(1000);
      instancePage.instanceName(
        "test-sample-with-less-than-sixty-three-characters"
      );
      instancePage.clickCancelButton();
    });

    it("1033 | Cancel Reserve VM", function () {
      homePage.hardwareCatalog();
      homePage.clickfourthGenIntelProd();
      cy.wait(1000);
      instancePage.instanceType("small");
      instancePage.instanceName("cancel-vm-test");
      cy.wait(1000);
      instancePage.clickCancelButton();
    });

    it("1034 | Reserve VM with special char name", function () {
      homePage.hardwareCatalog();
      homePage.clickfourthGenIntelProd();
      cy.wait(1000);
      instancePage.instanceType("small");
      instancePage.instanceName("Machine-897$");
      cy.wait(1000);
      cy.get('[intc-id="InstancenameInvalidMessage"]').contains(
        "Only lower case alphanumeric and hypen(-) allowed for Instance name:."
      );
      cy.wait(1000);
      instancePage.instanceName("lower-case-name");
      instancePage.clickCancelButton();
    });

    it("1035 | Verify instance details tab", function () {
      homePage.clickcomputePageButton();
      computePage.searchInstance("test1");
      cy.get('[intc-id="test1HyperLinkTable"]').contains("test1").click();
      cy.get('.d-flex.flex-column.col-md-3')
        .contains("Instance Type:")
        .scrollIntoView({ duration: 2000 });
      cy.contains("4th Generation Intel® Xeon® Scalable processor").should("be.visible");
    });

    it("1036 | Verify instance networking tab", function () {
      homePage.clickcomputePageButton();
      computePage.searchInstance("test1");
      cy.get('[intc-id="test1HyperLinkTable"]').contains("test1").click();
      computePage.clicknetworkSection();
      cy.get('[intc-id="vNetnameColumn2"]')
        .contains("vNet name")
        .scrollIntoView({ duration: 2000 });
      cy.get('[intc-id="regionLabel"]').then((reg) => {
        var region = reg.text().trim();
        cy.get('[intc-id="fillTable"]').eq(1).should("contain.text", region);
      });
    });

    it("1037 | Verify instance security tab", function () {
      homePage.clickcomputePageButton();
      computePage.searchInstance("test1");
      cy.get('[intc-id="test1HyperLinkTable"]').contains("test1").click();
      computePage.clicksecuritySection();
      cy.get('[intc-id="KeynameColumn1"]')
        .contains("Key name")
        .scrollIntoView({ duration: 2000 });
      cy.get('[intc-id="fillTable"]').contains("test").should("be.visible");
    });

    it("1046 | Verify close icon for how to connect", function () {
      homePage.clickcomputePageButton();
      computePage.searchInstance("test1");
      cy.get('[intc-id="test1HyperLinkTable"]').contains("test1").click({ force: true });
      computePage.clickhowToConnectButton();
      computePage.clickcloseHowtoConnectIcon();
    });

    it("1047 | Verify instance edit page", function () {
      homePage.clickcomputePageButton();
      cy.get('[intc-id="test1HyperLinkTable"]').contains("test1").click();
      computePage.clickactionsButton();
      computePage.clickeditActionsButton();
      cy.contains("After saving your changes, follow the next steps to complete the instance key update process.");
    });

    it("1048 | Cancel deleting an instance using actions", function () {
      homePage.clickcomputePageButton();
      cy.get('[intc-id="test1HyperLinkTable"]').contains("test1").click();
      computePage.clickactionsButton();
      computePage.clickdeleteActionsButton();
      computePage.checkConfirmDeleteModal();
      computePage.cancelDeleteInstance();
    });

    it("1049 | Delete an instance using actions", function () {
      homePage.clickcomputePageButton();
      cy.get('[intc-id="test1HyperLinkTable"]').contains("test1").click();
      computePage.clickactionsButton();
      computePage.clickdeleteActionsButton();
      computePage.checkConfirmDeleteModal();
      cy.getValue('[intc-id="deleteConfirmationName"]').then(text => {
        cy.get('[intc-id="NameInput"]').should("be.visible").clear().type(text);
      })
      computePage.confirmDeleteInstance();
    });

    it("1150 | Verify Contact Support URL from Could not launch modal", function () {
      homePage.hardwareCatalog();
      cy.get('[intc-id="btn-hardwarecatalog-select Intel® Max Series GPU"]').click();
      cy.wait(1000);
      instancePage.instanceName("gputest");
      computePage.launchInstance();
      cy.wait(5000);
      cy.get(".modal-header").should("be.visible").as('modal')
      cy.get("@modal").should("be.visible").contains("Could not launch your instance");
      cy.get('[intc-id="error-modal-contact-support-link"]').should(
        "have.attr",
        "href",
        "https://www.intel.com/content/www/us/en/support/contact-intel.html#support-intel-products_67709:59441:2314824"
      );
    });
  });
}
);

TestFilter(["Intel", "Premium"], () => {
  describe("2 - Reserve different types of machines - Specific to user", () => {
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

    // Test Cases
    VMdata.forEach((element) => {
      if (!element.skipped) {
        if (element.Smoke) {
          it(`${element.id} | ${element.name}`, () => {
            //Reserving all VM types
            homePage.hardwareCatalog();
            homePage.clickfourthGenIntelProd();
            cy.wait(1000);
            instancePage.instanceType(element.VMsize);
            instancePage.instanceName(element.instanceName);
            cy.wait(1000);
            computePage.launchInstance();
            cy.wait(40000);
          });
        }
      } else {
        it.skip(element.id + " | " + element.name, () => {
          cy.log("Skipped");
        });
      }
    });

    // Disabling Bare Metal coverage as products not available in Dev clusters.
    /*
    it("1055 | Reserve 3rd Generation Intel® Xeon® Scalable Processors", function () {
      homePage.hardwareCatalog()
      homePage.clickthirdGenIntelProd();
      //instancePage.instanceFamily(1);
      cy.wait(1000);
      instancePage.instanceName("3rd-gen");
      cy.wait(1000);
      computePage.launchInstance();
      cy.wait(1000);
    });
  
    it("1056 | Reserve Habana* Gaudi2 Deep Learning Server", function () {
      homePage.hardwareCatalog()
      homePage.clickhabanaGaudi2Prod();
      instancePage.instanceFamily(3);
      cy.wait(1000);
      instancePage.instanceName("gaudi2");
      cy.wait(1000);
      computePage.launchInstance();
      cy.wait(1000);
    });
 
    it("1057 | Reserve 4th Generation Intel® Xeon® Scalable processors", function () {
      homePage.hardwareCatalog()
      homePage.clickfourthGenIntelProd();
      //instancePage.instanceFamily(4);
      cy.wait(1000);
      instancePage.instanceName("4-gen");
      cy.wait(1000);
      computePage.launchInstance();
      cy.wait(1000);
    });

    it("1058 | Reserve Intel® Xeon® processors - Sapphire Rapids (HBM)", function () {
      homePage.hardwareCatalog()
      homePage.clickintelXeonProd();
      //instancePage.instanceFamily(5);
      cy.wait(1000);
      instancePage.instanceName("xeon-sapphire");
      cy.wait(1000);
      computePage.launchInstance();
      cy.wait(1000);
    });

    it("1059 | Reserve Intel® Max Series GPU (PVC)", function () {
      homePage.hardwareCatalog()
      homePage.clickintelMaxSeriesProd();
      //instancePage.instanceFamily(6);
      cy.wait(1000);
      instancePage.instanceName("max-series-gpu");
      cy.wait(1000);
      computePage.launchInstance();
      cy.wait(1000);
    });
    */
  });
});

TestFilter(["Premium", "Enterprise"], () => {
  describe("3 - How to connect external users verification", function () {
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

    it("1045 | Verify how to connect steps for windows option", function () {
      homePage.clickcomputePageButton();
      cy.get('[intc-id="vm-smallHyperLinkTable"]').contains("vm-small").click();
      computePage.clickhowToConnectButton();
      cy.contains("chmod 400 my-key.ssh").should("be.visible");
    });
  });
});

TestFilter(["Intel"], () => {
  describe("3 - How to Connect Intel user verification", function () {
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


    it("1039 | Verify copy button for step-1 in how to connect - Windows", function () {
      homePage.clickcomputePageButton();
      cy.get('[intc-id="vm-smallHyperLinkTable"]').contains("vm-small").click();
      computePage.clickhowToConnectButton();
      const selector = '.row.mt-0.align-items-center';
      const expectedContent =
        'Host 146.152.*.* ProxyCommand "C:Program FilesGitmingw64\binconnect.exe" -S internal-placeholder.com:1080 %h %p';
      cy.get(selector).eq(0).click();
      cy.copyToClipboard(selector, expectedContent);
    });

    it("1040 | Verify copy button for step-1 in how to connect - Linux/Mac", function () {
      homePage.clickcomputePageButton();
      cy.get('[intc-id="vm-smallHyperLinkTable"]').contains("vm-small").click();
      computePage.clickhowToConnectButton();
      computePage.clicklinuxOSOption();
      const selector = '.row.mt-0.align-items-center';
      const expectedContent =
        "Host 146.152.*.* ProxyCommand /usr/bin/nc -x internal-placeholder.com:1080 %h %p";
      cy.get(selector).eq(0).click();
      cy.copyToClipboard(selector, expectedContent);
    });

    it("1041 | Verify copy button for step-4 in how to connect", function () {
      homePage.clickcomputePageButton();
      cy.get('[intc-id="vm-smallHyperLinkTable"]').contains("vm-small").click();
      computePage.clickhowToConnectButton();
      const selector = '.row.mt-0.align-items-center';
      const expectedContent = "chmod 400 my-key.ssh";
      cy.get(selector).eq(1).click();
      cy.copyToClipboard(selector, expectedContent);
    });
  });
});

TestFilter(["epending"], () => {
  describe("Reserve Paid instance for Enterprise Pending user", function () {
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

    it("1060 | Verify reserving 4th Generation paid product throws error message", function () {
      homePage.hardwareCatalog();
      homePage.clickfourthGenIntelProd();
      instancePage.instanceFamily(1);
      instancePage.instanceType("medium");
      cy.wait(1000);
      instancePage.instanceName("test-paidprod");
      computePage.launchInstance();
      cy.wait(2000);
      cy.contains("Account confirmation required").should("be.visible");
      cy.contains(
        "Your enterprise account needs to be confirmed before launching this instance type. While the confirmation is in progress you can only use free instances."
      ).should("be.visible");
    });
  });
});

TestFilter(["IntelAll", "PremiumAll", "StandardAll", "ependingAll", "EnterpriseAll"], () => {
  describe("4 - Delete All Instances - Specific User", () => {
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

    it("1061 | Delete all instances", function () {
      homePage.clickcomputePageButton();
      cy.deleteAllInstances();
    });

    it("1062 | Delete all Keys", function () {
      homePage.clickKeysPageButton();
      cy.deleteFirstKey();
    });
  });
}
);

TestFilter(["IntelAll", "PremiumAll", "EnterpriseAll"], () => {
  describe("5 - Reserve Gaudi2 instance Group verification", () => {
    beforeEach(() => {
      cy.PrepareSession();
      cy.GetSession();
      cy.selectRegion(regionNum)
    });

    afterEach(() => {
      cy.TestClean();
    });

    after(() => {
      cy.TestClean();
    });

    /*  Disabling TC as there are no Gaudi 4 nodes available in dev clusters
        it("1127 | Reserve Gaudi2 Instance Group - when no BMH available", function () {
          homePage.clickcomputeGroupButton();
          computeGroupPage.launchInstanceGroup();
          cy.wait(1000);
          instanceGroupPage.instanceGroupName("gaudi-cluster");
          cy.wait(1000);
          instanceGroupPage.clickCreateKeyButton();
          accountKeyPage.addKeyName("test-key2");
          accountKeyPage.addKeyContent(publickey);
          instancePage.uploadKey()
          cy.wait(2000);
          instanceGroupPage.clickLaunchButton();
          cy.wait(8000);
          cy.get(".modal-dialog.modal-md.modal-dialog-centered").should("be.visible")
          cy.get(".modal-content").as('modal')
          cy.get("@modal").should("be.visible").contains("Instance groups unavailable");
          cy.get(".text-decoration-none.btn.btn-secondary").contains("Go Back").click({ force: true });
          cy.wait(2000);
        });
    
        it("1127 | Reserve Gaudi2 Instance Group - when no BMH available", function () {
          homePage.clickcomputeGroupButton();
          computeGroupPage.launchInstanceGroup();
          cy.wait(1000);
          instanceGroupPage.instanceGroupName("gaudi-cluster");
          cy.wait(1000);
          instanceGroupPage.clickCreateKeyButton();
          accountKeyPage.addKeyName("test-key2");
          accountKeyPage.addKeyContent(publickey);
          instancePage.uploadKey()
          cy.wait(2000);
          instanceGroupPage.clickLaunchButton();
          cy.wait(8000);
          cy.get(".modal-dialog.modal-lg.modal-dialog-centered").should("be.visible")
          cy.get(".modal-content").as('modal')
          cy.get("@modal").should("be.visible").contains("Could not launch your instance");
          cy.get(".btn.btn-primary").contains("Go back").click({ force: true });
          cy.wait(2000);
          instanceGroupPage.clickCancelButton();
        });
    */
    it("1127 | Cancel Delete Gaudi2 instance Group action", function () {
      homePage.clickcomputeGroupButton();
      cy.wait(1000);
      cy.get('[intc-id="data-view-empty"]')
        .should("be.visible")
        .contains("No instance groups found")
        .then((emptyMessage) => {
          if (emptyMessage) {
            assert.isOk("OK", "No instance groups available to delete.");
          } else {
            computeGroupPage.searchInstance("gaudi-cluster");
            computeGroupPage.deleteInstanceGroup();
            cy.wait(1000);
            computeGroupPage.cancelDeleteInstance();
          }
        });
    });

    it("1128 | Delete Gaudi2 instance Group action", function () {
      homePage.clickcomputeGroupButton();
      cy.wait(1000);
      cy.get('[intc-id="data-view-empty"]')
        .should("be.visible")
        .contains("No instance groups found")
        .then((emptyMessage) => {
          if (emptyMessage) {
            assert.isOk("OK", "No instance groups available to delete.");
          } else {
            computeGroupPage.searchInstance("gaudi-cluster");
            computeGroupPage.deleteInstanceGroup();
            cy.wait(1000);
            computeGroupPage.confirmDeleteInstance();
            homePage.clickaccountKeyPageButton();
            cy.deleteAllAccountKeys();
          }
        });
    });
  });
});
