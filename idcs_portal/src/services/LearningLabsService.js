// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import idcConfig from '../config/configurator'
import { AxiosInstance } from '../utils/AxiosInstance'

class LearningLabsService {
  getImageFromText(params) {
    const route = `${idcConfig.REACT_APP_API_LEARNING_LABS_SERVICE}/generate_image?prompts=${params.prompts}&height=${params.height}&width=${params.width}&num_inference_steps=${params.num_inference_steps}&guidance_scale=${params.guidance_scale}&batch_size=${1}&negative_prompts=${params.negative_prompts}&seed=${params.seed}&num_images_per_prompt=${1}`
    return AxiosInstance.get(route, {
      responseType: 'blob'
    })
  }

  async getMessageFromText(params) {
    const route = `${idcConfig.REACT_APP_API_LEARNING_LABS_SERVICE}/${params.model}/generate`
    const requestBody = {
      inputs: params.inputs,
      parameters: {
        model: params.model,
        prompt: params.prompt,
        max_new_tokens: params.max_new_tokens,
        temperature: params.temperature,
        top_k: params.top_k,
        top_p: params.top_p
      }
    }
    return AxiosInstance.post(route, requestBody)
  }
}

export default new LearningLabsService()
