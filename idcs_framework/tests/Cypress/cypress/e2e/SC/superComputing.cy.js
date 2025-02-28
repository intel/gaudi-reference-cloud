import accountKeyPage from "../../pages/IDC2.0/accountKeyPage";
import superComputingPage from "../../pages/IDC2.0/superComputingPage";
import TestFilter from "../../support/testFilter";
const instancePage = require("../../pages/IDC2.0/instancePage.js");
const homePage = require("../../pages/IDC2.0/homePage");
var publickey = Cypress.env("publicKey");

TestFilter(["PremiumAll"], () => {
    describe("Supercomputing cluster reservation", () => {
        beforeEach(() => {
            cy.PrepareSession();
            cy.GetSession();
        });

        afterEach(() => {
            cy.TestClean();
        });

        after(() => {
            cy.TestClean();
        });

        it("7051 | Create Super Computing Cluster with a General Compute Node group", function () {
            homePage.clickSuperComputingPage();
            cy.wait(2000);
            cy.contains("No clusters found").should("be.visible");
            cy.wait(2000);
            superComputingPage.createCluster();
            superComputingPage.addNodeGroup();
            superComputingPage.selectComputeInstance();
            cy.get('[intc-id="selected-option 4th Generation Intel® Xeon® Scalable processors (8468)"]').click();
            superComputingPage.inputNodes("1")
            superComputingPage.typeClusterName("mysc-cluster");
            superComputingPage.k8sVersion("1.29");
            superComputingPage.inputVolumeSize("2");
            instancePage.clickCreateKeyButton();
            accountKeyPage.addKeyName("test-key");
            accountKeyPage.addKeyContent(publickey);
            instancePage.uploadKey()
            cy.wait(2000);
            superComputingPage.launchCluster();
            cy.wait(50000);
        });

        it("7052 | Create Cluster with duplicate name", function () {
            homePage.clickSuperComputingPage();
            cy.wait(2000);
            superComputingPage.createCluster();
            superComputingPage.typeClusterName("mysc-cluster");
            superComputingPage.k8sVersion("1.28");
            superComputingPage.inputVolumeSize("2");
            cy.wait(9000);
            cy.get(".modal-header")
                .contains("Could not launch your cluster")
                .within(($form) => {
                    cy.wrap($form).should("be.visible");
                    cy.wait(4000);
                });
            cy.get(".btn.btn-primary").contains("Go back").click({ force: true });
            cy.wait(3000);
            superComputingPage.typeClusterName("mysc-cluster2");
            cy.wait(2000);
            superComputingPage.launchCluster();
            cy.wait(50000);
        });

        it("7053 | Search for Cluster and Clear filter- Negative", function () {
            homePage.clickSuperComputingPage();
            cy.wait(2000);
            superComputingPage.searchInput("test_cluster")
            cy.wait(5000);
            cy.get('[intc-id="data-view-empty"]')
                .should("be.visible")
                .contains("No clusters found");
            superComputingPage.clearFilter();
        });

        it("7054 | Verify SC Security tab - Edit default firewall rule", function () {
            homePage.clickSuperComputingPage();
            superComputingPage.searchInput("mysc-cluster");
            cy.get('[intc-id="mysc-clusterHyperLinkTable"]')
                .contains("mysc-cluster")
                .click();
            superComputingPage.clickSecurityTab();
            superComputingPage.editRule();
            superComputingPage.typeSourceIP("192.134.0.10/32");
            superComputingPage.selectTCP();
            superComputingPage.saveRuleEdit();
            cy.wait(3000);
        });

        it("7055 | Cancel Launch SC cluster", function () {
            homePage.clickSuperComputingPage();
            cy.wait(2000);
            superComputingPage.createFirstCluster();
            cy.wait(2000);
            superComputingPage.selectComputeInstance();
            cy.get('[intc-id="selected-option 4th Generation Intel® Xeon® Scalable processors (8468)"]').click();
            superComputingPage.inputNodes("1")
            superComputingPage.typeClusterName("mysc-cluster");
            superComputingPage.k8sVersion("1.29");
            superComputingPage.inputVolumeSize("2");
            cy.wait(2000);
            superComputingPage.cancelCluster();
        });

        it("7056 | Delete SC default firewall rule", function () {
            homePage.clickSuperComputingPage();
            superComputingPage.searchInput("mysc-cluster");
            cy.get('[intc-id="mysc-clusterHyperLinkTable"]')
                .contains("mysc-cluster")
                .click();
            superComputingPage.clickSecurityTab();
            superComputingPage.deleteRule();
            superComputingPage.checkConfirmDeleteModal();
            cy.getValue('[intc-id="deleteConfirmationName"]').then(text => {
                cy.get('[intc-id="NameInput"]').should("be.visible").clear().type(text);
            });
            superComputingPage.confirmDeleteRule();
            cy.contains("Only Active security rules can be deleted");
        });

        it("7057 | Add Worker Node Groups", function () {
            homePage.clickSuperComputingPage();
            cy.wait(2000);
            superComputingPage.searchInput("mysc-cluster")
            cy.get('[intc-id="mysc-clusterHyperLinkTable"]').contains("mysc-cluster").click();
            cy.get('[intc-id="Worker Node Groups (1)Tab"]')
                .contains("Worker Node Groups")
                .scrollIntoView({ duration: 2000 })
                .click();
            cy.wait(3000);
            superComputingPage.computeNodeRadio();
            superComputingPage.addNodeGroup();
            cy.wait(3000);
            superComputingPage.inputGroupName("group01");
            superComputingPage.nodesQuantity("1");
            superComputingPage.launchGroup();
            cy.wait(40000);
        });

        it("7058 | Verify Storage tab", function () {
            homePage.clickSuperComputingPage();
            cy.wait(2000);
            superComputingPage.searchInput("mysc-cluster");
            cy.get('[intc-id="mysc-clusterHyperLinkTable"]')
                .contains("mysc-cluster")
                .click();
            cy.wait(2000);
            superComputingPage.storageTab();
            cy.contains("2000GB");
            cy.contains("weka");
            cy.wait(50000);
        });

        it("7059 | Add Single Worker Node", function () {
            homePage.clickSuperComputingPage();
            cy.wait(2000);
            superComputingPage.searchInput("mysc-cluster");
            cy.get('[intc-id="mysc-clusterHyperLinkTable"]')
                .contains("mysc-cluster")
                .click();
            cy.get('[intc-id="Worker Node Groups (2)Tab"]')
                .contains("Worker Node Groups")
                .scrollIntoView({ duration: 2000 })
                .click();
            cy.wait(3000);
            superComputingPage.addNode();
            superComputingPage.confirmChangeNodes();
            cy.wait(50000);
        });

        it("7060 | Verify Cluster details tab", function () {
            homePage.clickSuperComputingPage();
            cy.wait(2000);
            superComputingPage.searchInput("mysc-cluster");
            cy.get('[intc-id="mysc-clusterHyperLinkTable"]')
                .contains("mysc-cluster")
                .click();
            cy.contains("You are already running the latest version of Kubernetes").should("be.visible");
            cy.wait(3000);
        });

        it("7061 | Delete Single Worker Node", function () {
            homePage.clickSuperComputingPage();
            cy.wait(2000);
            superComputingPage.searchInput("mysc-cluster");
            cy.get('[intc-id="mysc-clusterHyperLinkTable"]')
                .contains("mysc-cluster")
                .click();
            cy.get('[intc-id="Worker Node Groups (2)Tab"]')
                .contains("Worker Node Groups")
                .scrollIntoView({ duration: 2000 })
                .click();
            cy.wait(3000);
            superComputingPage.removeNode();
            superComputingPage.confirmChangeNodes();
            cy.wait(40000);
        });

        it("7062 | Delete Worker Node Group", function () {
            homePage.clickSuperComputingPage();
            cy.wait(2000);
            superComputingPage.searchInput("mysc-cluster");
            cy.get('[intc-id="mysc-clusterHyperLinkTable"]')
                .contains("mysc-cluster")
                .click();
            cy.get('[intc-id="Worker Node Groups (2)Tab"]')
                .contains("Worker Node Groups")
                .scrollIntoView({ duration: 2000 })
                .click();
            superComputingPage.deleteNodeGroup();
            superComputingPage.checkConfirmDeleteModal();
            cy.getValue('[intc-id="deleteConfirmationName"]').then(text => {
                cy.get('[intc-id="NameInput"]').should("be.visible").clear().type(text);
            })
            superComputingPage.confirmDeleteGroup();
            cy.wait(40000);
        });

        it("7063 | Upgrade cluster K8s version", function () {
            homePage.clickSuperComputingPage();
            cy.wait(2000);
            superComputingPage.searchInput("mysc-cluster2")
            cy.get('[intc-id="mysc-cluster2HyperLinkTable"]').contains("mysc-cluster2").click();
            superComputingPage.detailsTab();
            cy.wait(6000);
            superComputingPage.k8sUpgrade();
            superComputingPage.confirmUpgrade();
            cy.wait(3000);
        });

        it("7064 | Add Load Balancer", function () {
            homePage.clickSuperComputingPage();
            cy.wait(2000);
            superComputingPage.searchInput("mysc-cluster");
            cy.get('[intc-id="mysc-clusterHyperLinkTable"]')
                .contains("mysc-cluster")
                .click();
            cy.get('[intc-id="Load Balancers (0)Tab"]')
                .contains("Load Balancers (0)")
                .scrollIntoView({ duration: 2000 })
                .click();
            cy.get('[intc-id="data-view-empty"]')
                .should("be.visible")
                .contains("No Load Balancers found")
                .scrollIntoView({ duration: 2000 });
            superComputingPage.addLoadBalancer();
            superComputingPage.lbInputName("testlb");
            superComputingPage.portSelect();
            superComputingPage.portOption80();
            superComputingPage.selectLBType();
            superComputingPage.publicLB();
            superComputingPage.launchLB();
            cy.wait(3000);
        });

        it("7065| Delete Load Balancer", function () {
            homePage.clickSuperComputingPage();
            cy.wait(2000);
            superComputingPage.searchInput("mysc-cluster");
            cy.get('[intc-id="mysc-clusterHyperLinkTable"]')
                .contains("mysc-cluster")
                .click();
            cy.get('[intc-id="Load Balancers (1)Tab"]')
                .contains("Load Balancers (1)")
                .scrollIntoView({ duration: 2000 })
                .click();
            superComputingPage.deleteLB();
            superComputingPage.checkConfirmDeleteModal();
            cy.getValue('[intc-id="deleteConfirmationName"]').then(text => {
                cy.get('[intc-id="NameInput"]').should("be.visible").clear().type(text);
            })
            superComputingPage.confirmDeleteLB();
            cy.wait(3000);
        });

        it("7066 | Delete SC Cluster from Actions", function () {
            homePage.clickSuperComputingPage();
            cy.wait(2000);
            superComputingPage.searchInput("mysc-cluster2");
            cy.get('[intc-id="mysc-cluster2HyperLinkTable"]')
                .contains("mysc-cluster2")
                .scrollIntoView({ duration: 2000 })
                .click();
            superComputingPage.actionsDropDown();
            superComputingPage.actionDelete();
            superComputingPage.checkConfirmDeleteModal();
            cy.getValue('[intc-id="deleteConfirmationName"]').then(text => {
                cy.get('[intc-id="NameInput"]').should("be.visible").clear().type(text);
            })
            superComputingPage.confirmDelete();
            cy.wait(20000);
        });

        it("7067 | Delete SC Cluster from table grid", function () {
            homePage.clickSuperComputingPage();
            cy.wait(2000);
            superComputingPage.searchInput("mysc-cluster");
            superComputingPage.deleteCluster();
            superComputingPage.checkConfirmDeleteModal();
            cy.getValue('[intc-id="deleteConfirmationName"]').then(text => {
                cy.get('[intc-id="NameInput"]').should("be.visible").clear().type(text);
            })
            superComputingPage.confirmDelete();
            cy.wait(6000);
        });
    });
});
