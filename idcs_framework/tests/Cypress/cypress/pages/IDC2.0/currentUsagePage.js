class currentUsage {

    elements = {

        filterproductsInput: () => cy.get('[intc-id="FilterInput"]'),
        clearFiltersButton: () => cy.get('[intc-id="clearFilterButton"]'),
    }

    typeFilterproducts(filterval) {
        this.elements.filterproductsInput().clear().type(filterval);
    }

    clearFilters() {
        this.elements.clearFiltersButton().click({ force: true });
    }
}

module.exports = new currentUsage();