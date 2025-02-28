import TestFilter from "../../../support/testFilter";
const paymentMethods = require("../../../pages/IDC2.0/managePaymentMethodsPage");
const homePage = require("../../../pages/IDC2.0/homePage");
import "cypress-grep";

TestFilter(["PremiumAll"], () => {
  describe("Payment methods", function () {
    beforeEach(() => {
      cy.PrepareSession();
      cy.GetSession();
    });

    afterEach(() => {
      cy.TestClean();
    });

    after(() => {
      cy.TestClean();
    });

    it("1132 | Add a Credit Card / Change credit card", function () {
      homePage.clickuserAccountTab();
      homePage.clickPaymentMethods();
      cy.get('.btn.btn-outline-primary')
        .invoke("text")
        .then((buttonText) => {
          if (buttonText.includes("Add card")) {
            // This will be invoked when adding credit card for the first time
            paymentMethods.clickAddCardButton();
            paymentMethods.cardNumber("4242424242424242");
            paymentMethods.monthInput("07");
            paymentMethods.yearInput("28");
            paymentMethods.cvcInput("123");
            paymentMethods.firstName("visa");
            paymentMethods.lastName("User");
            paymentMethods.email("abe@test.com");
            paymentMethods.country("US");
            paymentMethods.addressLine1("xyz street");
            paymentMethods.city("LA");
            paymentMethods.state("CA");
            paymentMethods.zipCode("12345");
            paymentMethods.addCreditCard();
            cy.wait(5000);
            cy.contains("Success: Credit card added to account").should(
              "be.visible"
            );
            cy.contains("visa User").should("be.visible"); // Validating name
            cy.contains("Ends in 4242").should("be.visible"); // Validating card number
            cy.contains("Exp 7/2028").should("be.visible"); // Validating expiry month and year
          } else if (buttonText.includes("Change card")) {
            // This will be invoked if a credit card is already added
            paymentMethods.changeCreditCard();
            paymentMethods.cardNumber("6011111111111117");
            paymentMethods.monthInput("06");
            paymentMethods.yearInput("27");
            paymentMethods.cvcInput("432");
            paymentMethods.firstName("discover");
            paymentMethods.lastName("User");
            paymentMethods.email("discover@test.com");
            paymentMethods.country("US");
            paymentMethods.addressLine1("xyz street");
            paymentMethods.city("LA");
            paymentMethods.state("CA");
            paymentMethods.zipCode("12345");
            paymentMethods.addCreditCard();
            cy.wait(4000);
            cy.contains("Credit card added to account").should(
              "be.visible"
            );
            cy.contains("discover User").should("be.visible"); // Validating name
            cy.contains("Ends in 1117").should("be.visible"); // Validating card number
            cy.contains("Exp 6/2027").should("be.visible"); // Validating expiry month and year
          }
        });
    });

    it("1132 | Verify card number max length", function () {
      homePage.clickuserAccountTab();
      homePage.clickPaymentMethods();
      paymentMethods.changeCreditCard();
      paymentMethods.cardNumber("1234123412341234123");
    });

    it("1133 | Verify credit card invalid Month field with value more than 12 months", () => {
      homePage.clickuserAccountTab();
      homePage.clickPaymentMethods();
      paymentMethods.changeCreditCard();
      paymentMethods.monthInput("13");
      cy.contains("Invalid Month").should("be.visible");
    });

    it("1134 | Verify credit card invalid Month field with characters", function () {
      homePage.clickuserAccountTab();
      homePage.clickPaymentMethods();
      paymentMethods.changeCreditCard();
      paymentMethods.monthInput("w");
      cy.contains("Invalid Month").should("be.visible");
    });

    it("1135 | Verify credit card required Month field", function () {
      homePage.clickuserAccountTab();
      homePage.clickPaymentMethods();
      paymentMethods.changeCreditCard();
      paymentMethods.monthInput("23");
      cy.get('[intc-id="MonthInput"]').clear();
      cy.contains("Month is required").should("be.visible");
    });

    it("1136 | Verify credit card required Year field", function () {
      homePage.clickuserAccountTab();
      homePage.clickPaymentMethods();
      paymentMethods.changeCreditCard();
      paymentMethods.yearInput("13");
      cy.get('[intc-id="YearInput"]').clear();
      cy.contains("Year is required").should("be.visible");
    });

    it("1137 | Verify credit card invalid Year field", function () {
      homePage.clickuserAccountTab();
      homePage.clickPaymentMethods();
      paymentMethods.changeCreditCard();
      paymentMethods.yearInput("22");
      cy.contains("Invalid Year.").should("be.visible");
    });

    it("1138 | Verify credit card required CVC field", function () {
      homePage.clickuserAccountTab();
      homePage.clickPaymentMethods();
      paymentMethods.changeCreditCard();
      paymentMethods.cvcInput("12");
      cy.get('[intc-id="CVCInput"]').clear();
      cy.contains("CVC is required").should("be.visible");
    });

    it("1139| Verify first name with special chars and numbers in add credit card", function () {
      homePage.clickuserAccountTab();
      homePage.clickPaymentMethods();
      paymentMethods.changeCreditCard();
      paymentMethods.firstName("123#$!");
      cy.contains("Only letters from A-Z or a-z are allowed.").should(
        "be.visible"
      );
    });

    it("1140 | Verify first name required message in add credit card", function () {
      homePage.clickuserAccountTab();
      homePage.clickPaymentMethods();
      paymentMethods.changeCreditCard();
      paymentMethods.firstName("asd");
      cy.get('[intc-id="FirstNameInput"]').clear();
      cy.contains("First Name is required").should("be.visible");
    });

    it("1141 | Verify last name with special chars and numbers in add credit card", function () {
      homePage.clickuserAccountTab();
      homePage.clickPaymentMethods();
      paymentMethods.changeCreditCard();
      paymentMethods.lastName("123#$!");
      cy.contains("Only letters from A-Z or a-z are allowed.").should(
        "be.visible"
      );
    });

    it("1142 | Verify last name required message in add credit card", function () {
      homePage.clickuserAccountTab();
      homePage.clickPaymentMethods();
      paymentMethods.changeCreditCard();
      paymentMethods.lastName("asd");
      cy.get('[intc-id="LastNameInput"]').clear();
      cy.contains("Last Name is required").should("be.visible");
    });

    it("1143 | Verify company name in add credit card", function () {
      homePage.clickuserAccountTab();
      homePage.clickPaymentMethods();
      paymentMethods.changeCreditCard();
      paymentMethods.companyName("Intel Corporation");
    });

    it("1144 | Verify phone number in add credit card", function () {
      homePage.clickuserAccountTab();
      homePage.clickPaymentMethods();
      paymentMethods.changeCreditCard();
      paymentMethods.phoneNumber("12345635235235");
    });

    it("1145 | Verify phone number in add credit card", function () {
      homePage.clickuserAccountTab();
      homePage.clickPaymentMethods();
      paymentMethods.changeCreditCard();
      paymentMethods.phoneNumber("12345635235235");
    });

    it("1146 | Verify select country (USA) in add credit card", function () {
      homePage.clickuserAccountTab();
      homePage.clickPaymentMethods();
      paymentMethods.changeCreditCard();
      paymentMethods.country("US");
      cy.get('[intc-id="CountrySelect"]').should("have.value", "US");
      cy.get('[intc-id="CountrySelect"]').should(
        "contain",
        "United States of America"
      );
    });

    it("1147 | Verify select country (Spain) in add credit card", function () {
      homePage.clickuserAccountTab();
      homePage.clickPaymentMethods();
      paymentMethods.changeCreditCard();
      paymentMethods.country("Spain");
      cy.get('[intc-id="CountrySelect"]').should("have.value", "ES");
      cy.get('[intc-id="CountrySelect"]').should("contain", "Spain");
    });

    it("1148 | Verify Address line1 in add credit card", function () {
      homePage.clickuserAccountTab();
      homePage.clickPaymentMethods();
      paymentMethods.changeCreditCard();
      paymentMethods.addressLine1("Address line 1 test");
    });

    it("1149 | Verify Address line1 required error message in add credit card", function () {
      homePage.clickuserAccountTab();
      homePage.clickPaymentMethods();
      paymentMethods.changeCreditCard();
      paymentMethods.addressLine1("12");
      cy.get('[intc-id="Addressline1Input"]').clear();
      cy.contains("Address line 1 is required").should("be.visible");
    });

    it("1150 | Verify Address line2 in add credit card", function () {
      homePage.clickuserAccountTab();
      homePage.clickPaymentMethods();
      paymentMethods.changeCreditCard();
      paymentMethods.addressLine2("address line 2");
    });

    it("1151 | Verify city required in add credit card", function () {
      homePage.clickuserAccountTab();
      homePage.clickPaymentMethods();
      paymentMethods.changeCreditCard();
      paymentMethods.city("12");
      cy.get('[intc-id="CityInput"]').clear();
      cy.contains("City is required").should("be.visible");
    });

    it("1152 | Verify zipcode allows chars in add credit card", function () {
      homePage.clickuserAccountTab();
      homePage.clickPaymentMethods();
      paymentMethods.changeCreditCard();
      paymentMethods.zipCode("12345");
      cy.get('[intc-id="ZIPcodeInput"]').should("have.value", "12345");
      paymentMethods.zipCode("adawe");
      cy.get('[intc-id="ZIPcodeInput"]').should("have.value", "adawe");
    });

    it("1153 | Select Oregon state in add a credit card page", function () {
      homePage.clickuserAccountTab();
      homePage.clickPaymentMethods();
      paymentMethods.changeCreditCard();
      paymentMethods.country("US");
      paymentMethods.state("OR");
      cy.get('[intc-id="CountrySelect"]').should("have.value", "US");
      cy.get('[intc-id="CountrySelect"]').should(
        "contain",
        "United States of America"
      );
      cy.get('[intc-id="StateSelect"]').should("have.value", "OR");
      cy.get('[intc-id="StateSelect"]').should("contain", "Oregon");
    });

    it("1154 | Verify cancel button in add a credit card page", function () {
      homePage.clickuserAccountTab();
      homePage.clickPaymentMethods();
      paymentMethods.changeCreditCard();
      paymentMethods.cancelAddCreditCard();
    });

    it("1155 | Verify JCB card not allowed", function () {
      homePage.clickuserAccountTab();
      homePage.clickPaymentMethods();
      paymentMethods.changeCreditCard();
      paymentMethods.cardNumber("3530111333300000");
      cy.contains("Card is not allowed.");
    });

    it("1156 | Verify Dinners Club card not allowed", function () {
      homePage.clickuserAccountTab();
      homePage.clickPaymentMethods();
      paymentMethods.changeCreditCard();
      paymentMethods.cardNumber("30569309025904");
      cy.contains("Card is not allowed.");
    });

    it("1157 | Verify Amex requires 4 digits CVC ", function () {
      homePage.clickuserAccountTab();
      homePage.clickPaymentMethods();
      paymentMethods.changeCreditCard();
      paymentMethods.cardNumber("378282246310005");
      paymentMethods.monthInput("10");
      paymentMethods.yearInput("29");
      paymentMethods.cvcInput("123");
      paymentMethods.firstName("Amex Test")
      cy.contains("Invalid CVC");
    });

    it("1158 | Verify view credit details button", function () {
      homePage.clickuserAccountTab();
      homePage.clickPaymentMethods();
      paymentMethods.viewCreditDetails();
      cy.contains("Cloud Credits");
    });

    it("1159 | Verify Mastercard change", function () {
      homePage.clickuserAccountTab();
      homePage.clickPaymentMethods();
      paymentMethods.changeCreditCard();
      paymentMethods.cardNumber("5555555555554444");
      paymentMethods.monthInput("07");
      paymentMethods.yearInput("28");
      paymentMethods.cvcInput("4324");
      paymentMethods.firstName("mastercard");
      paymentMethods.lastName("User");
      paymentMethods.email("master@test.com");
      paymentMethods.country("US");
      paymentMethods.addressLine1("Trinity Ave");
      paymentMethods.city("New York City");
      paymentMethods.state("NY");
      paymentMethods.zipCode("10041");
      paymentMethods.addCreditCard();
      cy.wait(4000);
      cy.contains("Credit card added to account").should(
        "be.visible"
      );
      cy.contains("mastercard User").should("be.visible"); // Validating name
      cy.contains("Ends in 4444").should("be.visible"); // Validating card number
      cy.contains("Exp 7/2028").should("be.visible"); // Validating expiry month and year
    });

    it("1160 | Verify redeem coupon button", function () {
      homePage.clickuserAccountTab();
      homePage.clickPaymentMethods();
      paymentMethods.redeemCoupon();
      cy.contains("Redeem coupon");
    });
  });
});
