// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { render, screen, fireEvent } from '@testing-library/react'
import AuthWrapper from '../../../utils/authWrapper/AuthWrapper'
import StorageLaunchContainer from '../../../containers/storage/StorageLaunchContainer'
import { mockIntelUser } from '../../mocks/authentication/authHelper'
import { act } from 'react'
import idcConfig from '../../../config/configurator'
import { mockStorage } from '../../mocks/storage/Storage'
import { clearAxiosMock } from '../../mocks/utils'
import { createMemoryHistory } from 'history'
import { Router } from 'react-router-dom'
import { mockVendorsApi } from '../../mocks/shared/vendors'

const TestComponent = ({ history }) => {
  return (
    <AuthWrapper>
      <Router location={history.location} navigator={history}>
        <StorageLaunchContainer />
      </Router>
    </AuthWrapper>
  )
}

describe('Storage container Feature Flag', () => {
  let history = null

  beforeEach(() => {
    clearAxiosMock()
    history = createMemoryHistory()
  })

  beforeEach(() => {
    mockIntelUser()
  })

  beforeAll(() => {
    idcConfig.REACT_APP_FEATURE_STORAGE = 0
  })

  it('Render comming soon banner', async () => {
    await act(async () => {
      render(<TestComponent history={history} />)
    })

    expect(await screen.findByTestId('text-comming-soon')).toBeInTheDocument()
  })
})

describe('Storag form validation', () => {
  let history = null

  beforeEach(() => {
    clearAxiosMock()
    history = createMemoryHistory()
    mockVendorsApi()
  })

  beforeEach(() => {
    mockIntelUser()
  })

  beforeEach(() => {
    mockStorage()
  })

  beforeAll(() => {
    idcConfig.REACT_APP_FEATURE_STORAGE = 1
    idcConfig.REACT_APP_FEATURE_STORAGE_VAST = 0
  })

  it('Show navigation botton buttons', async () => {
    await act(async () => {
      render(<TestComponent history={history} />)
    })

    expect(await screen.findByTestId('btn-storagelaunch-navigationBottom Create')).toBeInTheDocument()

    expect(await screen.findByTestId('btn-storagelaunch-navigationBottom Cancel')).toBeInTheDocument()
  })

  it('Launch button', async () => {
    await act(async () => {
      render(<TestComponent history={history} />)
    })

    expect(await screen.findByTestId('btn-storagelaunch-navigationBottom Create')).toBeVisible()
  })

  it('Show form elements', async () => {
    await act(async () => {
      render(<TestComponent history={history} />)
    })

    expect(await screen.findByTestId('NameInput')).toBeInTheDocument()
  })

  it('Validate if form is valid for submition', async () => {
    await act(async () => {
      render(<TestComponent history={history} />)
    })

    const inputName = screen.getByTestId('NameInput')
    fireEvent.change(inputName, ('name', { target: { value: 'test' } }))
    expect(inputName).toHaveValue('test')
  })
})
