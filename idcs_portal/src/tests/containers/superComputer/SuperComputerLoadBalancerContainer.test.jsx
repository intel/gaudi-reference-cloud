// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { render, screen } from '@testing-library/react'
import AuthWrapper from '../../../utils/authWrapper/AuthWrapper'
import { Router, Routes, Route, useSearchParams } from 'react-router-dom'
import SuperComputerLoadBalancerContainer from '../../../containers/superComputer/SuperComputerLoadBalancerContainer'
import { mockIntelUser } from '../../mocks/authentication/authHelper'
import { act } from 'react'
import { clusterDetail, clusterEmptyDetail } from '../../mocks/superComputer/SuperComputer'
import { clearAxiosMock } from '../../mocks/utils'
import { createMemoryHistory } from 'history'
import useSuperComputerStore from '../../../store/superComputer/SuperComputerStore'

jest.mock('react-router-dom', () => {
  return {
    ...jest.requireActual('react-router-dom'),
    useSearchParams: jest.fn(),
    useNavigate: jest.fn()
  }
})

const TestComponent = ({ history, name }) => {
  history.push(`/supercomputer/d/${name}`)
  return (
    <AuthWrapper>
      <AuthWrapper>
        <Router location={history.location} navigator={history}>
          <Routes>
            <Route path="/supercomputer/d/:param" element={<SuperComputerLoadBalancerContainer history={history} />} />
          </Routes>
        </Router>
      </AuthWrapper>
    </AuthWrapper>
  )
}

describe('show empty view when no load balancers', () => {
  let history = null

  beforeEach(() => {
    clearAxiosMock()
    history = createMemoryHistory()
  })

  beforeEach(() => {
    mockIntelUser()
  })

  beforeEach(() => {
    useSearchParams.mockReturnValue([new URLSearchParams(''), jest.fn()])
  })

  it('Show no load balancers', async () => {
    const item = clusterEmptyDetail()
    useSuperComputerStore.setState({ clusterDetail: item })
    await act(async () => {
      render(<TestComponent history={history} />)
    })
    expect(await screen.findByTestId('data-view-empty')).toBeInTheDocument()
  })
})

describe('show grid with load balancers', () => {
  let history = null

  beforeEach(() => {
    clearAxiosMock()
    history = createMemoryHistory()
  })

  beforeEach(() => {
    mockIntelUser()
  })

  beforeEach(() => {
    useSearchParams.mockReturnValue([new URLSearchParams(''), jest.fn()])
  })

  it('Show grid', async () => {
    await act(async () => {
      const item = clusterDetail()
      useSuperComputerStore.setState({ clusterDetail: item })
      render(<TestComponent history={history} />)
    })

    expect(await screen.findByTestId('RowTable1')).toBeInTheDocument()
  })
})
