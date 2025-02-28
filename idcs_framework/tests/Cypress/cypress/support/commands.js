/// <reference types="cypress" />
export { };
//CUSTOM COMMANDS...
// ***********************************************
// This example commands.js shows you how to
// create various custom commands and overwrite
// existing commands.
//
// For more comprehensive examples of custom
// commands please read more here:
// https://on.cypress.io/custom-commands
// ***********************************************
//
//
// -- This is a parent command --
// Cypress.Commands.add('login', (email, password) => { ... })
//
//
// -- This is a child command --
// Cypress.Commands.add('drag', { prevSubject: 'element'}, (subject, options) => { ... })
//
//
// -- This is a dual command --
// Cypress.Commands.add('dismiss', { prevSubject: 'optional'}, (subject, options) => { ... })
//
//
// -- This will overwrite an existing command --
// Cypress.Commands.overwrite('visit', (originalFn, url, options) => { ... })

import "cypress-file-upload";
import ClipboardJS from "clipboard";
const computePage = require("../pages/IDC2.0/computePage");
const storagePage = require("../pages/IDC2.0/storagePage");
import accountKeyPage from "../pages/IDC2.0/accountKeyPage";
const auth = require("../fixtures/IDC2.0/login");
const couponData = require("../fixtures/IDC2.0/coupon");

require("@reportportal/agent-js-cypress/lib/commands/reportPortalCommands");
require("cy-verify-downloads").addCustomCommand();

Cypress.on("uncaught:exception", (err, runnable) => {
  // returning false here prevents Cypress from
  // failing the test
  return false;
});

Cypress.Commands.add("saveLocalStorageCacheFirstTimeLogin", () => {
  cy.log("save");
  localStorage.setItem("completedSteps", "9");
  localStorage.setItem("firstRunStatus", "Completed");
  cy.reload();
  Object.keys(localStorage).forEach((key) => {
    LOCAL_STORAGE_MEMORY[key] = localStorage[key];
  });

  const token = localStorage.getItem("authToken").slice(1, -1);
  Cypress.env("token", token);
});

Cypress.Commands.add("saveLocalStorageCache", () => {
  cy.log("save");
  Object.keys(localStorage).forEach((key) => {
    LOCAL_STORAGE_MEMORY[key] = localStorage[key];
  });
});

Cypress.Commands.add("restoreLocalStorageCache", () => {
  cy.log("restore");
  Object.keys(LOCAL_STORAGE_MEMORY).forEach((key) => {
    localStorage.setItem(key, LOCAL_STORAGE_MEMORY[key]);
  });
});

Cypress.Commands.add("clearLocalStorageCache", () => {
  localStorage.clear();
  LOCAL_STORAGE_MEMORY = {};
});

Cypress.Commands.add("deleteAllAccountKeys", () => {
  cy.get('[intc-id="ButtonTable Delete key"]').then(($buttons) => {
    if ($buttons.length === 0) {
      cy.log('No more items to delete.');
      return;
    }
    cy.wrap($buttons[0]).click({ force: true });
    cy.handleDeleteModal();
    cy.wait(2000);
    cy.get('[intc-id="ButtonTable Delete key"]', { timeout: 10000 }).should('have.length.lessThan', $buttons.length).then(($updatedButtons) => {
      if ($updatedButtons.length > 0) {
        cy.deleteAllAccountKeys();
      } else {
        cy.log('All items have been deleted.');
        assert.isOk("OK", "deletion completed");
      }
    });
  });
});

Cypress.Commands.add("checkRemainingCredit", () => {
  cy.get('.ms-auto.text-end.fw-semibold').eq(1).should("be.visible");
  return cy.get('.ms-auto.text-end.fw-semibold').eq(1).invoke('text').then(text => {
    return text.trim() == "0 USD";
  });
});

Cypress.Commands.add('getValue', (selector) => {
  return cy.get(selector).invoke('text');
})

Cypress.Commands.add('handleDeleteModal', () => {
  cy.get('[intc-id="deleteConfirmModal"]').should("be.visible");
  cy.getValue('[intc-id="deleteConfirmationName"]').then(text => {
    cy.get('[intc-id="NameInput"]').should("be.visible").clear().type(text);
  })
  cy.get('.btn.btn-danger').should("be.visible").contains("Delete").click({ force: true });
})

