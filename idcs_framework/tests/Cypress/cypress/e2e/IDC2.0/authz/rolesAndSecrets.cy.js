import rolesPage from "../../../pages/IDC2.0/rolesPage";
import credencialsPage from "../../../pages/IDC2.0/credencialPage";
import TestFilter from "../../../support/testFilter";
const homePage = require("../../../pages/IDC2.0/homePage");

TestFilter(["IntelAll", "PremiumAll", "ependingAll", "EnterpriseAll"], () => {
    describe("My Roles tab", () => {
        beforeEach(() => {
            cy.PrepareSession()
            cy.GetSession()
        });

        afterEach(() => {
            cy.TestClean()
        })

        after(() => {
            cy.TestClean()
        })

        it("4010 | Create empty role with no assigned permissions", function () {
            homePage.rolesTab();
            rolesPage.addRole();
            rolesPage.roleName("myrole1");
            rolesPage.createRole();
            cy.wait(2000);
            cy.get('[intc-id="myrole1HyperLinkTable"]').should("be.visible");
        });

        it("4011 | Create Role with duplicate name", function () {
            homePage.rolesTab();
            rolesPage.addRole();
            rolesPage.roleName("myrole1");
            rolesPage.createRole();
            cy.wait(2000);
            cy.get(".modal-header").should("be.visible").as('modal')
            cy.get("@modal").should("be.visible").contains("Could not create your role");
            cy.contains("Failed to create role alias already exists. Please try again or contact support if the issue continues.");
            cy.get(".btn.btn-primary").contains("Go back").click({ force: true });
        });

        it("4012 | Create Role using a higher max lenght name", function () {
            homePage.rolesTab();
            rolesPage.addRole();
            rolesPage.roleName("mynewrolehasmorethan50charsacceptedornotthisisavingmore");
            rolesPage.createRole();
            cy.contains("mynewrolehasmorethan50charsacceptedornotthisisavin");
        });

        it("4013 | Delete Role from dropdown actions", function () {
            homePage.rolesTab();
            cy.get('[intc-id="mynewrolehasmorethan50charsacceptedornotthisisavinHyperLinkTable"]').click();
            rolesPage.clickDropdown();
            rolesPage.deleteFromActions();
            rolesPage.checkConfirmDeleteModal();
            cy.getValue('[intc-id="deleteConfirmationName"]').then(text => {
                cy.get('[intc-id="NameInput"]').should("be.visible").clear().type(text);
            })
            rolesPage.confirmDeleteRole();
        });

        it("4014 | Create Role with invalid name", function () {
            homePage.rolesTab();
            rolesPage.addRole();
            rolesPage.roleName("#$%*&*@INVALID");
            cy.contains("Only lower case alphanumeric and hypen(-) allowed for Role Name:.");
        });

        it("4015 | Cancel Create Role creation", function () {
            homePage.rolesTab();
            rolesPage.addRole();
            rolesPage.roleName("myrole3");
            rolesPage.cancelCreateRole();
        });

        it("4016 | Verify Users tab when no Users are associated to the role", function () {
            homePage.rolesTab();
            cy.get('[intc-id="myrole1HyperLinkTable"]').should("be.visible").click({ force: true });
            cy.wait(2000);
            rolesPage.usersTab();
            cy.contains("No users associated with this role");
        });
        /* disable LC rule test
        it("4017 | Create Role with only Lifecycle Rules permissions", function () {
            homePage.rolesTab();
            rolesPage.addRole();
            rolesPage.roleName("lifecyclerules-role");
            cy.wrap('[intc-id="UserRoleCreateServiceTitle-lifecyclerule"]').then((btn) => {
                cy.get(btn).click({ force: true });
            });
            cy.wait(3000);
            cy.get('.form-check-input').contains("Get a lifecycle rule").check({ force: true });
            rolesPage.createRole();
        });
        */
        it("4018 | Cancel Edit Role", function () {
            homePage.rolesTab();
            cy.get('[intc-id="myrole1HyperLinkTable"]').click();
            rolesPage.clickDropdown();
            rolesPage.editFromActions();
            rolesPage.cancelUpdateRole();
        });

        it("4019 | Edit Role to use a different permission", function () {
            homePage.rolesTab();
            cy.get('[intc-id="myrole1HyperLinkTable"]').click({ force: true });
            rolesPage.clickDropdown();
            rolesPage.editFromActions();
            rolesPage.clickFileStorage();
            cy.get('[intc-id="selectAllCheckboxCheckbox"]').check({ force: true });
            rolesPage.saveUpdateRole();
        });

        it("4020 | Search Role by Name - Negative", function () {
            homePage.rolesTab();
            rolesPage.SearchRole("myrole3");
            cy.contains("No roles found")
            rolesPage.clearFilter();
        });

        it("4021 | Search Role by Name - Positive", function () {
            homePage.rolesTab();
            rolesPage.SearchRole("myrole1");
            cy.get('[intc-id="myrole1HyperLinkTable"]').should("be.visible");
            cy.get('[intc-id="defaultHyperLinkTable"]').should("not.exist");
        });

        it("4022 | Verify Default role is available", function () {
            homePage.rolesTab();
            rolesPage.SearchRole("default");
            cy.get('[intc-id="defaultHyperLinkTable"]').should("be.visible").click();
            rolesPage.clickDropdown();
            rolesPage.editFromActions();
            rolesPage.clickObjectPrincipal();
            rolesPage.clickObjectStorage();
            rolesPage.clickObjectRules();
            rolesPage.clickFileStorage();
        });

        it("4023 | Cancel Delete Role", function () {
            homePage.rolesTab();
            cy.get('[intc-id="ButtonTable Delete role"]').eq(1).click();
            rolesPage.checkConfirmDeleteModal();
            cy.getValue('[intc-id="deleteConfirmationName"]').then(text => {
                cy.get('[intc-id="NameInput"]').should("be.visible").clear().type(text);
            })
            rolesPage.cancelDeleteRole();
        });

        it("4024 | Delete Role from table grid", function () {
            homePage.rolesTab();
            cy.get('[intc-id="ButtonTable Delete role"]').eq(1).click();
            rolesPage.checkConfirmDeleteModal();
            cy.getValue('[intc-id="deleteConfirmationName"]').then(text => {
                cy.get('[intc-id="NameInput"]').should("be.visible").clear().type(text);
            })
            rolesPage.confirmDeleteRole();
        });
    });
});

