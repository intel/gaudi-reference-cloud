class iksPage {

    elements = {

        // IKS Cluster home
        createClusterBtn: () => cy.get('[intc-id="btn-iksLaunch-cluster"]').contains("Launch cluster"),
        firstClusterBtn: () => cy.get('.btn.btn-primary').contains("Launch cluster", { matchCase: false }),

        // K8s cluster configuration
        clusterName: () => cy.get('[intc-id="ClusternameInput"]'),
        k8sVersion: () => cy.get('[intc-id="Selectclusterkubernetesversion-form-select"]'),
        version129: () => cy.get('[intc-id="Selectclusterkubernetesversion-form-select-option-129"]'),
        version128: () => cy.get('[intc-id="Selectclusterkubernetesversion-form-select-option-128"]'),
        launchCluster: () => cy.get('[intc-id="btn-iksLaunchCluster-Launch"]'),
        cancelCluster: () => cy.get('[intc-id="btn-iksLaunchCluster-Cancel"]'),

        // My Cluster page
        searchInput: () => cy.get('[intc-id="searchCluster"]'),
        clickAddGroup: () => cy.get('[intc-id="AddgroupHyperLinkTable"]'),
        copyKubeConfig: () => cy.get('[intc-id="ButtonTable Copy"]'),
        downloadKubeConfig: () => cy.get('[intc-id="ButtonTable Download"]'),
        deleteBtn: () => cy.get('[intc-id="ButtonTable Delete cluster instance"]'),
        nameInput: () => cy.get('[intc-id="NameInput"]'),
        confirmDeleteModal: () => cy.get('[intc-id="deleteConfirmModal"]'),
        cancelDelete: () => cy.get('[intc-id="btn-confirm-Deletecluster instance-cancel"]'),
        confirmDelete: () => cy.get('[intc-id="btn-confirm-Deletecluster instance-delete"]'),
        clearFilterBtn: () => cy.get('[intc-id="ClearfiltersEmptyViewButton"]'),

        // Cluster Details Actions
        detailsTab: () => cy.get('[intc-id="DetailsTab"]'),
        workersTab: () => cy.get('[intc-id="Worker Node Groups (0)Tab"]'),
        loadBalancerTab: () => cy.get('[intc-id="Load Balancers (0)Tab"]'),
        storageTab: () => cy.get('[intc-id="StorageTab"]'),
        actionsDropDown: () => cy.get('[intc-id="myReservationActionsDropdownButton"] > .dropdown-toggle'),
        addNodeGroupDropDown: () => cy.get('[intc-id="myReservationActionsDropdownItemButton0"]'),
        addLoadBalancer: () => cy.get('[intc-id="myReservationActionsDropdownItemButton1"]'),
        actionDelete: () => cy.get('[intc-id="myReservationActionsDropdownItemButton2"]').contains("Delete"),
        dropdownDelete: () => cy.get('.dropdown-item').contains("Delete"),
        detailsCopyKubeConfig: () => cy.get('[intc-id="btn-details-tab-kubeconfig-Copy"]'),
        detailsDownloadKubeConfig: () => cy.get('[intc-id="btn-details-tab-kubeconfig-Download"]'),

        // Add node group page
        CPU: () => cy.get('[intc-id="CPU-radio-select"]'),
        GPU: () => cy.get('[intc-id="GPU-radio-select"]'),
        AI: () => cy.get('[intc-id="AI-radio-select"]'),
        nodeFamily: () => cy.get('[intc-id="Nodefamily-form-select"]'),
        nodeType: () => cy.get('[intc-id="Nodetype-form-select"]'),
        groupName: () => cy.get('[intc-id="NodegroupnameInput"]'),
        nodeQty: () => cy.get('[intc-id="Nodequantity-form-select"]'),
        nodeQty1: () => cy.get('[intc-id="Nodequantity-form-select-option-1"]'),
        nodeQty2: () => cy.get('[intc-id="Nodequantity-form-select-option-2"]'),
        uploadKey: () => cy.get('[intc-id="SelectkeysbtnExtra"]'),
        keyCheckbox: () => cy.get('[intc-id="checkbox testkey"]'),
        launchGroup: () => cy.get('[intc-id="btn-computelaunch-navigationBottom Launch node group - "]'),
        cancelGroup: () => cy.get('[intc-id="btn-computelaunch-navigationBottom Cancel - "]'),

        // Worker Node Groups tab
        addWorkerNodeGroupBtn: () => cy.get('[intc-id="btn-iksMyClusters-addWorkerNodeGroup"]'),
        deleteNodeGroup: () => cy.get('.btn.btn-outline-primary').contains("Delete"),
        confirmDeleteGroup: () => cy.get('[intc-id="btn-confirm-Deletenode group-delete"]'),
        cancelDeleteGroup: () => cy.get('[intc-id="btn-confirm-Deletenode group-cancel"]'),
        addNode: () => cy.get('[intc-id="btn-iksMyClusters-addNode"]'),
        removeNode: () => cy.get('[intc-id="btn-iksMyClusters-removeNode"]'),
        confirmChangeN: () => cy.get('[intc-id="btn-confirm-ChangeNode Group Count-Continue"]'),
        cancelChangeN: () => cy.get('[intc-id="btn-confirm-ChangeNode Group Count-cancel"]'),

        // k8s upgrade
        k8sUpgrade: () => cy.get('[intc-id="btn-details-tab-k8sversion-Upgrade"]').contains("Upgrade"),
        selectK8sUpgrade: () => cy.get('[intc-id="Selectthek8sVersionSelect"]'),
        cancelUpgrade: () => cy.get('[intc-id="btn-iksUpgradeClusterModal-cancel"]'),
        confirmUpgrade: () => cy.get('[intc-id="btn-iksUpgradeClusterModal-upgrade"]'),

        // Load Balancer
        lbalancerTab: () => cy.get('[intc-id="Load Balancers (0)Tab"]').contains("Load Balancers"),
        addLoadBalancer: () => cy.get('[intc-id="btn-iksMyClusters-addVip"]'),
        lbInputName: () => cy.get('[intc-id="NameInput"]'),
        portSelect: () => cy.get('[intc-id="Port-form-select"]'),
        lbType: () => cy.get('[intc-id="Type-form-select"]'),
        portOption80: () => cy.get('[intc-id="selected-option 80"]'),
        portOption443: () => cy.get('[intc-id="selected-option 443"]'),
        privateLB: () => cy.get('[intc-id="Type-form-select-option-private"]'),
        publicLB: () => cy.get('[intc-id="Type-form-select-option-public"]'),
        launchLB: () => cy.get('[intc-id="btn-iksLaunchVip-Launch"]'),
        cancelLB: () => cy.get('[intc-id="btn-iksLaunchVip-Cancel"]'),
        deleteLB: () => cy.get('[intc-id="ButtonTable Delete load balancer"]'),
        cancelDeleteLB: () => cy.get('[intc-id="btn-confirm-Deleteload balancer-cancel"]'),
        confirmDeleteLB: () => cy.get('[intc-id="btn-confirm-Deleteload balancer-delete"]'),

        // Add Storage
        addStorageBtn: () => cy.get('.btn.btn-outline-primary').contains("Add Storage"),

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

        // Metrics tab
        metricsTab: () => cy.get('[intc-id="MetricsTab"]'),
        viewMetricsDropDown: () => cy.get('[intc-id="View-form-select"]'),
        refreshMetrics: () => cy.get('[intc-id="metricsRefreshButton"]'),
        apiServerMetrics: () => cy.get('[intc-id="ClusterMetricsTitle-apiserver"]'),
        etcdMetrics: () => cy.get('[intc-id="ClusterMetricsTitle-etcd"]'),

    }

    createFirstCluster() {
        this.elements.firstClusterBtn().click({ force: true });
    }

    createCluster() {
        this.elements.firstClusterBtn().click({ force: true });
    }

    typeClusterName(name) {
        this.elements.clusterName().clear().type(name);
    }

    k8sVersion128() {
        this.elements.k8sVersion().click({ force: true });
        this.elements.version128().scrollIntoView();
        this.elements.version128().click({ force: true });
    }

    k8sVersion129() {
        this.elements.k8sVersion().click({ force: true });
        this.elements.version129().scrollIntoView();
        this.elements.version129().click({ force: true });
    }

    clickCPU() {
        this.elements.CPU().should("be.visible").check({ force: true })
    }

    clickGPU() {
        this.elements.GPU().should("be.visible").check({ force: true })
    }

    clickAI() {
        this.elements.AI().should("be.visible").check({ force: true })
    }

    launchCluster() {
        this.elements.launchCluster().click({ force: true });
    }

    cancelCluster() {
        this.elements.cancelCluster().click({ force: true });
    }

    learnHowItWorks() {
        this.elements.learnHowItWorks().click();
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

    clickAddGroup() {
        this.elements.clickAddGroup().click({ force: true });
    }

    copyKubeConfig() {
        this.elements.copyKubeConfig().click({ force: true });
    }

    downloadKubeConfig() {
        this.elements.downloadKubeConfig().click({ force: true });
    }

    deleteCluster() {
        this.elements.deleteBtn().click();
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

    actionDeleteDropDown() {
        this.elements.actionsDropDown().click({ force: true });
        this.elements.dropdownDelete().click({ force: true });
    }

    addWorkerNodeGroupDropDown() {
        this.elements.addNodeGroupDropDown().click({ force: true });
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

    deleteNameInput(name) {
        this.elements.nameInput().clear().type(name);
    }

    checkConfirmDeleteModal() {
        this.elements.confirmDeleteModal().should("be.visible");
    }

    uploadKey() {
        this.elements.uploadKey().click({ force: true });
    }

    addLoadBalancer() {
        this.elements.addLoadBalancer().click({ force: true });
    }

    actionDelete() {
        this.elements.actionDelete().click();
    }

    selectNodeFamily() {
        this.elements.nodeFamily().click();
    }

    selectNodeType() {
        this.elements.nodeType().click();
    }

    inputGroupName(name) {
        this.elements.groupName().scrollIntoView();
        this.elements.groupName().clear().type(name);
    }

    selectNodeQty() {
        this.elements.nodeQty1().click({ force: true });
    }

    keyCheckbox() {
        this.elements.keyCheckbox().click({ force: true });
    }

    launchGroup() {
        this.elements.launchGroup().scrollIntoView();
        this.elements.launchGroup().click({ force: true });
    }

    cancelGroup() {
        this.elements.cancelGroup().click();
    }

    wokerNodeGroupsTab() {
        this.elements.wokerNodeGroupsTab().click();
    }

    addWorkerNodeGroup() {
        this.elements.addWorkerNodeGroupBtn().click({ force: true });
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
        this.elements.lbalancerTab().click({ force: true });
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
        this.elements.lbType().click({ force: true });
        this.elements.privateLB().click({ force: true });
    }

    publicLB() {
        this.elements.lbType().click({ force: true });
        this.elements.publicLB().click({ force: true });
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

    addStorage() {
        this.elements.addStorageBtn().click({ force: true });
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
        this.elements.protocolDropDown().click({ force: true });
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

    clickMetricsTab() {
        this.elements.metricsTab().click({ force: true });
    }

    refreshMetrics() {
        this.elements.refreshMetrics().click({ force: true });
    }

    viewDropDown() {
        this.elements.viewMetricsDropDown().click({ force: true });
    }

    apiServerMetrics() {
        this.elements.apiServerMetrics().click({ force: true });
    }

    etcdMetrics() {
        this.elements.etcdMetrics().click({ force: true });
    }
}

module.exports = new iksPage();