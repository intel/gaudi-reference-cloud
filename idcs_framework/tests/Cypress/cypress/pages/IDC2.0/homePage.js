class homePage {
  elements = {
    //Account credit widget
    remainingCredit: () => cy.get('.ms-auto.text-end.fw-semibold').contains("Remaining Credits"),
    usedCredit: () => cy.get('[intc-id="usedCreditsLabel"]'),

    //Current month usage widget
    viewUsage: () => cy.get('[intc-id="link-hardwareusagewidgetr-view-usage"]'),
    estimatedUsedTime: () => cy.get('[intc-id="totalUsageTimeLabel"]'),
    estimatedCost: () => cy.get('[intc-id="totalAmountLabel"]'),
    // Dashboard container
    dashboardContainer: () => cy.get(".section.dashboard-container.gap-s8"),
    hardwareCatalogTittle: () => cy.contains("Available hardware"),

    // Help menu options
    helpButton: () => cy.get('[aria-label="Support menu"]'),
    documentationButton: () => cy.get('[intc-id="help-menu-Browse-documentation"]'),
    sendFeedback: () => cy.get('[intc-id="help-menu-Send-Feedback"]'),
    knowledgeBaseButton: () =>
      cy.get('[intc-id="help-menu-Knowledge-Base"]').contains("Knowledge Base"),
    submitTicketButton: () =>
      cy
        .get('[intc-id="help-menu-Submit-a-ticket"]')
        .contains("Submit a ticket"),
    communityHelpButton: () =>
      cy.get('[intc-id="help-menu-Community"]').contains("Community"),
    contactSupportButton: () => cy.get('[intc-id="help-menu-Contact-support"]'),

    // User Profile Menu
    rolesTab: () => cy.get('.tap-inactive.nav-link').contains("Roles"),
    credentialsTab: () => cy.get('.tap-inactive.nav-link').contains("Credentials"),
    userAccountTab: () => cy.get('[intc-id="userMenu"]'),
    invoicesButton: () => cy.get('[intc-id="invoicesHeaderButton"]'),
    profileSettings: () => cy.get('[intc-id="accountSettingsHeaderButton"]'),
    switchAccounts: () => cy.get('[intc-id="switchAccountsHeaderButton"]'),
    cloudCreditsButton: () => cy.get('[intc-id="cloudCreditsHeaderButton"]'),
    usageButton: () => cy.get('[intc-id="currentMonthUsageHeaderButton"]'),
    paymentMethodsButton: () => cy.get('[intc-id="paymentMethodsHeaderButton"]'),

    signOutButton: () => cy.get('[intc-id="signOutHeaderButton"]'),
    region: () => cy.get('[intc-id="regionLabel"]'),
    notificationButton: () => cy.get('[intc-id="notificationsHeaderButton"]'),

    //sideMenu mappings
    expandSideMenu: () => cy.get('[aria-label="Expand side menu"]'),
    collapasedSideMenu: () => cy.get('[aria-label="Collapse side menu"]'),
    homeSideMenu: () => cy.get('[aria-label="Go to Home Page"]'),
    subMenu: () => cy.get('.submenu.collapse.show'),

    catalogSideMenu: () => cy.get('[aria-label="Go to Catalog Page"]'),
    hardwareCatalogButton: () => cy.get('[intc-id="sidebarnavLink/hardware"]'),
    softwareCatalogButton: () => cy.get('[intc-id="sidebarnavLink/software"]'),

    //Compute 
    computeSideMenu: () => cy.get('[aria-label="Go to Compute Page"]'),
    computePageButton: () => cy.get('[intc-id="sidebarnavLink/compute"]'),
    computeGroupButton: () => cy.get('[intc-id="sidebarnavLink/compute-groups"]'),
    loadbalancerButton: () => cy.get('[intc-id="sidebarnavLink/load-balancer"]'),
    accountKeyButton: () => cy.get('[intc-id="sidebarnavLink/security/publickeys"]'),

    // IKS
    kubernetesSideMenu: () => cy.get('[aria-label="Go to Kubernetes Page"]'),
    iksOverviewPageButton: () => cy.get('[intc-id="sidebarnavLink/cluster/overview"]'),
    iksPageButton: () => cy.get('[intc-id="sidebarnavLink/cluster"]'),

    // Super computing
    superComputingSideMenu: () => cy.get('[aria-label="Go to Supercomputing Page"]'),
    superComputingOverview: () => cy.get('[intc-id="sidebarnavLink/supercomputer/overview"]'),
    superComputingPage: () => cy.get('[intc-id="sidebarnavLink/supercomputer"]'),

    // Preview 
    previewSideMenu: () => cy.get('[aria-label="Go to Preview Page"]'),
    previewComputeButton: () => cy.get('[intc-id="sidebarnavLink/preview/compute"]'),
    previewKeyButton: () => cy.get('[intc-id="sidebarnavLink/preview/security/publickeys"]'),
    previewHardwareCatalogButton: () => cy.get('[intc-id="sidebarnavLink/preview/hardware"]'),
    previewStorageButton: () => cy.get('[intc-id="sidebarnavLink/preview/storage"]'),

    // Storage
    storageSideMenu: () => cy.get('[aria-label="Go to Storage Page"]'),
    storagePageButton: () => cy.get('[intc-id="sidebarnavLink/storage"]'),
    bucketsPageButton: () => cy.get('[intc-id="sidebarnavLink/buckets"]'),

    metricsSideMenu: () => cy.get('[aria-label="Go to Metrics Page"]'),
    aiPlaySideMenu: () => cy.get('[intc-id="sidebarnavLink/aiplayground"]'),
    learningButton: () => cy.get('[intc-id="sidebarnavLink/learning/notebooks"]'),
    labsButton: () => cy.get('[intc-id="sidebarnavLink/learning/labs"]'),
    docsSideMenu: () => cy.get('[intc-id="sidebarnavLink/docs"]'),

    // Products
    thirdGenIntelProd: () =>
      cy.get(
        '[intc-id="btn-hardwarecatalog-select 3rd Generation Intel® Xeon® Scalable Processors"]'
      ),
    intelDataCenterGPUProd: () =>
      cy.get(
        '[intc-id="btn-hardwarecatalog-select Intel® Data Center GPU Flex Series on latest Intel® Xeon® processors"]'
      ),
    habanaGaudi2Prod: () =>
      cy.get(
        '[intc-id="btn-hardwarecatalog-select Gaudi2® Deep Learning Server"]'
      ),
    fourthGenIntelProd: () =>
      cy.get(
        '[intc-id="btn-hardwarecatalog-select 4th Generation Intel® Xeon® Scalable processors"]'
      ),
    intelXeonProd: () =>
      cy.get(
        '[intc-id="btn-hardwarecatalog-select Intel® Xeon® processors, codenamed Sapphire Rapids with high bandwidth memory (HBM)"]'
      ),
    intelMaxSeriesProd: () =>
      cy.get('[intc-id="btn-hardwarecatalog-select Intel® Max Series GPU"]'),
    habanaGaudi2Cluster: () =>
      cy.get(
        '[intc-id="btn-hardwarecatalog-select Gaudi2® Deep Learning Server"]'
      ),
    searchFilter: () => cy.get('[intc-id="Filter-Text"]'),
    releasedCategory: () => cy.get('[intc-id="Checkbox-item Released"]'),
    computeSideMenu: () => cy.get('[aria-label="Go to Compute Page"]'),

    // Available hardware filters
    allProducts: () => cy.get('[aria-label="Toggle filter all products"]'),
    cpuProcessor: () => cy.get('[aria-label="Toggle filter CPU products"]'),
    gpuProcessor: () => cy.get('[aria-label="Toggle filter GPU products"]'),
    AIprocessor: () => cy.get('[aria-label="Toggle filter AI products"]'),
    coreComputeOption: () => cy.get('[aria-label="Toggle filter Core compute products"]'),
    aiPC: () => cy.get('[aria-label="Toggle filter AI PC products"]'),

    // footer of mainpage
    companyOverviewLink: () =>
      cy.get(
        '[href="https://www.intel.com/content/www/us/en/company-overview/company-overview.html"]'
      ),
    contactIntelLink: () =>
      cy.get(
        '[href="https://www.intel.com/content/www/us/en/support/contact-intel.html"]'
      ),
    newsRoomLink: () =>
      cy.get(
        '[href="https://www.intel.com/content/www/us/en/newsroom/home.html"]'
      ),
    investorsLink: () => cy.get('[href="https://www.intc.com/"]'),
    careersLink: () =>
      cy.get(
        '[href="https://www.intel.com/content/www/us/en/jobs/jobs-at-intel.html"]'
      ),
    corporateRespLink: () =>
      cy.get(
        '[href="https://www.intel.com/content/www/us/en/corporate-responsibility/corporate-responsibility.html"]'
      ),
    diversityIncLink: () =>
      cy.get(
        '[href="https://www.intel.com/content/www/us/en/diversity/diversity-at-intel.html"]'
      ),
    publicPolicyLink: () =>
      cy.get(
        '[href="https://www.intel.com/content/www/us/en/company-overview/public-policy.html"]'
      ),
  };

  expandSideMenu() {
    this.elements.expandSideMenu().click({ force: true });
  }

  collapasedSideMenu() {
    this.elements.collapasedSideMenu().should("be.visible").click({ force: true });
  }

  checkDashboard() {
    this.elements.dashboardContainer().should("be.visible");
  }

  checkHardwarePage() {
    this.elements.hardwareCatalogTittle().should("be.visible");
  }

  hardwareCatalog() {
    this.elements.catalogSideMenu().click({ force: true });
    this.elements.subMenu().should("be.visible");
    this.elements.hardwareCatalogButton().click({ force: true });
  }

  softwareCatalog() {
    this.elements.catalogSideMenu().click({ force: true });
    this.elements.subMenu().should("be.visible");
    this.elements.softwareCatalogButton().click({ force: true });
  }

  learning() {
    this.elements.learningButton().click({ force: true });
  }

  labs() {
    this.elements.labsButton().click({ force: true });
  }

  rolesTab() {
    this.elements.userAccountTab().click({ force: true });
    this.elements.profileSettings().click({ force: true });
    this.elements.rolesTab().click({ force: true });
  }

  credentialsTab() {
    this.elements.userAccountTab().click({ force: true });
    this.elements.profileSettings().click({ force: true });
    this.elements.credentialsTab().click({ force: true });
  }

  viewCreditsPage() {
    this.elements.viewCredits().click({ force: true });
  }

  viewCouponPage() {
    this.elements.userAccountTab().click({ force: true });
    this.elements.cloudCreditsButton().click({ force: true });
  }

  viewUsages() {
    this.elements.viewUsage().click({ force: true });
  }

  getRemainingCredit() {
    const credit = Number(this.elements.remainingCredit().replace("USD", ""));
    return credit;
  }

  clickhelpButton() {
    this.elements.helpButton().should("be.visible").click({ force: true });
  }

  clickknowledgeBaseButton() {
    this.elements.knowledgeBaseButton().click({ force: true });
  }

  clickDocumentationButton() {
    this.elements.documentationButton().click({ force: true });
  }

  clicksubmitTicketButton() {
    this.elements.submitTicketButton().click({ force: true });
  }

  clickContactSupportButton() {
    this.elements.contactSupportButton().click({ force: true });
  }

  clickCommunityButton() {
    this.elements.communityHelpButton().click({ force: true });
  }

  clickuserAccountTab() {
    this.elements.userAccountTab().click({ multiple: true });
  }

  clickInvoices() {
    this.elements.invoicesButton().click({ force: true });
  }

  clickPaymentMethods() {
    this.elements.paymentMethodsButton().click({ force: true });
  }

  clickCurrentUsage() {
    this.elements.usageButton().click();
  }

  clickCloudCredits() {
    this.elements.cloudCreditsButton().click();
  }

  clickSendFeedback() {
    this.elements.sendFeedback().click();
  }

  clickProfileSettings() {
    this.elements.profileSettings().click();
  }

  signOut() {
    this.elements.signOutButton().click();
  }

  clickRegion() {
    this.elements.region().click();
  }

  clickNotificationButton() {
    this.elements.notificationButton().click();
  }

  clickhomePageButton() {
    this.elements.homeSideMenu().click({ force: true });
  }

  clickcomputePageButton() {
    this.elements.computeSideMenu().click({ force: true });
    this.elements.subMenu().should("be.visible");
    this.elements.computePageButton().click({ force: true });
  }

  clickIKSPageButton() {
    this.elements.kubernetesSideMenu().click({ force: true });
    this.elements.subMenu().should("be.visible");
    this.elements.iksPageButton().click({ force: true });
  }

  clickIKSOverviewButton() {
    this.elements.kubernetesSideMenu().click({ force: true });
    this.elements.subMenu().should("be.visible");
    this.elements.iksOverviewPageButton().click({ force: true });
  }

  clickSC_OverviewButton() {
    this.elements.superComputingSideMenu().click({ force: true });
    this.elements.subMenu().should("be.visible");
    this.elements.superComputingOverview().click({ force: true });
  }

  clickSuperComputingPage() {
    this.elements.superComputingSideMenu().click({ force: true });
    this.elements.subMenu().should("be.visible");
    this.elements.superComputingPage().click({ force: true });
    cy.contains("Feature Unavailable").should("be.visible");
    cy.get(".btn.btn-link").eq(1).contains("3").should("be.visible").click({ force: true });
  }

  clickStorageButton() {
    this.elements.storageSideMenu().click({ force: true });
    this.elements.subMenu().should("be.visible");
    this.elements.storagePageButton().click({ force: true });
  }

  clickBucketStorageButton() {
    this.elements.storageSideMenu().click({ force: true });
    this.elements.subMenu().should("be.visible");
    this.elements.bucketsPageButton().click({ force: true });
  }

  clickcomputeGroupButton() {
    this.elements.computeSideMenu().click({ force: true });
    this.elements.subMenu().should("be.visible");
    this.elements.computeGroupButton().click({ force: true });
  }

  clickKeysPageButton() {
    this.elements.computeSideMenu().click({ force: true });
    this.elements.subMenu().should("be.visible");
    this.elements.accountKeyButton().click({ force: true });
  }

  clickLoadbalancerPageButton() {
    this.elements.computeSideMenu().click({ force: true });
    this.elements.subMenu().should("be.visible");
    this.elements.loadbalancerButton().click({ force: true });
  }

  clickpreviewComputePageButton() {
    this.elements.previewSideMenu().click({ force: true });
    this.elements.subMenu().should("be.visible");
    this.elements.previewComputeButton().click({ force: true })
  }

  clickpreviewKeyPageButton() {
    this.elements.previewSideMenu().click({ force: true });
    this.elements.subMenu().should("be.visible");
    this.elements.previewKeyButton().click({ force: true });
  }

  clickpreviewStoragePageButton() {
    this.elements.previewSideMenu().click({ force: true });
    this.elements.subMenu().should("be.visible");
    this.elements.previewStorageButton().click({ force: true });
  }

  clickpreviewHardwareCatalog() {
    this.elements.previewSideMenu().click({ force: true });
    this.elements.subMenu().should("be.visible");
    this.elements.previewHardwareCatalogButton().click({ force: true });
  }

  clickcompressLeftBar() {
    this.elements.compressLeftBar().click({ force: true });
  }

  clickthirdGenIntelProd() {
    this.elements.thirdGenIntelProd().click({ force: true });
  }

  clickintelDataCenterGPUProd() {
    this.elements.intelDataCenterGPUProd().click({ force: true });
  }

  clickhabanaGaudi2Prod() {
    this.elements.habanaGaudi2Prod().click({ force: true });
  }

  clickhabanaGaudi2Cluster() {
    this.elements.habanaGaudi2Cluster().click({ force: true });
  }

  clickfourthGenIntelProd() {
    this.elements.fourthGenIntelProd().click({ force: true });
  }

  clickintelXeonProd() {
    this.elements.intelXeonProd().click({ force: true });
  }

  clickintelMaxSeriesProd() {
    this.elements.intelMaxSeriesProd().click({ force: true });
  }

  searchFilter(filterval) {
    this.elements.searchFilter().scrollIntoView().should("be.visible");
    this.elements.searchFilter().clear().type(filterval);
  }

  searchBoxIsVisible() {
    this.elements.searchFilter().scrollIntoView().should("be.visible");
  }

  clickAI_PC() {
    this.elements.aiPC().scrollIntoView().should("be.visible");
    this.elements.aiPC().click({ force: true });
  }

  clickcpuProcessor() {
    this.elements.cpuProcessor().scrollIntoView().should("be.visible");
    this.elements.cpuProcessor().click({ force: true });
  }

  clickgpuProcessor() {
    this.elements.gpuProcessor().scrollIntoView().should("be.visible");
    this.elements.gpuProcessor().click({ force: true });
  }

  clickAIprocessor() {
    this.elements.AIprocessor().scrollIntoView().should("be.visible");
    this.elements.AIprocessor().click({ force: true });
  }

  clickCoreCompute() {
    this.elements.coreComputeOption().scrollIntoView().should("be.visible");
    this.elements.coreComputeOption().click({ force: true });
  }

  clickAllProductsFilter() {
    this.elements.allProducts().scrollIntoView().should("be.visible");
    this.elements.allProducts().click({ force: true });
  }

  clickcompanyOverviewLink() {
    this.elements.companyOverviewLink().click();
  }

  clickcontactIntelLink() {
    this.elements.contactIntelLink().click();
  }

  clicknewsRoomLink() {
    this.elements.newsRoomLink().click();
  }

  clickinvestorsLink() {
    this.elements.investorsLink().click();
  }

  clickcareersLink() {
    this.elements.careersLink().click();
  }

  clickcorporateRespLink() {
    this.elements.corporateRespLink().click();
  }

  clickdiversityIncLink() {
    this.elements.diversityIncLink().click();
  }

  clickpublicPolicyLink() {
    this.elements.publicPolicyLink().click();
  }
}

module.exports = new homePage();
