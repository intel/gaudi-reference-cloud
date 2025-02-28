// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { render, screen } from '@testing-library/react'
import AuthWrapper from '../../../utils/authWrapper/AuthWrapper'
import { Router, Routes, Route } from 'react-router-dom'
import SuperComputerWorkerNodesContainer from '../../../containers/superComputer/SuperComputerWorkerNodesContainer'
import { mockIntelUser } from '../../mocks/authentication/authHelper'
import { act } from 'react'
import { clusterEmptyDetail } from '../../mocks/superComputer/SuperComputer'
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
            <Route path="/supercomputer/d/:param" element={<SuperComputerWorkerNodesContainer history={history} />} />
          </Routes>
        </Router>
      </AuthWrapper>
    </AuthWrapper>
  )
}

describe('show Ai node group', () => {
  let history = null

  beforeEach(() => {
    clearAxiosMock()
    history = createMemoryHistory()
  })

  beforeEach(() => {
    mockIntelUser()
  })

  it('Show IA Node', async () => {
    await act(async () => {
      const items = clusterEmptyDetail()
      useSuperComputerStore.setState({ clusterDetail: items })
      render(<TestComponent history={history} />)
    })

    expect(await screen.findByTestId('NodegrouptypeInputLabel')).toBeInTheDocument()
  })
})

describe('show General compute empty', () => {
  let history = null

  beforeEach(() => {
    clearAxiosMock()
    history = createMemoryHistory()
  })

  beforeEach(() => {
    mockIntelUser()
  })

  it('Show Action to add general compute', async () => {
    await act(async () => {
      const items = clusterEmptyDetail()
      useSuperComputerStore.setState({ clusterDetail: items })
      useSuperComputerStore.setState({ nodeTabNumber: 1 })
      render(<TestComponent history={history} />)
    })

    expect(await screen.findByTestId('btnAddNodeGroup')).toBeInTheDocument()
  })
})
