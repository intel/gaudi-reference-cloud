class computeGroupPage {

    elements = {
        launchInstanceGroup: () => cy.get('[intc-id="btn-navigate Launch Instance"]'),
        deleteInstanceGroup: () => cy.get('[intc-id="ButtonTable Delete instance group"]'),
        instanceTypeSortButton: () => cy.get('[intc-id="TypesortTableButton"]'),
        createdAtSortButton: () => cy.get('[intc-id="Created atsortTableButton"]'),
        stateSortButton: () => cy.get('[intc-id="StatesortTableButton"]'),
        ipSortButton: () => cy.get('[intc-id="IpsortTableButton"]'),
        instanceNameSortButton: () => cy.get('[intc-id="Instance Group NamesortTableButton"]'),
        previousPageButton: () => cy.get('.pagination > :nth-child(1)'),
        nextPageButton: () => cy.get('.pagination > :nth-child(3)'),
        pageSelect: () => cy.get('[intc-id="Select"]'),
        computeConsoleH1: () => cy.get('[intc-id="computeMyReservationsTitle"]').contains('Compute Console'),
        searchInstanceInput: () => cy.get('[intc-id="searchInstances"]'),
        cancelSearchInstanceButton: () => cy.get('[class="btn border-start-0 border cursor-default"]'),

        // Delete reconfirm dialog    
        nameInput: () => cy.get('[intc-id="NameInput"]'),
        confirmDeleteModal: () => cy.get('[intc-id="deleteConfirmModal"]'),
        cancelDelete: () => cy.get('.btn.btn-secondary').contains('Cancel'),
        confirmDelete: () => cy.get('.btn-danger').contains('Delete'),

        // Instance group details section
        howToConnectButton: () => cy.get('[intc-id="ButtonTable Connect via SSH"]'),
        actionsButton: () => cy.get('[intc-id="myReservationActionsDropdownButton"] > .dropdown-toggle'),
        deleteActionsButton: () => cy.get('[intc-id="myReservationActionsDropdownItemButton0"]'),
        closingInstanceDetailsSection: () => cy.get('[intc-id="btn-compute-groups-close-details"]'),
        detailsSection: () => cy.get('.m-0 > .nav > :nth-child(1) > .nav-link'),
        networkSection: () => cy.get('.m-0 > .nav > :nth-child(2) > .nav-link'),
        securitySection: () => cy.get('.m-0 > .nav > :nth-child(3) > .nav-link'),

        // How to connect popup page
        windowsOSOption: () => cy.get('[value="Windows"]'),
        linuxOSOption: () => cy.get('[value="linux"]'),
        MacOSOption: () => cy.get('[value="macOS"]'),
        copyConfigureSSHCommandButton: () => cy.get(':nth-child(1) > .col-12 > .m-2'),
        copykeyVisibilityCommandButton: () => cy.get(':nth-child(5) > .m-2'),
        copyConnecttoInstanceCommandButton: () => cy.get(':nth-child(7) > .m-2'),
        docHowtoConnect: () => cy.get('.text-decoration-underline'),
        closeHowtoConnectPageButton: () => cy.get('.modal-footer > .btn'),
        closeHowtoConnectIcon: () => cy.get('.modal-header > .btn-close'),
    }

    launchInstanceGroup() {
        this.elements.launchInstanceGroup().click({ force: true });
    }

    searchInstance(searchInstance) {
        this.elements.searchInstanceInput().type(searchInstance);
    }

    cancelSearchInstance() {
        this.elements.cancelSearchInstanceButton().click({ force: true });
    }

    deleteInstanceGroup() {
        this.elements.deleteInstanceGroup().click({ force: true });
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

    deleteNameInput(name) {
        this.elements.nameInput().clear().type(name);
    }

    checkConfirmDeleteModal() {
        this.elements.confirmDeleteModal().should("be.visible");
    }

    clickclosingInstanceDetailsSection() {
        this.elements.closingInstanceDetailsSection().click({ force: true });
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

module.exports = new computeGroupPage();