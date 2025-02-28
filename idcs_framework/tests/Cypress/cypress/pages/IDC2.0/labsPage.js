class labsPage {

    elements = {
        //Labs
        promntInput: () => cy.get('[intc-id="PromptInput"]'),
        generateButton: () => cy.get('[intc-id="btn-textToImage-navigationButtons Generate"]'),
        optionsButton: () => cy.get('[intc-id="btn-textToImage-navigationButtons Options"]'),
        resetButton: () => cy.get('[intc-id="btn-textToImage-navigationButtons Reset"]'),
        saveButton: () => cy.get('[intc-id="btn-textToImage-modalButtons Save"]'),
        resetModalButton: () => cy.get('[intc-id="btn-textToImage-modalButtons Reset"]'),
        seedInput: () => cy.get('[intc-id="SeedInput"]'),
        negativePromnt: () => cy.get('[intc-id="NegativePromptTextArea"]'),
    }

    // Labs page

    promntInput(name) {
        this.elements.promntInput().scrollIntoView();
        this.elements.promntInput().clear().type(name);
    }

    generateImage() {
        this.elements.generateButton().scrollIntoView();
        this.elements.generateButton().click({ force: true });
    }

    clickOptions() {
        this.elements.optionsButton().click({ force: true });
    }

    clickReset() {
        this.elements.resetButton().click({ force: true });
    }

    seedInput(name) {
        this.elements.seedInput().clear().type(name);
    }

    saveModal() {
        this.elements.saveButton().click({ force: true });
    }

    resetModal() {
        this.elements.resetModalButton().click({ force: true });
    }

    negativePromntInput(name) {
        this.elements.negativePromnt().clear().type(name);
    }
}

module.exports = new labsPage();