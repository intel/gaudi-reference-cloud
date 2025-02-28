class previewStorage {

    elements = {
        // Preview Storage page
        requestBucketBtn: () => cy.get('[intc-id="RequestBucketEmptyViewButton"]'),
        bucketSizeDropDown: () => cy.get('[intc-id="Bucketsize-form-select"]'),
        select10GB: () => cy.get('[intc-id="Bucketsize-form-select-option-10GB"]'),
        select20GB: () => cy.get('[intc-id="Bucketsize-form-select-option-20GB"]'),
        select50GB: () => cy.get('[intc-id="Bucketsize-form-select-option-50GB"]'),
        select100GB: () => cy.get('[intc-id="Bucketsize-form-select-option-100GB"]'),

        requestBtn: () => cy.get('[intc-id="btn-PreviewObjectStorage-navigationBottom Request"]'),
        cancelRequestBtn: () => cy.get('[intc-id="btn-PreviewObjectStorage-navigationBottom Cancel"]'),

        howToUse: () => cy.get('[intc-id="btn-preview-show-how-to-use-bucket"]'),
        editBucket: () => cy.get('[intc-id="ButtonTable Edit bucket"]'),
        deleteBucket: () => cy.get('[intc-id="ButtonTable Delete object storage"]'),
        extendBucket: () => cy.get('[intc-id="ButtonTable Extend bucket"]'),

        // Delete reconfirm dialog
        nameInput: () => cy.get('[intc-id="NameInput"]'),
        confirmDeleteModal: () => cy.get('[intc-id="deleteConfirmModal"]'),
        cancelDelete: () => cy.get('[intc-id="btn-confirm-Deleteobject storage-cancel"]').contains('Cancel'),
        confirmDelete: () => cy.get('[intc-id="btn-confirm-Deleteobject storage-delete"]').contains('Delete'),

        // Extend reconfirm dialog
        cancelExtend: () => cy.get('[intc-id="btn-confirm-Extendstorage reservation-cancel"]').contains('Cancel'),
        confirmExtend: () => cy.get('[intc-id="btn-confirm-Extendstorage reservation-Request"]').contains('Request'),

        // Request extension modal
        requestExtensionBtn: () => cy.get('[intc-id="btn-reservation-extension-Request"]'),
        extensionInvalidFeedback: () => cy.get('[intc-id="Extensionday(s)InvalidMessage"]'),

        // How to use page
        generateKey: () => cy.get('[intc-id="PreviewObjectStorageGenerateKeySubheader"]'),
        cancelGenerateKey: () => cy.get('[intc-id="btn-confirm-Warning!-cancel"]'),
        confirmGenerateKey: () => cy.get('[intc-id="btn-confirm-Warning!-Generate"]'),

        // Instance table:    
        bucketTable: () => cy.get('.table'),
    }

    requestStorage() {
        this.elements.requestBucketBtn().click({ force: true });
    }

    cancelRequest() {
        this.elements.cancelRequestBtn().click({ force: true });
    }

    requestExtendFromTable() {
        this.elements.extendBucket().click({ force: true });
    }

    checkStorageTableIsVisible() {
        this.elements.bucketTable().should("be.visible");
    }

    selectSize10GB() {
        this.elements.bucketSizeDropDown().click({ force: true });
        this.elements.select10GB().click({ force: true });
    }

    selectSize20GB() {
        this.elements.bucketSizeDropDown().click({ force: true });
        this.elements.select20GB().click({ force: true });
    }

    selectSize50GB() {
        this.elements.bucketSizeDropDown().click({ force: true });
        this.elements.select50GB().click({ force: true });
    }

    selectSize100GB() {
        this.elements.bucketSizeDropDown().click({ force: true });
        this.elements.select100GB().click({ force: true });
    }

    requestBucket() {
        this.elements.requestBtn().click({ force: true });
    }

    clickHowToUse() {
        this.elements.howToUse().click({ force: true });
    }

    deleteBucketFromTable() {
        this.elements.deleteBucket().click({ force: true });
    }

    editBucketFromTable() {
        this.elements.editBucket().click({ force: true });
    }

    cancelDelete() {
        this.elements.cancelDelete().click({ force: true });
    }

    confirmDelete() {
        this.elements.confirmDelete().click({ force: true });
    }

    deleteNameInput(name) {
        this.elements.nameInput().clear().type(name);
    }

    checkConfirmDeleteModal() {
        this.elements.confirmDeleteModal().should("be.visible");
    }

    requestExtend() {
        this.elements.requestExtensionBtn().click({ force: true });
    }

    confirmExtend() {
        this.elements.confirmExtend().click({ force: true });
    }

    cancelExtend() {
        this.elements.cancelExtend().click({ force: true });
    }

    clickGenerateKey() {
        this.elements.generateKey().scrollIntoView();
        this.elements.generateKey().click({ force: true });
    }

    cancelGenerateKey() {
        this.elements.cancelGenerateKey().click({ force: true });
    }

    confirmGenerateKey() {
        this.elements.confirmGenerateKey().click({ force: true });
    }
}

module.exports = new previewStorage();