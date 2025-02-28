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
import {
  mockObjectStorageUsers,
  mockBaseObjectStorageUsersStore,
  mockObjectStorage
} from '../../mocks/objectStorage/ObjectStorage'
import ObjectStorageUsersReservationsDetailsContainer from '../../../containers/objectStorage/ObjectStorageUsersReservationsDetailsContainer'

jest.mock('react-router-dom', () => {
  return {
    ...jest.requireActual('react-router-dom'),
    useSearchParams: () => [new URLSearchParams('tab=Details'), jest.fn()]
  }
})

const TestComponent = ({ history, name }) => {
  history.push(`/buckets/users/d/${name}`)
  return (
    <AuthWrapper>
      <Router location={history.location} navigator={history}>
        <Routes>
          <Route
            path="/buckets/users/d/:param"
            element={<ObjectStorageUsersReservationsDetailsContainer history={history} />}
          />
        </Routes>
      </Router>
    </AuthWrapper>
  )
}

describe('Object Storage users reservations container unit test cases', () => {
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
    mockObjectStorageUsers()
  })

  beforeAll(() => {
    idcConfig.REACT_APP_FEATURE_OBJECT_STORAGE = 1
  })

  it('Validates button actions when user status is "Terminating"', async () => {
    const name = mockBaseObjectStorageUsersStore().users[1].metadata.name
    await act(async () => {
      render(<TestComponent history={history} name={name} />)
    })

    // No action buttons should be there in user details.
    const actionButtons = screen.queryByTestId('myUsersActionsDropdownButton')
    expect(actionButtons).toBeFalsy()
  })

  it('Displays user details', async () => {
    const name = mockBaseObjectStorageUsersStore().users[0].metadata.name
    await act(async () => {
      render(<TestComponent history={history} name={name} />)
    })
    expect(await screen.findByText(`Principal: ${name}`)).toBeInTheDocument()
    expect(screen.getByTestId('myUsersActionsDropdownButton')).toBeInTheDocument()

    userEvent.click(await screen.findByText(/^Permissions/i))
    expect(await screen.findByText('Buckets:')).toBeInTheDocument()
  })
})
