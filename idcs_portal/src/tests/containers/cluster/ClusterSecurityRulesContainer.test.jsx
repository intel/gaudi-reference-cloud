// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { render, screen } from '@testing-library/react'
import AuthWrapper from '../../../utils/authWrapper/AuthWrapper'
import { Router, Routes, Route, useSearchParams } from 'react-router-dom'
import ClusterSecurityRulesContainer from '../../../containers/cluster/ClusterSecurityRulesContainer'
import { mockIntelUser } from '../../mocks/authentication/authHelper'
import { act } from 'react'
import { mockGetSecurityRules, mockSecurityRules } from '../../mocks/cluster/Cluster'
import { clearAxiosMock } from '../../mocks/utils'
import { createMemoryHistory } from 'history'
import idcConfig from '../../../config/configurator'
import useClusterStore from '../../../store/clusterStore/ClusterStore'

jest.mock('react-router-dom', () => {
  return {
    ...jest.requireActual('react-router-dom'),
    useSearchParams: jest.fn(),
    useNavigate: jest.fn()
  }
})

const TestComponent = ({ history, name }) => {
  history.push(`/cluster/d/${name}`)
  return (
    <AuthWrapper>
      <AuthWrapper>
        <Router location={history.location} navigator={history}>
          <Routes>
            <Route path="/cluster/d/:param" element={<ClusterSecurityRulesContainer history={history} />} />
          </Routes>
        </Router>
      </AuthWrapper>
    </AuthWrapper>
  )
}

describe('show no security rules', () => {
  let history = null

  beforeEach(() => {
    clearAxiosMock()
    history = createMemoryHistory()
  })

  beforeEach(() => {
    mockIntelUser()
  })

  beforeAll(() => {
    idcConfig.REACT_APP_FEATURE_SC_SECURITY = 1
  })

  it('Show no security rules', async () => {
    await act(async () => {
      useSearchParams.mockReturnValue([new URLSearchParams('tab=security'), jest.fn()])
      render(<TestComponent history={history} />)
    })

    expect(await screen.findByTestId('data-view-empty')).toBeInTheDocument()
  })
})

describe('show grid with security rules', () => {
  let history = null

  beforeEach(() => {
    clearAxiosMock()
    history = createMemoryHistory()
  })

  beforeEach(() => {
    mockIntelUser()
  })

  beforeEach(() => {
    mockGetSecurityRules()
  })

  beforeAll(() => {
    idcConfig.REACT_APP_FEATURE_SC_SECURITY = 1
  })

  beforeEach(() => {
    useSearchParams.mockReturnValue([new URLSearchParams('tab=security'), jest.fn()])
  })

  it('Show grid', async () => {
    await act(async () => {
      const items = mockSecurityRules()
      useClusterStore.setState({ clusterSecurityRules: items })
      render(<TestComponent history={history} />)
    })

    expect(await screen.findByTestId('RowTable1')).toBeInTheDocument()
    expect(await screen.findByTestId('TypesortTableButton')).toBeInTheDocument()
  })
})