TestFilter(["IntelAll"], () => {
    describe("Credentials tab", () => {
        beforeEach(() => {
            cy.PrepareSession()
            cy.GetSession()
        });

        afterEach(() => {
            cy.TestClean()
        })

        after(() => {
            cy.TestClean()
        })

        it("7020 | Generate Secret with Invalid Name", function () {
            homePage.credentialsTab();
            credencialsPage.generateSecretEmptyView();
            credencialsPage.secretName("@#@$@%@-secret");
            cy.contains("Only lower case alphanumeric and hypen(-) allowed for Secret name:.")
        });

        it("7021 | Cancel Create Client Secret", function () {
            homePage.credentialsTab();
            credencialsPage.generateSecretEmptyView();
            credencialsPage.secretName("mysecret");
            credencialsPage.cancelCreateSecret();
        });

        it("7022 | Create a new valid Client Secret", function () {
            homePage.credentialsTab();
            credencialsPage.generateSecretEmptyView();
            credencialsPage.secretName("mysecret");
            credencialsPage.createSecret();
            cy.get(".modal-header").should("be.visible").as('modal')
            cy.get("@modal").should("be.visible").contains("Client Secret");
            cy.contains("Client Secret generated successfully, it will be visible only once. Make sure to store it properly.")
            credencialsPage.tokenClose();
        });

        it("7023 | Create Secret with duplicate name", function () {
            homePage.credentialsTab();
            credencialsPage.generateSecret();
            credencialsPage.secretName("mysecret");
            credencialsPage.createSecret();
            cy.contains("AppClient Name already Exists, Please use different Name mysecret");
        });

        it("7024 | Delete Client Secret from table grid", function () {
            homePage.credentialsTab();
            credencialsPage.deleteSecret();
            credencialsPage.checkConfirmDeleteModal();
            cy.getValue('[intc-id="deleteConfirmationName"]').then(text => {
                cy.get('[intc-id="NameInput"]').should("be.visible").clear().type(text);
            })
            credencialsPage.confirmDeleteSecret();
        });
    });
});
