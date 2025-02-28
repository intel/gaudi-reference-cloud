class credentialPage {
    elements = {
        // Account Access Management
        generateSecretEmpty: () => cy.get('[intc-id="GenerateClientSecretEmptyViewButton"]'),
        generateSecret: () => cy.get('.btn.btn-primary').contains("Generate client secret", { matchCase: false }),
        createSecret: () => cy.get('[intc-id="btn-user-credential-navigationBottom Generate Secret"]'),
        cancelCreateSecret: () => cy.get('[intc-id="btn-user-credential-navigationBottom Cancel"]'),
        tokenClose: () => cy.get('[intc-id="TokenClose"]'),
        secretInputName: () => cy.get('[intc-id="SecretnameInput"]'),
        deleteSecret: () => cy.get('[intc-id="ButtonTable Delete client secret"]').contains("Delete"),

        // Delete reconfirm dialog
        nameInput: () => cy.get('[intc-id="NameInput"]'),
        confirmDeleteModal: () => cy.get('[intc-id="deleteConfirmModal"]'),
        cancelDelete: () => cy.get('[intc-id="btn-confirm-Deleteclient secret-cancel"]').contains('Cancel'),
        confirmDelete: () => cy.get('[intc-id="btn-confirm-Deleteclient secret-delete"]').contains('Delete'),
    };

    generateSecretEmptyView() {
        this.elements.generateSecretEmpty().click({ force: true });
    }

    generateSecret() {
        this.elements.generateSecret().click({ force: true });
    }

    createSecret() {
        this.elements.createSecret().click({ force: true });
    }

    secretName(name) {
        this.elements.secretInputName().should("be.visible").clear().type(name);
    }

    deleteSecret() {
        this.elements.deleteSecret().click({ force: true });
    }

    tokenClose() {
        this.elements.tokenClose().click({ force: true });
    }

    cancelCreateSecret() {
        this.elements.cancelCreateSecret().click({ force: true });
    }

    checkConfirmDeleteModal() {
        this.elements.confirmDeleteModal().should("be.visible");
    }

    cancelDeleteSecret() {
        this.elements.cancelDelete().click({ force: true });
    }

    confirmDeleteSecret() {
        this.elements.confirmDelete().click({ force: true });
    }
}

module.exports = new credentialPage();
