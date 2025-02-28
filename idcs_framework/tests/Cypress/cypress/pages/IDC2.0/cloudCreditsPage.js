class cloudCredits {

    elements = {
        // main page
        redeemCouponButton: () => cy.get('[intc-id="redeemCouponButton"]'),
        remainingCredits: () => cy.get('[intc-id="remainingCreditsLabel"]'),
        usedCredits: () => cy.get('[intc-id="usedCreditsLabel"]'),
        filterCreditsInput: () => cy.get('[intc-id="FiltercreditsInput"]'),
        credentialTypesortButton: () => cy.get('[intc-id="Credit TypesortTableButton"]'),
        obtainedOnsortButton: () => cy.get('[intc-id="Obtained onsortTableButton"]'),
        expirationDatesortButton: () => cy.get('[intc-id="Expiration datesortTableButton"]'),
        totalCreditAmountsortButton: () => cy.get('[intc-id="Total credit amountsortTableButton"]'),
        amountUsedsortButton: () => cy.get('[intc-id="Amount usedsortTableButton"]'),
        amountRemainingsortButton: () => cy.get('[intc-id="Amount remainingsortTableButton"]'),

        // After reddem coupon
        couponCodeInput: () => cy.get('[intc-id="CouponCodeInput"]'),
        redeemButton: () => cy.get('[intc-id="btn-managecouponcode-Redeem"]'),
        cancelCouponCodeButton: () => cy.get('[intc-id="btn-managecouponcode-RedeemCancel"]'),
    }

    remainingCredits() {
        this.elements.remainingCredits().click({ force: true });
    }

    usedCredits() {
        this.elements.usedCredits().click({ force: true });
    }

    redeemCoupon() {
        this.elements.redeemCouponButton().click({ force: true });
    }

    filterCredits(filterval) {
        this.elements.filterCreditsInput().type(filterval);
    }

    credentialTypesort() {
        this.elements.credentialTypesortButton().click({ force: true });
    }

    obtainedOnsort() {
        this.elements.obtainedOnsortButton().click({ force: true });
    }

    expirationDatesort() {
        this.elements.expirationDatesortButton().click({ force: true });
    }

    totalCreditAmountsort() {
        this.elements.totalCreditAmountsortButton().click({ force: true });
    }

    amountUsedsort() {
        this.elements.amountUsedsortButton().click({ force: true });
    }

    amountRemainingsort() {
        this.elements.amountRemainingsortButton().click({ force: true });
    }

    typeCouponCode(code) {
        this.elements.couponCodeInput().type(code);
    }

    clickRedeemButton() {
        this.elements.redeemButton().click({ force: true });
    }

    clickCancelCouponCodeButton() {
        this.elements.cancelCouponCodeButton().click({ force: true });
    }
}

module.exports = new cloudCredits();