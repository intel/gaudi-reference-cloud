import React from 'react'
import { BrowserRouter } from 'react-router-dom'
import './App.scss'
import Routing from './containers/routing/Routing'
import { MsalProvider } from '@azure/msal-react'
import AccessControlWrapper from './utility/wrapper/AccessControlWrapper'
import AuthWrapper from './utility/wrapper/AuthWrapper'
import { AppRolesEnum } from './utility/Enums'
import 'bootstrap/dist/js/bootstrap.bundle.min.js'
import AuthenticationSpinner from './utility/authenticationSpinner/AuthenticationSpinner'
import { type IPublicClientApplication } from '@azure/msal-browser'
import useAppStore from './store/appStore/AppStore'
import DarkModeContainer from './utility/darkMode/DarkModeContainer'
import NoAccessError from './pages/error/AccessDenied'
// For production mode, devtools is disabled

interface AppProps {
  msalInstance: IPublicClientApplication
  children?: React.ReactNode
}

const App = ({ msalInstance }: AppProps): React.ReactElement => {
  const firstLoadComplete = useAppStore((state) => state.firstLoadComplete)

  const content = <Routing />

  if (!firstLoadComplete) {
    return (
      <>
        <DarkModeContainer />
        <AuthenticationSpinner />
      </>
    )
  }

  return (
    <MsalProvider instance={msalInstance}>
      <DarkModeContainer />
      <AuthWrapper>
        <AccessControlWrapper allowedRoles={[AppRolesEnum.GlobalAdmin]} renderNoAccess={() => <NoAccessError />}>
          <BrowserRouter>
            {content}
          </BrowserRouter>
        </AccessControlWrapper>
      </AuthWrapper>
    </MsalProvider>
  )
}

export default App
