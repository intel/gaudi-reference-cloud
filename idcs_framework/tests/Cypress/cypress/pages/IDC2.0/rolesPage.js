class roles {
    elements = {
        // Account Access Management
        addRoleBtn: () => cy.get('[intc-id="btn-navigate-create-new-role"]'),
        editRoleBtn: () => cy.get('[intc-id="ButtonTable Edit role"]'),
        deleteRoleBtn: () => cy.get('[intc-id="ButtonTable Delete role"]'),
        searchRoleInput: () => cy.get('[intc-id="searchRoles"]'),
        rolesNameInput: () => cy.get('[intc-id="RoleNameInput"]'),
        objectPrincipal: () => cy.get('[intc-id="UserRoleUpdateServiceTitle-principal"]'),
        objectStorage: () => cy.get('[intc-id="UserRoleUpdateServiceTitle-objectstorage"]'),
        objectRules: () => cy.get('[intc-id="UserRoleUpdateServiceTitle-lifecyclerule"]'),
        fileStorage: () => cy.get('[intc-id="UserRoleUpdateServiceTitle-filestorage"]'),
        createRoleBtn: () => cy.get('[intc-id="UserRoleCreate-btn-navigationBottom Create"]'),
        cancelCreateBtn: () => cy.get('[intc-id="UserRoleCreate-btn-navigationBottom Cancel"]'),
        clearFilter: () => cy.get('[intc-id="ClearfiltersEmptyViewButton"]'),
        rolesTab: () => cy.get('.tap-inactive.nav-link').contains("Roles"),
        permissionTab: () => cy.get('[intc-id="PermissionTab"]'),
        usersTab: () => cy.get('.tap-inactive.nav-link').contains("Users"),
        saveUpdate: () => cy.get('[intc-id="UserRoleUpdate-btn-navigationBottom Update"]'),
        cancelUpdateBtn: () => cy.get('[intc-id="UserRoleUpdate-btn-navigationBottom Cancel"]'),

        // Edit roles
        actionsDropdown: () => cy.get('[intc-id="myReservationActionsDropdownButton"]').contains("Actions"),
        editActionSelect: () => cy.get('[intc-id="myReservationActionsDropdownItemButton0"]').contains("Edit"),
        deleteActionSelect: () => cy.get('[intc-id="myReservationActionsDropdownItemButton1"]').contains("Delete"),

        // Delete reconfirm dialog
        nameInput: () => cy.get('[intc-id="NameInput"]'),
        confirmDeleteModal: () => cy.get('[intc-id="deleteConfirmModal"]'),
        cancelDelete: () => cy.get('[intc-id="btn-confirm-Deleterole-cancel"]').contains('Cancel'),
        confirmDelete: () => cy.get('[intc-id="btn-confirm-Deleterole-delete"]').contains('Delete'),
    };

    addRole() {
        this.elements.addRoleBtn().click({ force: true });
    }

    roleName(name) {
        this.elements.rolesNameInput().should("be.visible").clear().type(name);
    }

    editRole() {
        this.elements.editRoleBtn().click({ force: true });
    }

    deleteRole() {
        this.elements.deleteRoleBtn().click({ force: true });
    }

    saveUpdateRole() {
        this.elements.saveUpdate().scrollIntoView();
        this.elements.saveUpdate().click({ force: true });
    }

    cancelUpdateRole() {
        this.elements.cancelUpdateBtn().click({ force: true });
    }

    cancelCreateRole() {
        this.elements.cancelCreateBtn().click({ force: true });
    }

    SearchRole(name) {
        this.elements.searchRoleInput().clear().type(name);
    }

    checkConfirmDeleteModal() {
        this.elements.confirmDeleteModal().should("be.visible");
    }

    cancelDeleteRole() {
        this.elements.cancelDelete().click({ force: true });
    }

    confirmDeleteRole() {
        this.elements.confirmDelete().click({ force: true });
    }

    clickObjectPrincipal() {
        this.elements.objectPrincipal().click({ force: true });
        this.elements.objectPrincipal().click({ force: true });
    }

    clickObjectStorage() {
        this.elements.objectStorage().click({ force: true });
        this.elements.objectStorage().click({ force: true });
    }

    clickObjectRules() {
        this.elements.objectRules().click({ force: true });
        this.elements.objectRules().click({ force: true });
    }

    clickFileStorage() {
        this.elements.fileStorage().click({ force: true });
        this.elements.fileStorage().click({ force: true });
    }

    clearFilter() {
        this.elements.clearFilter().click({ force: true });
    }

    createRole() {
        this.elements.createRoleBtn().scrollIntoView();
        this.elements.createRoleBtn().click({ force: true });
    }

    rolesTab() {
        this.elements.rolesTab().click({ force: true });
    }

    permissionTab() {
        this.elements.permissionTab().click({ force: true });
    }

    usersTab() {
        this.elements.usersTab().should("be.visible").click({ force: true });
    }

    clickDropdown() {
        this.elements.actionsDropdown().click({ force: true });
    }

    editFromActions() {
        this.elements.editActionSelect().click({ force: true });
    }

    deleteFromActions() {
        this.elements.deleteActionSelect().click({ force: true });
    }
}

module.exports = new roles();
