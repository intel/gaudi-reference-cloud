// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import idcConfig from '../config/configurator'
import { FetchErrorResponse, FetchInstance } from '../utils/FetchInstance'
import useUserStore from '../store/userStore/UserStore'

function isValidJSON(str) {
  try {
    JSON.parse(str)
    return true
  } catch (error) {
    return false
  }
}

function isFinalMessage(streamedJSON) {
  try {
    if (streamedJSON.result.response.details.finish_reason === 'FINISH_REASON_EOS_TOKEN') {
      return true
    } else {
      return false
    }
  } catch (error) {
    return false
  }
}

class MaasService {
  async generateStream(params, onChunkReceived, delayMs = 25) {
    const url = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/maas/generatestream`
    const cloudAccountNumber = useUserStore.getState().user.cloudAccountNumber

    // Create Body to fetch
    const requestBody = {
      cloudAccountId: cloudAccountNumber,
      model: params.model,
      request: {
        params: {
          maxNewTokens: params.maxNewTokens,
          temperature: params.temperature
        },
        prompt: params.prompt
      },
      productName: params.productName,
      productId: params.productId
    }
    const response = await FetchInstance(url, requestBody, 'POST')

    if (!response.body) {
      throw new Error('ReadableStream is not supported.')
    }

    if (!response.ok) {
      const data = await response.json()
      throw new FetchErrorResponse({
        data: data.error,
        status: response.status,
        statusText: response.statusText
      })
    }

    const reader = response.body.getReader()
    const decoder = new TextDecoder()
    const delay = (ms) => new Promise((resolve) => setTimeout(resolve, ms))

    const readStream = async () => {
      const { done, value } = await reader.read()
      if (done) return

      const chunk = decoder.decode(value, { stream: true })
      let text = ''
      if (isValidJSON(chunk)) {
        const streamedJSON = JSON.parse(chunk)
        if (!isFinalMessage(streamedJSON)) text = streamedJSON.result.response.token.text
      } else {
        if (chunk.includes('\n')) {
          const jsonArray = chunk.split('\n')
          jsonArray.forEach((textJSON) => {
            if (isValidJSON(textJSON)) {
              const streamedJSON = JSON.parse(textJSON)
              if (!isFinalMessage(streamedJSON)) text = text + streamedJSON.result.response.token.text
            }
          })
        }
      }
      await delay(delayMs)
      onChunkReceived(text)
      await readStream()
    }

    await readStream()
  }
}

export default new MaasService()
