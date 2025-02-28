// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { render, screen } from '@testing-library/react'
import AuthWrapper from '../../../utils/authWrapper/AuthWrapper'
import { Router } from 'react-router-dom'
import SuperComputerLaunchContainer from '../../../containers/superComputer/SuperComputerLaunchContainer'
import { mockIntelUser } from '../../mocks/authentication/authHelper'
import { act } from 'react'
import {
  mockGetProductCatalog,
  mockGetInstanceTypesSuperComputer,
  mockGetRuntimes,
  mockGetVnets,
  mockGetClusters,
  mockGetProductCatalogEmpty
} from '../../mocks/superComputer/SuperComputer'
import { mockVendorsApi } from '../../mocks/shared/vendors'
import { mockPublicKeys } from '../../mocks/publicKeys/publicKeys'
import { clearAxiosMock } from '../../mocks/utils'
import { createMemoryHistory } from 'history'
import userEvent from '@testing-library/user-event'

const TestComponent = ({ history }) => {
  return (
    <AuthWrapper>
      <Router location={history.location} navigator={history}>
        <SuperComputerLaunchContainer />
      </Router>
    </AuthWrapper>
  )
}

describe('Load form', () => {
  let history = null

  beforeEach(() => {
    clearAxiosMock()
    history = createMemoryHistory()
  })

  beforeEach(() => {
    mockIntelUser()
  })

  beforeEach(() => {
    mockGetInstanceTypesSuperComputer()
    mockGetProductCatalog()
    mockPublicKeys()
    mockGetRuntimes()
    mockGetVnets()
    mockGetClusters()
    mockVendorsApi()
  })

  it('Render component with products', async () => {
    await act(async () => {
      render(<TestComponent history={history} />)
    })
    expect(await screen.findByTestId('superComputeTitle')).toBeInTheDocument()
  })
})

describe('Check user not whitelisted', () => {
  let history = null

  beforeEach(() => {
    clearAxiosMock()
    history = createMemoryHistory()
  })

  beforeEach(() => {
    mockIntelUser()
  })

  beforeEach(() => {
    mockGetInstanceTypesSuperComputer()
    mockGetProductCatalogEmpty()
    mockPublicKeys()
    mockGetRuntimes()
    mockGetVnets()
    mockGetClusters()
    mockVendorsApi()
  })

  it('Render modal product empty', async () => {
    await act(async () => {
      render(<TestComponent history={history} />)
    })

    expect(await screen.findByTestId('Empty-Catalog')).toBeInTheDocument()
  })
})

describe('Validate form', () => {
  let history = null

  beforeEach(() => {
    clearAxiosMock()
    history = createMemoryHistory()
  })

  beforeEach(() => {
    mockIntelUser()
  })

  beforeEach(() => {
    mockGetInstanceTypesSuperComputer()
    mockGetProductCatalog()
    mockPublicKeys()
    mockGetRuntimes()
    mockGetVnets()
    mockGetClusters()
    mockVendorsApi()
  })

  it('Show error when form is not valid', async () => {
    await act(async () => {
      render(<TestComponent history={history} />)
    })

    const btnLaunchSuperComputer = screen.getByTestId('btn-computelaunch-navigationBottom Launch')
    userEvent.click(btnLaunchSuperComputer)
    expect(await screen.getByTestId('ClusternameInput')).toHaveClass('is-invalid')
    expect(await screen.getByTestId('Clusterkubernetesversion-form-select')).toHaveClass('is-invalid')
    expect(await screen.getByTestId('Volumesize(TB)Input')).toHaveClass('is-invalid')
  })
})

describe('Invalid input name', () => {
  let history = null

  beforeEach(() => {
    clearAxiosMock()
    history = createMemoryHistory()
  })

  beforeEach(() => {
    mockIntelUser()
  })

  beforeEach(() => {
    mockGetInstanceTypesSuperComputer()
    mockGetProductCatalog()
    mockPublicKeys()
    mockGetRuntimes()
    mockGetVnets()
    mockGetClusters()
    mockVendorsApi()
  })

  it('Show error when input has invalid character', async () => {
    await act(async () => {
      render(<TestComponent history={history} />)
    })

    const inputClusternameInput = screen.getByTestId('ClusternameInput')
    userEvent.type(inputClusternameInput, 'INVALID-CHARACTER')
    expect(await screen.getByTestId('ClusternameInput')).toHaveClass('is-invalid')
  })
})
