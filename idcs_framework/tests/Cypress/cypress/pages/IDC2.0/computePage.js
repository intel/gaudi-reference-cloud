class computePage {

    elements = {
        // Compute Console Page
        launchInstance: () => cy.get('.btn.btn-primary').contains('Launch'),
        deleteInstance: () => cy.get('[intc-id="ButtonTable Delete instance"]'),
        editInstance: () => cy.get('[intc-id="ButtonTable Edit instance"]'),
        instanceTypeSortButton: () => cy.get('[intc-id="Instance TypesortTableButton"]'),
        createdAtSortButton: () => cy.get('[intc-id="Created atsortTableButton"]'),
        stateSortButton: () => cy.get('[intc-id="StatesortTableButton"]'),
        ipSortButton: () => cy.get('[intc-id="IpsortTableButton"]'),
        instanceNameSortButton: () => cy.get('[intc-id="Instance NamesortTableButton"]'),
        previousPageButton: () => cy.get('[intc-id="goToPreviousPageButton"]'),
        nextPageButton: () => cy.get('[intc-id="goNextPageButton"]'),
        pageSelect: () => cy.get('[intc-id="Rowsperpage-form-select"]'),
        clearFilter: () => cy.get('[intc-id="ClearfiltersEmptyViewButton"]').contains('Clear filters'),
        searchInstanceInput: () => cy.get('[intc-id="searchInstances"]'),
        instanceTable: () => cy.get('.tableContainer'),

        // Delete reconfirm dialog
        nameInput: () => cy.get('[intc-id="NameInput"]'),
        confirmDeleteModal: () => cy.get('[intc-id="deleteConfirmModal"]'),
        cancelDelete: () => cy.get('[intc-id="btn-confirm-Deleteinstance-cancel"]').contains('Cancel'),
        confirmDelete: () => cy.get('[intc-id="btn-confirm-Deleteinstance-delete"]').contains('Delete'),

        // Instance details section
        howToConnectButton: () => cy.get('[intc-id="btn-computereservation-how-to-connect"]'),
        actionsButton: () => cy.get('[intc-id="myReservationActionsDropdownButton"] > .dropdown-toggle'),
        editActionsButton: () => cy.get('[intc-id="myReservationActionsDropdownItemButton0"]'),
        deleteActionsButton: () => cy.get('[intc-id="myReservationActionsDropdownItemButton1"]'),
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
        closeHowtoConnectPageButton: () => cy.get('.modal-footer > .btn'),
        closeHowtoConnectIcon: () => cy.get('.modal-header > .btn-close'),

        // Edit instance page
        editInstanceName: () => cy.get('[intc-id="InstancenameInput"]'),
    }

    editInstanceName() {
        this.elements.editInstanceName().should("be.disabled");
    }

    launchInstance() {
        this.elements.launchInstance().click({ force: true });
    }

    checkInstanceIsReady() {
        cy.wait(30000)
        cy.contains("Ready");
    }

    clearFilter() {
        this.elements.clearFilter().click({ force: true });
    }

    searchInstance(name) {
        this.elements.searchInstanceInput().should("be.visible");
        this.elements.searchInstanceInput()
            .scrollIntoView({ duration: 2000 })
            .clear()
            .type(name);
    }

    instanceTableIsVisible() {
        this.elements.instanceTable().should("be.visible");
    }

    deleteInstance() {
        this.elements.deleteInstance().click({ force: true });
    }

    editInstance() {
        this.elements.editInstance().click({ force: true });
    }

    deleteNameInput(name) {
        this.elements.nameInput().clear().type(name);
    }

    checkConfirmDeleteModal() {
        this.elements.confirmDeleteModal().should("be.visible");
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

    clickcreatedAtSortButton() {
        this.elements.createdAtSortButton().click({ force: true });
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

    clickeditActionsButton() {
        this.elements.editActionsButton().click({ force: true });
    }

    clickdeleteActionsButton() {
        this.elements.deleteActionsButton().click({ force: true });
    }

    clickdetailsSection() {
        this.elements.detailsSection().click({ force: true });
    }

    clicknetworkSection() {
        this.elements.networkSection().click({ force: true });
    }

    clicksecuritySection() {
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
}

module.exports = new computePage();