// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { render, screen, waitFor } from '@testing-library/react'
import AuthWrapper from '../../../utils/authWrapper/AuthWrapper'
import { mockIntelUser } from '../../mocks/authentication/authHelper'
import { act } from 'react'
import idcConfig from '../../../config/configurator'
import { clearAxiosMock } from '../../mocks/utils'
import { createMemoryHistory } from 'history'
import { Router } from 'react-router-dom'
import userEvent from '@testing-library/user-event'
import {
  mockEmptyObjectStorageUsers,
  mockObjectStorageUsers,
  mockBaseObjectStorageUsersStore,
  mockObjectStorage
} from '../../mocks/objectStorage/ObjectStorage'
import ObjectStorageUsersReservationsContainer from '../../../containers/objectStorage/ObjectStorageUsersReservationsContainer'

const TestComponent = ({ history }) => {
  return (
    <AuthWrapper>
      <Router location={history.location} navigator={history}>
        <ObjectStorageUsersReservationsContainer />
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

  it('Checks Empty object storage users component when there are no users.', async () => {
    mockEmptyObjectStorageUsers()
    await act(async () => {
      render(<TestComponent history={history} />)
    })
    expect(await screen.findByText('No principals found')).toBeInTheDocument()

    const launchUserLink = screen.getByTestId('CreateprincipalEmptyViewButton')
    await waitFor(() => {
      expect(launchUserLink.getAttribute('href')).toBe('/buckets/users/reserve')
    })
  })

  it('Checks users list when there are users avaliable.', async () => {
    await act(async () => {
      render(<TestComponent history={history} />)
    })

    const deleteButton = screen.getAllByTestId('ButtonTable Delete principal')

    await waitFor(() => {
      expect(deleteButton.length).toBe(mockBaseObjectStorageUsersStore().users.length - 1)
      expect(deleteButton[0]).toBeInTheDocument()
    })
  })

  it('Search by name should show expected users, and hide the one that does not match', async () => {
    await act(async () => {
      render(<TestComponent history={history} />)
    })

    const searchInput = await screen.findByTestId('searchUsers')
    expect(searchInput).toBeInTheDocument()

    const usersOnGrid = mockBaseObjectStorageUsersStore().users
    const userOne = usersOnGrid.find((user) => user.metadata.name === 'test')
    const userTwo = usersOnGrid.find((user) => user.metadata.name === 'test1')

    // Before Search
    expect(screen.getByText(userOne.metadata.name)).toBeVisible()
    expect(screen.getByText(userTwo.metadata.name)).toBeVisible()

    // Search
    userEvent.type(searchInput, userTwo.metadata.name)

    await waitFor(() => {
      expect(screen.queryByText(userOne.metadata.name)).toBeNull()
    })
    expect(screen.getByText(userTwo.metadata.name)).toBeVisible()
    userEvent.clear(searchInput)
  })

  it('Checks User Delete button actions with Dialog box validation', async () => {
    await act(async () => {
      render(<TestComponent history={history} />)
    })

    const deleteButton = await screen.findAllByTestId('ButtonTable Delete principal')

    await waitFor(() => {
      expect(deleteButton[0]).toBeInTheDocument()
    })

    userEvent.click(deleteButton[0])
    await screen.findByRole('dialog')
    expect(await screen.findByText((content, element) => content.startsWith('Delete principal'))).toBeInTheDocument()
  })
})
