import accountKeyPage from "../../../pages/IDC2.0/accountKeyPage.js";
import storagePage from "../../../pages/IDC2.0/storagePage.js";
import TestFilter from "../../../support/testFilter.js";
const computePage = require("../../../pages/IDC2.0/computePage.js");
const instancePage = require("../../../pages/IDC2.0/instancePage.js");
const homePage = require("../../../pages/IDC2.0/homePage.js");
const cloudCredits = require("../../../pages/IDC2.0/cloudCreditsPage.js");
var publickey = Cypress.env("publicKey");
var regionNum = Cypress.env("region");

TestFilter(["IntelAll", "PremiumAll", "EnterpriseAll"], () => {
  describe("1 Storage - Verify volumes", () => {
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

    it("2009 | Reserve VM for storage E2E", function () {
      homePage.hardwareCatalog();
      homePage.clickfourthGenIntelProd();
      cy.wait(1000);
      instancePage.instanceType("small");
      instancePage.instanceName("vmstore");
      instancePage.clickCreateKeyButton();
      accountKeyPage.addKeyName("test-key");
      accountKeyPage.addKeyContent(publickey);
      instancePage.uploadKey()
      cy.wait(1000);
      computePage.launchInstance();
      cy.wait(45000);
      computePage.checkInstanceIsReady();
    });

    it("2010 | Create 2TB Volume", function () {
      homePage.clickStorageButton();
      storagePage.createFirstVolume();
      storagePage.volumeNameInput("staas");
      storagePage.inputStorageSize("2");
      storagePage.launchVolume();
      cy.wait(20000);
      storagePage.volumeTableIsVisible();
      cy.contains("staas")
    });

    it("2011 | Search for volume and Clear filter- Negative", function () {
      homePage.clickStorageButton();
      storagePage.volumeTableIsVisible();
      storagePage.searchVolume("volume1");
      cy.wait(3000);
      cy.contains("The applied filter criteria did not match any items.");
      storagePage.clearFilter();
    });

    it("2012 | Create volume with duplicate name", function () {
      homePage.clickStorageButton();
      storagePage.createVolume();
      storagePage.volumeNameInput("staas");
      cy.wait(2000);
      storagePage.inputStorageSize("5");
      storagePage.launchVolume();
      cy.wait(6000);
      cy.get(".modal-header")
        .contains("Could not create your volume")
        .within(($form) => {
          cy.wrap($form).should("be.visible");
          cy.wait(2000);
        });
      cy.get(".btn.btn-primary").contains("Go back").click({ force: true });
      cy.wait(3000);
      storagePage.clearVolumeName();
      storagePage.volumeNameInput("testvolume");
      storagePage.launchVolume();
      cy.wait(40000);
    });

    it("2013 | Edit Volume Size from 2TB to 5TB", function () {
      homePage.clickStorageButton();
      storagePage.volumeTableIsVisible();
      storagePage.searchVolume("staas");
      storagePage.editVolume();
      storagePage.inputStorageSize("5");
      if (Cypress.env('accountType') === "Intel") {
        cy.get('.d-flex.flex-row.valid-feedback').should("contain", "$0.15");
      } else {
        cy.get('.d-flex.flex-row.valid-feedback').should("contain", "$0.50");
      }
      storagePage.saveVolumeEdit();
      storagePage.volumeTableIsVisible();
      storagePage.searchVolume("staas");
      cy.get('[intc-id="staasHyperLinkTable"]').contains("staas").click();
      cy.contains("5TB");
      cy.wait(10000);
    });

    it("2014 | Verify How to Mount instructions", function () {
      homePage.clickStorageButton();
      storagePage.searchVolume("staas");
      cy.get('[intc-id="staasHyperLinkTable"]').contains("staas").click();
      storagePage.howToMountButton();
      storagePage.singleInstanceSelect();
      cy.get(".dropdown-item").contains("vmstore").should("be.visible").click();
      storagePage.howToConnectInstance();
      let selector = '.row.mt-0.align-items-center';
      let expectedContent = "chmod 400 my-key.ssh";
      if (Cypress.env("accountType") === "Intel") {
        cy.get(selector).eq(1).click();
      } else {
        cy.get(selector).eq(0).click();
      }
      cy.copyToClipboard(selector, expectedContent);
      storagePage.howToMountClose();
    });

    it("2015 | Verify How to unMount instructions", function () {
      homePage.clickStorageButton();
      storagePage.searchVolume("staas");
      cy.get('[intc-id="staasHyperLinkTable"]').contains("staas").click();
      storagePage.howToUnMountButton();
      storagePage.singleInstanceSelect();
      cy.get(".dropdown-item").contains("vmstore").should("be.visible").click();
      storagePage.howToConnectStorage();
      cy.wait(2000);
      cy.contains("sudo umount /mnt/test").should("be.visible");
      storagePage.howToMountClose();
    });

    it("2016 | Verify volume details tab", function () {
      homePage.clickStorageButton();
      storagePage.searchVolume("staas");
      cy.get('[intc-id="staasHyperLinkTable"]').contains("staas").click();
      cy.contains("Volume information");
      cy.contains("5TB").should("be.visible");
      cy.contains("Ready").should("be.visible");
    });

    it("2017 | Verify volume security tab", function () {
      homePage.clickStorageButton();
      storagePage.searchVolume("staas");
      cy.get('[intc-id="staasHyperLinkTable"]').contains("staas").click();
      storagePage.securitySection();
      // Temporal change due to VAST compatibility available in region 1
      if (regionNum?.toString() !== "1") {
        cy.get('[aria-label="Copy User"]').should("be.visible").scrollIntoView({ duration: 2000 });
        cy.get('.btn.btn-outline-primary.btn-sm').contains("Generate password").should("be.visible");
        storagePage.generatePassword();
      } else {
        cy.contains("Allowed IP Range");
      }
    });

    it("2018 | Delete 5TB volume from table grid", function () {
      homePage.clickStorageButton();
      storagePage.volumeTableIsVisible();
      storagePage.searchVolume("staas");
      cy.wait(2000);
      storagePage.deleteBtn();
      storagePage.checkConfirmDeleteModal();
      cy.getValue('[intc-id="deleteConfirmationName"]').then(text => {
        cy.get('[intc-id="NameInput"]').should("be.visible").clear().type(text);
      })
      storagePage.confirmDelete();
    });

    it("2019 | Verify Minimum allowed Storage Size input", function () {
      homePage.clickStorageButton();
      storagePage.createVolume();
      storagePage.volumeNameInput("minsize");
      storagePage.inputStorageSize("0");
      cy.get('[intc-id="StorageSize(TB)InvalidMessage"]').contains("Value less than 1 is not allowed.")
    });

    it("2020 | Verify Maximum allowed Storage Size input", function () {
      homePage.clickStorageButton();
      storagePage.createVolume();
      storagePage.volumeNameInput("maxsize");
      storagePage.inputStorageSize("2001");
      cy.get('[intc-id="StorageSize(TB)InvalidMessage"]').contains("Value more than 100 is not allowed.")
    });

    it("2021 | Verify Volume creation required fields", function () {
      homePage.clickStorageButton();
      storagePage.createVolume();
      storagePage.volumeNameInput("staas4");
      storagePage.inputStorageSize("5");
      storagePage.clearVolumeName();
      storagePage.clearVolumeSize();
      cy.contains("Name is required")
      cy.get('[intc-id="StorageSize(TB)InvalidMessage"]').contains("Storage Size (TB) is required")
      cy.get('[intc-id="btn-storagelaunch-navigationBottom Create"]').should('be.enabled')
    });

    it("2022 | Verify Storage volume creation using an invalid name", function () {
      homePage.clickStorageButton();
      storagePage.createVolume();
      storagePage.volumeNameInput("Name@@#$@!$%");
      cy.contains("Only lower case alphanumeric and hypen(-) allowed for Name:.")
    });

    it("2023 | Verify volume max allowed quota check", function () {
      homePage.clickStorageButton();
      storagePage.createVolume();
      storagePage.volumeNameInput("staas3");
      storagePage.inputStorageSize("2");
      storagePage.launchVolume();
      cy.wait(6000);
      cy.get(".modal-header")
        .contains("Could not create your volume")
        .within(($form) => {
          cy.wrap($form).should("be.visible");
          cy.wait(2000);
        });
      cy.get(".btn.btn-primary").contains("Go back").click({ force: true });
    });

    it("2024 | Delete all instances and account keys used for Storage Volumes", function () {
      homePage.clickcomputePageButton();
      cy.wait(1000);
      computePage.searchInstance("vmstore");
      computePage.deleteInstance();
      computePage.checkConfirmDeleteModal();
      cy.getValue('[intc-id="deleteConfirmationName"]').then(text => {
        cy.get('[intc-id="NameInput"]').should("be.visible").clear().type(text);
      })
      computePage.confirmDeleteInstance();
    });

    it("2025 | Delete Pub Keys", function () {
      homePage.clickKeysPageButton();
      cy.deleteFirstKey();
    });

    it("2026 | Delete all Storage Volumes", function () {
      homePage.clickStorageButton();
      cy.wait(2000);
      cy.deleteAllVolumes();
    });
  });
});


