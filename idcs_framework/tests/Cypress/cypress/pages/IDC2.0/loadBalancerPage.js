class loadBalancerPage {

    elements = {
        // Load balancer table
        launchLB: () => cy.get('[intc-id="btn-navigate-Create-load-balancer"]'),
        deleteLB: () => cy.get('[intc-id="ButtonTable Delete load balancer"]'),
        editLB: () => cy.get('[intc-id="ButtonTable Edit balancer"]'),
        createdAtSortButton: () => cy.get('[intc-id="Created atsortTableButton"]'),
        stateSortButton: () => cy.get('[intc-id="StatesortTableButton"]'),
        ipSortButton: () => cy.get('[intc-id="Virtual IPsortTableButton"]'),
        nameSortButton: () => cy.get('[intc-id="NamesortTableButton"]'),
        previousPageButton: () => cy.get('[intc-id="goToPreviousPageButton"]'),
        nextPageButton: () => cy.get('[intc-id="goNextPageButton"]'),
        pageSelect: () => cy.get('[intc-id="Rowsperpage-form-select"]'),
        clearFilter: () => cy.get('[intc-id="ClearfiltersEmptyViewButton"]').contains('Clear filters'),
        searchBalancerInput: () => cy.get('[intc-id="searchLoadBalancers"]'),
        lbTable: () => cy.get('.tableContainer'),

        // Delete reconfirm dialog
        nameInput: () => cy.get('[intc-id="NameInput"]'),
        confirmDeleteModal: () => cy.get('[intc-id="deleteConfirmModal"]'),
        cancelDelete: () => cy.get('[intc-id="btn-confirm-Deleteload balancer-cancel"]').contains('Cancel'),
        confirmDelete: () => cy.get('[intc-id="btn-confirm-Deleteload balancer-delete"]').contains('Delete'),

        // Launch Load balancer form
        nameInput: () => cy.get('[intc-id="NameInput"]'),
        sourceIp: () => cy.get('[intc-id="SourceIPInput"]'),
        addSourceIp: () => cy.get('[intc-id="btn-loadBalancer-addSourceIP"]'),
        deleteSourceIp: () => cy.get('[intc-id="btn-loadBalancer-delete-source-ip"]'),
        listenerPort: () => cy.get('[intc-id="ListenerPortInput"]'),
        instancePort: () => cy.get('[intc-id="InstancePortInput"]'),
        monitorType: () => cy.get('[intc-id="Monitortype-form-select"]'),
        selectTCP: () => cy.get('[intc-id="Monitortype-form-select-option-TCP"]'),
        selectHTTP: () => cy.get('[intc-id="Monitortype-form-select-option-HTTP"]'),
        selectHTTPS: () => cy.get('[intc-id="Monitortype-form-select-option-HTTPS"]'),
        addListener: () => cy.get('[intc-id="btn-loadBalancer-addListener"]'),
        deleteListener: () => cy.get('[intc-id="btn-loadBalancer-delete-listener"]'),
        selectAll: () => cy.get('[intc-id="Instances-Input-option-selectAll"]'),

        launchBtn: () => cy.get('[intc-id="btn-LoadBalancer-navigationBottomLaunch"]'),
        cancelBtn: () => cy.get('[intc-id="btn-LoadBalancer-navigationBottomCancel"]'),
        saveEditBtn: () => cy.get('[intc-id="btn-LoadBalancer-navigationBottomSave"]'),

        // Load balancer details section
        actionsButton: () => cy.get('[intc-id="loadBalancerReservationActionsDropdownButton"] > .dropdown-toggle'),
        editActionsButton: () => cy.get('[intc-id="loadBalancerReservationActionsDropdownButton0"]'),
        deleteActionsButton: () => cy.get('[intc-id="loadBalancerReservationActionsDropdownButton1"]').contains("Delete"),
        detailsTab: () => cy.get('[intc-id="DetailsTab"]'),
        sourceTab: () => cy.get('[intc-id="Source IPsTab"]'),
        listenersTab: () => cy.get('[intc-id="ListenersTab"]'),
    }

    launchLBfromGrid() {
        this.elements.launchLB().click({ force: true });
    }

    clearFilter() {
        this.elements.clearFilter().click({ force: true });
    }

    searchBalancer(name) {
        this.elements.searchBalancerInput().should("be.visible");
        this.elements.searchBalancerInput().clear().type(name);
    }

    lbTableIsVisible() {
        this.elements.lbTable().should("be.visible");
    }

    deleteLb() {
        this.elements.deleteLB().click({ force: true });
    }

    deleteNameInput(name) {
        this.elements.nameInput().clear().type(name);
    }

    checkConfirmDeleteModal() {
        this.elements.confirmDeleteModal().should("be.visible");
    }

    editLb() {
        this.elements.editLB().click({ force: true });
    }

    saveEdit() {
        this.elements.saveEditBtn().click({ force: true });
    }

    cancelDeleteLB() {
        this.elements.cancelDelete().click({ force: true });
    }

    confirmDeleteLB() {
        this.elements.confirmDelete().click({ force: true });
    }

    clickcreatedAtSortButton() {
        this.elements.createdAtSortButton().click({ force: true });
    }

    clickstateSortButton() {
        this.elements.stateSortButton().click({ force: true });
    }

    clickIpSortButton() {
        this.elements.ipSortButton().click({ force: true });
    }

    clickNameSort() {
        this.elements.nameSortButton().click({ force: true });
    }

    clickPreviousPageButton() {
        this.elements.previousPageButton().click({ force: true });
    }

    clickNextPageButton() {
        this.elements.nextPageButton().click({ force: true });
    }

    clickPageSelect() {
        this.elements.pageSelect().click({ force: true });
    }

    clickActionsButton() {
        this.elements.actionsButton().click({ force: true });
    }

    clickDeleteActionsButton() {
        this.elements.actionsButton().click({ force: true });
        cy.get('.dropdown-item').should("be.visible").contains("Delete").click({ force: true });
    }

    clickEditActionsButton() {
        this.elements.actionsButton().click({ force: true });
        this.elements.editActionsButton().click({ force: true });
    }

    clickDetailsTab() {
        this.elements.detailsTab().click({ force: true });
    }

    clickSourceIPsTab() {
        this.elements.sourceTab().click({ force: true });
    }

    clickListenersTab() {
        this.elements.listenersTab().click({ force: true });
    }

    balancerName(name) {
        this.elements.nameInput().clear().type(name);
    }

    clearName() {
        this.elements.nameInput().clear();
    }

    getBalancerName() {
        return this.elements.nameInput();
    }

    inputSourceIP(ip) {
        this.elements.sourceIp().clear().type(ip);
    }

    addListener() {
        this.elements.addListener().click({ force: true });
    }

    addSourceIp() {
        this.elements.addSourceIp().click({ force: true });
    }

    deleteListener(num) {
        this.elements.deleteListener().eq(num).click({ force: true });
    }

    deleteSourceIp(num) {
        this.elements.deleteSourceIp().eq(num).click({ force: true });
    }

    clickListenersTab() {
        this.elements.listenersTab().click({ force: true });
    }

    listenerPortInput(port) {
        this.elements.listenerPort().clear().type(port);
    }

    instancePortInput(port) {
        this.elements.instancePort().clear().type(port);
    }

    selectTCP() {
        this.elements.monitorType().click({ force: true });
        this.elements.selectTCP().click({ force: true });
    }

    selectHTTP() {
        this.elements.monitorType().click({ force: true });
        this.elements.selectHTTP().click({ force: true });
    }

    selectHTTPS() {
        this.elements.monitorType().click({ force: true });
        this.elements.selectHTTPS().click({ force: true });
    }

    selectAllInstances() {
        this.elements.selectAll().check();
    }

    launchBalancer() {
        this.elements.launchBtn().scrollIntoView().should("be.visible");
        this.elements.launchBtn().click({ force: true });
    }

    cancelLaunch() {
        this.elements.cancelBtn().click({ force: true });
    }
}

module.exports = new loadBalancerPage();