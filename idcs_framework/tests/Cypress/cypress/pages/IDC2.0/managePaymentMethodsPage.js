class paymentMethods {

    elements = {
        // main payment methods page
        managePaymentMethodsTittle: () => cy.get('[intc-id="managePaymentMethodsTitle"]'),
        addCardButton: () => cy.get('.btn.btn-outline-primary').contains('Add card'),
        changeCreditCardButton: () => cy.get('[intc-id="btn-managepayment-changecard"]').contains('Change card'),
        viewCreditDetailsButton: () => cy.get('[aria-label="View Credit Details"]').contains('View Credit Details'),
        redeemCouponButton: () => cy.get('[aria-label="Redeem Coupon"]').contains('Redeem Coupon'),


        // Add a credit card page
        cardNumberInput: () => cy.get('[intc-id="CardnumberInput"]'),
        inputMM: () => cy.get('[intc-id="MonthInput"]'),
        inputYY: () => cy.get('[intc-id="YearInput"]'),
        inputCVC: () => cy.get('[intc-id="CVCInput"]'),
        firstNameInput: () => cy.get('[intc-id="FirstNameInput"]'),
        lastNameInput: () => cy.get('[intc-id="LastNameInput"]'),
        emailInput: () => cy.get('[intc-id="EmailInput"]'),
        companyNameInput: () => cy.get('[intc-id="CompanynameInput"]'),
        phoneNumberInput: () => cy.get('[intc-id="PhoneInput"]'),
        countrySelect: () => cy.get('[intc-id="CountrySelect"]'),
        countryUSA: () => cy.get('[intc-id="USOption"]'),
        addressLine1Input: () => cy.get('[intc-id="Addressline1Input"]'),
        addressLine2Input: () => cy.get('[intc-id="Addressline2Input"]'),
        cityInput: () => cy.get('[intc-id="CityInput"]'),
        stateSelect: () => cy.get('[intc-id="StateSelect"]'),
        zipcodeInput: () => cy.get('[intc-id="ZIPcodeInput"]'),

        addCreditCardButton: () => cy.get('[intc-id="btn-credit-AddCreditPayment"]'),
        cancelButton: () => cy.get('[intc-id="btn-credit-CancelCreditCard"]'),
    }

    confirmManagePaymentPage() {
        this.elements.managePaymentMethodsTittle().should("exist");
    }

    clickAddCardButton() {
        this.elements.addCardButton().click({ force: true });
    }

    clickClosePayAlert() {
        this.elements.closePaymentAlert().click({ force: true });
    }

    changeCreditCard() {
        this.elements.changeCreditCardButton().click({ force: true });
    }

    viewCreditDetails() {
        this.elements.viewCreditDetailsButton().click({ force: true });
    }

    redeemCoupon() {
        this.elements.redeemCouponButton().click({ force: true });
    }

    cardNumber(cardNum) {
        this.elements.cardNumberInput().type(cardNum);
    }

    monthInput(month) {
        this.elements.inputMM().type(month);
    }

    yearInput(year) {
        this.elements.inputYY().type(year);
    }

    cvcInput(cvc) {
        this.elements.inputCVC().clear().type(cvc);
    }

    firstName(fname) {
        this.elements.firstNameInput().clear().type(fname);
    }

    lastName(lname) {
        this.elements.lastNameInput().clear().type(lname);
    }

    email(email) {
        this.elements.emailInput().clear().type(email);
    }

    companyName(company) {
        this.elements.companyNameInput().clear().type(company);
    }

    phoneNumber(num) {
        this.elements.phoneNumberInput().clear().type(num);
    }

    country(value) {
        this.elements.countrySelect().select(value);
    }

    selectUSA() {
        this.elements.countryUSA().click({ force: true });
    }

    addressLine1(line1) {
        this.elements.addressLine1Input().clear().type(line1);
    }

    addressLine2(line2) {
        this.elements.addressLine2Input().clear().type(line2);
    }

    city(city) {
        this.elements.cityInput().clear().type(city);
    }

    state(state) {
        this.elements.stateSelect().select(state);
    }

    zipCode(code) {
        this.elements.zipcodeInput().clear().type(code);
    }

    addCreditCard() {
        this.elements.addCreditCardButton().click({ force: true });
    }

    cancelAddCreditCard() {
        this.elements.cancelButton().click({ force: true });
    }
}


module.exports = new paymentMethods();