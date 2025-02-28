// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { render, screen, waitFor, within } from '@testing-library/react'
import AuthWrapper from '../../../utils/authWrapper/AuthWrapper'
import { mockIntelUser } from '../../mocks/authentication/authHelper'
import { act } from 'react'
import idcConfig from '../../../config/configurator'
import { clearAxiosMock } from '../../mocks/utils'
import { createMemoryHistory } from 'history'
import { Router } from 'react-router-dom'
import userEvent from '@testing-library/user-event'
import LoadBalancerReservationsContainer from '../../../containers/loadBalancer/LoadBalancerReservationsContainer'
import {
  mockBaseLoadBalancerStore,
  mockEmptyLoadBalancer,
  mockLoadBalancer
} from '../../mocks/loadBalancer/LoadBalancer'
import { mockInstanceTypes, mockInstances } from '../../mocks/compute/instances'
import { mockMachineImages } from '../../mocks/compute/machineImages'
import { mockVendorsApi } from '../../mocks/shared/vendors'
import { mockProductCatalog } from '../../mocks/compute/productCatalog'

const TestComponent = ({ history }) => {
  return (
    <AuthWrapper>
      <Router location={history.location} navigator={history}>
        <LoadBalancerReservationsContainer />
      </Router>
    </AuthWrapper>
  )
}

describe('Load balancer reservations container unit test cases', () => {
  let history = null

  beforeEach(() => {
    clearAxiosMock()
    history = createMemoryHistory()
  })

  beforeEach(() => {
    mockIntelUser()
  })

  beforeEach(() => {
    mockInstances()
    mockInstanceTypes()
    mockMachineImages()
    mockVendorsApi()
    mockProductCatalog()
    mockLoadBalancer()
  })

  beforeAll(() => {
    idcConfig.REACT_APP_FEATURE_LOAD_BALANCER = 1
  })

  it('Checks Empty load balancer component when there are no balancers.', async () => {
    mockEmptyLoadBalancer()
    await act(async () => {
      render(<TestComponent history={history} />)
    })
    expect(await screen.findByText('No load balancers found')).toBeInTheDocument()

    const launchBalancerLink = screen.getByTestId('LaunchLoadBalancerEmptyViewButton')
    await waitFor(() => {
      expect(launchBalancerLink.getAttribute('href')).toBe('/load-balancer/reserve')
    })
  })

  it('Checks Load balancer list when there are balancers avaliable.', async () => {
    await act(async () => {
      render(<TestComponent history={history} />)
    })

    const deleteButton = screen.getAllByTestId('ButtonTable Delete load balancer')

    await waitFor(() => {
      expect(deleteButton.length).toBe(mockBaseLoadBalancerStore().items.length - 1)
      expect(deleteButton[0]).toBeInTheDocument()
    })
  })

  it('Search by name should show expected balancers, and hide the one that does not match', async () => {
    await act(async () => {
      render(<TestComponent history={history} />)
    })

    const searchInput = await screen.findByTestId('searchLoadBalancers')
    expect(searchInput).toBeInTheDocument()

    const balancersOnGrid = mockBaseLoadBalancerStore().items
    const balancerOne = balancersOnGrid.find((balancer) => balancer.metadata.name === 'my-lb-1')
    const balancerTwo = balancersOnGrid.find((balancer) => balancer.metadata.name === 'my-lb-2')

    // Before Search
    expect(screen.getByText(balancerOne.metadata.name)).toBeVisible()
    expect(screen.getByText(balancerTwo.metadata.name)).toBeVisible()

    // Search
    userEvent.type(searchInput, balancerOne.metadata.name)

    await waitFor(() => {
      expect(screen.queryByText(balancerTwo.metadata.name)).toBeNull()
    })
    expect(screen.getByText(balancerOne.metadata.name)).toBeVisible()
    userEvent.clear(searchInput)
  })

  it('Validates button actions when balancer status is "Deleting"', async () => {
    await act(async () => {
      render(<TestComponent history={history} />)
    })
    const status = 'Deleting'

    const row = screen.getByText(status).closest('tr')
    await waitFor(() => {
      expect(within(row).queryByTestId('ButtonTable Delete load balancer')).not.toBeInTheDocument()
    })
  })

  it('Checks Load Balancer Delete button actions with Dialog box validation', async () => {
    await act(async () => {
      render(<TestComponent history={history} />)
    })

    const deleteButton = screen.getAllByTestId('ButtonTable Delete load balancer')

    await waitFor(() => {
      expect(deleteButton[0]).toBeInTheDocument()
    })

    userEvent.click(deleteButton[0])
    await screen.findByRole('dialog')
    expect(
      await screen.findByText((content, element) => content.startsWith('Delete load balancer'))
    ).toBeInTheDocument()
  })
})
