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
import ObjectStorageReservationsContainer from '../../../containers/objectStorage/ObjectStorageReservationsContainer'
import {
  mockEmptyObjectStorage,
  mockObjectStorage,
  mockBaseObjectStorageStore
} from '../../mocks/objectStorage/ObjectStorage'

const TestComponent = ({ history }) => {
  return (
    <AuthWrapper>
      <Router location={history.location} navigator={history}>
        <ObjectStorageReservationsContainer />
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

  it('Checks Empty object storage component when there are no buckets.', async () => {
    mockEmptyObjectStorage()
    await act(async () => {
      render(<TestComponent history={history} />)
    })

    expect(await screen.findByText('No buckets found')).toBeInTheDocument()

    const launchBucketLink = screen.getByTestId('CreatebucketEmptyViewButton')
    await waitFor(() => {
      expect(launchBucketLink.getAttribute('href')).toBe('/buckets/reserve')
    })
  })

  it('Checks Buckets list when there are buckets avaliable.', async () => {
    await act(async () => {
      render(<TestComponent history={history} />)
    })

    const deleteButton = screen.getAllByTestId('ButtonTable Delete bucket')

    await waitFor(() => {
      expect(deleteButton.length).toBe(mockBaseObjectStorageStore().items.length - 1)
      expect(deleteButton[0]).toBeInTheDocument()
    })
  })

  it('Search by name should show expected buckets, and hide the one that does not match', async () => {
    await act(async () => {
      render(<TestComponent history={history} />)
    })

    const searchInput = await screen.findByTestId('searchBuckets')
    expect(searchInput).toBeInTheDocument()

    const bucketsOnGrid = mockBaseObjectStorageStore().items
    const bucketOne = bucketsOnGrid.find((bucket) => bucket.metadata.name === '549881761988-test1')
    const bucketTwo = bucketsOnGrid.find((bucket) => bucket.metadata.name === '549881761988-test2')

    // Before Search
    expect(screen.getByText(bucketOne.metadata.name)).toBeVisible()
    expect(screen.getByText(bucketTwo.metadata.name)).toBeVisible()

    // Search
    userEvent.type(searchInput, bucketOne.metadata.name)

    await waitFor(() => {
      expect(screen.queryByText(bucketTwo.metadata.name)).toBeNull()
    })
    expect(screen.getByText(bucketOne.metadata.name)).toBeVisible()
    userEvent.clear(searchInput)
  })

  it('Checks Bucket Delete button actions with Dialog box validation', async () => {
    await act(async () => {
      render(<TestComponent history={history} />)
    })

    const deleteButton = await screen.findAllByTestId('ButtonTable Delete bucket')

    await waitFor(() => {
      expect(deleteButton[0]).toBeInTheDocument()
    })

    userEvent.click(deleteButton[0])
    await screen.findByRole('dialog')
    expect(await screen.findByText((content, element) => content.startsWith('Delete bucket'))).toBeInTheDocument()
  })
})
