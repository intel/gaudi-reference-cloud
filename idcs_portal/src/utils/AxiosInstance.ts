// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import axios, { type AxiosRequestConfig } from 'axios'
import { msalInstance, LoginRequest } from '../AuthConfig'
import { type RedirectRequest, type SilentRequest } from '@azure/msal-browser'
import idcConfig from '../config/configurator'

export const parseJwt = (token: string): any => {
  try {
    return JSON.parse(atob(token.split('.')[1]))
  } catch (e) {
    return null
  }
}

export const AxiosInstance = axios.create({
  timeout: Number(idcConfig.REACT_APP_AXIOS_TIMEOUT),
  headers: {
    'Content-Type': 'application/json',
    Accept: 'application/json'
  }
})

let abortController = new AbortController()

export const resetAbortController = (): void => {
  abortController = new AbortController()
}

export const abortGetRequests = (): void => {
  abortController.abort()
}

export const getAccessToken = async (requestAdditionalSettings: any = null): Promise<string> => {
  try {
    const accessTokenRequest = requestAdditionalSettings
      ? {
          scopes: [idcConfig.REACT_APP_AZURE_CLIENT_API_SCOPE],
          ...requestAdditionalSettings
        }
      : {
          scopes: [idcConfig.REACT_APP_AZURE_CLIENT_API_SCOPE]
        }

    const accessTokenResponse = await msalInstance.acquireTokenSilent(accessTokenRequest as SilentRequest)

    const accessToken = accessTokenResponse.accessToken

    const decodedJwt = parseJwt(accessToken)

    const timeStampInSeconds = Math.floor(Date.now() / 1000)
    if (decodedJwt !== null && decodedJwt.exp < timeStampInSeconds) {
      const response = await msalInstance.ssoSilent(LoginRequest as RedirectRequest)
      return response.accessToken
    }

    return accessToken
  } catch {
    try {
      const response = await msalInstance.ssoSilent(LoginRequest as RedirectRequest)
      return response.accessToken
    } catch (error) {
      // If the ssoSilent fails it means we need interaction, the Auth wrapper will catch that error and start a redirect to the login page
      return ''
    }
  }
}

const disableCacheInAxiosConfig = (config: AxiosRequestConfig): void => {
  config.params = config.params
    ? {
        ...config.params,
        t: new Date().getTime()
      }
    : {
        t: new Date().getTime()
      }
}

AxiosInstance.interceptors.request.use(
  async (config) => {
    const accessToken = await getAccessToken()
    config.headers.Authorization = 'Bearer ' + accessToken
    const shouldUpdateConfig =
      config.method === 'post' ||
      config.method === 'get' ||
      config.method === 'delete' ||
      config.method === 'put' ||
      config.method === 'patch'
    const shouldCancelOnRegionSwitch = config.method === 'get'
    if (shouldUpdateConfig) {
      disableCacheInAxiosConfig(config)
    }
    if (shouldCancelOnRegionSwitch) {
      config.signal = abortController.signal
    }
    return config
  },
  () => {
    throw Error('Cannot get access token')
  }
)
