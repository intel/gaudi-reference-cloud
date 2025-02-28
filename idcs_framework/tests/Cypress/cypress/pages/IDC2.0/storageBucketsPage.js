class storageBuckets {
    elements = {
        // Compute Console Page
        firstBucketBtn: () => cy.get('[intc-id="CreatebucketEmptyViewButton"]').contains("Create bucket"),
        createBucket: () => cy.get('[intc-id="btn-navigate-Create-bucket"]').contains("Create bucket"),
        managePrincipalButton: () => cy.get('[intc-id="btn-navigate-Manage-users-and-permissions"]'),

        // Create Bucket Form
        bucketNamePrepend: () => cy.get('[intc-id="StorageBucketNamePrependMessage"]'),
        bucketNameInput: () => cy.get('[intc-id="NameInput"]'),
        bucketDescription: () => cy.get('[intc-id="DescriptionTextArea"]'),
        enableVersioningCheckbox: () => cy.get('[intc-id="Checkbox"]'),
        submitBucketButton: () => cy.get('[intc-id="btn-ObjectStorage-navigationBottom Create"]'),
        cancelBucketCreation: () => cy.get('[intc-id="btn-ObjectStorage-navigationBottom Cancel"]'),
        invalidBucketNameMessage: () => cy.get('[intc-id="NameInvalidMessage"]'),
        // Bucket table
        bucketTable: () => cy.get('.tableContainer'),
        bucketName: () => cy.get('.text-decoration-underline.btn.btn-link'),
        emptyTable: () => cy.get('[intc-id="data-view-empty"]'),
        searchBucketInput: () => cy.get('[intc-id="searchBuckets"]'),
        deleteBucketButton: () => cy.get('[intc-id="ButtonTable Delete bucket"]'),
        clearFilter: () => cy.get('[intc-id="ClearfiltersEmptyViewButton"]'),

        // Details tab
        detailsTab: () => cy.get('[intc-id="DetailsTab"]'),
        actionsDropdown: () => cy.get('.dropdown-toggle').contains("Actions"),
        deleteFromActions: () => cy.get('[intc-id="myReservationActionsDropdownItemButton0"]'),
        // Principal tab
        principalTab: () => cy.get('[intc-id="PrincipalsTab"]'),
        linkToPrincipalPage: () => cy.get('[intc-id="btn-navigate-Manage-users-and-permissions"]'),
        // Delete reconfirm dialog
        nameInput: () => cy.get('[intc-id="NameInput"]'),
        confirmDeleteModal: () => cy.get('[intc-id="deleteConfirmModal"]'),
        confirmDeleteButton: () => cy.get('[intc-id="btn-confirm-Deletebucket-delete"]'),
        cancelDeleteButton: () => cy.get('[intc-id="btn-confirm-Deletebucket-cancel"]'),
        deleteBucketPrincipalConfirmButton: () => cy.get('[intc-id="btn-confirm-Deleted bucket principals-Ok"]'),
    }

    createBucket() {
        this.elements.createBucket().click({ force: true });
    }

    createFirstBucket() {
        this.elements.firstBucketBtn().click({ force: true });
    }

    submitBucket() {
        this.elements.submitBucketButton().should('be.visible');
        this.elements.submitBucketButton().click();
    }

    cancelBucket() {
        this.elements.cancelBucketCreation().should('be.visible');
        this.elements.cancelBucketCreation().click();
    }

    enterBucketName(name) {
        this.elements.bucketNameInput().clear().type(name);
    }

    enterBucketDescription(name) {
        this.elements.bucketDescription().clear().type(name);
    }

    enableVersioning() {
        this.elements.enableVersioningCheckbox().check()
    }

    disableVersioning() {
        this.elements.enableVersioningCheckbox().uncheck();
    }

    checkBucketTableVisible() {
        this.elements.bucketTable().should("be.visible");
    }

    clearFilter() {
        this.elements.clearFilter().click({ force: true });
    }

    checkEmptyBucketTable() {
        this.elements.emptyTable().should("be.visible");
    }

    findBucket(name) {
        this.elements.bucketName().contains(name).should("be.visible").click({ force: true });
    }

    checkInvalidBucketNameMessage() {
        this.elements.invalidBucketNameMessage().should("be.visible");
        cy.contains("Only lower case alphanumeric and hypen(-) allowed for Name:.");
    }

    getBucketNamePrepend() {
        this.elements.bucketNamePrepend().should("not.have.value", null);
        return this.elements.bucketNamePrepend();
    }

    searchBucket(value) {
        this.elements.searchBucketInput().should("be.visible");
        this.elements.searchBucketInput().clear().type(value);
    }

    deleteBucket() {
        this.elements.deleteBucketButton().first().click();
    }

    confirmDeleteBucket() {
        this.elements.confirmDeleteButton().click({ force: true });
    }

    getBucketNameInput() {
        return this.elements.bucketNameInput();
    }

    deleteBucketPrincipalConfirm() {
        this.elements.deleteBucketPrincipalConfirmButton().click({ force: true });
    }

    clickActionsButton() {
        this.elements.actionsDropdown().click()
    }

    deleteFromActions() {
        this.elements.deleteFromActions().click();
    }

    deleteNameInput(name) {
        this.elements.nameInput().clear().type(name);
    }

    checkConfirmDeleteModal() {
        this.elements.confirmDeleteModal().should("be.visible");
    }

    clickPrincipalTab() {
        this.elements.principalTab().click();
    }

    clickDetailsTab() {
        this.elements.detailsTab().click();
    }

    clickLinkToPrincipalPage() {
        this.elements.linkToPrincipalPage().click();
    }

    clickManagePrincipalButton() {
        this.elements.managePrincipalButton().click();
    }
    deleteAllBuckets() {
        this.elements.deleteBucketButton().then($items => {
            const remainingItems = $items.length;
            $items[0].click();
            this.checkConfirmDeleteModal();
            cy.getValue('[intc-id="deleteConfirmationName"]').then(text => {
                cy.get('[intc-id="NameInput"]').should("be.visible").clear().type(text);
            })
            this.confirmDeleteBucket();
            cy.wait(4000);
            this.deleteBucketPrincipalConfirm();
            if (remainingItems > 1) {
                this.checkBucketTableVisible();
                this.deleteAllBuckets();
            } else {
                this.checkEmptyBucketTable();
            }
        });
    }
}

module.exports = new storageBuckets();