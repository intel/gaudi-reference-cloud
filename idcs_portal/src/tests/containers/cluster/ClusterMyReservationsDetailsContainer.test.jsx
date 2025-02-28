// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { render, screen } from '@testing-library/react'
import AuthWrapper from '../../../utils/authWrapper/AuthWrapper'
import { Router, Routes, Route, useSearchParams } from 'react-router-dom'
import { mockIntelUser } from '../../mocks/authentication/authHelper'
import { act } from 'react'
import { mockGetIksReservations, mockIKSReservations, mockSecurityRules } from '../../mocks/cluster/Cluster'
import { clearAxiosMock } from '../../mocks/utils'
import { createMemoryHistory } from 'history'
import idcConfig from '../../../config/configurator'
import ClusterMyReservationsDetailsContainer from '../../../containers/cluster/ClusterMyReservationsDetailsContainer'

jest.mock('react-router-dom', () => {
  return {
    ...jest.requireActual('react-router-dom'),
    useSearchParams: jest.fn(),
    useNavigate: jest.fn()
  }
})

const TestComponent = ({ history, name }) => {
  const route = `/cluster/d/${name}`
  history.push(route)
  return (
    <AuthWrapper>
      <Router location={history.location} navigator={history}>
        <Routes>
          <Route path="/cluster/d/:param" element={<ClusterMyReservationsDetailsContainer history={history} />} />
        </Routes>
      </Router>
    </AuthWrapper>
  )
}

describe('Show Detail with security rules', () => {
  let history = null

  beforeAll(() => {
    idcConfig.REACT_APP_FEATURE_SC_SECURITY = 1
  })

  beforeEach(() => {
    clearAxiosMock()
    history = createMemoryHistory()
  })

  beforeEach(() => {
    mockIntelUser()
  })

  beforeEach(() => {
    mockGetIksReservations()
    mockSecurityRules()
  })

  it('Show reserve details', async () => {
    const name = mockIKSReservations().clusters[0].name
    await act(async () => {
      idcConfig.REACT_APP_FEATURE_SC_SECURITY = 1
      useSearchParams.mockReturnValue([new URLSearchParams('tab=Details'), jest.fn()])
      render(<TestComponent history={history} name={name} />)
    })
    expect(await screen.findByTestId('myReservationActionsDropdownButton')).toBeInTheDocument()
    expect(await screen.findByTestId('DetailsTab')).toBeInTheDocument()
    expect(await screen.findByTestId('Worker Node Groups (0)Tab')).toBeInTheDocument()
    expect(await screen.findByTestId('Load Balancers (0)Tab')).toBeInTheDocument()
    expect(await screen.findByTestId('Storage (0)Tab')).toBeInTheDocument()
    expect(await screen.findByTestId('SecurityTab')).toBeInTheDocument()
  })

  it('show no worker node', async () => {
    const name = mockIKSReservations().clusters[0].name
    await act(async () => {
      idcConfig.REACT_APP_FEATURE_SC_SECURITY = 1
      useSearchParams.mockReturnValue([new URLSearchParams('tab=workerNodeGroups'), jest.fn()])
      render(<TestComponent history={history} name={name} />)
    })
    expect(await screen.findByTestId('btn-iksMyClusters-addWorkerNodeGroup')).toBeInTheDocument()
  })

  it('show no load balancer with no nodegroups', async () => {
    const name = mockIKSReservations().clusters[0].name
    await act(async () => {
      idcConfig.REACT_APP_FEATURE_SC_SECURITY = 1
      useSearchParams.mockReturnValue([new URLSearchParams('tab=loadBalancers'), jest.fn()])
      render(<TestComponent history={history} name={name} />)
    })
    expect(await screen.findByTestId('btn-iksMyClusters-addWorkerNodeGroup')).toBeInTheDocument()
  })
})
