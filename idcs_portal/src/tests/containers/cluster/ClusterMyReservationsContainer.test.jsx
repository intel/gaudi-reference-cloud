// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { render, screen } from '@testing-library/react'
import AuthWrapper from '../../../utils/authWrapper/AuthWrapper'
import { Router } from 'react-router-dom'
import { mockIntelUser } from '../../mocks/authentication/authHelper'
import { act } from 'react'
import {
  mockGetClustersEmpty,
  mockGeSuperComputerReservations,
  mockGetIksReservations
} from '../../mocks/cluster/Cluster'
import { clearAxiosMock } from '../../mocks/utils'
import { createMemoryHistory } from 'history'
import ClusterMyReservationsContainer from '../../../containers/cluster/ClusterMyReservationsContainer'
import userEvent from '@testing-library/user-event'

const TestComponent = ({ history }) => {
  return (
    <AuthWrapper>
      <Router location={history.location} navigator={history}>
        <ClusterMyReservationsContainer />
      </Router>
    </AuthWrapper>
  )
}

describe('No Iks reservations', () => {
  let history = null

  beforeEach(() => {
    clearAxiosMock()
    history = createMemoryHistory()
  })

  beforeEach(() => {
    mockIntelUser()
  })

  beforeEach(() => {
    mockGetClustersEmpty()
  })

  it('Show empty view', async () => {
    await act(async () => {
      render(<TestComponent history={history} />)
    })
    expect(await screen.findByTestId('data-view-empty')).toBeInTheDocument()
  })
})

describe('No Show supercomputer items', () => {
  let history = null

  beforeEach(() => {
    clearAxiosMock()
    history = createMemoryHistory()
  })

  beforeEach(() => {
    mockIntelUser()
  })

  beforeEach(() => {
    mockGeSuperComputerReservations()
  })

  it('Show empty view', async () => {
    await act(async () => {
      render(<TestComponent history={history} />)
    })
    expect(await screen.findByTestId('data-view-empty')).toBeInTheDocument()
  })
})

describe('Show Iks items', () => {
  let history = null

  beforeEach(() => {
    clearAxiosMock()
    history = createMemoryHistory()
  })

  beforeEach(() => {
    mockIntelUser()
  })

  beforeEach(() => {
    mockGetIksReservations()
  })

  it('Show items', async () => {
    await act(async () => {
      render(<TestComponent history={history} />)
    })

    const btnDeleteIksItem = screen.getByTestId('ButtonTable Delete cluster instance')
    userEvent.click(btnDeleteIksItem)

    expect(await screen.findByTestId('deleteConfirmModal')).toBeInTheDocument()
    expect(await screen.findByTestId('RowTable1')).toBeInTheDocument()
    expect(await screen.findByTestId('ButtonTable Delete cluster instance')).toBeInTheDocument()
    expect(await screen.findByTestId('iks-ui-stg-1HyperLinkTable')).toBeInTheDocument()
  })
})
