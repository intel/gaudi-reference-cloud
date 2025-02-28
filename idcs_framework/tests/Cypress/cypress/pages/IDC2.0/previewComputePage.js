class previewCompute {

    elements = {
        // Preview instance page
        requestInstanceEmpty: () => cy.get('[intc-id="RequestinstanceEmptyViewButton"]'),
        requestInstance: () => cy.get('[intc-id="btn-navigate Launch Instance"]'),
        deleteInstance: () => cy.get('[intc-id="ButtonTable Delete instance"]'),
        editInstance: () => cy.get('[intc-id="ButtonTable Edit instance"]'),
        connectBtn: () => cy.get('[intc-id="ButtonTable Connect"]'),
        extendBtn: () => cy.get('[intc-id="ButtonTable Entend instance"]'),
        instanceNameInput: () => cy.get('[intc-id="InstancenameInput"]'),
        machineImageLabel: () => cy.get('[intc-id="MachineimageInputLabel"]'),

        instanceTypeSortButton: () => cy.get('[intc-id="Instance TypesortTableButton"]'),
        reservationStartSort: () => cy.get('[intc-id="Reservation StartsortTableButton"]'),
        reservationEndSort: () => cy.get('[intc-id="Reservation EndsortTableButton"]'),
        stateSortButton: () => cy.get('[intc-id="StatesortTableButton"]'),
        ipSortButton: () => cy.get('[intc-id="IpsortTableButton"]'),
        instanceNameSortButton: () => cy.get('[intc-id="Instance NamesortTableButton"]'),
        previousPageButton: () => cy.get('[intc-id="goToPreviousPageButton"]'),
        nextPageButton: () => cy.get('[intc-id="goNextPageButton"]'),
        pageSelect: () => cy.get('[intc-id="Rowsperpage-form-select"]'),
        searchInstanceInput: () => cy.get('[intc-id="searchInstances"]'),

        // Delete reconfirm dialog
        nameInput: () => cy.get('[intc-id="NameInput"]'),
        confirmDeleteModal: () => cy.get('[intc-id="deleteConfirmModal"]'),
        cancelDelete: () => cy.get('[intc-id="btn-confirm-Deleteinstance-cancel"]').contains('Cancel'),
        confirmDelete: () => cy.get('[intc-id="btn-confirm-Deleteinstance-delete"]').contains('Delete'),

        // Request extension modal
        extensionDaysInput: () => cy.get('[intc-id="Extensionday(s)Input"]'),
        requestExtensionBtn: () => cy.get('[intc-id="btn-reservation-extension-Request"]'),
        extensionInvalidFeedback: () => cy.get('[intc-id="Extensionday(s)InvalidMessage"]'),
        closeExtendModal: () => cy.get('[aria-label="Close request extention modal"]'),

        // Instance details section
        openTerminalBtn: () => cy.get('[intc-id="btn-computereservation-open-cloud-terminal"]'),
        howToConnectButton: () => cy.get('[intc-id="btn-computereservation-how-to-connect"]'),
        actionsButton: () => cy.get('[intc-id="myReservationActionsDropdownButton"] > .dropdown-toggle'),
        connectActionsButton: () => cy.get('[intc-id="myReservationActionsDropdownItemButton0"]'),
        editActionsButton: () => cy.get('[intc-id="myReservationActionsDropdownItemButton1"]'),
        extendActionsButton: () => cy.get('[intc-id="myReservationActionsDropdownItemButton2"]'),
        deleteActionsButton: () => cy.get('[intc-id="myReservationActionsDropdownItemButton3"]'),
        closingInstanceDetailsSection: () => cy.get('[intc-id="btn-computereservation-close details"]'),
        secondaryOwnerInput: () => cy.get('[intc-id="SecondaryowneremailInput"]'),
        detailsSection: () => cy.get('[intc-id="DetailsTab"]'),
        networkSection: () => cy.get('[intc-id="NetworkingTab"]'),
        securitySection: () => cy.get('[intc-id="SecurityTab"]'),

        // How to connect popup page
        windowsOSOption: () => cy.get('[intc-id="WindowsRadioButton"]'),
        linuxOSOption: () => cy.get('[intc-id="LinuxRadioButton"]'),
        copyConfigureSSHCommandButton: () => cy.get(':nth-child(1) > .col-12 > .m-2'),
        copykeyVisibilityCommandButton: () => cy.get(':nth-child(5) > .m-2'),
        copyConnecttoInstanceCommandButton: () => cy.get(':nth-child(7) > .m-2'),
        docHowtoConnect: () => cy.get('.text-decoration-underline'),
        closeEditKeys: () => cy.get('[intc-id="UpdateInstanceSshClose"]'),
        closeHowtoConnectPageButton: () => cy.get('.modal-footer > .btn'),
        closeHowtoConnectIcon: () => cy.get('[intc-id="HowToConnectClose"]'),

        // Request instance form
        requestBtn: () => cy.get('[intc-id="btn-computelaunch-navigationBottom Request - singlenode"]'),
        cancelBtn: () => cy.get('[intc-id="btn-computelaunch-navigationBottom Cancel - singlenode"]'),
        instanceFamilyDropdown: () => cy.get('[intc-id="Instancefamily-form-select"]'),
        instanceTypeDropdown: () => cy.get('[intc-id="Instance type-form-select"]'),
        machineImageTittle: () => cy.get('[intc-id="MachineimageInputLabel"]'),
        machineImageDropdown: () => cy.get('[intc-id="Machine image-form-select"]'),
        instanceNameInput: () => cy.get('[intc-id="InstancenameInput"]'),
        intendedUsedInput: () => cy.get('[intc-id="IntendeduseTextArea"]'),
        useCase: () => cy.get('[intc-id="Usecase-form-select"]'),
        deployAIuseCase: () => cy.get('[intc-id="Usecase-form-select-option-DeployAIworkloadindevelopercloud"]'),
        hardwarePerformance: () => cy.get('[intc-id="Usecase-form-select-option-Hardwareperformanceevaluation"]'),
        gaudi3: () => cy.get('[intc-id="checkBoxTable-Instancetype-grid-pre-bm-gnr-gaudi3"]'),

        // Edit instance page
        editInstanceName: () => cy.get('[intc-id="InstancenameInput"]'),
        saveEditInstanceButton: () => cy.get('.btn.btn-primary').contains("Save"),
        uploadKeyButton: () => cy.get('[intc-id="SelectkeysbtnExtra"]'),
        selectAllKeys: () => cy.get('[intc-id="SelectkeysbtnSelectAll"]'),
        cancelEditBtn: () => cy.get('.btn.btn-link').contains("Cancel"),

        // Instance table:    
        instanceTable: () => cy.get('.table'),
        emptyTable: () => cy.get('[intc-id="data-view-empty"]'),

        // Instance configuration options
        CPU: () => cy.get('[intc-id="CPU-radio-select"]'),
        GPU: () => cy.get('[intc-id="GPU-radio-select"]'),
        AI: () => cy.get('[intc-id="AI-radio-select"]'),
        AI_PC: () => cy.get('[intc-id="AI PC-radio-select"]'),

    }

    requestInstance() {
        this.elements.requestInstanceEmpty().click({ force: true });
    }

    requestFromTable() {
        this.elements.requestInstance().click({ force: true });
    }

    instanceName(name) {
        this.elements.instanceNameInput().scrollIntoView({ duration: 1000 }).should("be.visible");
        this.elements.instanceNameInput().clear().type(name, { force: true });
    }

    cancelRequest() {
        this.elements.cancelBtn().scrollIntoView();
        this.elements.cancelBtn().click({ force: true });
    }

    clickConnect() {
        this.elements.connectBtn().click({ force: true });
    }

    clickExtend() {
        this.elements.extendBtn().click({ force: true });
    }

    machineImageDropdown() {
        this.elements.machineImageDropdown().scrollIntoView().should("be.visible");
    }

    closeExtendModal() {
        this.elements.closeExtendModal().click({ force: true });
    }

    deleteNameInput(name) {
        this.elements.nameInput().clear().type(name);
    }

    secondayOwnerInput(name) {
        this.elements.secondaryOwnerInput().scrollIntoView();
        this.elements.secondaryOwnerInput().clear().type(name);
    }

    checkConfirmDeleteModal() {
        this.elements.confirmDeleteModal().should("be.visible");
    }

    inputExtensionDays(days) {
        this.elements.extensionDaysInput().clear().type(days);
    }

    confirmRequestExtension() {
        this.elements.requestExtensionBtn().click({ force: true });
    }

    openTerminal() {
        this.elements.openTerminalBtn().click({ force: true });
    }

    searchInstance(searchInstance) {
        this.elements.searchInstanceInput().clear().type(searchInstance);
    }

    deleteInstance() {
        this.elements.deleteInstance().click({ force: true });
    }

    closeEditKeysModal() {
        this.elements.closeEditKeys().click({ force: true });
    }

    editInstance() {
        this.elements.editInstance().click({ force: true });
    }

    cancelEdit() {
        this.elements.cancelEditBtn().click({ force: true });
    }

    cancelDeleteInstance() {
        this.elements.cancelDelete().click({ force: true });
    }

    confirmDeleteInstance() {
        this.elements.confirmDelete().click({ force: true });
    }

    clickinstanceTypeSortButton() {
        this.elements.instanceTypeSortButton().click({ force: true });
    }

    selectGaudi3() {
        this.elements.gaudi3().scrollIntoView();
        this.elements.gaudi3().check();
    }

    reservationStartSort() {
        this.elements.reservationStartSort().click({ force: true });
    }

    reservationEndSort() {
        this.elements.reservationEndSort().click({ force: true });
    }

    clickstateSortButton() {
        this.elements.stateSortButton().click({ force: true });
    }

    clickipSortButton() {
        this.elements.ipSortButton().click({ force: true });
    }

    clickinstanceNameSortButton() {
        this.elements.instanceNameSortButton().click({ force: true });
    }

    clickpreviousPageButton() {
        this.elements.previousPageButton().click({ force: true });
    }

    clicknextPageButton() {
        this.elements.nextPageButton().click({ force: true });
    }

    clickpageSelect() {
        this.elements.pageSelect().click({ force: true });
    }

    clickhowToConnectButton() {
        this.elements.howToConnectButton().click({ force: true });
    }

    clickactionsButton() {
        this.elements.actionsButton().click({ force: true });
    }

    clickConnectActionsButton() {
        this.elements.connectActionsButton().click({ force: true });
    }

    clickExtendActionsButton() {
        this.elements.extendActionsButton().click({ force: true });
    }

    clickeditActionsButton() {
        this.elements.editActionsButton().click({ force: true });
    }

    clickdeleteActionsButton() {
        this.elements.deleteActionsButton().click({ force: true });
    }

    clickDetailsTab() {
        this.elements.detailsSection().click({ force: true });
    }

    clickNetworkingTab() {
        this.elements.networkSection().click({ force: true });
    }

    clickSecurityTab() {
        this.elements.securitySection().click({ force: true });
    }

    clickwindowsOSOption() {
        this.elements.windowsOSOption().click({ force: true });
    }

    clicklinuxOSOption() {
        this.elements.linuxOSOption().click({ force: true });
    }

    clickMacOSOption() {
        this.elements.MacOSOption().click({ force: true });
    }

    copyConfigureSSHCommand() {
        this.elements.copyConfigureSSHCommandButton().click({ force: true });
    }

    copykeyVisibilityCommand() {
        this.elements.copykeyVisibilityCommandButton().click({ force: true });
    }

    copyConnecttoInstanceCommand() {
        this.elements.copyConnecttoInstanceCommandButton().click({ force: true });
    }

    clickdocHowtoConnect() {
        this.elements.docHowtoConnect().click({ force: true });
    }

    clickcloseHowtoConnectPageButton() {
        this.elements.closeHowtoConnectPageButton().click({ force: true });
    }

    clickcloseHowtoConnectIcon() {
        this.elements.closeHowtoConnectIcon().click({ force: true });
    }

    machineImageLabel() {
        this.elements.machineImageTittle().scrollIntoView({ duration: 1000 });
    }

    intendedUsed(typeInput) {
        this.elements.intendedUsedInput().scrollIntoView();
        this.elements.intendedUsedInput().clear().type(typeInput);
    }

    selectDeployAI_useCase() {
        this.elements.useCase().scrollIntoView();
        this.elements.useCase().click({ force: true });
        this.elements.deployAIuseCase().click({ force: true })
    }

    selectHardwarePerformance_useCase() {
        this.elements.useCase().scrollIntoView();
        this.elements.useCase().click({ force: true });
        this.elements.hardwarePerformance().select();
    }

    editInstanceNameIsDisabled() {
        this.elements.editInstanceName().should("be.disabled");
    }

    checkInstanceTableVisible() {
        this.elements.instanceTable().should("be.visible");
    }

    checkEmptyInstanceTable() {
        this.elements.emptyTable().should("be.visible");
    }

    saveEditInstance() {
        this.elements.saveEditInstanceButton().should("be.visible");
        this.elements.saveEditInstanceButton().click({ force: true });
    }

    clickUploadKeyButton() {
        this.elements.uploadKeyButton().scrollIntoView();
        this.elements.uploadKeyButton().click({ force: true });
    }

    clickSelectAllKeys() {
        this.elements.selectAllKeys().scrollIntoView();
        this.elements.selectAllKeys().click({ force: true });
    }
}

module.exports = new previewCompute();