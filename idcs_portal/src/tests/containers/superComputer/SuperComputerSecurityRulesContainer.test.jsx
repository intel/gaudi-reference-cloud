// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { render, screen } from '@testing-library/react'
import AuthWrapper from '../../../utils/authWrapper/AuthWrapper'
import { Router, Routes, Route, useSearchParams } from 'react-router-dom'
import SuperComputerSecurityRulesContainer from '../../../containers/superComputer/SuperComputerSecurityRulesContainer'
import { mockIntelUser } from '../../mocks/authentication/authHelper'
import { act } from 'react'
import {
  mockGetReservations,
  mockGetInstanceTypesSuperComputer,
  mockGetProductCatalog,
  mockSecurityRules,
  clusterDetail
} from '../../mocks/superComputer/SuperComputer'
import { clearAxiosMock } from '../../mocks/utils'
import { createMemoryHistory } from 'history'
import { mockVendorsApi } from '../../mocks/shared/vendors'
import idcConfig from '../../../config/configurator'
import useSuperComputerStore from '../../../store/superComputer/SuperComputerStore'

jest.mock('react-router-dom', () => {
  return {
    ...jest.requireActual('react-router-dom'),
    useSearchParams: jest.fn(),
    useNavigate: jest.fn()
  }
})

const TestComponent = ({ history, name }) => {
  history.push(`/supercomputer/d/${name}`)
  return (
    <AuthWrapper>
      <AuthWrapper>
        <Router location={history.location} navigator={history}>
          <Routes>
            <Route path="/supercomputer/d/:param" element={<SuperComputerSecurityRulesContainer history={history} />} />
          </Routes>
        </Router>
      </AuthWrapper>
    </AuthWrapper>
  )
}

describe('show empty view when no security rules', () => {
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

  beforeEach(() => {
    useSearchParams.mockReturnValue([new URLSearchParams(''), jest.fn()])
  })

  it('Show no security rules', async () => {
    await act(async () => {
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
    mockGetReservations()
    mockGetInstanceTypesSuperComputer()
    mockVendorsApi()
    mockGetProductCatalog()
    mockSecurityRules()
    clusterDetail()
  })

  beforeAll(() => {
    idcConfig.REACT_APP_FEATURE_SC_SECURITY = 1
  })

  beforeEach(() => {
    useSearchParams.mockReturnValue([new URLSearchParams(''), jest.fn()])
  })

  it('Show grid', async () => {
    await act(async () => {
      const items = clusterDetail()
      useSuperComputerStore.setState({ clusterDetail: items })
      render(<TestComponent history={history} />)
    })

    expect(await screen.findByTestId('RowTable1')).toBeInTheDocument()
    expect(await screen.findByTestId('ButtonTable Edit')).toBeInTheDocument()
  })
})
