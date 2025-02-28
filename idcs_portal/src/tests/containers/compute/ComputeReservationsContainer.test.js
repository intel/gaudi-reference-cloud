// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

// eslint-disable-next-line react/display-name
import { render, screen, waitFor, within } from '@testing-library/react'
import { Router } from 'react-router-dom'
import { createMemoryHistory } from 'history'
import { act } from 'react'
import AuthWrapper from '../../../utils/authWrapper/AuthWrapper'
import { clearAxiosMock } from '../../mocks/utils'
import ComputeReservationsContainers from '../../../containers/compute/ComputeReservationsContainers'
import { mockIntelUser } from '../../mocks/authentication/authHelper'
import {
  mockBaseInstancesStore,
  mockEmptyInstances,
  mockInstanceTypes,
  mockInstances
} from '../../mocks/compute/instances'
import { mockMachineImages } from '../../mocks/compute/machineImages'
import { mockVendorsApi } from '../../mocks/shared/vendors'
import { mockProductCatalog } from '../../mocks/compute/productCatalog'
import userEvent from '@testing-library/user-event'

const TestComponent = ({ history }) => {
  return (
    <AuthWrapper>
      <Router location={history.location} navigator={history}>
        <ComputeReservationsContainers />
      </Router>
    </AuthWrapper>
  )
}

