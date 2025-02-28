// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { render, screen } from '@testing-library/react'
import AuthWrapper from '../../../utils/authWrapper/AuthWrapper'
import { Router } from 'react-router-dom'
import SuperComputerReservationsContainer from '../../../containers/superComputer/SuperComputerReservationsContainer'
import { mockIntelUser } from '../../mocks/authentication/authHelper'
import { act } from 'react'
import { mockGetNoReservations, mockGetReservations, mockGetIksClusters } from '../../mocks/superComputer/SuperComputer'
import { clearAxiosMock } from '../../mocks/utils'
import { createMemoryHistory } from 'history'
import idcConfig from '../../../config/configurator'

const TestComponent = ({ history }) => {
  return (
    <AuthWrapper>
      <Router location={history.location} navigator={history}>
        <SuperComputerReservationsContainer />
      </Router>
    </AuthWrapper>
  )
}

describe('Show no reservations', () => {
  let history = null

  beforeEach(() => {
    clearAxiosMock()
    history = createMemoryHistory()
  })

  beforeEach(() => {
    mockIntelUser()
  })

  beforeEach(() => {
    mockGetNoReservations()
  })

  it('Show empty view', async () => {
    await act(async () => {
      render(<TestComponent history={history} />)
    })
    expect(await screen.findByTestId('data-view-empty')).toBeInTheDocument()
  })
})

describe('show grid with reservation', () => {
  let history = null

  beforeEach(() => {
    clearAxiosMock()
    history = createMemoryHistory()
  })

  beforeEach(() => {
    mockIntelUser()
  })

  beforeAll(() => {
    idcConfig.REACT_APP_FEATURE_SC_SECURITY = 0
  })

  beforeEach(() => {
    mockGetReservations()
  })

  it('Show grid', async () => {
    await act(async () => {
      render(<TestComponent history={history} />)
    })
    expect(await screen.findByTestId('launch-SCCluster')).toBeInTheDocument()
    expect(await screen.findByTestId('ClusterNameColumn1')).toBeInTheDocument()
    expect(await screen.findByTestId('RowTable1')).toBeInTheDocument()
  })

  it('Show options in grid', async () => {
    await act(async () => {
      render(<TestComponent history={history} />)
    })

    expect(await screen.findByTestId('RowTable1')).toBeInTheDocument()
    expect(await screen.findByTestId('ButtonTable DownloadReadOnly')).toBeInTheDocument()
    expect(await screen.findByTestId('ButtonTable Delete cluster instance')).toBeInTheDocument()
  })
})

describe('Validate only Super Computer Items', () => {
  let history = null

  beforeEach(() => {
    clearAxiosMock()
    history = createMemoryHistory()
  })

  beforeEach(() => {
    mockIntelUser()
  })

  beforeAll(() => {
    idcConfig.REACT_APP_FEATURE_SC_SECURITY = 0
  })

  beforeEach(() => {
    mockGetIksClusters()
  })

  it('Show empty grid when there is no super computer items', async () => {
    await act(async () => {
      render(<TestComponent history={history} />)
    })

    expect(await screen.findByTestId('data-view-empty')).toBeInTheDocument()
  })
})
