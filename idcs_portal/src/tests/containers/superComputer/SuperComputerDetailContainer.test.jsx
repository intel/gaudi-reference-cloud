// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { render, screen } from '@testing-library/react'
import AuthWrapper from '../../../utils/authWrapper/AuthWrapper'
import { Router, Routes, Route, useSearchParams } from 'react-router-dom'
import SuperComputerDetailContainer from '../../../containers/superComputer/SuperComputerDetailContainer'
import { mockIntelUser } from '../../mocks/authentication/authHelper'
import { act } from 'react'
import {
  mockGetReservations,
  mockGetInstanceTypesSuperComputer,
  mockGetProductCatalog,
  mockReservations,
  mockGetSecurityRules
} from '../../mocks/superComputer/SuperComputer'
import { clearAxiosMock } from '../../mocks/utils'
import { createMemoryHistory } from 'history'
import { mockVendorsApi } from '../../mocks/shared/vendors'
import idcConfig from '../../../config/configurator'

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
            <Route path="/supercomputer/d/:param" element={<SuperComputerDetailContainer history={history} />} />
          </Routes>
        </Router>
      </AuthWrapper>
    </AuthWrapper>
  )
}

describe('Show Detail no security', () => {
  let history = null

  beforeAll(() => {
    idcConfig.REACT_APP_FEATURE_SC_SECURITY = 0
  })

  beforeEach(() => {
    clearAxiosMock()
    history = createMemoryHistory()
  })

  beforeEach(() => {
    useSearchParams.mockReturnValue([new URLSearchParams(''), jest.fn()])
  })

  beforeEach(() => {
    mockIntelUser()
  })

  beforeEach(() => {
    mockGetReservations()
    mockGetInstanceTypesSuperComputer()
    mockVendorsApi()
    mockGetProductCatalog()
  })

  it('Show valid reserve', async () => {
    const name = mockReservations().clusters[0].name
    await act(async () => {
      render(<TestComponent history={history} name={name} />)
    })
    expect(await screen.findByTestId('SuperComputerActionsDropdownButton')).toBeInTheDocument()
    expect(await screen.findByTestId('SuperComputerActionsDropdownButton')).toBeInTheDocument()
    expect(await screen.findByTestId('Worker Node Groups (1)Tab')).toBeInTheDocument()
    expect(await screen.findByTestId('Load Balancers (2)Tab')).toBeInTheDocument()
    expect(await screen.findByTestId('StorageTab')).toBeInTheDocument()
    expect(await screen.findByTestId('DetailsTab')).toBeInTheDocument()
  })
})

describe('Show Detail with security rules', () => {
  let history = null

  beforeAll(() => {
    idcConfig.REACT_APP_FEATURE_SC_SECURITY = 1
  })

  beforeEach(() => {
    clearAxiosMock()
    history = createMemoryHistory()
  })

  beforeEach(() => {
    useSearchParams.mockReturnValue([new URLSearchParams(''), jest.fn()])
  })

  beforeEach(() => {
    mockIntelUser()
  })

  beforeEach(() => {
    mockGetReservations()
    mockGetSecurityRules()
    mockGetInstanceTypesSuperComputer()
    mockVendorsApi()
    mockGetProductCatalog()
  })

  it('Show valid reserve with security rules', async () => {
    const name = mockReservations().clusters[0].name
    await act(async () => {
      idcConfig.REACT_APP_FEATURE_SC_SECURITY = 1
      render(<TestComponent history={history} name={name} />)
    })
    expect(await screen.findByTestId('SuperComputerActionsDropdownButton')).toBeInTheDocument()
    expect(await screen.findByTestId('Worker Node Groups (1)Tab')).toBeInTheDocument()
    expect(await screen.findByTestId('Load Balancers (2)Tab')).toBeInTheDocument()
    expect(await screen.findByTestId('StorageTab')).toBeInTheDocument()
    expect(await screen.findByTestId('DetailsTab')).toBeInTheDocument()
    expect(await screen.findByTestId('SecurityTab')).toBeInTheDocument()
  })
})
