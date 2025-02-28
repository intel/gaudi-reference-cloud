class storageLifeCylePage {
    elements = {
        // Add Lifecycle Rule form
        addRule: () => cy.get('[intc-id="btn-ObjectStorageRuleLaunch-navigationBottom Add"]'),
        cancelAddRule: () => cy.get('[intc-id="btn-ObjectStorageRuleLaunch-navigationBottom Cancel"]'),
        ruleNameInput: () => cy.get('[intc-id="NameInput"]'),
        prefix: () => cy.get('[intc-id="PrefixInput"]'),
        deleteMarkerCheckbox: () => cy.get('[intc-id="-Radio-option-DeleteMarker"]'),
        expiryDaysCheckbox: () => cy.get('[intc-id="-Radio-option-ExpiryDays"]'),
        expiryDaysInput: () => cy.get('[intc-id="Input"]'),
        nonExpiryDays: () => cy.get('[intc-id="NoncurrentexpirydaysInput"]'),
        invalidRuleNameMessage: () => cy.get('[intc-id="NameInvalidMessage"]'),

        // Edit Lifecycle Rule form
        saveRule: () => cy.get('[intc-id="btn-ObjectStorageRuleEdit-navigationBottom Save"]'),
        cancelEditRule: () => cy.get('[intc-id="btn-ObjectStorageRuleEdit-navigationBottom Cancel"]'),

        // LifecycleRule tab actions
        rulesTab: () => cy.get('[intc-id="Lifecycle RulesTab"]'),
        createRule: () => cy.get('[intc-id="btn-create-bucket-rule"]'),
        editFromActions: () => cy.get('[intc-id="ButtonTable Edit Rule"]'),
        deleteFromActions: () => cy.get('[intc-id="ButtonTable Delete rule"]'),

        // Delete confirmation
        nameInput: () => cy.get('[intc-id="NameInput"]'),
        confirmDeleteModal: () => cy.get('[intc-id="deleteConfirmModal"]'),
        confirmDeleteButton: () => cy.get('[intc-id="btn-confirm-Deleterule-delete"]'),
        cancelDeleteButton: () => cy.get('[intc-id="btn-confirm-Deleterule-cancel"]'),

        rulesTable: () => cy.get('.tableContainer'),
        emptyTable: () => cy.get('[intc-id="data-view-empty"]'),
    }

    addRule() {
        this.elements.addRule().click();
    }

    addRuleIsEnabled() {
        this.elements.addRule().should("be.enabled");
    }

    rulesTab() {
        this.elements.rulesTab().click({ force: true });
    }

    cancelAddRule() {
        this.elements.cancelAddRule().click();
    }

    ruleNameInput(name) {
        this.elements.ruleNameInput().clear().type(name);
    }

    getRuleName() {
        return this.elements.ruleNameInput();
    }

    inputPrefix(text) {
        this.elements.prefix().clear().type(text);
    }

    deleteMarkerCheck() {
        this.elements.deleteMarkerCheckbox().check()
    }

    uncheckDeleteMarker() {
        this.elements.deleteMarkerCheckbox().uncheck();
    }

    expiryDaysCheck() {
        this.elements.expiryDaysCheckbox().check()
    }

    unCheckExpiryDays() {
        this.elements.expiryDaysCheckbox().uncheck()
    }

    expiryDaysInput(days) {
        this.elements.expiryDaysInput().type(days);
    }

    clearExpiryDays() {
        this.elements.expiryDaysInput().clear();
    }

    nonExpiryDays(days) {
        this.elements.nonExpiryDays().type(days);
    }

    checkInvalidRuleNameMessage() {
        this.elements.invalidRuleNameMessage().should("be.visible");
    }

    saveEditRule() {
        this.elements.saveRule().click({ force: true });
    }

    editRuleIsDisabled() {
        this.elements.editRule().should("be.disabled");
    }

    cancelEditRule() {
        this.elements.cancelEditRule().click();
    }

    createRule() {
        this.elements.createRule().click();
    }

    editRule() {
        this.elements.editFromActions().click();
    }

    deleteRule() {
        return this.elements.deleteFromActions().click();
    }

    deleteNameInput(name) {
        this.elements.nameInput().clear().type(name);
    }

    checkConfirmDeleteModal() {
        this.elements.confirmDeleteModal().should("be.visible");
    }

    confirmDelete() {
        this.elements.confirmDeleteButton().click();
    }

    clearPrefix() {
        this.elements.prefix().clear();
    }

    cancelDelete() {
        this.elements.cancelDeleteButton().click();
    }

    checkRulesTableVisible() {
        this.elements.rulesTable().should("be.visible");
    }

    checkEmptyRulesTable() {
        this.elements.emptyTable().should("be.visible");
    }

    deleteAllLifeCycleRules() {
        this.elements.deleteFromActions().then($items => {
            const remainingItems = $items.length;
            $items[0].click();
            this.confirmDelete();
            cy.wait(4000);
            this.deleteBucketPrincipalConfirm();
            if (remainingItems > 1) {
                this.checkRulesTableVisible();
                this.deleteAllLifeCycleRules();
            } else {
                this.checkEmptyRulesTable();
            }
        });
    }
}

module.exports = new storageLifeCylePage();