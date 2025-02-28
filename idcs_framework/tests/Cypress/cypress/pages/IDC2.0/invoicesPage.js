class invoices {

    elements = {
        currentMonthUsageLink: () => cy.get('[intc-id="usagesButton"]'),
        filterProductsInput: () => cy.get('[intc-id="FilterProductsInput"]'),
        startDateInput: () => cy.get('[intc-id="StartDateInput"]'),
        endDateInput: () => cy.get('[intc-id="EndDateInput"]'),
        clearFilterButton: () => cy.get('[intc-id="clearFilterButton"]'),

        invoicesIdSortButton: () => cy.get('[intc-id="Invoice IDsortTableButton"]'),
        billingPeriodSortButton: () => cy.get('[intc-id="Billing periodsortTableButton"]'),
        periodStartSortButton: () => cy.get('[intc-id="Period startsortTableButton"]'),
        periodEndSortButton: () => cy.get('[intc-id="Period endsortTableButton"]'),
        dueDateSortButton: () => cy.get('[intc-id="Due datesortTableButton"]'),
        statusSortButton: () => cy.get('[intc-id="StatussortTableButton"]'),
        totalAmountSortButton: () => cy.get('[intc-id="Total amountsortTableButton"]'),
        amountpaidSortButton: () => cy.get('[intc-id="Amount paidsortTableButton"]'),
        amountDueSortButton: () => cy.get('[intc-id="Amount duesortTableButton"]'),

    }

    ClickcurrentMonthUsageLink() {
        this.elements.currentMonthUsageLink().click({ force: true });
    }

    filterProducts(filterval) {
        this.elements.filterProductsInput().type(filterval);
    }

    startDate(startdate) {
        this.elements.startDateInput().type(startdate);
    }

    endDate(enddate) {
        this.elements.endDateInput().type(enddate);
    }

    clearFilter() {
        this.elements.clearFilterButton().click({ force: true });
    }

    invoicesIdSort() {
        this.elements.invoicesIdSortButton().click({ force: true });
    }

    billingPeriodSort() {
        this.elements.billingPeriodSortButton().click({ force: true });
    }

    periodStartSort() {
        this.elements.periodStartSortButton().click({ force: true });
    }

    periodEndSort() {
        this.elements.periodEndSortButton().click({ force: true });
    }

    dueDateSort() {
        this.elements.dueDateSortButton().click({ force: true });
    }

    statusSort() {
        this.elements.statusSortButton().click({ force: true });
    }

    totalAmountSort() {
        this.elements.totalAmountSortButton().click({ force: true });
    }

    amountpaidSort() {
        this.elements.amountpaidSortButton().click({ force: true });
    }

    amountDueSort() {
        this.elements.amountDueSortButton().click({ force: true });
    }
}

module.exports = new invoices();