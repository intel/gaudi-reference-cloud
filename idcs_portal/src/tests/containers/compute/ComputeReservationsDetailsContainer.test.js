// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

// eslint-disable-next-line react/display-name
import { render, screen, waitFor, within } from '@testing-library/react'
import { Router, Routes, Route, useSearchParams } from 'react-router-dom'
import { createMemoryHistory } from 'history'
import { act } from 'react'
import AuthWrapper from '../../../utils/authWrapper/AuthWrapper'
import { clearAxiosMock } from '../../mocks/utils'
import ComputeReservationsDetailsContainer from '../../../containers/compute/ComputeReservationsDetailsContainer'
import { mockIntelUser } from '../../mocks/authentication/authHelper'
import { mockBaseInstancesStore, mockInstanceTypes, mockInstances } from '../../mocks/compute/instances'
import { mockMachineImages } from '../../mocks/compute/machineImages'
import { mockVendorsApi } from '../../mocks/shared/vendors'
import { mockProductCatalog } from '../../mocks/compute/productCatalog'
import userEvent from '@testing-library/user-event'

jest.mock('react-router-dom', () => {
  return {
    ...jest.requireActual('react-router-dom'),
    useSearchParams: jest.fn()
  }
})

const TestComponent = ({ history, name }) => {
  history.push(`/compute/d/${name}`)
  return (
    <AuthWrapper>
      <Router location={history.location} navigator={history}>
        <Routes>
          <Route path="/compute/d/:param" element={<ComputeReservationsDetailsContainer history={history} />} />
        </Routes>
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
    useSearchParams.mockReturnValue([new URLSearchParams('tab=Details'), jest.fn()])
  })

  beforeEach(() => {
    mockIntelUser()
    mockInstances()
    mockInstanceTypes()
    mockMachineImages()
    mockVendorsApi()
    mockProductCatalog()
  })

  it('Displays instance details when we click on specific Instance', async () => {
    const name = mockBaseInstancesStore().items[0].metadata.name
    await act(async () => {
      render(<TestComponent history={history} name={name} />)
    })
    const instanceName = mockBaseInstancesStore().items[0].metadata.name
    expect(await screen.findByText(`Instance: ${instanceName}`)).toBeInTheDocument()
    expect(screen.getByTestId('btn-computereservation-how-to-connect')).toBeInTheDocument()
    expect(await screen.findByText('Instance type information')).toBeInTheDocument()
    expect(screen.getByTestId('myReservationActionsDropdownButton')).toBeInTheDocument()

    await userEvent.click(await screen.findByText(/^Networking/i))
    expect(await screen.findByText('Networking interfaces')).toBeInTheDocument()

    await userEvent.click(await screen.findByText(/^Security/i))
    expect(await screen.findByText('Instance Public Keys')).toBeInTheDocument()
  })

  describe('Validates Button Actions', () => {
    it('Validates button actions when instance status is "Terminating"', async () => {
      const name = mockBaseInstancesStore().items[4].metadata.name
      await act(async () => {
        render(<TestComponent history={history} name={name} />)
      })

      // No action buttons should be there in Instance details.
      const actionButtons = screen.queryByTestId('myReservationActionsDropdownButton')
      expect(actionButtons).toBeFalsy()
    })

    it('Validates button actions when instance status is "Ready"', async () => {
      const name = mockBaseInstancesStore().items[1].metadata.name
      await act(async () => {
        render(<TestComponent history={history} name={name} />)
      })

      await waitFor(() => {
        expect(screen.getByTestId('myReservationActionsDropdownButton')).toBeInTheDocument()
      })
      const actionButtons = screen.getByTestId('myReservationActionsDropdownButton').getElementsByTagName('button')[0]
      await userEvent.click(actionButtons)
      // Checking whether the right action buttons are available.
      await waitFor(() => {
        const actionButtonsInside = screen.getByTestId('myReservationActionsDropdownButton')
        expect(within(actionButtonsInside).getByText('Edit')).toBeTruthy()
        expect(within(actionButtonsInside).getByText('Delete')).toBeTruthy()
      })
    })

    it('Validates button actions when instance status is "Provisioning"', async () => {
      const name = mockBaseInstancesStore().items[0].metadata.name
      await act(async () => {
        render(<TestComponent history={history} name={name} />)
      })
      await waitFor(() => {
        expect(screen.getByTestId('myReservationActionsDropdownButton')).toBeInTheDocument()
      })
      const actionButtons = screen.getByTestId('myReservationActionsDropdownButton').getElementsByTagName('button')[0]
      await userEvent.click(actionButtons)
      // Checking whether the right action buttons are available.
      await waitFor(() => {
        const actionButtonsInside = screen.getByTestId('myReservationActionsDropdownButton')
        expect(within(actionButtonsInside).getByText('Edit')).toBeTruthy()
        expect(within(actionButtonsInside).getByText('Delete')).toBeTruthy()
      })
    })

    it('Validates button actions when instance status is "Stopped"', async () => {
      const name = mockBaseInstancesStore().items[2].metadata.name
      await act(async () => {
        render(<TestComponent history={history} name={name} />)
      })

      await waitFor(() => {
        expect(screen.getByTestId('myReservationActionsDropdownButton')).toBeInTheDocument()
      })
      const actionButtons = screen.getByTestId('myReservationActionsDropdownButton').getElementsByTagName('button')[0]
      await userEvent.click(actionButtons)
      // Checking whether the right action buttons are available.
      const actionButtonsInside = screen.getByTestId('myReservationActionsDropdownButton')
      expect(within(actionButtonsInside).getByText('Edit')).toBeTruthy()
      expect(within(actionButtonsInside).queryByText('Delete')).toBeTruthy()
    })

    it('Validates button actions when instance status is "Failed"', async () => {
      const name = mockBaseInstancesStore().items[3].metadata.name
      await act(async () => {
        render(<TestComponent history={history} name={name} />)
      })

      await waitFor(() => {
        expect(screen.getByTestId('myReservationActionsDropdownButton')).toBeInTheDocument()
      })
      const actionButtons = screen.getByTestId('myReservationActionsDropdownButton').getElementsByTagName('button')[0]
      await userEvent.click(actionButtons)
      // Checking whether the right action buttons are available.
      const actionButtonsInside = screen.getByTestId('myReservationActionsDropdownButton')
      expect(within(actionButtonsInside).queryByText('Edit')).toBe(null)
      expect(within(actionButtonsInside).getByText('Delete')).toBeTruthy()
    })
  })

  it('Validates How to connect button for Provisioning Instance', async () => {
    const name = mockBaseInstancesStore().items.find((x) => x.status.phase === 'Provisioning').metadata.name
    await act(async () => {
      render(<TestComponent history={history} name={name} />)
    })
    global.navigator.clipboard = { writeText: jest.fn() }

    const howToConnectButton = screen.getByTestId('btn-computereservation-how-to-connect')
    expect(howToConnectButton).toBeInTheDocument()

    await userEvent.click(howToConnectButton)
    expect(await screen.findByRole('dialog')).toBeInTheDocument()
    expect(screen.getByText('How to connect to your instance')).toBeInTheDocument()
    expect(
      screen.getByText(
        'We are currently awaiting the completion of instance initialization. Once the instance is ready, we will show you how to connect to your instance.'
      )
    ).toBeInTheDocument()
  })

  it('Validates How to connect button for Failed Instance', async () => {
    const name = mockBaseInstancesStore().items.find((x) => x.status.phase === 'Failed').metadata.name
    await act(async () => {
      render(<TestComponent history={history} name={name} />)
    })

    global.navigator.clipboard = { writeText: jest.fn() }

    const howToConnectButton = screen.getByTestId('btn-computereservation-how-to-connect')
    expect(howToConnectButton).toBeInTheDocument()

    await userEvent.click(howToConnectButton)
    expect(await screen.findByRole('dialog')).toBeInTheDocument()
    expect(screen.getByText('How to connect to your instance')).toBeInTheDocument()
    expect(
      screen.getByText('The instance could not be initialized. Please try creating a new one.')
    ).toBeInTheDocument()
  })

  it('Validates How to connect button for Ready Instance', async () => {
    const name = mockBaseInstancesStore().items.find((x) => x.status.phase === 'Ready').metadata.name
    await act(async () => {
      render(<TestComponent history={history} name={name} />)
    })
    global.navigator.clipboard = { writeText: jest.fn() }

    const howToConnectButton = screen.getByTestId('btn-computereservation-how-to-connect')
    expect(howToConnectButton).toBeInTheDocument()

    await userEvent.click(howToConnectButton)
    expect(await screen.findByRole('dialog')).toBeInTheDocument()
    expect(screen.getByText('How to connect to your instance')).toBeInTheDocument()
    expect(screen.getByText('Select your OS:')).toBeInTheDocument()

    const windowsRadioButton = screen.getByTestId('WindowsRadioButton')
    expect(windowsRadioButton).toBeInTheDocument()
    await userEvent.click(windowsRadioButton)
    expect(windowsRadioButton).toBeChecked()

    const LinuxRadioButton = screen.getByTestId('LinuxRadioButton')
    expect(LinuxRadioButton).toBeInTheDocument()
    expect(LinuxRadioButton).not.toBeChecked()
    await userEvent.click(LinuxRadioButton)
    expect(LinuxRadioButton).toBeChecked()

    const dialog = await screen.findByRole('dialog')
    expect(dialog).toBeInTheDocument()

    const preCode = dialog.getElementsByClassName('code-line')
    expect(preCode.length).toEqual(3)

    expect(preCode[1]).toHaveTextContent('chmod 400 my-key.ssh')
    expect(preCode[2]).toHaveTextContent('ssh -J guest-dev9@10.165.62.252 ubuntu@172.16.0.172')

    const copyButton1 = preCode[1].getElementsByTagName('button')[0]
    expect(copyButton1).toHaveTextContent('Copy')

    const copyButton2 = preCode[2].getElementsByTagName('button')[0]
    expect(copyButton2).toHaveTextContent('Copy')

    await userEvent.click(copyButton1)
    expect(navigator.clipboard.writeText).toHaveBeenCalledWith('chmod 400 my-key.ssh')

    await userEvent.click(copyButton2)
    expect(navigator.clipboard.writeText).toHaveBeenCalledWith('ssh -J guest-dev9@10.165.62.252 ubuntu@172.16.0.172')

    const howToConnectClose = within(dialog).getByTestId('HowToConnectClose')
    expect(howToConnectClose).toBeInTheDocument()
    await userEvent.click(howToConnectClose)
    await waitFor(() => {
      expect(screen.queryAllByRole('dialog').length).toEqual(0)
    })
  })

  describe('Validates Tab Params', () => {
    it('Validates default tab for detail page', async () => {
      const name = mockBaseInstancesStore().items.find((x) => x.status.phase === 'Ready').metadata.name
      await act(async () => {
        render(<TestComponent history={history} name={name} />)
      })

      await waitFor(() => {
        expect(screen.getByTestId('DetailsTab')).toHaveClass('tap-active')
      })
      expect(await screen.findByText('Instance type information')).toBeInTheDocument()
    })

    it('Validates active tab for incorrect tab params', async () => {
      const name = mockBaseInstancesStore().items.find((x) => x.status.phase === 'Ready').metadata.name
      const randomTab = 'tab=RandomTab'
      useSearchParams.mockReturnValue([new URLSearchParams(randomTab), jest.fn()])
      await act(async () => {
        render(<TestComponent history={history} name={name} />)
      })

      await waitFor(() => {
        expect(screen.getByTestId('DetailsTab')).toHaveClass('tap-active')
      })
      expect(await screen.findByText('Instance type information')).toBeInTheDocument()
    })

    it('Validates active tab for correct tab params', async () => {
      const name = mockBaseInstancesStore().items.find((x) => x.status.phase === 'Ready').metadata.name
      const randomTab = 'tab=Networking'
      useSearchParams.mockReturnValue([new URLSearchParams(randomTab), jest.fn()])
      await act(async () => {
        render(<TestComponent history={history} name={name} />)
      })

      await waitFor(() => {
        expect(screen.getByTestId('NetworkingTab')).toHaveClass('tap-active')
      })
      expect(await screen.findByText('Networking interfaces')).toBeInTheDocument()
    })
  })
})
