// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { render, screen } from '@testing-library/react'
import AuthWrapper from '../../../utils/authWrapper/AuthWrapper'
import { mockIntelUser } from '../../mocks/authentication/authHelper'
import { act } from 'react'
import idcConfig from '../../../config/configurator'
import { clearAxiosMock } from '../../mocks/utils'
import { createMemoryHistory } from 'history'
import { Router, Routes, Route } from 'react-router-dom'
import userEvent from '@testing-library/user-event'
import ObjectStorageReservationsDetailsContainer from '../../../containers/objectStorage/ObjectStorageReservationsDetailsContainer'
import { mockObjectStorage, mockBaseObjectStorageStore } from '../../mocks/objectStorage/ObjectStorage'

jest.mock('react-router-dom', () => {
  return {
    ...jest.requireActual('react-router-dom'),
    useSearchParams: () => [new URLSearchParams('tab=Details'), jest.fn()]
  }
})

const TestComponent = ({ history, name }) => {
  history.push(`/buckets/d/${name}`)
  return (
    <AuthWrapper>
      <Router location={history.location} navigator={history}>
        <Routes>
          <Route path="/buckets/d/:param" element={<ObjectStorageReservationsDetailsContainer history={history} />} />
        </Routes>
      </Router>
    </AuthWrapper>
  )
}

describe('Object Storage reservations container unit test cases', () => {
  let history = null

  beforeEach(() => {
    clearAxiosMock()
    history = createMemoryHistory()
  })

  beforeEach(() => {
    mockIntelUser()
  })

  beforeEach(() => {
    mockObjectStorage()
  })

  beforeAll(() => {
    idcConfig.REACT_APP_FEATURE_OBJECT_STORAGE = 1
  })

  it('Validates button actions when bucket status is "Terminating"', async () => {
    const name = mockBaseObjectStorageStore().items[2].metadata.name
    await act(async () => {
      render(<TestComponent history={history} name={name} />)
    })

    // No action buttons should be there in bucket details.
    const actionButtons = screen.queryByTestId('myReservationActionsDropdownButton')
    expect(actionButtons).toBeFalsy()
  })

  it('Displays bucket details when we click on specific bucket', async () => {
    const name = mockBaseObjectStorageStore().items[0].metadata.name
    await act(async () => {
      render(<TestComponent history={history} name={name} />)
    })

    expect(await screen.findByText('Bucket information')).toBeInTheDocument()
    expect(screen.getByTestId('myReservationActionsDropdownButton')).toBeInTheDocument()

    userEvent.click(await screen.findByText(/^Principals/i))
    expect(await screen.findByTestId('btn-navigate-action-Manage-users-and-permissions')).toBeInTheDocument()
  })
})
