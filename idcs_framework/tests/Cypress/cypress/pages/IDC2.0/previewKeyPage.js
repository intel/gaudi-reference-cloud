class previewKeys {

    elements = {
        // Main preview key page
        pageNumber: () => cy.get('[intc-id="Select"]'),
        deleteKeyButton: () => cy.get('[intc-id="ButtonTable Delete"]'),
        uploadKeyButton: () => cy.get('[intc-id="btn-sshview-UploadKey"]').contains('Upload key'),
        firstKeyUploadBtn: () => cy.get('[intc-id="UploadkeyEmptyViewButton"]'),

        // Delete reconfirm dialog
        nameInput: () => cy.get('[intc-id="NameInput"]'),
        confirmDeleteModal: () => cy.get('[intc-id="deleteConfirmModal"]'),
        cancelKeyDeletion: () => cy.get('[intc-id="btn-confirm-Deletekey-cancel"]').contains('Cancel'),
        confirmDeleteKey: () => cy.get('[intc-id="btn-confirm-Deletekey-delete"]').contains('Delete'),

        closeDeleteWindow: () => cy.get('[aria-label="Close"]'),
        typeSortTableButton: () => cy.get('[intc-id="TypesortTableButton"]'),
        nameSortTableButton: () => cy.get('[intc-id="NamesortTableButton"]'),
        previousPage: () => cy.get('[intc-id="goToPreviousPageButton"]'),
        nextPage: () => cy.get('[intc-id="goNextPageButton"]'),
        searchFilter: () => cy.get('[intc-id="searchKeys"]'),
        clearFilterButton: () => cy.get('[intc-id="ClearfiltersEmptyViewButton"]'),
        keyTable: () => cy.get('[intc-id="sshPublicKeysTable"]'),
        emptyKeyTable: () => cy.get('[intc-id="data-view-empty"]'),
        // ssh key creation page
        keyName: () => cy.get('[intc-id="KeyNameInput"]'),
        associatedEmail: () => cy.get('[intc-id="AssociatedEmailInput"]'),
        keyNameError: () => cy.get('[intc-id="KeyNameInvalidMessage"]'),
        createSSHDetailBar: () => cy.get('.accordion-button'),
        keyContents: () => cy.get('[intc-id="PasteyourkeycontentsTextArea"]'),
        createKey: () => cy.get('[intc-id="btn-ssh-createpublickey"]'),
        cancelKeyCreation: () => cy.get('[intc-id="btn-ssh-cancelPublicKey"]'),
        windowsOption: () => cy.get('[intc-id="WindowsRadioButton"]'),
        linuxOption: () => cy.get('[intc-id="LinuxRadioButton"]'),
        copyGenerateKeyCommand: () => cy.get('.btn.btn-secondary').contains("Copy").eq(0),
        copyOpenPublicKeyCommand: () => cy.get('.btn.btn-secondary').contains("Copy").eq(1),
        sshKeyDocumentationLink: () => cy.get('.valid-feedback').contains("/docs/guides/ssh_keys.html"),
        closeWarningBar: () => cy.get('[aria-label="Close alert"]'),
        duplicatekeyToast: () => cy.get('.toast-container.position-body.p-3.mb-5').contains('already exists'),

        // Co-development Service Agreement
        rejectCoDevelopment: () => cy.get('[intc-id="btn-confirm-Co-developmentServices Agreement-cancel"]'),
        acceptCoDevelopment: () => cy.get('[intc-id="btn-confirm-Co-developmentServices Agreement-Accept"]'),
        coDevelopmentSwitch: () => cy.get('[intc-id="Co-developmentSwitch"]'),

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

    uploadFirstKey() {
        this.elements.firstKeyUploadBtn().click();
    }

    deleteNameInput(name) {
        this.elements.nameInput().clear().type(name);
    }

    rejectCoDevelop() {
        this.elements.rejectCoDevelopment().click();
    }

    acceptCoDevelop() {
        this.elements.acceptCoDevelopment().click();
    }

    switchCoDevelop() {
        this.elements.firstKeyUploadBtn().check();
    }
    checkConfirmDeleteModal() {
        this.elements.confirmDeleteModal().should("be.visible");
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

    clickTypeSortTable() {
        this.elements.typeSortTableButton().click();
    }

    clicknameSortTable() {
        this.elements.nameSortTableButton().click();
    }

    addKeyName(keyName) {
        this.elements.keyName().clear().type(keyName);
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

    addEmail(email) {
        this.elements.associatedEmail().type(email);
    }

    clearAssociatedEmail() {
        this.elements.associatedEmail().clear();
    }

    usePlaceholderEmail() {
        this.elements.associatedEmail().invoke('attr', 'placeholder').should('be.a', 'string').then((placeholder) => {
            this.addEmail(placeholder);
        })
    }

    setAssociatedEmail() {
        if (Cypress.env('accountType') === "Intel") {
            this.addEmail('sys_devcloudgen1@intel.com');
        } else {
            this.addEmail('premium.cy.user@mailslurp.net');
        }
    }

    createKey() {
        this.elements.createKey().click();
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

    clickCloseWarningBar() {
        this.elements.closeWarningBar().click();
    }
    clicksearchFilter(filterval) {
        this.elements.searchFilter().type(filterval);
    }

    clickClearFilterButton() {
        this.elements.clearFilterButton().click();
    }

    duplicateKeyToastMessage() {
        this.elements.duplicatekeyToast().should("be.visible");
    }

    checkVisibleKeyTable() {
        this.elements.keyTable().should("be.visible");
    }

    checkEmptyKeyTable() {
        this.elements.emptyKeyTable()
            .should("be.visible")
            .contains("No keys found");
    }

    checkEmptySearchFilter() {
        this.elements.searchFilter().should("have.value", '');
    }
}

module.exports = new previewKeys();