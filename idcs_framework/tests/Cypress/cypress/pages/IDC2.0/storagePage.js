class storagePage {

    elements = {

        // My storage volumes home
        storageConsoleTittle: () => cy.get('[intc-id="storageMyReservationsTitle"]'),
        createVolumeBtn: () => cy.get('[intc-id="btn-navigate-create-volume"]').contains("Create volume"),
        createFirstVolume: () => cy.get('[intc-id="CreatevolumeEmptyViewButton"]'),
        searchVolume: () => cy.get('[intc-id="searchVolumes"]'),
        deleteBtn: () => cy.get('[intc-id="ButtonTable Delete storage"]'),
        editBtn: () => cy.get('[intc-id="ButtonTable Edit Storage"]'),
        clearFilter: () => cy.get('[intc-id="ClearfiltersEmptyViewButton"]'),
        volumeTable: () => cy.get('.tableContainer'),

        // Create Storage volume form
        volumeNameInput: () => cy.get('[intc-id="NameInput"]'),
        inputStorageSize: () => cy.get('[intc-id="StorageSize(TB)Input"]'),
        launchVolume: () => cy.get('[intc-id="btn-storagelaunch-navigationBottom Create"]'),
        cancelVolume: () => cy.get('[intc-id="btn-storagelaunch-navigationBottom Cancel"]'),
        searchVolumeInput: () => cy.get('[intc-id="searchVolumes"]'),

        // Edit volume
        cancelEdit: () => cy.get('[intc-id="btn-StorageEdit-navigationBottom Cancel"]'),
        saveEdit: () => cy.get('[intc-id="btn-StorageEdit-navigationBottom Edit"]'),

        // Delete reconfirm dialog
        nameInput: () => cy.get('[intc-id="NameInput"]'),
        confirmDeleteModal: () => cy.get('[intc-id="deleteConfirmModal"]'),
        cancelDelete: () => cy.get('[intc-id="btn-confirm-Deletestorage-cancel"]').contains('Cancel'),
        confirmDelete: () => cy.get('[intc-id="btn-confirm-Deletestorage-delete"]').contains('Delete'),

        // My volume details section
        howToMountButton: () => cy.get('[intc-id="btn how-to-connect"]').contains('How to mount'),
        howToUnMountButton: () => cy.get('[intc-id="btn how-to-connect"]').contains('How to unmount'),
        editActionsButton: () => cy.get('[intc-id="myReservationActionsDropdownItemButton0"]'),
        deleteActionsButton: () => cy.get('[intc-id="myReservationActionsDropdownItemButton1"]'),
        detailsSection: () => cy.get('[intc-id="DetailsTab"]'),
        securitySection: () => cy.get('[intc-id="SecurityTab"]'),
        volumeUser: () => cy.get('.d-flex.align-items-center.text-wrap.gap-s4'),
        copyUser: () => cy.get('[aria-label="Copy User"]'),
        copyPasswd: () => cy.get('[aria-label="Copy Password"]'),
        generatePwd: () => cy.get('.btn.btn-outline-primary.btn-sm').contains("Generate password"),

        // How to mount modal
        singleInstanceTab: () => cy.get('[intc-id="nav-link.small.tap-inactive"]').contains('Instance group'),
        singleInstanceSelect: () => cy.get('[intc-id="Instance-form-select"]'),
        instanceGroupTab: () => cy.get('[intc-id="nav-link.small.tap-inactive"]').contains('Instance group'),
        instanceGroupSelect: () => cy.get('[intc-id="InstanceGroup-form-select"]'),
        howToMountClose: () => cy.get('[intc-id="HowToConnectClose"]'),
        howToConnectInstance: () => cy.get('[intc-id="howToConnectInstanceAccordion"]'),
        howToConnectStorage: () => cy.get('[intc-id="howToConnectStorageAccordion"]'),
    }

    clickStorageConsoleTittle() {
        this.elements.storageConsoleTittle().click();
    }

    createVolume() {
        this.elements.createVolumeBtn().click({ force: true });
    }

    volumeTableIsVisible() {
        this.elements.volumeTable().should('be.visible');
    }

    createFirstVolume() {
        this.elements.createFirstVolume().click({ force: true });
    }

    launchVolumeIsEnabled() {
        this.elements.launchVolume().should('be.enabled');
    }

    clearFilter() {
        this.elements.clearFilter().click();
    }

    cancelVolumeEdit() {
        this.elements.clearFilter().click();
    }

    saveVolumeEdit() {
        this.elements.saveEdit().click();
    }

    searchVolume(name) {
        this.elements.searchVolumeInput().clear().type(name);
    }

    clickStorageConsoleTittle() {
        this.elements.storageConsoleTittle().click();
    }

    deleteBtn() {
        this.elements.deleteBtn().click({ force: true });
    }

    editVolume() {
        this.elements.editBtn().click({ force: true });
    }

    volumeNameInput(name) {
        this.elements.volumeNameInput().clear().type(name);
    }

    clearVolumeName() {
        this.elements.volumeNameInput().clear();
    }

    clearVolumeSize() {
        this.elements.inputStorageSize().clear();
    }

    inputStorageSize(size) {
        this.elements.inputStorageSize().type(size);
    }

    launchVolume() {
        this.elements.launchVolume().should("be.visible").click({ force: true });
    }

    cancelVolume() {
        this.elements.cancelVolume().should("be.visible").click();
    }

    cancelDelete() {
        this.elements.cancelDelete().click({ force: true });
    }

    confirmDelete() {
        this.elements.confirmDelete().click({ force: true });
    }

    howToMountButton() {
        this.elements.howToMountButton().click();
    }

    howToUnMountButton() {
        this.elements.howToUnMountButton().click();
    }

    detailsSection() {
        this.elements.detailsSection().click();
    }

    editActionsButton() {
        this.elements.editActionsButton().click();
    }

    deleteActionsButton() {
        this.elements.deleteActionsButton().click();
    }

    deleteNameInput(name) {
        this.elements.nameInput().clear().type(name);
    }

    checkConfirmDeleteModal() {
        this.elements.confirmDeleteModal().should("be.visible");
    }

    securitySection() {
        this.elements.securitySection().click();
    }

    getVolumeUser() {
        this.elements.volumeUser().click();
    }

    copyUser() {
        this.elements.copyUser().click();
    }

    generatePassword() {
        this.elements.generatePwd().click();
    }

    copyPasswd() {
        this.elements.copyPasswd().click();
    }

    singleInstanceTab() {
        this.elements.singleInstanceTab().click();
    }

    singleInstanceSelect() {
        this.elements.singleInstanceSelect().click({ force: true });
    }

    instanceGroupTab() {
        this.elements.instanceGroupTab().click();
    }

    instanceGroupSelect() {
        this.elements.instanceGroupSelect().click();
    }

    howToMountClose() {
        this.elements.howToMountClose().click();
    }

    howToConnectInstance() {
        this.elements.howToConnectInstance().click();
    }

    howToConnectStorage() {
        this.elements.howToConnectStorage().click();
    }
}

module.exports = new storagePage();