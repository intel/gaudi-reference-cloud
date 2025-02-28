// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { render, screen, waitFor } from '@testing-library/react'
import SoftwareCatalogContainer from '../../../containers/software/SoftwareCatalogContainer'
import AuthWrapper from '../../../utils/authWrapper/AuthWrapper'
import { act } from 'react'
import { mockIntelUser } from '../../mocks/authentication/authHelper'
import { mockSoftwareEmpty, mockSoftwareList } from '../../mocks/software/Software'
import { mockVendorsApi } from '../../mocks/shared/vendors'
import { clearAxiosMock } from '../../mocks/utils'
import { BrowserRouter } from 'react-router-dom'
import idcConfig from '../../../config/configurator'

const TestComponent = () => {
  return (
    <BrowserRouter>
      <AuthWrapper>
        <SoftwareCatalogContainer />
      </AuthWrapper>
    </BrowserRouter>
  )
}

describe('Software catalog empty', () => {
  beforeEach(() => {
    mockIntelUser()
    mockVendorsApi()
    mockSoftwareEmpty()
  })

  beforeAll(() => {
    idcConfig.REACT_APP_FEATURE_SOFTWARE = 1
  })

  it('Render component with no products', async () => {
    await act(async () => {
      render(<TestComponent />)
    })
    expect(await screen.findByTestId('no-products')).toBeInTheDocument()
  })
})

describe('Software catalog', () => {
  beforeEach(() => {
    clearAxiosMock()
    mockIntelUser()
    mockVendorsApi()
    mockSoftwareList()
  })
  beforeAll(() => {
    idcConfig.REACT_APP_FEATURE_SOFTWARE = 0
  })

  it('Render component product', async () => {
    await act(async () => {
      render(<TestComponent />)
    })

    await waitFor(() => {
      screen.findByTestId('h6-btn-software-Test')
    })
  })

  it('Render button select', async () => {
    await act(async () => {
      render(<TestComponent />)
    })

    await waitFor(() => {
      screen.findByTestId('btn-software-select test')
    })
  })
})
