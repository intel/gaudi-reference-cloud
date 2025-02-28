/// <reference types="cypress-mailslurp" />
import TestFilter from "../../support/testFilter";
import multiUser from "../../pages/IDC2.0/multiuserPage";
import homePage from "../../pages/IDC2.0/homePage";

TestFilter(["PremiumAll"], () => {
  describe("Multi user account owner verification - Premium user", () => {
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

    it("3010 | Ensure Premium User has Account Access Management enabled", function () {
      homePage.clickuserAccountTab();
      cy.get('[class="badge header-badge header-badge-bg-premium text-capitalize"]')
        .contains("Premium")
        .should("be.visible")
        .click();
      cy.wait(5000);
      cy.get('[intc-id="btn-profile-addMember"]')
        .should("be.visible")
        .contains("Grant access");
    });

    it("3011 | Validate Cancel Grant access to a member", function () {
      homePage.clickuserAccountTab();
      homePage.clickProfileSettings();
      multiUser.addMember();
      cy.wait(1000);
      multiUser.cancelAddMember();
      cy.wait(1000);
    });

    it("3012 | Verify send invitation with valid OTP - User with No Cloud Account", function () {
      const inboxId = Cypress.env("premiumAdminInboxId");
      homePage.clickuserAccountTab();
      homePage.clickProfileSettings();
      multiUser.addMember();
      multiUser.emailInput(Cypress.env("premium2MemberEmail"));
      cy.wait(1000);
      cy.get('[intc-id="InvitationexpirationdateInput"]').click({ force: true }).type("2024-12-12");
      multiUser.addNote("This is a test invite");
      multiUser.grantMemberAccess();
      cy.wait(22000);
      cy.mailslurp()
        .then(mailslurp =>
          mailslurp.waitForLatestEmail(inboxId, 40000, true))
        .then(email => {
          expect(email.subject).contain('Your OTP to Grant Invitation Access on Intel Cloud Services')
          cy.wait(4000);
          const code = email.body.match(/(\d{6})/g)[4]
          console.log(code)
          cy.get(".modal-content").contains("We sent a code to").as('modal')
          multiUser.verificationCodeInput(code)
          multiUser.verifyCode();
          cy.wait(50000);
        })
    });

    it("3013 | Verify send invitation to an already existing member", function () {
      const inboxId = Cypress.env("premiumAdminInboxId");
      homePage.clickuserAccountTab();
      homePage.clickProfileSettings();
      multiUser.addMember();
      multiUser.emailInput(Cypress.env("premium2MemberEmail"));
      cy.wait(1000);
      cy.get('[intc-id="InvitationexpirationdateInput"]').click({ force: true }).type("2024-12-12");
      multiUser.addNote("This is a test invite");
      multiUser.grantMemberAccess();
      cy.wait(22000);
      cy.mailslurp()
        .then(mailslurp =>
          mailslurp.waitForLatestEmail(inboxId, 40000, true))
        .then(email => {
          expect(email.subject).contain('Your OTP to Grant Invitation Access on Intel Cloud Services')
          cy.wait(4000);
          const code = email.body.match(/(\d{6})/g)[4]
          console.log(code)
          cy.get(".modal-content").contains("We sent a code to").as('modal')
          multiUser.verificationCodeInput(code)
          multiUser.verifyCode();
          cy.wait(30000);
          cy.wrap(".toast-container.position-body.p-3.mb-5").then(($toast) => {
            cy.get($toast).contains("memberEmail already exist");
          })
        })
    });

    it("3014 | Verify invalid OTP - Alphanumeric value", function () {
      homePage.clickuserAccountTab();
      homePage.clickProfileSettings();
      multiUser.addMember();
      multiUser.emailInput(`${Date.now()}@test.com`);
      cy.wait(1000);
      cy.get('[intc-id="InvitationexpirationdateInput"]').click({ force: true }).type("2024-12-12");
      multiUser.addNote("This is a test invite");
      multiUser.grantMemberAccess();
      cy.wait(22000);
      cy.get(".modal-content").contains("We sent a code to").as('modal')
      multiUser.verificationCodeInput("AABC589")
      multiUser.verifyCode();
      cy.wait(50000);
      cy.get('[intc-id="VerificationcodeInvalidMessage"]').contains('Incorrect verification code');
    });

    it("3015 | Verify send invitation with an invalid E-mail format", function () {
      homePage.clickuserAccountTab();
      homePage.clickProfileSettings();
      multiUser.addMember();
      multiUser.emailInput("invalid-email");
      cy.wait(3000);
      cy.get('[intc-id="EmailInvalidMessage"]').contains("Invalid email address.");
      multiUser.cancelAddMember();
      cy.wait(1000);
    });

    it("3016 | Verify previous Date for Invitation expiration", function () {
      homePage.clickuserAccountTab();
      homePage.clickProfileSettings();
      multiUser.addMember();
      multiUser.emailInput(`${Date.now()}@test.com`);
      cy.wait(1000);
      cy.get('[intc-id="InvitationexpirationdateInput"]').click({ force: true }).type("2023-12-12");
      cy.wait(5000);
      cy.get('[intc-id="InvitationexpirationdateInvalidMessage"]').contains("Date must be greater than today");
      multiUser.cancelAddMember();
    });

    it("3017 | Validate Multiuser documentation link", function () {
      homePage.clickuserAccountTab();
      homePage.clickProfileSettings();
      cy.get('.btn.btn-link.px-0.text-start').should(
        "have.attr",
        "href",
        "/docs/guides/multi_user_accounts.html"
      );
    });

    it("3018 | Verify Search for member invitations - Invalid input", function () {
      homePage.clickuserAccountTab();
      homePage.clickProfileSettings();
      multiUser.searchMember("wrong")
      cy.wait(5000);
      cy.get('.add-break-line.mt-3.lead').contains('The applied filter criteria did not match any items');
      multiUser.clearFilter();
      cy.wait(1000);
    });

    it("3019 | Verify Resend invitation action", function () {
      homePage.clickuserAccountTab();
      homePage.clickProfileSettings();
      cy.get("body").then(($body) => {
        cy.wait(5000);
        if ($body.find('[intc-id="GrantaccessEmptyViewButton"]').length > 0) {
          assert.isOk('OK', 'No invitations');
        } else {
          cy.contains('td', 'Resend')
            .parent()
            .within($tr => {
              cy.wrap($tr).then(() => {
                multiUser.resendInvitation();
                cy.wait(40000);
              })
            })
        }
      })
    });

    it("3020 | Verify Remove invitation action", function () {
      homePage.clickuserAccountTab();
      homePage.clickProfileSettings();
      cy.get("body").then(($body) => {
        cy.wait(5000);
        if ($body.find('[intc-id="GrantaccessEmptyViewButton"]').length > 0) {
          assert.isOk('OK', 'No invitations');
        } else {
          cy.contains('td', 'Remove invitation')
            .parent()
            .within($tr => {
              cy.wrap($tr).then(() => {
                multiUser.removeInvitation();
              })
            })
          cy.get(".modal-content").contains("Remove invitation").as('modal')
          multiUser.confirmRemoveInvitation();
          cy.wait(40000);
        }
      })
    });

    it("3021 | Verify send new OTP code - invitation to Standard User with CA", function () {
      const inboxId = Cypress.env("premiumAdminInboxId");
      homePage.clickuserAccountTab();
      homePage.clickProfileSettings();
      multiUser.addMember();
      multiUser.emailInput(Cypress.env("standardMemberEmail"));
      cy.wait(1000);
      cy.get('[intc-id="InvitationexpirationdateInput"]').click({ force: true }).type("2024-12-12");
      multiUser.addNote("This is an invite to a Standard User");
      multiUser.grantMemberAccess();
      cy.wait(22000);
      cy.get(".modal-content").contains("We sent a code to").as('modal')
      multiUser.sendNewCode();
      cy.mailslurp()
        .then(mailslurp =>
          mailslurp.waitForLatestEmail(inboxId, 40000, true))
        .then(email => {
          expect(email.subject).contain('Your OTP to Grant Invitation Access on Intel Cloud Services')
          cy.wait(4000);
          const code = email.body.match(/(\d{6})/g)[4]
          console.log(code)
          multiUser.verificationCodeInput(code)
          multiUser.verifyCode();
          cy.wait(50000);
        })
    });

    it("3022 | Verify send invitation with valid OTP to Premium User with CA", function () {
      const inboxId = Cypress.env("premiumAdminInboxId");
      homePage.clickuserAccountTab();
      homePage.clickProfileSettings();
      multiUser.addMember();
      multiUser.emailInput(Cypress.env("premiumMemberEmail"));
      cy.wait(1000);
      cy.get('[intc-id="InvitationexpirationdateInput"]').click({ force: true }).type("2024-12-12");
      multiUser.addNote("Invitation to Premium User");
      multiUser.grantMemberAccess();
      cy.wait(22000);
      cy.mailslurp()
        .then(mailslurp =>
          mailslurp.waitForLatestEmail(inboxId, 40000, true))
        .then(email => {
          expect(email.subject).contain('Your OTP to Grant Invitation Access on Intel Cloud Services')
          cy.wait(4000);
          const code = email.body.match(/(\d{6})/g)[4]
          console.log(code)
          cy.get(".modal-content").contains("We sent a code to").as('modal')
          multiUser.verificationCodeInput(code)
          multiUser.verifyCode();
          cy.wait(50000);
        })
    });

    it("3023 | Verify send invitation with valid OTP to Enterprise User with CA", function () {
      const inboxId = Cypress.env("premiumAdminInboxId");
      homePage.clickuserAccountTab();
      homePage.clickProfileSettings();
      multiUser.addMember();
      multiUser.emailInput(Cypress.env("enterpriseMemberEmail"));
      cy.wait(1000);
      cy.get('[intc-id="InvitationexpirationdateInput"]').click({ force: true }).type("2024-12-12");
      multiUser.addNote("Invitation to Enterprise User");
      multiUser.grantMemberAccess();
      cy.wait(22000);
      cy.mailslurp()
        .then(mailslurp =>
          mailslurp.waitForLatestEmail(inboxId, 40000, true))
        .then(email => {
          expect(email.subject).contain('Your OTP to Grant Invitation Access on Intel Cloud Services')
          cy.wait(4000);
          const code = email.body.match(/(\d{6})/g)[4]
          console.log(code)
          cy.get(".modal-content").contains("We sent a code to").as('modal')
          multiUser.verificationCodeInput(code)
          multiUser.verifyCode();
          cy.wait(50000);
        })
    });

    it("3024 | Verify Revoke access action", function () {
      homePage.clickuserAccountTab();
      homePage.clickProfileSettings();
      cy.get("body").then(($body) => {
        cy.wait(5000);
        if ($body.find('[intc-id="GrantaccessEmptyViewButton"]').length > 0) {
          assert.isOk('OK', 'No invitations');
        } else {
          cy.contains('td', 'Revoke')
            .parent()
            .within($tr => {
              cy.wrap($tr).then(() => {
                multiUser.revokeAccess();
              })
            })
          cy.get(".modal-content").contains("Revoke access").as('modal')
          multiUser.confirmRevokeInvitation();
          cy.wait(40000);
        }
      })
    })
  });
});

