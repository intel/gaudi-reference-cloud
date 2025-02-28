import TestFilter from "../../../support/testFilter";
const computePage = require("../../../pages/IDC2.0/computePage.js");
const homePage = require("../../../pages/IDC2.0/homePage.js");
const instancePage = require("../../../pages/IDC2.0/instancePage.js");
const balancerPage = require("../../../pages/IDC2.0/loadBalancerPage.js");
var publickey = Cypress.env("publicKey");
import accountKeyPage from "../../../pages/IDC2.0/accountKeyPage";
var regionNum = Cypress.env("region");

TestFilter(["IntelAll", "PremiumAll", "StandardAll", "ependingAll", "EnterpriseAll"], () => {
    describe("Load Balancer verification", () => {
        beforeEach(() => {
            cy.PrepareSession();
            cy.GetSession();
            cy.selectRegion(regionNum);
        });

        afterEach(() => {
            cy.TestClean();
        });

        after(() => {
            cy.TestClean();
        });

        it("9010 | Reserve VM for Load Balancer", function () {
            homePage.hardwareCatalog();
            homePage.clickfourthGenIntelProd();
            cy.wait(1000);
            instancePage.instanceType("small");
            cy.wait(1000);
            instancePage.instanceName("lbinstance");
            cy.wait(1000);
            instancePage.clickCreateKeyButton();
            accountKeyPage.addKeyName("balancer-key");
            accountKeyPage.addKeyContent(publickey);
            instancePage.uploadKey()
            cy.wait(2000);
            computePage.launchInstance();
            cy.wait(40000);
        });

        it("9011 | Create Load balancer with Invalid name", function () {
            homePage.clickLoadbalancerPageButton();
            balancerPage.launchLBfromGrid();
            cy.wait(2000);
            balancerPage.balancerName("$#@#@#@");
            cy.wait(2000);
            cy.contains("Only lower case alphanumeric and hypen(-) allowed for Name:.");
        });

        it("9012 | Create Load balancer with Long name", function () {
            homePage.clickLoadbalancerPageButton();
            balancerPage.launchLBfromGrid();
            cy.wait(2000);
            balancerPage.balancerName("loadbalancerlongnamepreventsnomorethan50character1234");
            balancerPage.getBalancerName().should("have.value", "loadbalancerlongnamepreventsnomorethan50character1");
        });

        it("9013 | Create Load balancer with 1 Source IP as Any", function () {
            homePage.clickLoadbalancerPageButton();
            balancerPage.launchLBfromGrid();
            cy.wait(2000);
            balancerPage.balancerName("lbtest");
            balancerPage.inputSourceIP("any");
            balancerPage.listenerPortInput("8090");
            balancerPage.instancePortInput("8080");
            balancerPage.selectTCP();
            balancerPage.selectAllInstances();
            balancerPage.launchBalancer();
            cy.wait(15000);
        });

        it("9014 | Verify create Load balancer without Source IP is not allowed", function () {
            homePage.clickLoadbalancerPageButton();
            balancerPage.launchLBfromGrid();
            cy.wait(2000);
            balancerPage.balancerName("balancer");
            balancerPage.deleteSourceIp(0);
            balancerPage.listenerPortInput("8443");
            balancerPage.instancePortInput("8080");
            balancerPage.selectTCP();
            balancerPage.selectAllInstances();
            balancerPage.launchBalancer();
            cy.wait(2000);
            cy.contains("At-least one source IP is required.");
        });

        it("9015 | Verify create Load balancer without Listener is not allowed", function () {
            homePage.clickLoadbalancerPageButton();
            balancerPage.launchLBfromGrid();
            cy.wait(2000);
            balancerPage.balancerName("balancer");
            balancerPage.inputSourceIP("any");
            balancerPage.deleteListener(0);
            balancerPage.launchBalancer();
            cy.wait(2000);
            cy.contains("At-least one listener is required.");
            balancerPage.cancelLaunch();
        });

        it("9016 | Create Load balancer with 2 Source IPs", function () {
            homePage.clickLoadbalancerPageButton();
            balancerPage.launchLBfromGrid();
            cy.wait(2000);
            balancerPage.balancerName("mybalancer4");
            balancerPage.addSourceIp();
            cy.get('[intc-id="SourceIPInput"]').eq(0).should("be.visible").clear().type("192.168.1.1");
            cy.wait(2000)
            cy.get('[intc-id="SourceIPInput"]').eq(1).should("be.visible").clear().type("174.160.20.5")
            cy.contains("Up to 20 source IPs max. (18) remaining");
            balancerPage.listenerPortInput("8080");
            balancerPage.instancePortInput("6000");
            balancerPage.selectTCP();
            balancerPage.selectAllInstances();
            balancerPage.launchBalancer();
            cy.wait(30000);
        });

        it("9017 | Create Load balancer without instance selected", function () {
            homePage.clickLoadbalancerPageButton();
            balancerPage.launchLBfromGrid();
            cy.wait(2000);
            balancerPage.balancerName("mybalancer3");
            balancerPage.inputSourceIP("any");
            balancerPage.listenerPortInput("8010");
            balancerPage.instancePortInput("9010");
            balancerPage.selectTCP();
            balancerPage.launchBalancer();
            cy.wait(2000);
            cy.contains("Instances is required");
        });

        it("9018 | Create Load balancer with Duplicate name", function () {
            homePage.clickLoadbalancerPageButton();
            balancerPage.launchLBfromGrid();
            cy.wait(2000);
            balancerPage.balancerName("lbtest");
            balancerPage.inputSourceIP("any");
            balancerPage.listenerPortInput("8090");
            balancerPage.instancePortInput("8080");
            balancerPage.selectHTTP();
            balancerPage.selectAllInstances();
            balancerPage.launchBalancer();
            cy.wait(2000);
            cy.get(".modal-header").should("be.visible")
            cy.contains("Could not create your load balancer");
            cy.get(".btn.btn-primary").contains("Go back").click({ force: true });
        });

        it("9019 | Search Load balancer - Positive", function () {
            homePage.clickLoadbalancerPageButton();
            balancerPage.lbTableIsVisible();
            balancerPage.searchBalancer("lbtest");
            cy.get('[intc-id="lbtestHyperLinkTable"]').should("be.visible");
        });

        it("9020 | Search Load balancer - Negative", function () {
            homePage.clickLoadbalancerPageButton();
            balancerPage.lbTableIsVisible();
            balancerPage.searchBalancer("nonexisting-lb");
            cy.contains("The applied filter criteria did not match any items.");
            balancerPage.clearFilter();
        });

        it("9021 | Verify Load Balancer quota max", function () {
            homePage.clickLoadbalancerPageButton();
            balancerPage.launchLBfromGrid();
            cy.wait(2000);
            balancerPage.balancerName("newload");
            balancerPage.inputSourceIP("any");
            balancerPage.listenerPortInput("8090");
            balancerPage.instancePortInput("8080");
            balancerPage.selectTCP();
            balancerPage.selectAllInstances();
            balancerPage.launchBalancer();
            cy.wait(2000);
            cy.get(".modal-header").should("be.visible").as('modal')
            cy.get("@modal").should("be.visible").contains("Could not create your load balancer");
            cy.get(".btn.btn-primary").contains("Go back").click({ force: true });
        });

        it("9022 | Cancel Reserve Load balancer", function () {
            homePage.clickLoadbalancerPageButton();
            balancerPage.launchLBfromGrid();
            cy.wait(2000);
            balancerPage.balancerName("lb-cancel");
            balancerPage.inputSourceIP("any");
            balancerPage.cancelLaunch();
            cy.wait(2000);
        });

        it("9023 | Verify Load balancer details tab", function () {
            homePage.clickLoadbalancerPageButton();
            balancerPage.lbTableIsVisible();
            cy.wait(2000);
            balancerPage.searchBalancer("lbtest");
            cy.get('[intc-id="lbtestHyperLinkTable"]').should("be.visible").click({ force: true });
            cy.contains("Load Balancer information");
            cy.wait(2000);
        });

        it("9024 | Verify Load balancer Source IPS tab", function () {
            homePage.clickLoadbalancerPageButton();
            balancerPage.lbTableIsVisible();
            cy.wait(2000);
            balancerPage.searchBalancer("lbtest");
            cy.get('[intc-id="lbtestHyperLinkTable"]').should("be.visible").click({ force: true });
            balancerPage.clickSourceIPsTab();
            cy.contains("Source IPs information");
            cy.contains("any");
        });

        it("9025 | Verify Load balancer Listeners tab", function () {
            homePage.clickLoadbalancerPageButton();
            balancerPage.lbTableIsVisible();
            cy.wait(2000);
            balancerPage.searchBalancer("lbtest");
            cy.get('[intc-id="lbtestHyperLinkTable"]').should("be.visible").click({ force: true });
            balancerPage.clickListenersTab();
            cy.contains("roundRobin")
            cy.contains("tcp")
        });

        it("9026 | Verify Load balancer Invalid ip input", function () {
            homePage.clickLoadbalancerPageButton();
            balancerPage.launchLBfromGrid();
            cy.wait(2000);
            balancerPage.balancerName("mybalancer3");
            balancerPage.inputSourceIP("192.90.x..7");
            cy.contains("Invalid IP");
        });

        it("9027 | Verify Load balancer Invalid ports input", function () {
            homePage.clickLoadbalancerPageButton();
            balancerPage.launchLBfromGrid();
            cy.wait(2000);
            balancerPage.balancerName("test-lb");
            balancerPage.inputSourceIP("192.176.90.2");
            balancerPage.listenerPortInput("xyz");
            balancerPage.instancePortInput("xyz");
            cy.get('[intc-id="ListenerPortInvalidMessage"]').should("be.visible").contains("Value less than 1 is not allowed.");
            cy.get('[intc-id="InstancePortInvalidMessage"]').should("be.visible").contains("Value less than 1 is not allowed.");
        });

        it("9028 | Verify Edit Load balancer from table grid (remove Listener and add a new one)", function () {
            homePage.clickLoadbalancerPageButton();
            balancerPage.lbTableIsVisible();
            balancerPage.searchBalancer("mybalancer4");
            cy.get('[intc-id="mybalancer4HyperLinkTable"]').should("be.visible");
            balancerPage.editLb();
            balancerPage.deleteListener(0);
            balancerPage.addListener();
            balancerPage.listenerPortInput("80");
            balancerPage.instancePortInput("9000");
            balancerPage.selectHTTPS();
            balancerPage.saveEdit();
            balancerPage.lbTableIsVisible();
            balancerPage.searchBalancer("mybalancer4");
            cy.get('[intc-id="mybalancer4HyperLinkTable"]').should("be.visible").click({ force: true });
            balancerPage.clickListenersTab();
            cy.contains('8090').should('not.exist');
            cy.contains('9000');
        });

        it("9029 | Verify Edit Load balancer from table grid (remove source IP and set monitor to HTTP)", function () {
            homePage.clickLoadbalancerPageButton();
            balancerPage.lbTableIsVisible();
            balancerPage.searchBalancer("mybalancer4");
            cy.get('[intc-id="mybalancer4HyperLinkTable"]').should("be.visible");
            balancerPage.editLb();
            balancerPage.deleteSourceIp(0);
            balancerPage.addSourceIp();
            cy.wait(2000)
            cy.get('[intc-id="SourceIPInput"]').eq(0).should("be.visible").clear().type("192.168.123.132");
            balancerPage.selectHTTP();
            balancerPage.saveEdit();
            balancerPage.lbTableIsVisible();
            balancerPage.searchBalancer("mybalancer4");
            cy.get('[intc-id="mybalancer4HyperLinkTable"]').should("be.visible").click({ force: true });
            balancerPage.clickSourceIPsTab();
            cy.contains('192.168.1.1').should('not.exist');
            cy.contains('192.168.123.132');
            cy.contains('174.160.20.5');
        });

        it("9030 | Verify user is not allowed to Save Edit with No listeners", function () {
            homePage.clickLoadbalancerPageButton();
            balancerPage.lbTableIsVisible();
            balancerPage.searchBalancer("lbtest");
            cy.get('[intc-id="lbtestHyperLinkTable"]').should("be.visible");
            balancerPage.editLb();
            balancerPage.deleteListener(0);
            balancerPage.saveEdit();
            cy.contains("At-least one listener is required.");
            balancerPage.cancelLaunch();
        });

        it("9031 | Delete Load Balancer from table grid", function () {
            homePage.clickLoadbalancerPageButton();
            balancerPage.lbTableIsVisible();
            balancerPage.searchBalancer("lbtest");
            cy.wait(1000);
            cy.get('[intc-id="lbtestHyperLinkTable"]').should("be.visible")
            balancerPage.deleteLb();
            balancerPage.checkConfirmDeleteModal();
            cy.getValue('[intc-id="deleteConfirmationName"]').then(text => {
                cy.get('[intc-id="NameInput"]').should("be.visible").clear().type(text);
            })
            balancerPage.confirmDeleteLB();
            cy.contains("Load balancer marked for deletion.");
        });

        it("9032 | Cancel Delete Load Balancer from table grid", function () {
            homePage.clickLoadbalancerPageButton();
            balancerPage.lbTableIsVisible();
            balancerPage.searchBalancer("mybalancer4");
            cy.wait(1000);
            cy.get('[intc-id="mybalancer4HyperLinkTable"]').should("be.visible")
            balancerPage.deleteLb();
            balancerPage.checkConfirmDeleteModal();
            balancerPage.cancelDeleteLB();
        });

        it("9033 | Delete Load Balancer from details Action", function () {
            homePage.clickLoadbalancerPageButton();
            balancerPage.lbTableIsVisible();
            balancerPage.searchBalancer("mybalancer4");
            cy.wait(1000);
            cy.get('[intc-id="mybalancer4HyperLinkTable"]').should("be.visible").click({ force: true });
            balancerPage.clickDeleteActionsButton();
            balancerPage.checkConfirmDeleteModal();
            cy.getValue('[intc-id="deleteConfirmationName"]').then(text => {
                cy.get('[intc-id="NameInput"]').should("be.visible").clear().type(text);
            })
            balancerPage.confirmDeleteLB();
        });

        it("9034 | Delete LB instance", function () {
            homePage.clickcomputePageButton();
            cy.wait(1000);
            computePage.searchInstance("lbinstance");
            computePage.deleteInstance();
            computePage.checkConfirmDeleteModal();
            cy.getValue('[intc-id="deleteConfirmationName"]').then(text => {
                cy.get('[intc-id="NameInput"]').should("be.visible").clear().type(text);
            })
            computePage.confirmDeleteInstance();
        });

        it("9035 | Delete LB Keys", function () {
            homePage.clickKeysPageButton();
            cy.deleteFirstKey();
        });
    });
}
);

