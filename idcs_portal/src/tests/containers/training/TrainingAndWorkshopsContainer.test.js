// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { render, screen, waitFor } from '@testing-library/react'
import TrainingAndWorkshopsContainer from '../../../containers/trainingAndWorkshops/TrainingAndWorkshopsContainer'
import AuthWrapper from '../../../utils/authWrapper/AuthWrapper'
import { act } from 'react'
import { mockIntelUser } from '../../mocks/authentication/authHelper'
import { mockTrainings } from '../../mocks/training/Training'
import { clearAxiosMock } from '../../mocks/utils'

const TestComponent = () => {
  return (
    <AuthWrapper>
      <TrainingAndWorkshopsContainer />
    </AuthWrapper>
  )
}

describe('Training and Workshops catalog empty', () => {
  beforeEach(() => {
    mockIntelUser()
  })

  it('Render component with no products', async () => {
    await act(async () => {
      render(<TestComponent />)
    })
    expect(await screen.findByTestId('no-products')).toBeInTheDocument()
  })
})

describe('Training and Workshops catalog', () => {
  beforeEach(() => {
    clearAxiosMock()
    mockIntelUser()
    mockTrainings()
  })

  it('Render component product', async () => {
    await act(async () => {
      render(<TestComponent />)
    })

    await waitFor(() => {
      screen.findByTestId('h6-btn-training-Test')
    })
  })

  it('Render button select', async () => {
    await act(async () => {
      render(<TestComponent />)
    })

    await waitFor(() => {
      screen.findByTestId('btn-training-select Test')
    })
  })

  it('Render text filter', async () => {
    await act(async () => {
      render(<TestComponent />)
    })

    await waitFor(() => {
      screen.findByTestId('Filter-Text')
    })
  })

  it('Render checkbox filter', async () => {
    await act(async () => {
      render(<TestComponent />)
    })

    await waitFor(() => {
      screen.findByTestId('Checkbox-item')
    })
  })
})
