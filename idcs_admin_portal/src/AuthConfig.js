import { PublicClientApplication, LogLevel, EventType } from '@azure/msal-browser'
import idcConfig from './config/configurator'
import useAppStore from './store/appStore/AppStore'

export const MsalConfig = {
  auth: {
    clientId: idcConfig.REACT_APP_AZURE_CLIENT_ID,
    authority: idcConfig.REACT_APP_AZURE_CLIENT_AUTHORITY,
    redirectUri: '/'
  },
  cache: {
    cacheLocation: 'localStorage', // Configures cache location. "sessionStorage" is more secure, but "localStorage" gives you SSO between tabs.
    storeAuthStateInCookie: false // Set this to "true" if you are having issues on IE11 or Edge
  },
  system: {
    loggerOptions: {
      loggerCallback: (level, message, containsPii) => {
        if (containsPii || idcConfig.REACT_APP_ENABLE_MSAL_LOGGING) {
          return
        }
        switch (level) {
          case LogLevel.Error:
            break
          case LogLevel.Info:
            break
          case LogLevel.Verbose:
            break
          case LogLevel.Warning:
            break
          default:
            break
        }
      }
    }
  }
}

/**
 * Scopes you add here will be prompted for user consent during sign-in.
 * By default, MSAL.js will add OIDC scopes (openid, profile, email) to any login request.
 * For more information about OIDC scopes, visit:
 * https://docs.microsoft.com/en-us/azure/active-directory/develop/v2-permissions-and-consent#openid-connect-scopes
 */
export const LoginRequest = {
  scopes: ['openid', idcConfig.REACT_APP_AZURE_CLIENT_API_SCOPE]
}

export const RedirectLogout = {
  postLogoutRedirectUri: window.location.origin
}

/**
 * MSAL should be instantiated outside of the component tree to prevent it from being re-instantiated on re-renders.
 * For more, visit: https://github.com/AzureAD/microsoft-authentication-library-for-js/blob/dev/lib/msal-react/docs/getting-started.md
 */
export const msalInstance = new PublicClientApplication(MsalConfig)

const loginUser = async () => {
  const accounts = msalInstance.getAllAccounts()
  const activeAccount = msalInstance.getActiveAccount()
  if (accounts.length === 0 || !activeAccount) {
    await msalInstance.loginRedirect(LoginRequest)
  } else {
    msalInstance.setActiveAccount(activeAccount)
    useAppStore.getState().setFirstLoadComplete(true)
  }
}

const init = async () => {
  await msalInstance.initialize()

  msalInstance.addEventCallback(async (event) => {
    if (event.eventType === EventType.HANDLE_REDIRECT_END) {
      await loginUser()
    }
    if (
      (event.eventType === EventType.LOGIN_SUCCESS ||
        event.eventType === EventType.ACQUIRE_TOKEN_SUCCESS ||
        event.eventType === EventType.SSO_SILENT_SUCCESS) &&
      event.payload.account
    ) {
      const { account } = event.payload
      msalInstance.setActiveAccount(account)
      useAppStore.getState().setFirstLoadComplete(true)
    }
  })

  msalInstance
    .handleRedirectPromise()
    .then(() => {})
    .catch(() => {})
}

init()
