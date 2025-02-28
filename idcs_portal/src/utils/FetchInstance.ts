// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { msalInstance, LoginRequest } from '../AuthConfig'
import { type RedirectRequest, type SilentRequest } from '@azure/msal-browser'
import idcConfig from '../config/configurator'

export class FetchErrorResponse extends Error {
  response: any
  constructor(response: any) {
    super(`Request failed with status code ${response.status}`)
    this.name = 'FetchError'
    this.response = response
    Object.setPrototypeOf(this, FetchErrorResponse.prototype)
  }
}

export const parseJwt = (token: string): any => {
  try {
    return JSON.parse(atob(token.split('.')[1]))
  } catch (e) {
    return null
  }
}

export const FetchInstance = async (url: string, requestBody: any, method: string): Promise<any> => {
  const accessToken = await getAccessToken()
  return await fetch(url, {
    method,
    body: JSON.stringify(requestBody),
    headers: {
      'Content-Type': 'application/json',
      Authorization: 'Bearer ' + accessToken
    }
  })
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
