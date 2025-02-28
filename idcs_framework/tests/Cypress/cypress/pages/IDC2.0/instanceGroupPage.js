class instanceGroupPage {

    elements = {
        backtoInstanceTypesButton: () => cy.get('[intc-id="navigationTop ⟵ Back"]'),
        instanceFamilyDropdown: () => cy.get('[intc-id="family:Select"]'),
        instanceTypeDropdown: () => cy.get('[intc-id="Type: *form-select"]'),
        machineImageDropdown: () => cy.get('[data-bs-toggle="dropdown"]'),
        gaudiProductPageURL: () => cy.get('[href="https://habana.ai/products/gaudi2/"]'),
        developerGaudiURL: () => cy.get('[href="https://developer.habana.ai/"]'),
        groupNameInput: () => cy.get('[intc-id="Groupname:Input"]'),
        selectKeys: () => cy.get('[class="list-group-item border-0 pb-0"]'),
        createKeyButton: () => cy.get('[intc-id="SelectkeysbtnExtra"]'),
        refreshList: () => cy.get('[intc-id="SelectkeysbtnRefresh"]'),
        launchButton: () => cy.get('[intc-id="btn-computelaunch-navigationBottom Launch - cluster"]'),
        cancelButton: () => cy.get('[intc-id="btn-computelaunch-navigationBottom Cancel - cluster"]'),
        compareInstanceTypesButton: () => cy.get('[intc-id="Type:btnLabel"]'),

        // Gaudi2 dropDown instance types
        gaudiCluster4Nodes: () => cy.get('[intc-id="selected-option 4 Node cluster of 8 Gaudi2® HL-225H mezzanine cards with 3rd Gen Xeon® Platinum 8380 Processors"]'),
    }

    clickbacktoInstanceTypesButton(){
        this.elements.backtoInstanceTypesButton().click();
    }

    instanceFamily(index){
        this.elements.instanceFamilyDropdown().select(index);
    }

    clickgaudiCluster4Nodes(){
        this.elements.instanceTypeDropdown().click({force:true});
    }

    machineImage(){
        this.elements.machineImageDropdown().click({force:true});
    }

    clickDeveloperGaudiURL(){
        this.elements.developerGaudiURL().click();
    }

    clickGaudiProductPageURL(){
        this.elements.gaudiProductPageURL().click();
    }

    instanceGroupName(name){
        this.elements.groupNameInput().clear().type(name);
    }

    clickSelectKeys(key){
        this.elements.selectKeys().select(key);
    }

    clickCreateKeyButton(){
        this.elements.createKeyButton().click();
    }

    clickRefreshListButton(){
        this.elements.refreshList().click();
    }

    clickLaunchButton(){
        this.elements.launchButton().click({force:true});
    }

    clickCancelButton(){
        this.elements.cancelButton().click({force:true});
    }

}

module.exports = new instanceGroupPage();