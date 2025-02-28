class accountKeys {

    elements = {
        // Main account key page
        pageNumber: () => cy.get('[intc-id="Select"]'),
        deleteKeyButton: () => cy.get('[intc-id="ButtonTable Delete"]'),
        uploadKeyButton: () => cy.get('.btn.btn-primary').contains('Upload key'),
        createFirstKey: () => cy.get('[intc-id="UploadkeyEmptyViewButton"]').contains('Upload key'),
        emptyViewUploadKey: () => cy.get('[intc-id="btn-sshview-UploadKeyOnNoKeysAvailable"]').contains('Upload key'),
        cancelKeyDeletion: () => cy.get('[intc-id="btn-confirm-Deletekey-cancel"]').contains('Cancel'),
        confirmDeleteKey: () => cy.get('[intc-id="btn-confirm-Deletekey-delete"]').contains('Delete'),
        closeDeleteWindow: () => cy.get('[aria-label="Close"]'),
        typeSortTableButton: () => cy.get('[intc-id="TypesortTableButton"]'),
        nameSortTableButton: () => cy.get('[intc-id="NamesortTableButton"]'),
        previousPage: () => cy.get('[intc-id="goToPreviousPageButton"]'),
        nextPage: () => cy.get('[intc-id="goNextPageButton"]'),
        copyKeyBtn: () => cy.get('[intc-id="ButtonTable Copy key"]'),

        // ssh key creation page
        keyName: () => cy.get('[intc-id="KeyNameInput"]'),
        keyNameError: () => cy.get('[intc-id="KeyName:InvalidMessage"]'),
        createSSHDetailBar: () => cy.get('.accordion-button'),
        keyContents: () => cy.get('[intc-id="PasteyourkeycontentsTextArea"]'),
        createKey: () => cy.get('[intc-id="btn-ssh-createpublickey"]'),
        cancelKeyCreation: () => cy.get('[intc-id="btn-ssh-cancelPublicKey"]'),
        windowsOption: () => cy.get('[intc-id="WindowsRadioButton"]'),
        linuxOption: () => cy.get('[intc-id="LinuxRadioButton"]'),
        copyGenerateKeyCommand: () => cy.get(':nth-child(2) > .col-12 > .m-2'),
        copyOpenPublicKeyCommand: () => cy.get(':nth-child(4) > .col-12 > .m-2'),
        sshKeyDocumentationLink: () => cy.get('.valid-feedback').contains("/docs/guides/ssh_keys.html"),
        backToKeysButton: () => cy.get('[class="text-decoration-none"]').contains('Back to keys list'),
        closeWarningBar: () => cy.get('[aria-label="Close alert"]'),
        duplicatekeyToast: () => cy.get('.toast-container.position-body.p-3.mb-5').contains('already exists'),
        invalidkeyToast: () => cy.get('.toast-container.position-body.p-3.mb-5').eq(1).contains('SshPublicKey should have at least algorithm and publickey')  // For Training Upload Key
    }

    clickPageNumber() {
        this.elements.pageNumber().click();
    }

    clickDeleteButton() {
        this.elements.deleteKeyButton().click();
    }

    uploadKey() {
        this.elements.uploadKeyButton().click();
    }

    cancelKeyDeletion() {
        this.elements.cancelKeyDeletion().click();
    }

    confirmDeleteKey() {
        this.elements.confirmDeleteKey().click();
    }

    clickCloseDeleteWindow() {
        this.elements.closeDeleteWindow().click();
    }

    clickCopyKey() {
        this.elements.copyKeyBtn().click();
    }

    duplicateKeyToastMessage() {
        this.elements.duplicatekeyToast().should("be.visible");
    }

    invalidKeyToastMessage() {
        this.elements.invalidkeyToast().should("be.visible");
    }

    ClickTypeSortTable() {
        this.elements.typeSortTableButton().click();
    }

    clicknameSortTable() {
        this.elements.nameSortTableButton().click();
    }

    addKeyName(keyName) {
        this.elements.keyName().scrollIntoView();
        this.elements.keyName().type(keyName);
    }

    clearKeyName() {
        this.elements.keyName().clear();
    }

    clickKeyNameError() {
        this.elements.keyNameError().click();
    }

    clickCreateSSHDetailBar() {
        this.elements.createSSHDetailBar().click();
    }

    addKeyContent(keyContent) {
        this.elements.keyContents().type(keyContent);
    }

    clearKeyContent() {
        this.elements.keyContents().clear();
    }

    createKey() {
        this.elements.createKey().click();
    }

    createFirstKey() {
        this.elements.createFirstKey().click();
    }

    cancelKeyCreate() {
        this.elements.cancelKeyCreation().click();
    }

    clickWindowsOption() {
        this.elements.windowsOption().click();
    }

    clickLinuxOption() {
        this.elements.linuxOption().click();
    }

    clickCopyGenerateKeyCommand() {
        this.elements.copyGenerateKeyCommand().click();
    }

    clickCopyOpenPublicKeyCommand() {
        this.elements.copyOpenPublicKeyCommand().click();
    }

    clickSSHKeyDocLink() {
        this.elements.sshKeyDocumentationLink().click();
    }

    clickBackToKeysButton() {
        this.elements.backToKeysButton().click();
    }

    clickCloseWarningBar() {
        this.elements.closeWarningBar().click();
    }

}

module.exports = new accountKeys();