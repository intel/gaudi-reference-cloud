// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import SuperComputerHomePage from '../../components/superComputer/superComputerHomePage/SuperComputerHomePage'

const SuperComputerHomePageContainer = () => {
  const initialState = {
    subtitle: 'Train and deploy your AI workloads at scale',
    steps: [
      {
        title: 'Scalable High-Performance Computing',
        description: 'Spin up clusters with hundreds of bare metal nodes optimized for intensive workloads.'
      },
      {
        title: 'Low-Latency, High-Throughput Networking',
        description:
          'Minimize communication overhead between nodes. Ensure efficient data exchange for distributed DL frameworks.'
      },
      {
        title: 'Customizable Node Combinations',
        description:
          'Choose from a variety of CPU and high-bandwidth GPUs to match the computational demands of your AI workloads.'
      },
      {
        title: 'Root Access and Direct Hardware Control',
        description:
          'Direct hardware interaction, allowing you to install custom software, optimize drivers, and leverage specialized libraries'
      }
    ],
    purpose:
      'Manage your cluster seamlessly with Kubernetes for containerized workflows. Gain full root access for ultimate control and flexibility.',
    useCases: [
      'Large Scale Image Recognition: Train complex deep learning models on massive image datasets. Hundreds of bare metal nodes with high-bandwidth GPUs accelerate training times and deliver superior accuracy.',
      'Accelerated Natural Language Processing: Develop cutting-edge NLP models for tasks like sentiment analysis, machine translation, and chatbot development. Leverage the scalable compute power and low-latency network fabric to process massive amounts of text data efficiently.',
      'Real-Time Recommendation Engines: Build highly responsive recommendation systems that personalize user experiences in real-time. Bare metal clusters provide the raw power and control needed to handle large user bases and deliver accurate recommendations with minimal latency.'
    ]
  }

  return (
    <SuperComputerHomePage
      subtitle={initialState.subtitle}
      steps={initialState.steps}
      useCases={initialState.useCases}
      purpose={initialState.purpose}
    />
  )
}

export default SuperComputerHomePageContainer