describe('Compute Reservations Container - My Instances page', () => {
  let history = null

  beforeEach(() => {
    clearAxiosMock()
    history = createMemoryHistory()
  })

  beforeEach(() => {
    mockIntelUser()
    mockInstances()
    mockInstanceTypes()
    mockMachineImages()
    mockVendorsApi()
    mockProductCatalog()
  })

  it('Checks Instances list when there are instances avaliable.', async () => {
    await act(async () => {
      render(<TestComponent history={history} />)
    })

    const editButton = screen.getAllByTestId('ButtonTable Edit instance')
    const deleteButton = screen.getAllByTestId('ButtonTable Delete instance')

    await waitFor(() => {
      expect(editButton.length).toBe(3)
      expect(editButton[0]).toBeInTheDocument()

      expect(deleteButton.length).toBe(4)
      expect(deleteButton[0]).toBeInTheDocument()
    })
  })

  it('Search by name should show expected instances, and hide the one that does not match', async () => {
    await act(async () => {
      render(<TestComponent history={history} />)
    })

    const searchInput = await screen.findByTestId('searchInstances')
    expect(searchInput).toBeInTheDocument()

    const searchInstanceInput = screen.getByTestId('searchInstances')
    expect(searchInstanceInput).toBeInTheDocument()

    const instancesOnGrid = mockBaseInstancesStore().items
    const instanceOne = instancesOnGrid.find((instance) => instance.metadata.name === 'test-one')
    const instanceTwo = instancesOnGrid.find((instance) => instance.metadata.name === 'test-two')
    // Before Search
    expect(screen.getByText(instanceOne.metadata.name)).toBeVisible()
    expect(screen.getByText(instanceTwo.metadata.name)).toBeVisible()
    // Search
    await userEvent.type(searchInput, instanceTwo.metadata.name)

    await waitFor(() => {
      expect(screen.queryByText(instanceOne.metadata.name)).toBeNull()
    })
    expect(screen.getByText(instanceTwo.metadata.name)).toBeVisible()
    await userEvent.clear(searchInput)
  })

  it('Search terminating instance by name should show expected instance, and hide the one that does not match', async () => {
    await act(async () => {
      render(<TestComponent history={history} />)
    })

    const searchInput = await screen.findByTestId('searchInstances')
    expect(searchInput).toBeInTheDocument()

    const searchInstanceInput = screen.getByTestId('searchInstances')
    expect(searchInstanceInput).toBeInTheDocument()

    const instancesOnGrid = mockBaseInstancesStore().items
    const readyInstance = instancesOnGrid.find((instance) => instance.status.phase === 'Ready')
    const terminatingInstance = instancesOnGrid.find((instance) => instance.status.phase === 'Terminating')
    // Before Search
    expect(screen.getByText(readyInstance.metadata.name)).toBeVisible()
    expect(screen.getByText(terminatingInstance.metadata.name)).toBeVisible()
    // Search
    await userEvent.type(searchInput, terminatingInstance.metadata.name)

    await waitFor(() => {
      expect(screen.queryByText(readyInstance.metadata.name)).toBeNull()
    })
    expect(screen.getByText(terminatingInstance.metadata.name)).toBeVisible()
    await userEvent.clear(searchInput)
  })

  it('Search by status should show expected instances, and hide the one that does not match', async () => {
    await act(async () => {
      render(<TestComponent history={history} />)
    })

    const searchInput = await screen.findByTestId('searchInstances')
    expect(searchInput).toBeInTheDocument()

    const searchInstanceInput = screen.getByTestId('searchInstances')
    expect(searchInstanceInput).toBeInTheDocument()

    const instancesOnGrid = mockBaseInstancesStore().items
    const readyInstance = instancesOnGrid.find((instance) => instance.status.phase === 'Ready')
    const failedInstance = instancesOnGrid.find((instance) => instance.status.phase === 'Failed')
    // Before Search
    expect(screen.getByText(readyInstance.status.phase)).toBeVisible()
    expect(screen.getByText(failedInstance.status.phase)).toBeVisible()
    // Search
    await userEvent.type(searchInput, failedInstance.status.phase)

    await waitFor(() => {
      expect(screen.queryByText(readyInstance.status.phase)).toBeNull()
    })
    expect(screen.getByText(failedInstance.status.phase)).toBeVisible()
    await userEvent.clear(searchInput)
  })

  it('Search by IP Address should show expected instances, and hide the one that does not match', async () => {
    await act(async () => {
      render(<TestComponent history={history} />)
    })

    const searchInput = await screen.findByTestId('searchInstances')
    expect(searchInput).toBeInTheDocument()

    const searchInstanceInput = screen.getByTestId('searchInstances')
    expect(searchInstanceInput).toBeInTheDocument()

    const instancesOnGrid = mockBaseInstancesStore().items
    const instanceOne = instancesOnGrid.find(
      (instance) => instance.status.interfaces[0].addresses[0] === '172.16.0.171'
    )
    const instanceTwo = instancesOnGrid.find(
      (instance) => instance.status.interfaces[0].addresses[0] === '172.16.0.172'
    )

    const ip1 = instanceOne.status.interfaces[0].addresses[0]
    const ip2 = instanceTwo.status.interfaces[0].addresses[0]
    // Before Search
    expect(screen.getByText(ip1)).toBeVisible()
    expect(screen.getByText(ip2)).toBeVisible()
    // Search
    await userEvent.type(searchInput, ip2)

    await waitFor(() => {
      expect(screen.queryByText(ip1)).toBeNull()
    })
    expect(screen.getByText(ip2)).toBeVisible()
    await userEvent.clear(searchInput)
  })

  describe('Validates Button Actions', () => {
    it('Validates button actions when instance status is "Terminating"', async () => {
      await act(async () => {
        render(<TestComponent history={history} />)
      })
      const status = 'Terminating'

      const row = screen.getByText(status).closest('tr')
      await waitFor(() => {
        expect(within(row).queryByTestId('ButtonTable Edit instance')).not.toBeInTheDocument()
        expect(within(row).queryByTestId('ButtonTable Delete instance')).not.toBeInTheDocument()
      })
    })

    it('Validates button actions when instance status is "Ready"', async () => {
      await act(async () => {
        render(<TestComponent history={history} />)
      })
      const status = 'Ready'

      const row = screen.getByText(status).closest('tr')

      await waitFor(() => {
        expect(within(row).getByTestId('ButtonTable Edit instance')).toBeInTheDocument()
        expect(within(row).getByTestId('ButtonTable Delete instance')).toBeInTheDocument()
      })
    })

    it('Validates button actions when instance status is "Provisioning"', async () => {
      await act(async () => {
        render(<TestComponent history={history} />)
      })
      const status = 'Provisioning'

      const row = screen.getByText(status).closest('tr')

      expect(within(row).getByLabelText('Edit instance')).toBeInTheDocument()
      expect(within(row).getByLabelText('Delete instance')).toBeInTheDocument()
      expect(within(row).getByTestId('Tooltip Message')).toBeInTheDocument()
    })

    it('Validates button actions when instance status is "Stopped"', async () => {
      await act(async () => {
        render(<TestComponent history={history} />)
      })
      const status = 'Stopped'

      const row = screen.getByText(status).closest('tr')
      expect(await within(row).findByLabelText('Edit instance')).toBeInTheDocument()
      expect(await within(row).findByLabelText('Delete instance')).toBeInTheDocument()
    })

    it('Validates button actions when instance status is "Failed"', async () => {
      await act(async () => {
        render(<TestComponent history={history} />)
      })
      const status = 'Failed'

      const row = screen.getByText(status).closest('tr')
      await waitFor(() => {
        expect(within(row).queryByTestId('ButtonTable Edit instance')).not.toBeInTheDocument()
      })
      expect(await within(row).findByLabelText('Delete instance')).toBeInTheDocument()
    })

    it('Checks Instances Delete button actions with Dialog box validation', async () => {
      await act(async () => {
        render(<TestComponent history={history} />)
      })

      const deleteButton = await screen.findAllByTestId('ButtonTable Delete instance')

      await waitFor(() => {
        expect(deleteButton[0]).toBeInTheDocument()
      })

      await userEvent.click(deleteButton[0])
      await screen.findByRole('dialog')
      expect(await screen.findByText((content, element) => content.startsWith('Delete instance'))).toBeInTheDocument()
    })
  })

  it('Checks Empty instances component when there are no instances.', async () => {
    mockEmptyInstances()
    await act(async () => {
      render(<TestComponent history={history} />)
    })
    expect(await screen.findByText('No instances found')).toBeInTheDocument()

    const launchInstanceLink = screen.getByTestId('LaunchinstanceEmptyViewButton')
    await waitFor(() => {
      expect(launchInstanceLink.getAttribute('href')).toBe('/compute/reserve')
    })
  })
})