Cypress.Commands.add("searchFilter", (searchvalue) => {
  cy.get('[intc-id="Filter-Text"]').scrollIntoView();
  cy.get('[intc-id="Filter-Text"]').type(searchvalue);
  cy.wait(4000);
  cy.get('.card-body').each(($item) => {
    const listItemText = $item.text();
    expect(listItemText).to.contain(searchvalue);
  });
});

Cypress.Commands.add("copyToClipboard", (selector, value) => {
  const clipboard = new ClipboardJS(selector);
  clipboard.on("success", (e) => {
    expect(e.text).to.equal(value);
  });
});

Cypress.Commands.add("elementExists", (selector) => {
  cy.get("body").then(($body) => {
    if ($body.find(selector, { timeout: 2000 }).length) {
      cy.wait(2000);
      return cy.get(selector);
    } else {
      // Throws no error when element not found
      assert.isOk("OK", "Element does not exist.");
    }
  });
});

Cypress.Commands.add("deleteAllInstances", () => {
  cy.get('[intc-id="ButtonTable Delete instance"]').then(($buttons) => {
    if ($buttons.length === 0) {
      cy.log('No more items to delete.');
      return;
    }
    cy.wrap($buttons[0]).click({ force: true });
    cy.handleDeleteModal();
    cy.wait(3000);
    cy.get('[intc-id="ButtonTable Delete instance"]', { timeout: 15000 }).should('have.length.lessThan', $buttons.length).then(($updatedButtons) => {
      if ($updatedButtons.length > 0) {
        cy.deleteAllInstances();
      } else {
        cy.log('All items have been deleted.');
        assert.isOk("OK", "deletion completed");
      }
    });
  });
});

Cypress.Commands.add("deleteAllVolumes", () => {
  cy.get('[intc-id="ButtonTable Delete storage"]').then(($buttons) => {
    if ($buttons.length === 0) {
      cy.log('No more items to delete.');
      return;
    }
    cy.wrap($buttons[0]).click({ force: true });
    cy.handleDeleteModal();
    cy.wait(2000);
    cy.get('[intc-id="ButtonTable Delete storage"]', { timeout: 10000 }).should('have.length.lessThan', $buttons.length).then(($updatedButtons) => {
      if ($updatedButtons.length > 0) {
        cy.deleteAllVolumes();
      } else {
        cy.log('All items have been deleted.');
        assert.isOk("OK", "deletion completed");
      }
    });
  });
});

Cypress.Commands.add("selectAccount", (cloudAccount) => {
  const accountOwner = cloudAccount;
  cy.get('[intc-id="AccountIDColumn1"]').should("be.visible").then(() => {
    cy.get('.btn.btn-link.text-decoration-none.p-0').contains(accountOwner).click({ force: true });
    cy.wait(10000);
  })
});

Cypress.Commands.add("selectRegion", (region) => {
  const selectedRegion = region || "1";
  if (selectedRegion !== "1") {
    const id = "-" + region;
    cy.get('[intc-id="regionLabel"]').should("be.visible").click({ force: true });
    cy.get('.dropdown-item').contains(id).should("be.visible").click({ force: true });
  } else
    cy.log("Using default region.");
});

Cypress.Commands.add("getCoupon", () => {
  let bodyParams;
  if (Cypress.env('accountType') === "Standard") {
    bodyParams = couponData.couponStandard;
  } else {
    bodyParams = couponData.coupon;
  }
  const token = Cypress.env('token');
  const authorization = `Bearer ${token}`;
  const options = {
    method: "POST",
    form: false,
    url: Cypress.env('globalEndpoint') + "v1/cloudcredits/coupons",
    headers: {
      authorization,
    },
    body: bodyParams,
  };
  cy.request(options).then((response) => {
    expect(response.body).to.have.property("code");
    const coupon = JSON.stringify(response.body.code);
    Cypress.env("newCoupon", coupon.slice(1, 15)), console.log(coupon);
  });
});

Cypress.Commands.add("deleteFirstKey", () => {
  cy.get('[intc-id="ButtonTable Delete key"]').first().click();
  cy.handleDeleteModal();
});

Cypress.Commands.add("deleteFirstInstance", () => {
  cy.get('[intc-id="ButtonTable Delete instance"]').first().click({ force: true });
  cy.handleDeleteModal();
  cy.contains("Terminating");
});
