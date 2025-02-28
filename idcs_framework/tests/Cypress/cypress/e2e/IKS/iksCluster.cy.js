import accountKeyPage from "../../pages/IDC2.0/accountKeyPage";
import iksPage from "../../pages/IDC2.0/iksPage";
import TestFilter from "../../support/testFilter";
const instancePage = require("../../pages/IDC2.0/instancePage.js");
const homePage = require("../../pages/IDC2.0/homePage");
var publickey = Cypress.env("publicKey");
var regionNum = Cypress.env("region");

TestFilter(["IntelAll", "PremiumAll"], () => {
  describe("IKS - Cluster reservations", () => {
    beforeEach(() => {
      cy.PrepareSession();
      cy.GetSession();
      cy.selectRegion(regionNum)
    });

    afterEach(() => {
      cy.TestClean();
    });

    after(() => {
      cy.TestClean();
    });

    it("2050 | Create IKS Cluster with invalid name", function () {
      homePage.clickIKSPageButton();
      iksPage.createFirstCluster();
      cy.wait(2000);
      iksPage.typeClusterName("@#$TEST#@");
      cy.contains("Only lower case alphanumeric and hypen(-) allowed for Cluster name:.")
    });

    it("2051 | Create IKS Cluster", function () {
      homePage.clickIKSPageButton();
      iksPage.createFirstCluster();
      cy.wait(2000);
      iksPage.typeClusterName("iksclus01");
      cy.wait(6000);
      iksPage.k8sVersion128();
      iksPage.launchCluster();
      cy.wait(50000);
    });

    it("2052 | Verify add Load Balancer without NodeGroups", function () {
      homePage.clickIKSPageButton();
      iksPage.searchInput("iksclus01");
      cy.get('[intc-id="iksclus01HyperLinkTable"]')
        .contains("iksclus01")
        .click();
      cy.get('[intc-id="Load Balancers (0)Tab"]')
        .contains("Load Balancers (0)")
        .scrollIntoView({ duration: 2000 })
        .click();
      cy.contains('To add a load balancer, please add a worker node group with at least one node first.');
      cy.wait(3000);
    });

    it("2053 | Create Cluster with duplicate name", function () {
      homePage.clickIKSPageButton();
      iksPage.createCluster();
      cy.wait(1000);
      iksPage.typeClusterName("iksclus01");
      cy.wait(1000);
      iksPage.k8sVersion129();
      iksPage.launchCluster();
      cy.wait(9000);
      cy.get(".modal-header")
        .contains("Could not launch your cluster")
        .within(($form) => {
          cy.wrap($form).should("be.visible");
          cy.wait(4000);
        });
      cy.get(".btn.btn-primary").contains("Go back").click({ force: true });
      cy.wait(3000);
      iksPage.typeClusterName("iksclus02");
      iksPage.launchCluster();
      cy.wait(50000);
    });

    it("2053 | Verify add Load Balancer without NodeGroups", function () {
      homePage.clickIKSPageButton();
      iksPage.searchInput("iksclus01");
      cy.get('[intc-id="iksclus01HyperLinkTable"]')
        .contains("iksclus01")
        .click();
      cy.get('[intc-id="Load Balancers (0)Tab"]')
        .contains("Load Balancers (0)")
        .scrollIntoView({ duration: 2000 })
        .click();
      cy.contains('To add a load balancer, please add a worker node group with at least one node first.');
      cy.wait(3000);
    });

    it("2054 | Search for Cluster and Clear filter- Negative", function () {
      homePage.clickIKSPageButton();
      cy.wait(2000);
      iksPage.searchInput("test_cluster")
      cy.wait(5000);
      cy.get('[intc-id="data-view-empty"]')
        .should("be.visible")
        .contains("No clusters found");
      iksPage.clearFilter();
    });

    it("2055 | Verify Cluster details tab", function () {
      homePage.clickIKSPageButton();
      iksPage.searchInput("iksclus01");
      cy.get('[intc-id="iksclus01HyperLinkTable"]')
        .contains("iksclus01")
        .click();
      cy.contains("1.28").should("be.visible");
      cy.wait(3000);
    });

    it("2056 | Add Load Balancer without Worker NodeGroups", function () {
      homePage.clickIKSPageButton();
      iksPage.searchInput("iksclus01");
      cy.get('[intc-id="iksclus01HyperLinkTable"]')
        .contains("iksclus01")
        .click();
      cy.get('[intc-id="Load Balancers (0)Tab"]')
        .contains("Load Balancers (0)")
        .scrollIntoView({ duration: 2000 })
        .click();
      cy.get('[intc-id="data-view-empty"]')
        .should("be.visible")
        .contains("No Load Balancers found")
        .scrollIntoView({ duration: 2000 });
      cy.contains("To add a load balancer, please add a worker node group with at least one node first.")
    });

    it("2057 | Add Worker Node Groups", function () {
      homePage.clickIKSPageButton();
      cy.wait(50000);   // Add wait for Cluster to be Act
      iksPage.searchInput("iksclus01")
      cy.contains("Active")
      cy.get('[intc-id="iksclus01HyperLinkTable"]').contains("iksclus01").click();
      cy.get('[intc-id="Worker Node Groups (0)Tab"]')
        .contains("Worker Node Groups")
        .scrollIntoView({ duration: 2000 })
        .click();
      cy.wait(3000);
      iksPage.addWorkerNodeGroup();
      cy.wait(3000);
      cy.get('[intc-id="checkBoxTable-Nodetype-grid-vm-spr-sml').should("be.visible").check({ force: true })
      iksPage.inputGroupName("group01");
      iksPage.selectNodeQty();
      iksPage.uploadKey();
      accountKeyPage.addKeyName("test-key");
      accountKeyPage.addKeyContent(publickey);
      instancePage.uploadKey();
      cy.wait(2000);
      iksPage.launchGroup();
      cy.wait(30000);
    });

    it("2058 | Verify Security tab - Edit default firewall rule", function () {
      homePage.clickIKSPageButton();
      iksPage.searchInput("iksclus02");
      cy.get('[intc-id="iksclus02HyperLinkTable"]')
        .contains("iksclus02")
        .click();
      cy.get('[intc-id="SecurityTab"]', { timeout: 0 }).should('be.visible').then((tab) => {
        if (tab.length === 0) {
          cy.log("Tab is not visible, skipping the test");
          this.skip();
        } else if (!tab.is(':visible')) {
          cy.log("Tab is not visible, skipping the test");
          this.skip();
        } else {
          iksPage.clickSecurityTab();
          iksPage.editRule();
          iksPage.typeSourceIP("192.134.0.10/32");
          iksPage.selectTCP();
          iksPage.saveRuleEdit();
          cy.wait(3000);
        }
      })
    });

    it("2059 | Add Single Worker Node", function () {
      homePage.clickIKSPageButton();
      iksPage.searchInput("iksclus01");
      cy.get('[intc-id="iksclus01HyperLinkTable"]')
        .contains("iksclus01")
        .click();
      cy.get('[intc-id="Worker Node Groups (1)Tab"]')
        .contains("Worker Node Groups")
        .scrollIntoView({ duration: 2000 })
        .click();
      cy.get('[intc-id="accordion-group01-nodegroup"]')
        .scrollIntoView({ duration: 2000 })
        .click();
      iksPage.addNode();
      iksPage.confirmChangeNodes();
      cy.wait(20000);
    });

    it("2060 | Delete Single Worker Node", function () {
      homePage.clickIKSPageButton();
      cy.wait(20000);   // Add wait for Cluster to be Active
      iksPage.searchInput("iksclus01");
      cy.contains("Active")
      cy.get('[intc-id="iksclus01HyperLinkTable"]')
        .contains("iksclus01")
        .click();
      cy.get('[intc-id="Worker Node Groups (1)Tab"]')
        .contains("Worker Node Groups")
        .scrollIntoView({ duration: 2000 })
        .click();
      iksPage.removeNode();
      iksPage.confirmChangeNodes();
    });

    it("2061 | Add Load Balancer", function () {
      homePage.clickIKSPageButton();
      iksPage.searchInput("iksclus01");
      cy.get('[intc-id="iksclus01HyperLinkTable"]')
        .contains("iksclus01")
        .click();
      cy.get('[intc-id="Load Balancers (0)Tab"]')
        .contains("Load Balancers (0)")
        .scrollIntoView({ duration: 2000 })
        .click();
      cy.get('[intc-id="data-view-empty"]')
        .should("be.visible")
        .contains("No Load Balancers found")
        .scrollIntoView({ duration: 2000 });
      iksPage.addLoadBalancer();
      iksPage.lbInputName("testlb");
      iksPage.portSelect();
      iksPage.portOption80();
      iksPage.publicLB();
      iksPage.launchLB();
      cy.wait(3000);
    });

    it("2062 | Delete Worker Node Group", function () {
      homePage.clickIKSPageButton();
      cy.wait(20000);   // Add wait for Cluster to be Active
      iksPage.searchInput("iksclus01");
      cy.contains("Active")
      cy.get('[intc-id="iksclus01HyperLinkTable"]')
        .contains("iksclus01")
        .click();
      cy.get('[intc-id="Worker Node Groups (1)Tab"]')
        .contains("Worker Node Groups")
        .scrollIntoView({ duration: 2000 })
        .click();
      iksPage.deleteNodeGroup();
      iksPage.checkConfirmDeleteModal();
      cy.getValue('[intc-id="deleteConfirmationName"]').then(text => {
        cy.get('[intc-id="NameInput"]').should("be.visible").clear().type(text);
      })
      iksPage.confirmDeleteGroup();
      cy.wait(40000);
    });

    it("2063 | Upgrade cluster K8s version", function () {
      homePage.clickIKSPageButton();
      iksPage.searchInput("iksclus01")
      cy.contains("Active")
      cy.get('[intc-id="iksclus01HyperLinkTable"]').contains("iksclus01").click();
      cy.wait(2000);
      iksPage.detailsTab()
      iksPage.k8sUpgrade();
      iksPage.confirmUpgrade();
      cy.wait(10000);
    });

    it("2064 | Delete default firewall rule", function () {
      homePage.clickIKSPageButton();
      iksPage.searchInput("iksclus02");
      cy.get('[intc-id="iksclus02HyperLinkTable"]')
        .contains("iksclus02")
        .click();
      cy.get('[intc-id="SecurityTab"]', { timeout: 0 }).should('be.visible').then((tab) => {
        if (tab.length === 0) {
          cy.log("Tab is not visible, skipping the test");
          this.skip();
        } else if (!tab.is(':visible')) {
          cy.log("Tab is not visible, skipping the test");
          this.skip();
        } else {
          iksPage.deleteRule();
          iksPage.checkConfirmDeleteModal();
          cy.getValue('[intc-id="deleteConfirmationName"]').then(text => {
            cy.get('[intc-id="NameInput"]').should("be.visible").clear().type(text);
          });
          iksPage.confirmDeleteRule();
          cy.contains("Only Active security rules can be deleted");
        }
      })
    });

    it("2065| Delete Load Balancer", function () {
      homePage.clickIKSPageButton();
      iksPage.searchInput("iksclus01");
      cy.get('[intc-id="iksclus01HyperLinkTable"]')
        .contains("iksclus01")
        .click();
      cy.get('[intc-id="Load Balancers (1)Tab"]')
        .contains("Load Balancers (1)")
        .scrollIntoView({ duration: 2000 })
        .click();
      iksPage.deleteLB();
      iksPage.checkConfirmDeleteModal();
      cy.getValue('[intc-id="deleteConfirmationName"]').then(text => {
        cy.get('[intc-id="NameInput"]').should("be.visible").clear().type(text);
      })
      iksPage.confirmDeleteLB();
      cy.wait(3000);
    });

    it("2066 | Delete IKS Cluster from Actions", function () {
      homePage.clickIKSPageButton();
      iksPage.searchInput("iksclus02");
      cy.get('[intc-id="iksclus02HyperLinkTable"]')
        .contains("iksclus02")
        .scrollIntoView({ duration: 2000 })
        .click();
      iksPage.actionDeleteDropDown();
      iksPage.checkConfirmDeleteModal();
      cy.getValue('[intc-id="deleteConfirmationName"]').then(text => {
        cy.get('[intc-id="NameInput"]').should("be.visible").clear().type(text);
      })
      iksPage.confirmDelete();
    });

    it("2067 | Delete IKS Cluster from table grid", function () {
      homePage.clickIKSPageButton();
      iksPage.searchInput("iksclus01");
      iksPage.deleteCluster();
      iksPage.checkConfirmDeleteModal();
      cy.getValue('[intc-id="deleteConfirmationName"]').then(text => {
        cy.get('[intc-id="NameInput"]').should("be.visible").clear().type(text);
      })
      iksPage.confirmDelete();
    });

    it("1068 | Delete Test Keys", function () {
      homePage.clickKeysPageButton();
      cy.deleteFirstKey();
    });
  });
});