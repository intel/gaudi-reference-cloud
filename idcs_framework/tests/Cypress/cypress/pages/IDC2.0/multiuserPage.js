class multiUser {
  elements = {
    // Account Access Management
    addMemberBtn: () => cy.get('[intc-id="btn-profile-addMember"]'),
    emailInput: () => cy.get('[intc-id="EmailInput"]'),
    invitationExpiration: () => cy.get('[intc-id="InvitationexpirationdateInput"]'),
    noteTextArea: () => cy.get('[intc-id="NoteTextArea"]'),
    cancelAddMemberBtn: () => cy.get('[intc-id="btn-accessmanagement-addMember-cancel"]'),
    grantMemberAccessBtn: () => cy.get('[intc-id="btn-accessmanagement-addMember-grant"]'),
    learnAboutAccountAccessDocs: () => cy.get('[class="btn.btn-link.px-0.text-start"]'),
    verificationCodeInput: () => cy.get('[intc-id="VerificationcodeInput"]'),
    verifyCode: () => cy.get('[intc-id="btn-accessmanagement-addMember-grant"]'),
    sendNewCode: () => cy.get('[intc-id="btn-accessmanagement-addMember-cancel"]'),
    validFeedback: () => cy.get('[class="valid-feedback"]'),
    clearFilter: () => cy.get('[intc-id="ClearfiltersEmptyViewButton"]'),
    searchMemberInvite: () => cy.get('[intc-id="btn-profile-saerchmember"]'),

    // Invitation actions
    searchInvitations: () => cy.get('[intc-id="btn-profile-saerchmember"]'),
    removeInvitation: () => cy.get('[intc-id="ButtonTable Remove invitation"]'),
    resendInvitation: () => cy.get('[intc-id="ButtonTable Resend invitation"]'),
    revokeAccess: () => cy.get('[intc-id="ButtonTable Revoke access"]'),

    // Remove invitation confirmation dialog
    cancelRemoveInvitation: () => cy.get('[intc-id="btn-accessmanagement-addMember-cancel"]'),
    confirmRemoveInvitation: () => cy.get('[intc-id="btn-accessmanagement-addMember-grant"]'),

    // Revoke invitation confirmation dialog
    cancelRevokeInvitation: () => cy.get('[intc-id="btn-accessmanagement-addMember-cancel"]'),
    confirmRevokeInvitation: () => cy.get('[intc-id="btn-accessmanagement-addMember-grant"]'),

    // My account
    searchInvitations: () => cy.get('[intc-id="searchInstances"]'),
    switchAccounts: () => cy.get('[intc-id="topNavBarSwitchAccounts"]'),

    // Membership accounts
    acceptInvitation: () => cy.get('[intc-id="btn-accounts-accept"]'),
    declineInvitation: () => cy.get('[intc-id="btn-accounts-decline"]'),

    // Accept invitation dialog
    invitationCodeInput: () => cy.get('[intc-id="Invitationcode:Input"]'),
    cancelInvitationConfirm: () => cy.get('[intc-id="btn-accessaccount-accept-invite-cancel"]'),
    confirmInvitation: () => cy.get('[intc-id="btn-accessaccount-accept-invite-confirm"]'),
    resendCode: () => cy.get(".btn.btn-primary.my-3").constains(" Resend"),

    // Decline invitation confirmation dialog
    confirmDecline: () => cy.get("btn.btn-secondary").contains('Cencel'),
    cancelDecline: () => cy.get("btn.btn-danger").contains('Decline'),
  };

  addMember() {
    this.elements.addMemberBtn().click({ force: true });
  }

  emailInput(email) {
    this.elements.emailInput().type(email);
  }

  setInvitationExpiration(date) {
    this.elements.invitationExpiration().click({ force: true }).type(date);
  }

  addNote(note) {
    this.elements.noteTextArea().type(note);
  }
  cancelAddMember() {
    this.elements.cancelAddMemberBtn().click({ force: true });
  }

  grantMemberAccess() {
    this.elements.grantMemberAccessBtn().click({ force: true });
  }

  setInvitationExpiration() {
    this.elements.invitationExpiration().click();
  }

  multiuserDocs() {
    this.elements.learnAboutAccountAccessDocs().click();
  }

  verificationCodeInput(otp) {
    this.elements.verificationCodeInput().type(otp);
  }

  verifyCode() {
    this.elements.verifyCode().click({ force: true });
  }

  sendNewCode() {
    this.elements.sendNewCode().click({ force: true });
  }

  validFeedbackMessage() {
    this.elements.validFeedback().click();
  }

  searchMember(name) {
    this.elements.searchMemberInvite().type(name);
  }

  // Invitation actions

  searchInvitations(name) {
    this.elements.searchInvitations().type(name);
  }

  clearFilter() {
    this.elements.clearFilter().click();
  }

  removeInvitation() {
    this.elements.removeInvitation().click();
  }

  resendInvitation() {
    this.elements.resendInvitation().click();
  }

  revokeAccess() {
    this.elements.revokeAccess().click();
  }

  // Remove invitation confirmation dialog

  cancelRemoveInvitation() {
    this.elements.cancelRemoveInvitation().click({ force: true });
  }

  confirmRemoveInvitation() {
    this.elements.confirmRemoveInvitation().click({ force: true });
  }

  // Revoke invitation confirmation dialog

  cancelRevokeInvitation() {
    this.elements.cancelRevokeInvitation().click();
  }

  confirmRevokeInvitation() {
    this.elements.confirmRevokeInvitation().click();
  }

  // My account

  searchInvitations(name) {
    this.elements.searchInvitations().type(name);
  }

  switchAccounts() {
    this.elements.switchAccounts().click({ force: true });
  }

  // Membership accounts

  acceptInvitation() {
    this.elements.acceptInvitation().click();
  }

  declineInvitation() {
    this.elements.declineInvitation().click();
  }

  // Accept invitation dialog

  invitationCodeInput(code) {
    this.elements.invitationCodeInput().type(code);
  }

  cancelInvitationConfirm() {
    this.elements.cancelInvitationConfirm().click({ force: true });
  }

  confirmInvitation() {
    this.elements.confirmInvitation().click();
  }

  resendCode() {
    this.elements.resendCode().click();
  }

  // Decline confirmation dialog

  confirmDecline() {
    this.elements.confirmDecline().click({ force: true });
  }

  cancelDecline() {
    this.elements.cancelDecline().click({ force: true });
  }

}

module.exports = new multiUser();
