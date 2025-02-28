import TestFilter from "../../../support/testFilter";
const homePage = require("../../../pages/IDC2.0/homePage");
const learningPage = require("../../../pages/IDC2.0/learningPage");
const labsPage = require("../../../pages/IDC2.0/labsPage");

TestFilter(["IntelAll", "PremiumAll", "StandardAll", "ependingAll", "EnterpriseAll"], () => {
  describe("Learning validation", () => {
    beforeEach(() => {
      cy.PrepareSession();
      cy.GetSession();
      cy.viewport(1920, 1080);
    });

    afterEach(() => {
      cy.TestClean();
    });

    after(() => {
      cy.TestClean();
    });

    it("1210 | Access How to use the GPU Migration Tool", function () {
      homePage.learning();
      cy.get('[intc-id="title-How to use the GPU Migration Tool"]').scrollIntoView().click({ force: true });
      cy.wait(1000);
      cy.get('[intc-id="h1-title-How to use the GPU Migration Tool"]').should("be.visible");
      cy.contains("Intel Gaudi PyTorch");
    });

    it("1211 | Verify Llama 2 Fine Tuning and Inference Training", function () {
      homePage.learning();
      cy.get('[intc-id="title-Llama 2 Fine Tuning and Inference with Hugging Face"]').scrollIntoView().click({ force: true });
      cy.wait(1000);
      cy.get('[intc-id="h1-title-Llama 2 Fine Tuning and Inference with Hugging Face"]').should("be.visible");
    });

    it("1212 | Verify Text to Image Stable Diffusion training", function () {
      homePage.learning();
      cy.get('[intc-id="title-Text-to-Image with Stable Diffusion"]').scrollIntoView().click({ force: true });
      cy.wait(1000);
      cy.get('[intc-id="h1-title-Text-to-Image with Stable Diffusion"]').should("be.visible");
    });

    it("1213 | Verify Essentials of SYCL training", function () {
      homePage.learning();
      cy.get('[intc-id="title-Essentials of SYCL"]').scrollIntoView().click({ force: true });
      cy.wait(1000);
      cy.get('[intc-id="h1-title-Essentials of SYCL"]').should("be.visible");
    });

    it("1214 | Verify Getting started with Intel Gaudi training", function () {
      homePage.learning();
      cy.get('[intc-id="title-Getting Started With Intel® Gaudi®"]').click({ force: true });
      cy.wait(1000);
      cy.get('[intc-id="h1-title-Getting Started With Intel® Gaudi®"]').should("be.visible");
      cy.contains("Intel Gaudi PyTorch");
    });

    it("1215 | Verify Image to Image Generation training - Launch Jupyter Notebook", function () {
      homePage.learning();
      cy.get('[intc-id="title-Image-to-Image Generation with Stable Diffusion"]').scrollIntoView().click({ force: true });
      cy.wait(1000);
      cy.get('[intc-id="h1-title-Image-to-Image Generation with Stable Diffusion"]').should("be.visible");
      cy.contains("PyTorch Optimizations from Intel");
    });

    it("1216 | Verify Inference with Stable Diffusion v2.1 - Launch Jupyter Notebook", function () {
      homePage.learning();
      cy.get('[intc-id="title-Inference with Stable Diffusion v2.1"]').scrollIntoView().click({ force: true });
      cy.wait(1000);
      cy.get('[intc-id="h1-title-Inference with Stable Diffusion v2.1"]').should("be.visible");
      cy.contains("Intel Gaudi Software");
    });

    it("1217 | Verify LLM Fine-tuning with QLoRA", function () {
      homePage.learning();
      cy.get('[intc-id="title-LLM Fine-tuning with QLoRA"]').scrollIntoView().click({ force: true });
      cy.wait(1000);
      cy.get('[intc-id="h1-title-LLM Fine-tuning with QLoRA"]').should("be.visible");;
    });

    it("1218 | Verify Labs page", function () {
      homePage.labs();
      cy.get('[intc-id="btn-training-select Text-to-Image with Stable Diffusion"]').scrollIntoView().click({ force: true });
      cy.wait(1000);
      cy.get('[intc-id="TextToImageTitle"]').should("be.visible");
      labsPage.promntInput("Cars flying in the space");
      labsPage.generateImage();
    });

  });
}
);
