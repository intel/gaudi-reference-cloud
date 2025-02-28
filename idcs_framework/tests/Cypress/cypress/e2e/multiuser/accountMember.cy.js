/// <reference types="cypress-mailslurp" />
import TestFilter from "../../support/testFilter.js";
import multiUser from "../../pages/IDC2.0/multiuserPage.js";

TestFilter(["PremiumAll"], () => {
  describe("Multi user accounts - Member verifications", () => {
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

    it("40012 | Decline invitation", function () {
      const memberEmail = Cypress.env("premiumMemberEmail");
      cy.get('[intc-id="AccountIDColumn1"]').should("be.visible").then(() => {
        cy.wait(20000);
        cy.contains('td', memberEmail)
          .parent()
          .within($tr => {
            cy.wrap($tr).then(() => {
              multiUser.declineInvitation();
              cy.get(".modal-content").contains("Decline").as('modal')
              multiUser.cancelDecline();
              cy.wait(40000);
            })
          })
      })
    });

    it("40012 | Accept invitation - input Incorrect code", function () {
      const memberEmail = Cypress.env("premiumMemberEmail");
      cy.get('[intc-id="AccountIDColumn1"]').should("be.visible").then(() => {
        cy.wait(20000);
        cy.contains('td', memberEmail)
          .parent()
          .within($tr => {
            cy.wrap($tr).then(() => {
              multiUser.acceptInvitation();
              cy.wait(15000);
              cy.get(".modal-content").contains("Accept invitation from cloud account").as('modal')
              multiUser.invitationCodeInput("AABC589")
              multiUser.confirmInvitation();
              cy.wait(40000);
              cy.get('[intc-id="Invitationcode:InvalidMessage"]').contains('Incorrect code');
              multiUser.cancelInvitationConfirm();
            })
          })
      })
    });

    it("4013 | Verify confirm invitation from Premium account login as Premium Member", function () {
      const inboxId = Cypress.env("premiumMemberInboxId");
      cy.wait(10000)
      cy.get('[intc-id="AccountIDColumn1"]').should("be.visible").then(() => {
        multiUser.acceptInvitation();
        cy.wait(15000);
        cy.mailslurp()
          .then(mailslurp =>
            mailslurp.waitForLatestEmail(inboxId, 40000, true))
          .then(email => {
            expect(email.subject).contain("You're invited to Intel Developer Cloud! Please respond")
            cy.wait(4000);
            const code = email.body.match(/(\d{8})/g)[4]
            console.log(code)
            cy.get(".modal-content").contains("Accept invitation from cloud account").as('modal')
            multiUser.invitationCodeInput(code)
            multiUser.confirmInvitation();
            cy.wait(40000);
            cy.wrap(".toast-container.position-body.p-3.mb-5").then(($toast) => {
              cy.get($toast).contains("Invitation confirmed!");
            })
          })
      })
    });

    it("4014 | Verify confirm invitation from Premium account login as Standard Member", function () {
      const inboxId = Cypress.env("standardMemberInboxId");
      cy.wait(10000)
      cy.get('[intc-id="AccountIDColumn1"]').should("be.visible").then(() => {
        multiUser.acceptInvitation();
        cy.wait(15000);
        cy.mailslurp()
          .then(mailslurp =>
            mailslurp.waitForLatestEmail(inboxId, 40000, true))
          .then(email => {
            expect(email.subject).contain("You're invited to Intel Developer Cloud! Please respond")
            cy.wait(4000);
            const code = email.body.match(/(\d{8})/g)[4]
            console.log(code)
            cy.get(".modal-content").contains("Accept invitation from cloud account").as('modal')
            multiUser.invitationCodeInput(code)
            multiUser.confirmInvitation();
            cy.wait(40000);
            cy.wrap(".toast-container.position-body.p-3.mb-5").then(($toast) => {
              cy.get($toast).contains("Invitation confirmed!");
            })
          })
      })
    });

    it("4015 | Verify confirm invitation from Premium account login as Enterprise User Member", function () {
      const inboxId = Cypress.env("enterpriseMemberInboxId");
      cy.wait(10000)
      cy.get('[intc-id="AccountIDColumn1"]').should("be.visible").then(() => {
        multiUser.acceptInvitation();
        cy.wait(15000);
        cy.mailslurp()
          .then(mailslurp =>
            mailslurp.waitForLatestEmail(inboxId, 40000, true))
          .then(email => {
            expect(email.subject).contain("You're invited to Intel Developer Cloud! Please respond")
            cy.wait(4000);
            const code = email.body.match(/(\d{8})/g)[4]
            console.log(code)
            cy.get(".modal-content").contains("Accept invitation from cloud account").as('modal')
            multiUser.invitationCodeInput(code)
            multiUser.confirmInvitation();
            cy.wait(40000);
            cy.wrap(".toast-container.position-body.p-3.mb-5").then(($toast) => {
              cy.get($toast).contains("Invitation confirmed!");
            })
          })
      })
    });
  });

  TestFilter(["Premium"], () => {
    describe("Compute Flow Verification - Using Member", () => {
      beforeEach(() => {
        cy.PrepareSession();
        cy.GetSession();
        cy.selectAccount(Cypress.env("premMemberCloudAccount"));
      });

      afterEach(() => {
        cy.TestClean();
      });

      after(() => {
        cy.TestClean();
      });
      // Calls existing Compute regression using a Member account from cy.selectAccout method.
      require('../IDC2.0/compute/computeFlow.cy.js')
    })
  })
});
