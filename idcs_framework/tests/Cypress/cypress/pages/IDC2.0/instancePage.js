class instancePage {

    elements = {
        machineImageDropdown: () => cy.get('[intc-id="Machineimage-form-select"]'),
        processorSpecificationLink: () => cy.get('[href="https://ark.intel.com/content/www/us/en/ark/products/231749/intel-xeon-platinum-8444h-processor-45m-cache-2-90-ghz.html"]'),
        processorSupportLink: () => cy.get('[href="https://www.intel.com/content/www/us/en/support/contact-support.html"]'),
        instanceNameInput: () => cy.get('[intc-id="InstancenameInput"]'),
        selectKeys: () => cy.get('[class="list-group-item border-0 pb-0"]'),
        keyCheckBox: () => cy.get(".form-check-input").eq(0),
        createKeyButton: () => cy.get('[intc-id="SelectkeysbtnExtra"]'),
        selectAllKeys: () => cy.get('[intc-id="SelectkeysbtnSelectAll"]'),
        instanceCreateKey: () => cy.get('[intc-id="btn-ssh-createpublickey"]'),
        launchButton: () => cy.get('[intc-id="btn-computelaunch-navigationBottom Launch - singlenode"]'),
        cancelButton: () => cy.get('[intc-id="btn-computelaunch-navigationBottom Cancel - singlenode"]'),
        submitRequestInstanceButton: () => cy.get('[intc-id="btn-computelaunch-navigationBottom Launch instance - singlenode"]'),
        submitPreviewInstanceButton: () => cy.get('[intc-id="btn-computelaunch-navigationBottom Request instance - singlenode"]'),
        oneClickConnect: () => cy.get('[intc-id="Switch"]'),
        CPU: () => cy.get('[intc-id="CPU-radio-select"]'),
        GPU: () => cy.get('[intc-id="GPU-radio-select"]'),
        AI: () => cy.get('[intc-id="AI-radio-select"]'),
        AI_PC: () => cy.get('[intc-id="AI PC-radio-select"]'),

        // 4Gen dropDown instance types
        fourthGenXeon: () => cy.get('[intc-id="checkBoxTable-Instancetype-grid-bm-spr-pl"]'),
        largeVM: () => cy.get('[intc-id="checkBoxTable-Instancetype-grid-vm-spr-lrg"]'),
        mediumVM: () => cy.get('[intc-id="checkBoxTable-Instancetype-grid-vm-spr-med"]'),
        smallVM: () => cy.get('[intc-id="checkBoxTable-Instancetype-grid-vm-spr-sml"]'),
        tinyVM: () => cy.get('[intc-id="checkBoxTable-Instancetype-grid-vm-spr-tiny"]'),

        // Preview instance
        sierraForest: () => cy.get('[intc-id="checkBoxTable-Instancetype-grid-pre-bm-srf-sp"]'),

        // Instance tags
        addTagBtn: () => cy.get('[intc-id="Instancetags-add-dictionary-item"]'),
        removeTag: () => cy.get('[intc-id="Instancetags-remove-dictionary-item-0"]'),
        keyInput: () => cy.get('[intc-id="Instancetags-input-dictionary-Key-0"]'),
        valueInput: () => cy.get('[intc-id="Instancetags-input-dictionary-Key-0"]'),
    }

    checkFirstKey() {
        this.elements.keyCheckBox().click({ force: true });
    }

    oneClick() {
        this.elements.oneClickConnect().check();
    }

    uploadKey() {
        this.elements.instanceCreateKey().scrollIntoView();
        this.elements.instanceCreateKey().click({ force: true });
    }

    checkLaunchIsDisabled() {
        this.elements.launchButton().contains("Launch").should("be.disabled");
    }

    checkLaunchIsEnabled() {
        this.elements.launchButton().contains("Launch").should("be.enabled");
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

    clickAIPC() {
        this.elements.AI_PC().should("be.visible").check({ force: true })
    }

    clickFourthGenXeon() {
        this.elements.fourthGenXeon().should("be.visible").check({ force: true })
    }

    clickLargeVM() {
        this.elements.largeVM().should("be.visible").check({ force: true })
    }

    clickMediumVM() {
        this.elements.mediumVM().should("be.visible").check({ force: true })
    }

    clickSmallVM() {
        this.elements.smallVM().should("be.visible").check({ force: true })
    }

    clickTinyVM() {
        this.elements.tinyVM().should("be.visible").check({ force: true })
    }

    instanceType(vmsize) {
        if (vmsize == 'tiny') {
            this.clickTinyVM();
        } else if (vmsize == 'small') {
            this.clickSmallVM();
        } else if (vmsize == 'medium') {
            this.clickMediumVM();
        } else if (vmsize == 'large') {
            this.clickLargeVM();
        }
    }

    machineImage() {
        this.elements.machineImageDropdown().click({ force: true });
    }

    clickProcessorSpecificationLink() {
        this.elements.processorSpecificationLink().click();
    }

    clickProcessorSupportLink() {
        this.elements.processorSupportLink().click();
    }

    instanceName(name) {
        this.elements.instanceNameInput().scrollIntoView().should('be.visible');
        this.elements.instanceNameInput().clear().type(name);
    }

    clickSelectKeys(key) {
        this.elements.instanceNameInput().scrollIntoView();
        this.elements.selectKeys().select(key);
    }

    clickCreateKeyButton() {
        this.elements.createKeyButton().scrollIntoView();
        this.elements.createKeyButton().click({ force: true });
    }

    clickSelectAllKeysButton() {
        this.elements.selectAllKeys().click();
    }

    clickLaunchButton() {
        this.elements.launchButton().click({ force: true });
    }

    clickCancelButton() {
        this.elements.cancelButton().scrollIntoView();
        this.elements.cancelButton().click({ force: true });
    }

    checkSierraForest() {
        this.elements.sierraForest().scrollIntoView();
        this.elements.sierraForest().check();
    }

    submitRequestInstance() {
        this.elements.submitRequestInstanceButton().scrollIntoView();
        this.elements.submitRequestInstanceButton().click({ force: true });
    }

    submitPreviewInstance() {
        this.elements.submitPreviewInstanceButton().should("be.visible");
        this.elements.submitPreviewInstanceButton().scrollIntoView();
        this.elements.submitPreviewInstanceButton().click({ force: true });
    }

    addTag() {
        this.elements.addTagBtn().click();
    }

    removeTag() {
        this.elements.removeTag().click({ force: true });
    }

    inputKey(name) {
        this.elements.keyInput().scrollIntoView().should("be.visible");
        this.elements.keyInput().clear().type(name);
    }

    inputValue(name) {
        this.elements.valueInput().scrollIntoView().should("be.visible");;
        this.elements.valueInput().clear().type(name);
    }
}

module.exports = new instancePage();