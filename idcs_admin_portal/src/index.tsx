import React from 'react'
import { createRoot } from 'react-dom/client'
import App from './App'
import * as serviceWorker from './serviceWorker'
import { msalInstance } from './AuthConfig'
import ErrorBoundary from './pages/error/ErrorBoundary'
import { ErrorBoundaryLevel } from './utility/Enums'

const rootElement = document.getElementById('root')
const root = createRoot(rootElement as HTMLElement)

// Triggers Frontend trace.
// WebTracer("idcs-frontend-console");

root.render(
  <React.StrictMode>
    <ErrorBoundary errorBoundaryLevel={ErrorBoundaryLevel.AppLevel}>
      <App msalInstance={msalInstance}/>
    </ErrorBoundary>
  </React.StrictMode>
)

// If you want your app to work offline and load faster, you can change
// unregister() to register() below. Note this comes with some pitfalls.
// Learn more about service workers: https://bit.ly/CRA-PWA
serviceWorker.unregister()
