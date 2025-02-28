class learningPage {

    elements = {
        trainingTitle: () => cy.get('[intc-id="trainingTitle"]'),
        launchJupyterLab: () => cy.get('[intc-id="btn-select-lab"]').contains("Launch JupyterLab"),
        openTraining: () => cy.get('[intc-id="btn_open_training"]'),

        // AI with Intel Gaudi 2 Accelerator
        startingIntelGaudi: () => cy.get('[intc-id="btn-training-select Getting Started With Intel® Gaudi®"]'),
        stableDiffusion: () => cy.get('[intc-id="btn-training-select Inference with Stable Diffusion v2.1"]'),
        retrievalAugmentedGaudi: () => cy.get('[intc-id="btn-training-select Retrieval Augmented Generation (RAG) using Intel® Gaudi®"]'),
        llama2FineTuning: () => cy.get('[intc-id="btn-training-select Llama 2 Fine Tuning and Inference with Hugging Face"]'),
        howToMigrationTool: () => cy.get('[intc-id="btn-training-select How to use the GPU Migration Tool"]'),

        // C ++ SYCL trainings
        essentialsOfSYCL: () => cy.get('[intc-id="btn-training-select Essentials of SYC"]'),
        performancePortability: () => cy.get('[intc-id="btn-training-select Performance, Portability and Productivity"]'),
        openMPOffloadBasics: () => cy.get('[intc-id="btn-training-OpenMP® Offload Basics"]'),
        migrateFromCUDAtoCSYCL: () => cy.get('[intc-id="btn-training-select Migrate from CUDA® to C++ with SYCL®"]'),
        introToGPUOptimization: () => cy.get('[intc-id="btn-training-select Introduction to GPU Optimization"]'),

        // AI with Max Series GPU
        AIKitXGBoostPredictive: () => cy.get('[intc-id="btn-training-select AI Kit XGBoost Predictive Modeling"]'),
        heterogeneousProgrammingNumba: () => cy.get('[intc-id="btn-training-select Heterogeneous Programming Using Data Parallel Extension for Numba® for AI and HPC"]'),
        machineLearningUsingOneAPI: () => cy.get('[intc-id="btn-training-select Machine Learning Using oneAPI"]'),
        genAIPlayground: () => cy.get('[intc-id="h6-btn-training-GenAI Playground"]'),
        pyTorsh24GPUs: () => cy.get('[intc-id="btn-training-select PyTorch on Intel® GPUs"]'),
        textToImageStableDiff: () => cy.get('[intc-id="btn-training-select Text-to-Image with Stable Diffusion"]'),
        imageToImageStableDiff: () => cy.get('[intc-id="btn-training-select Image-to-Image Generation with Stable Diffusion"]'),
        simpleLLMInference: () => cy.get('[intc-id="btn-training-select Simple LLM Inference: Playing with Language Models"]'),
        LLMFineTunningWithQLoRA: () => cy.get('[intc-id="btn-training-select LLM Fine-tuning with QLoRA"]'),
        retrievalAugmentedGeneration: () => cy.get('[intc-id="btn-training-select Retrieval Augmented Generation (RAG) with LangChain"]'),
        gemmaModelFineTunning: () => cy.get('[intc-id="btn-training-select Gemma Model Fine-tuning using SFT and LoRA"]'),
        optimizeCodeGeneration: () => cy.get('[intc-id="btn-training-select Optimize Code Generation with LLMs"]'),

        // Rendering Toolkit
        intelRenderingToolkit: () => cy.get('[intc-id="btn-training-select Intel® Rendering Toolkit Interactive Learning Path"]'),
        // Quatum computing
        quantumComputing: () => cy.get('[intc-id="btn-training-select Introduction to Quantum Computing Applications with the Intel Quantum SDK"]'),
        // Training details page
        launchJupyterNotebook: () => cy.get('.btn.btn-primary').contains("Launch Jupyter notebook"),
    }

    clickTrainingTitle() {
        this.elements.trainingTitle().click();
    }

    clickOpenTraining() {
        this.elements.openTraining().click();
    }

    launchJupyterLabHome() {
        this.elements.launchJupyterLab().click({ force: true });
    }

    clickAIKitXGBoostPredictive() {
        this.elements.AIKitXGBoostPredictive().click();
    }

    clickHeterogeneousProgrammingNumba() {
        this.elements.heterogeneousProgrammingNumba().click();
    }

    clickmachineLearningUsingOneAPI() {
        this.elements.machineLearningUsingOneAPI().click();
    }

    clickGenAIPlayground() {
        this.elements.genAIPlayground().click();
    }

    clickPyTorsh24() {
        this.elements.pyTorsh24GPUs().click();
    }

    clickEssentialsOfSYCL() {
        this.elements.essentialsOfSYCL().click();
    }

    clickPerformancePortability() {
        this.elements.performancePortability().click();
    }

    clickOpenMPOffloadBasics() {
        this.elements.openMPOffloadBasics().click();
    }

    clickMigrateFromCUDAtoCSYCL() {
        this.elements.migrateFromCUDAtoCSYCL().click();
    }

    clickTextToImageStableDiff() {
        this.elements.textToImageStableDiff().click();
    }

    clickImageToImageStableDiff() {
        this.elements.imageToImageStableDiff().click();
    }

    clickSimpleLLMInference() {
        this.elements.simpleLLMInference().click();
    }

    clickIntelRenderingToolkit() {
        this.elements.intelRenderingToolkit().click();
    }

    clickLLMFineTunningWithQLoRA() {
        this.elements.LLMFineTunningWithQLoRA().click();
    }

    clickRetrievalAugmentedGeneration() {
        this.elements.retrievalAugmentedGeneration().click();
    }

    clickGemmaModelFineTunning() {
        this.elements.gemmaModelFineTunning().click();
    }

    clickOptimizeCodeGeneration() {
        this.elements.optimizeCodeGeneration().click();
    }

    clickIntroToGPUOptimization() {
        this.elements.introToGPUOptimization().click();
    }

    clickStartingWithGaudi() {
        this.elements.startingIntelGaudi().click();
    }

    clickStableDiffusion() {
        this.elements.stableDiffusion().click();
    }

    clickRetrievalAugmentedGaudi() {
        this.elements.retrievalAugmentedGaudi().click();
    }

    clickLlama2FineTuning() {
        this.elements.llama2FineTuning().click();
    }

    clickHowToMigrationTool() {
        this.elements.howToMigrationTool().click();
    }

    clickLaunchJupyterNotebook() {
        this.elements.launchJupyterNotebook().click();
    }
}

module.exports = new learningPage();