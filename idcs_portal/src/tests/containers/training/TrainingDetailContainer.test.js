// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { render, screen, waitFor } from '@testing-library/react'
import TrainingDetailContainer from '../../../containers/trainingAndWorkshops/TrainingDetailContainer'
import AuthWrapper from '../../../utils/authWrapper/AuthWrapper'
import { act } from 'react'
import { mockIntelUser } from '../../mocks/authentication/authHelper'
import { mockTraining, mockTrainingEmpty, mockEnrollTraining, mockExpiry } from '../../mocks/training/Training'
import { mockPublicKeys } from '../../mocks/publicKeys/publicKeys'
import { createMemoryHistory } from 'history'
import { Router } from 'react-router-dom'
import { mockVendorsApi } from '../../mocks/shared/vendors'
import { clearAxiosMock } from '../../mocks/utils'
import userEvent from '@testing-library/user-event'

jest.mock('react-router-dom', () => ({
  ...jest.requireActual('react-router-dom'),
  useSearchParams: () => [new URLSearchParams({ id: '5ba105a9-1490-47fc-a5cc-5a23570de3eb' })]
}))

const TestComponent = ({ history }) => {
  return (
    <AuthWrapper>
      <Router location={history.location} navigator={history}>
        <TrainingDetailContainer />
      </Router>
    </AuthWrapper>
  )
}

describe('Training details empty', () => {
  let history = null

  beforeEach(() => {
    clearAxiosMock()
    history = createMemoryHistory()
    mockIntelUser()
    mockPublicKeys()
    mockVendorsApi()
    mockTrainingEmpty()
    mockExpiry()
  })

  it('Render empty view', async () => {
    await act(async () => {
      render(<TestComponent history={history} />)
    })

    await waitFor(() => {
      screen.findByTestId('data-view-empty')
    })
  })
})

describe('Training details view', () => {
  let history = null

  beforeEach(() => {
    clearAxiosMock()
    history = createMemoryHistory()
    mockIntelUser()
    mockPublicKeys()
    mockVendorsApi()
    mockTraining()
    mockEnrollTraining()
    mockExpiry()
  })

  it('Render view', async () => {
    await act(async () => {
      render(<TestComponent history={history} />)
    })
    await screen.findByTestId('h1-title-Test')
    await screen.findAllByTestId('btn-details-launch-training-jupiter')
  })

  it('Render onClick Modal enroll training', async () => {
    await act(async () => {
      render(<TestComponent history={history} />)
    })
    userEvent.click(screen.getAllByTestId('btn-details-launch-training-jupiter')[0])
    await waitFor(() => {
      screen.findByTestId('enroll-training')
    })
  })
})
