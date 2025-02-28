// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import ReactDOM from 'react-dom/client'
import App from './App'
import reportWebVitals from './reportWebVitals'
import { msalInstance } from './AuthConfig'
import ErrorBoundary from './pages/error/ErrorBoundary'
import { ErrorBoundaryLevel } from './utils/Enums'
import { QueryClient, QueryClientProvider } from 'react-query'

const root = ReactDOM.createRoot(document.getElementById('root') as HTMLElement)
const queryClient = new QueryClient()

root.render(
  <React.StrictMode>
    <ErrorBoundary errorBoundaryLevel={ErrorBoundaryLevel.AppLevel}>
      <QueryClientProvider client={queryClient} contextSharing={true}>
        <App msalInstance={msalInstance} />
      </QueryClientProvider>
    </ErrorBoundary>
  </React.StrictMode>
)

// If you want to start measuring performance in your app, pass a function
// to log results (for example: reportWebVitals(console.log))
// or send to an analytics endpoint. Learn more: https://bit.ly/CRA-vitals
reportWebVitals()
