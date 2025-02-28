class superComputingPage {

    elements = {

        // Supercomputing table grid
        createFirstCluster: () => cy.get('[intc-id="LaunchSupercomputingClusterEmptyViewButton"]'),
        createClusterBtn: () => cy.get('[intc-id="launch-SCCluster"]'),
        copyKubeConfig: () => cy.get('[intc-id="ButtonTable Copy"]').contains("Copy"),
        downloadKubeConfig: () => cy.get('[intc-id="ButtonTable Download"]').contains("Download"),
        searchInput: () => cy.get('[intc-id="searchCluster"]'),

        // Delete SC cluster
        nameInput: () => cy.get('[intc-id="NameInput"]'),
        confirmDeleteModal: () => cy.get('[intc-id="deleteConfirmModal"]'),
        confirmDeleteBtn: () => cy.get('[intc-id="btn-confirm-Deletecluster instance-delete"]'),
        cancelDeleteBtn: () => cy.get('[intc-id="btn-confirm-Deletecluster instance-cancel"]'),
        deleteClusterBtn: () => cy.get('[intc-id="ButtonTable Delete cluster instance"]'),

        // Launch Supercomputing cluster page
        clusterName: () => cy.get('[intc-id="ClusternameInput"]'),
        userDataURL: () => cy.get('[intc-id="UserdataURLInput"]'),
        kubernetesVersion: () => cy.get('[intc-id="Clusterkubernetesversion-form-select"]'),
        volumeInputField: () => cy.get('[intc-id="Volumesize(TB)Input"]'),
        uploadKeyBtn: () => cy.get('[intc-id="keysbtnExtra"]'),
        selectAllKeysBtn: () => cy.get('[intc-id="keysbtnSelectAll"]'),
        launchCluster: () => cy.get('[intc-id="btn-computelaunch-navigationBottom Launch"]'),
        cancelCluster: () => cy.get('[intc-id="btn-computelaunch-navigationBottom Cancel"]'),
        addNodeGroup: () => cy.get('[intc-id="btn-iksMyClusters-addNode"]'),
        computeInstanceType: () => cy.get('[intc-id="ComputeInstancetype-form-select"]'),
        nodesInput: () => cy.get('[intc-id="NodesInput"]'),
        removeGeneralCompute: () => cy.get('[aria-label="Delete node group"]'),

        // SC cluster details page
        detailsTab: () => cy.get('[intc-id="DetailsTab"]'),
        workersTab: () => cy.get('[intc-id="Worker Node Groups (1)Tab"]'),
        loadBalancerTab: () => cy.get('[intc-id="Load Balancers (0)Tab"]'),
        storageTab: () => cy.get('[intc-id="StorageTab"]'),
        cancelDelete: () => cy.get('[intc-id="btn-confirm-Delete Cluster Instance-cancel"]'),
        confirmDelete: () => cy.get('[intc-id="btn-confirm-Delete Cluster Instance-Delete"]'),
        clearFilterBtn: () => cy.get('[intc-id="ClearfiltersEmptyViewButton"]'),
        detailsCopyKubeConfig: () => cy.get('[intc-id="btn-details-tab-kubeconfig-Copy"]'),
        detailsDownloadKubeConfig: () => cy.get('[intc-id="btn-details-tab-kubeconfig-Download"]'),

        // Hourly cost estimation
        SCCostTotal: () => cy.get('[intc-id="SCCostTotal"]'),
        SCControlPlaneCost: () => cy.get('[intc-id="SCControlPlaneCost"]'),
        SCAINodesCost: () => cy.get('[intc-id="SCAINodesCost"]'),
        SCGCNodesCost: () => cy.get('[intc-id="SCGCNodesCost"]'),
        SCStorageCost: () => cy.get('[intc-id="SCStorageCost"]'),

        // Cluster Details Actions
        actionsDropDown: () => cy.get('[intc-id="SuperComputerActionsDropdownButton"]'),
        deleteFromActions: () => cy.get('[intc-id="myReservationActionsDropdownItemButton0"]'),

        // k8s upgrade
        k8sUpgrade: () => cy.get('[intc-id="btn-details-tab-k8sversion-Upgrade"]').contains("Upgrade"),
        selectK8sUpgrade: () => cy.get('[intc-id="SelectthekubernetesVersion-toggle-form-select"]'),
        cancelUpgrade: () => cy.get('[intc-id="btn-iksUpgradeClusterModal-cancel"]'),
        confirmUpgrade: () => cy.get('[intc-id="btn-iksUpgradeClusterModal-upgrade"]'),

        // Add node group page
        aiNodesRadio: () => cy.get('[intc-id="Nodegrouptype-Radio-option-AInodes"]'),
        computeNodesRadio: () => cy.get('[intc-id="Nodegrouptype-Radio-option-Computenodes"]'),
        searchNodes: () => cy.get('[intc-id="searchNodes"]'),

        nodeType: () => cy.get('[intc-id="Nodetype-form-select"]'),
        groupName: () => cy.get('[intc-id="NodegroupnameInput"]'),
        nodeQty: () => cy.get('[intc-id="NodequantityInput"]'),
        uploadKey: () => cy.get('[intc-id="SelectkeysbtnExtra"]'),
        keyCheckbox: () => cy.get('[intc-id="checkbox testkey"]'),
        launchGroup: () => cy.get('[intc-id="btn-computelaunch-navigationBottom Launch - "]'),
        cancelGroup: () => cy.get('[intc-id="btn-computelaunch-navigationBottom Cancel - "]'),

        // Worker Node Groups tab
        addWorkerNodeGroup: () => cy.get('.btn.btn-outline-primary').contains("Add node group"),
        deleteNodeGroup: () => cy.get('.btn.btn-outline-primary').contains("Delete"),
        confirmDeleteGroup: () => cy.get('[intc-id="btn-confirm-Delete Node Group-Delete"]'),
        cancelDeleteGroup: () => cy.get('[intc-id="btn-confirm-Delete Node Group-cancel"]'),
        addNode: () => cy.get('.btn.btn-outline-primary').contains("Add Node"),
        removeNode: () => cy.get('.btn.btn-outline-primary').contains("Delete Node"),
        confirmChangeN: () => cy.get('[intc-id="btn-confirm-Addnode-Add"]'),
        cancelChangeN: () => cy.get('[intc-id="btn-confirm-Addnode-cancel"]'),
        searchNodes: () => cy.get('[intc-id="searchNodes"]'),

        // Load Balancer
        addLoadBalancer: () => cy.get('[intc-id="btn-iksMyClusters-addVip"]'),
        lbInputName: () => cy.get('[intc-id="NameInput"]'),
        portSelect: () => cy.get('[intc-id="Port-form-select"]'),
        lbType: () => cy.get('[intc-id="Type-form-select"]'),
        portOption80: () => cy.get('[intc-id="Port-form-select-option-80"]'),
        portOption443: () => cy.get('[intc-id="Port-form-select-option-443"]'),
        privateLB: () => cy.get('[intc-id="selected-option private"]'),
        publicLB: () => cy.get('[intc-id="selected-option public"]'),
        launchLB: () => cy.get('[intc-id="btn-iksLaunchCluster-Launch"]'),
        cancelLB: () => cy.get('[intc-id="btn-iksLaunchCluster-Cancel"]'),
        deleteLB: () => cy.get('[intc-id="ButtonTable Delete load balancer"]'),
        cancelDeleteLB: () => cy.get('[intc-id="btn-confirm-Delete Load Balancer-cancel"]'),
        confirmDeleteLB: () => cy.get('[intc-id="btn-confirm-Delete Load Balancer-Delete"]'),

        // Security tab
        securityTab: () => cy.get('[intc-id="SecurityTab"]'),
        editFirewall: () => cy.get('[intc-id="ButtonTable Edit"]').contains("Edit"),
        deleteFirewall: () => cy.get('[intc-id="ButtonTable Delete"]').contains("Delete"),
        sourceIpInput: () => cy.get('[intc-id="SourceIPInput"]'),
        addSourceIPbtn: () => cy.get('[intc-id="btn-iksMyClusters-addSourceIp"]'),
        deleteSourceIPbtn: () => cy.get('[intc-id="btn-IKS-delete-source-ip"]'),
        protocolDropDown: () => cy.get('[intc-id="Protocol-form-select"]'),
        selectTCP: () => cy.get('[intc-id="Protocol-form-select-option-TCP"]'),
        selectUDP: () => cy.get('[intc-id="Protocol-form-select-option-UDP"]'),
        saveRule: () => cy.get('[intc-id="btn-ikscluster-ruleEdit-navigationBottom Save"]'),
        cancelEditRule: () => cy.get('[intc-id="btn-ikscluster-ruleEdit-navigationBottom Cancel"]'),
        confirmDeleteRule: () => cy.get('[intc-id="btn-confirm-DeleteSecurity Rule-delete"]'),
        cancelDeleteRule: () => cy.get('[intc-id="btn-confirm-DeleteSecurity Rule-cancel"]'),

    }

    createFirstCluster() {
        this.elements.createFirstCluster().should("be.visible").click({ force: true });
    }

    createCluster() {
        this.elements.createClusterBtn().should("be.visible").click({ force: true });
    }

    removeCompute() {
        this.elements.removeGeneralCompute().click({ force: true });
    }

    typeClusterName(name) {
        this.elements.clusterName()
            .scrollIntoView({ duration: 2000 })
            .clear()
            .type(name);
    }

    typeUserURLdata(name) {
        this.elements.userDataURL().clear().type(name);
    }

    inputNodes(name) {
        this.elements.nodesInput().clear().type(name);
    }

    nodesQuantity(node) {
        this.elements.nodeQty().clear().type(node);
    }

    searchNodes(name) {
        this.elements.searchNodes().clear().type(name);
    }

    k8sVersion(version) {
        this.elements.kubernetesVersion().click({ force: true });
        cy.get('.dropdown-item').contains(version).should("be.visible").click({ force: true });
    }

    computeNodeRadio() {
        this.elements.computeNodesRadio().should("be.visible").select();
    }

    confirmDeleteCluster() {
        this.elements.confirmDeleteBtn().click({ force: true });
    }

    cancelDelete() {
        this.elements.cancelDeleteBtn().click({ force: true });
    }

    deleteCluster() {
        this.elements.deleteClusterBtn().click({ force: true });
    }

    SCCostTotal() {
        this.elements.SCCostTotal().should("be.visible");
    }

    SCControlPlaneCost() {
        this.elements.SCControlPlaneCost().should("be.visible");
    }

    SCAINodesCost() {
        this.elements.SCAINodesCost().should("be.visible");
    }

    SCGCNodesCost() {
        this.elements.SCGCNodesCost().should("be.visible");
    }

    SCStorageCost() {
        this.elements.SCStorageCost().should("be.visible");
    }

    launchCluster() {
        this.elements.launchCluster().click({ force: true });
    }

    cancelCluster() {
        this.elements.cancelCluster().click({ force: true });
    }

    searchInput(name) {
        this.elements.searchInput()
            .scrollIntoView({ duration: 2000 })
            .clear()
            .type(name);
    }

    clearFilter() {
        this.elements.clearFilterBtn().click();
    }

    selectComputeInstance() {
        this.elements.computeInstanceType().click();
    }

    copyKubeConfig() {
        this.elements.copyKubeConfig().click({ force: true });
    }

    downloadKubeConfig() {
        this.elements.downloadKubeConfig().click({ force: true });
    }

    cancelDelete() {
        this.elements.cancelDelete().click();
    }

    confirmDelete() {
        this.elements.confirmDelete().click();
    }

    actionsDropDown() {
        this.elements.actionsDropDown().click({ force: true });
    }

    addNodeGroup() {
        this.elements.addNodeGroup().click({ force: true });
    }

    uploadKey() {
        this.elements.uploadKey().click({ force: true });
    }

    deleteNameInput(name) {
        this.elements.nameInput().clear().type(name);
    }

    checkConfirmDeleteModal() {
        this.elements.confirmDeleteModal().should("be.visible");
    }

    addLoadBalancer() {
        this.elements.addLoadBalancer().click({ force: true });
    }

    actionDelete() {
        this.elements.deleteFromActions().click({ force: true });
    }

    selectNodeFamily() {
        this.elements.nodeFamily().click();
    }

    selectNodeType() {
        this.elements.nodeType().click();
    }

    inputGroupName(name) {
        this.elements.groupName().type(name);
    }

    inputVolumeSize(size) {
        this.elements.volumeInputField().clear().type(size);
    }

    keyCheckbox() {
        this.elements.keyCheckbox().click({ force: true });
    }

    launchGroup() {
        this.elements.launchGroup().click();
    }

    cancelGroup() {
        this.elements.cancelGroup().click();
    }

    wokerNodeGroupsTab() {
        this.elements.wokerNodeGroupsTab().click();
    }

    addWorkerNodeGroupDropDown() {
        this.elements.addWorkerNodeGroup().click({ force: true });
    }

    deleteNodeGroup() {
        this.elements.deleteNodeGroup().click({ force: true });
    }

    confirmDeleteGroup() {
        this.elements.confirmDeleteGroup().click({ force: true });
    }

    cancelDeleteGroup() {
        this.elements.cancelDeleteGroup().click({ force: true });
    }

    k8sUpgrade() {
        this.elements.k8sUpgrade().click({ force: true });
    }

    chooseUpgradeVersion() {
        this.elements.selectK8sUpgrade().click({ force: true });
    }

    confirmUpgrade() {
        this.elements.confirmUpgrade().click({ force: true });
    }

    cancelUpgrade() {
        this.elements.cancelUpgrade().click({ force: true });
    }

    addNode() {
        this.elements.addNode().click({ force: true });
    }

    removeNode() {
        this.elements.removeNode().click();
    }

    confirmChangeNodes() {
        this.elements.confirmChangeN().click();
    }

    cancelChangeNodes() {
        this.elements.cancelChangeN().click();
    }

    lbalancerTab() {
        this.elements.loadBalancerTab()
            .scrollIntoView({ duration: 2000 })
            .click({ force: true });
    }

    detailsTab() {
        this.elements.detailsTab()
            .scrollIntoView({ duration: 2000 })
            .click({ force: true });
    }

    storageTab() {
        this.elements.storageTab()
            .scrollIntoView({ duration: 2000 })
            .click({ force: true });
    }

    wokerNodeGroupsTab() {
        this.elements.workersTab()
            .scrollIntoView({ duration: 2000 })
            .click({ force: true });
    }

    addLoadBalancer() {
        this.elements.addLoadBalancer().click();
    }

    lbInputName(name) {
        this.elements.lbInputName().type(name);
    }

    portSelect() {
        this.elements.portSelect().click();
    }

    selectLBType() {
        this.elements.lbType().click({ force: true });
    }

    portOption80() {
        this.elements.portOption80().click({ force: true });
    }

    portOption443() {
        this.elements.portOption443().click({ force: true });
    }

    privateLB() {
        this.elements.privateLB().click();
    }

    publicLB() {
        this.elements.publicLB().click();
    }

    launchLB() {
        this.elements.launchLB().click({ force: true });
    }

    cancelLB() {
        this.elements.cancelLB().click({ force: true });
    }

    deleteLB() {
        this.elements.deleteLB().click({ force: true });
    }

    cancelDeleteLB() {
        this.elements.cancelDeleteLB().click({ force: true });
    }

    confirmDeleteLB() {
        this.elements.confirmDeleteLB().click({ force: true });
    }

    clickSecurityTab() {
        this.elements.securityTab().click({ force: true });
    }

    editRule() {
        this.elements.editFirewall().click({ force: true });
    }

    deleteRule() {
        this.elements.deleteFirewall().click({ force: true });
    }

    typeSourceIP(source) {
        this.elements.sourceIpInput().clear().type(source);
    }

    addSourceIP() {
        this.elements.addSourceIPbtn().click({ force: true });
    }

    deleteSourceIP() {
        this.elements.deleteSourceIPbtn().click({ force: true });
    }

    selectTCP() {
        this.elements.protocolDropDown.click({ force: true });
        this.elements.selectTCP().click({ force: true });
    }

    selectUDP() {
        this.elements.protocolDropDown().click({ force: true });
        this.elements.selectUDP().click({ force: true });
    }

    saveRuleEdit() {
        this.elements.saveRule().click({ force: true });
    }

    cancelEditRule() {
        this.elements.cancelEditRule().click({ force: true });
    }

    confirmDeleteRule() {
        this.elements.confirmDeleteRule().click({ force: true });
    }

    cancelDeleteRule() {
        this.elements.cancelDeleteRule().click({ force: true });
    }
}

module.exports = new superComputingPage();