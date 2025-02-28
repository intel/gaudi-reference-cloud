class storagePrincipals {
    elements = {
        // Main Page
        createPrincipalButton: () => cy.get('[intc-id="btn-navigate-Create-users"]').contains("Create principal"),
        firstPrincipal: () => cy.get('[intc-id="CreateprincipalEmptyViewButton"]'),
        viewBucketsButton: () => cy.get('[intc-id="btn-navigate-Manage-users-and-permissions"]'),
        emptyPrincipalTable: () => cy.get('[intc-id="data-view-empty"]'),
        principalTable: () => cy.get('.table'),
        principalName: () => cy.get('.text-decoration-underline.btn.btn-link'),
        editPrincipalTableButton: () => cy.get('[intc-id="ButtonTable Edit principal"]'),
        deletePrincipalTableButton: () => cy.get('[intc-id="ButtonTable Delete principal"]'),
        confirmDeleteModal: () => cy.get('[intc-id="deleteConfirmModal"]'),
        confirmDeleteButton: () => cy.get('[intc-id="btn-confirm-Deleteprincipal-delete"]'),
        cancelDeleteButton: () => cy.get('[intc-id="btn-confirm-Deleteprincipal-cancel"]'),
        searchPrincipalInput: () => cy.get('[intc-id="searchUsers"]'),
        //  Create Principal Form
        nameInput: () => cy.get('[intc-id="NameInput"]'),
        invalidNameMessage: () => cy.get('[intc-id="NameInvalidMessage"]'),
        applyPermissionsForAllBuckets: () => cy.get('[intc-id="Applypermissions-Radio-option-Forallbuckets"]'),
        applyPermissionsPerBucket: () => cy.get('[intc-id="Applypermissions-Radio-option-Perbucket"]'),
        // Allowed Actions
        selectAllActionsCheckbox: () => cy.get('[intc-id="Allowedactions-Input-option-selectAll"]'),
        getBucketLocationCheckbox: () => cy.get('[intc-id="Allowedactions-Input-option-GetBucketLocation"]'),
        getBucketPolicyCheckbox: () => cy.get('[intc-id="Allowedactions-Input-option-GetBucketPolicy"]'),
        listBucketCheckbox: () => cy.get('[intc-id="Allowedactions-Input-option-ListBucket"]'),
        listBucketMultipartUploadsCheckbox: () => cy.get('[intc-id="Allowedactions-Input-option-ListBucketMultipartUploads"]'),
        listMultipartUploadPartsCheckbox: () => cy.get('[intc-id="Allowedactions-Input-option-ListMultipartUploadParts"]'),
        getBucketTaggingCheckbox: () => cy.get('[intc-id="Allowedactions-Input-option-GetBucketTagging"]'),
        // Allowed policies
        selectAllPoliciesCheckbox: () => cy.get('[intc-id="Allowedpolicies-Input-option-selectAll"]'),
        readAllowedCheckbox: () => cy.get('[intc-id="Allowedpolicies-Input-option-Read"]'),
        writeAllowedCheckbox: () => cy.get('[intc-id="Allowedpolicies-Input-option-Write"]'),
        deleteAllowedCheckbox: () => cy.get('[intc-id="Allowedpolicies-Input-option-Delete"]'),
        allowPoliciesPath: () => cy.get('[intc-id="AllowedpoliciespathInput"]'),

        submitCreatePrincipalBtn: () => cy.get('[intc-id="btn-objectStorageUsersLaunch-navigationBottom Create"]'),
        cancelCreatePrincipalBtn: () => cy.get('[intc-id="btn-objectStorageUsersLaunch-navigationBottom Cancel"]'),

        // Permissions tab
        permissionsTab: () => cy.get('[intc-id="PermissionsTab"]'),
        actionsDropdown: () => cy.get('[intc-id="myUsersActionsDropdownButton"]'),
        editActionsButton: () => cy.get('[intc-id="myUsersActionsDropdownItemButton0"]'),
        credentialsTab: () => cy.get('[intc-id="CredentialsTab"]'),

        // Edit principal form
        saveEditButton: () => cy.get('[intc-id="btn-objectStorageUserEdit-navigationBottom Save"]'),
        cancelEditButton: () => cy.get('[intc-id="btn-objectStorageUserEdit-navigationBottom Cancel"]')
    }

    createPrincipal() {
        this.elements.createPrincipalButton().click({ force: true });
    }

    cancelCreate() {
        this.elements.cancelCreatePrincipalBtn().click({ force: true });
    }

    firstPrincipal() {
        this.elements.firstPrincipal().click({ force: true });
    }

    searchPrincipal(name) {
        this.elements.searchPrincipalInput().clear().type(name);
    }

    enterPrincipalName(name) {
        this.elements.nameInput().clear().type(name);
    }

    principalName(name) {
        this.elements.principalName().should("be.visible").contains(name);
    }

    applyPermissionsForAllBucketsByDefault() {
        this.elements.applyPermissionsForAllBuckets().should("be.checked");
    }

    applyPermissionsForAllBuckets() {
        this.elements.applyPermissionsForAllBuckets().check();
    }

    applyPermissionsPerBucket() {
        this.elements.applyPermissionsPerBucket().check();
    }

    applySelectAllActions() {
        this.elements.selectAllActionsCheckbox().check();
    }

    applyGetBucketLocation() {
        this.elements.getBucketLocationCheckbox().check();
    }

    applyGetBucketPolicy() {
        this.elements.getBucketPolicyCheckbox().check();
    }

    applyListBucket() {
        this.elements.listBucketCheckbox().check();
    }

    applyListBucketMultipartUploads() {
        this.elements.listBucketMultipartUploadsCheckbox().check();
    }

    applyListMultipartUploadParts() {
        this.elements.listMultipartUploadPartsCheckbox().check();
    }

    applyGetBucketTagging() {
        this.elements.getBucketTaggingCheckbox().check();
    }

    applySelectAllPolicies() {
        this.elements.selectAllPoliciesCheckbox().check();
    }

    applyReadPolicy() {
        this.elements.readAllowedCheckbox().check();
    }

    applyWritePolicy() {
        this.elements.writeAllowedCheckbox().check();
    }

    applyDeletePolicy() {
        this.elements.deleteAllowedCheckbox().check();
    }

    uncheckListMultipartUploadParts() {
        this.elements.listMultipartUploadPartsCheckbox().uncheck();
    }

    uncheckGetBucketPolicy() {
        this.elements.getBucketPolicyCheckbox().uncheck();
    }

    uncheckGetBucketLocation() {
        this.elements.getBucketLocationCheckbox().uncheck();
    }

    uncheckGetBucketTagging() {
        this.elements.getBucketTaggingCheckbox().uncheck();
    }

    uncheckReadPolicy() {
        this.elements.readAllowedCheckbox().uncheck();
    }

    uncheckWritePolicy() {
        this.elements.writeAllowedCheckbox().uncheck();
    }

    inputAllowedPolicyPath(path) {
        this.elements.allowPoliciesPath().clear().type(path);
    }

    checkEmptyPrincipalTable() {
        this.elements.emptyPrincipalTable().should("be.visible");
    }

    checkPrincipalTableVisible() {
        this.elements.principalTable().should("be.visible");
    }

    submitCreatePrincipal() {
        this.elements.submitCreatePrincipalBtn().click();
    }

    clickPermissionsTab() {
        this.elements.permissionsTab().should("be.visible").click({ force: true });
    }

    clickCredntialsTab() {
        this.elements.credentialsTab().should("be.visible").click({ force: true });
    }

    clickActionsDropdown() {
        this.elements.actionsDropdown().click();
    }

    clickEditActionsButton() {
        this.elements.editActionsButton().click();
    }

    editPrincipalFromTable(idx) {
        this.elements.editPrincipalTableButton().eq(idx).click();
    }

    saveEdit() {
        this.elements.saveEditButton().click({ force: true });
    }

    cancelEditButton() {
        this.elements.cancelEditButton().click({ force: true });
    }

    checkSaveEditButtonDisabled() {
        this.elements.saveEditButton().should("be.disabled");
    }

    checkInvalidNameMessageVisible() {
        this.elements.invalidNameMessage().should("be.visible");
    }

    getPrincipalNameInput() {
        return this.elements.nameInput();
    }

    viewBuckets() {
        this.elements.viewBucketsButton().click();
    }

    deleteNameInput(name) {
        this.elements.nameInput().clear().type(name);
    }

    checkConfirmDeleteModal() {
        this.elements.confirmDeleteModal().should("be.visible");
    }

    confirmDeletePrincipal() {
        this.elements.confirmDeleteButton().click();
    }

    deletePrincipals() {
        this.elements.deletePrincipalTableButton().then($items => {
            const remainingItems = $items.length;
            $items[0].click();
            cy.get(".modal").contains("Delete principal").should("be.visible");
            cy.getValue('[intc-id="deleteConfirmationName"]').then(text => {
                cy.get('[intc-id="NameInput"]').should("be.visible").clear().type(text);
            })
            this.confirmDeletePrincipal();
            cy.wait(4000);
            if (remainingItems > 1) {
                this.checkPrincipalTableVisible();
                this.deletePrincipals();
            } else {
                this.checkEmptyPrincipalTable();
            }
        });
    }
}

module.exports = new storagePrincipals();