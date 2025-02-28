// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import AuthWrapper from '../../../utils/authWrapper/AuthWrapper'
import { mockIntelUser } from '../../mocks/authentication/authHelper'
import { act } from 'react'
import idcConfig from '../../../config/configurator'
import { clearAxiosMock } from '../../mocks/utils'
import { createMemoryHistory } from 'history'
import { Router } from 'react-router-dom'
import userEvent from '@testing-library/user-event'
import LoadBalancerLaunchContainer from '../../../containers/loadBalancer/LoadBalancerLaunchContainer'
import { mockInstanceTypes, mockInstances } from '../../mocks/compute/instances'
import { mockMachineImages } from '../../mocks/compute/machineImages'
import { mockVendorsApi } from '../../mocks/shared/vendors'
import { mockProductCatalog } from '../../mocks/compute/productCatalog'

const TestComponent = ({ history }) => {
  return (
    <AuthWrapper>
      <Router location={history.location} navigator={history}>
        <LoadBalancerLaunchContainer />
      </Router>
    </AuthWrapper>
  )
}

describe('Load Balancer Launch container unit test cases', () => {
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

  beforeAll(() => {
    idcConfig.REACT_APP_FEATURE_LOAD_BALANCER = 1
  })

  it('Render component if loads correctly', async () => {
    await act(async () => {
      render(<TestComponent history={history} />)
    })

    expect(await screen.findByText('Launch a Load Balancer')).toBeInTheDocument()
  })

  it('Show navigation bottom buttons', async () => {
    await act(async () => {
      render(<TestComponent history={history} />)
    })

    expect(await screen.findByTestId('btn-LoadBalancer-navigationBottomLaunch')).toBeInTheDocument()

    expect(await screen.findByTestId('btn-LoadBalancer-navigationBottomCancel')).toBeInTheDocument()
  })

  it('Show form elements', async () => {
    await act(async () => {
      render(<TestComponent history={history} />)
    })

    expect(await screen.findByTestId('NameInput')).toBeInTheDocument()
    expect(await screen.findByTestId('LoadBalancerNameValidMessage')).toBeInTheDocument()
    expect(await screen.findByTestId('LoadBalancerNameValidMessage')).toHaveTextContent(
      'Max length 50 characters. Letters, numbers and ‘- ‘ accepted.Name should start and end with an alphanumeric character.'
    )

    expect(await screen.findByTestId('SourceIPInput')).toBeInTheDocument()
    expect(await screen.findByTestId('LoadBalanceripsValidMessage')).toBeInTheDocument()
    expect(await screen.findByTestId('LoadBalanceripsValidMessage')).toHaveTextContent(
      'Use “any” to allow access from anywhere. Specify a single IP (ex.: 10.0.0.1) or CIDR-format (ex.: 10.0.0.1/24)'
    )

    expect(await screen.findByTestId('ListenerPortInput')).toBeInTheDocument()
    expect(await screen.findByTestId('InstancePortInput')).toBeInTheDocument()
    expect(await screen.findByTestId('Monitortype-form-select')).toBeInTheDocument()
    expect(await screen.findByTestId('Mode-form-select')).toBeInTheDocument()
    // expect(await screen.findByTestId('Selectortype-Radio-option-InstanceLabels')).toBeInTheDocument()
    expect(await screen.findByTestId('Selectortype-Radio-option-Instances')).toBeInTheDocument()
    expect(await screen.findByTestId('Selectortype-Radio-option-Instances')).toBeChecked()
  })

  describe('Load Balancer Launch container form validation', () => {
    it('Validate Error message for name input', async () => {
      await act(async () => {
        render(<TestComponent history={history} />)
      })
      const nameInput = screen.getByTestId('NameInput')
      userEvent.clear(nameInput)
      userEvent.type(nameInput, 'Name ')
      await waitFor(() => {
        expect(screen.getByTestId('NameInvalidMessage')).toBeInTheDocument()
        expect(screen.getByTestId('NameInvalidMessage')).toHaveTextContent(
          'Only lower case alphanumeric and hypen(-) allowed for Name:.'
        )
      })

      userEvent.clear(nameInput)
      userEvent.type(nameInput, 'name-')
      await waitFor(() => {
        expect(screen.getByTestId('NameInvalidMessage')).toBeInTheDocument()
        expect(screen.getByTestId('NameInvalidMessage')).toHaveTextContent(
          'Only lower case alphanumeric and hypen(-) allowed for Name:.'
        )
      })

      userEvent.clear(nameInput)
      userEvent.type(nameInput, '-name-')
      await waitFor(() => {
        expect(screen.getByTestId('NameInvalidMessage')).toBeInTheDocument()
        expect(screen.getByTestId('NameInvalidMessage')).toHaveTextContent(
          'Only lower case alphanumeric and hypen(-) allowed for Name:.'
        )
      })

      const instanceName64 = 'instancename-instancename-instancename-instancename-instancename' // 64 letters
      const instanceName50 = 'instancename-instancename-instancename-instancenam' // 50 letters

      userEvent.clear(nameInput)
      userEvent.type(nameInput, instanceName64)
      await waitFor(() => {
        expect(screen.getByTestId('NameInput')).toHaveValue(instanceName50)
      })

      userEvent.clear(nameInput)
      await waitFor(() => {
        expect(screen.getByTestId('NameInvalidMessage')).toBeInTheDocument()
        expect(screen.getByTestId('NameInvalidMessage')).toHaveTextContent('Name is required')
      })
    })

    it('Validate Error message for ip input', async () => {
      await act(async () => {
        render(<TestComponent history={history} />)
      })
      const ipInput = screen.getByTestId('SourceIPInput')

      userEvent.clear(ipInput)
      userEvent.type(ipInput, 'ANY')
      await waitFor(() => {
        expect(screen.getByTestId('SourceIPInvalidMessage')).toBeInTheDocument()
        expect(screen.getByTestId('SourceIPInvalidMessage')).toHaveTextContent('Invalid IP')
      })

      userEvent.clear(ipInput)
      userEvent.type(ipInput, '10.0.0')
      await waitFor(() => {
        expect(screen.getByTestId('SourceIPInvalidMessage')).toBeInTheDocument()
        expect(screen.getByTestId('SourceIPInvalidMessage')).toHaveTextContent('Invalid IP')
      })

      userEvent.clear(ipInput)
      userEvent.type(ipInput, '10.0.0.01')
      await waitFor(() => {
        expect(screen.getByTestId('SourceIPInvalidMessage')).toBeInTheDocument()
        expect(screen.getByTestId('SourceIPInvalidMessage')).toHaveTextContent('Invalid IP')
      })

      userEvent.clear(ipInput)
      await waitFor(() => {
        expect(screen.getByTestId('SourceIPInvalidMessage')).toBeInTheDocument()
        expect(screen.getByTestId('SourceIPInvalidMessage')).toHaveTextContent('Source IP is required')
      })
    })

    it('Validate Error message for Listener Port input', async () => {
      await act(async () => {
        render(<TestComponent history={history} />)
      })
      const portInput = screen.getByTestId('ListenerPortInput')

      userEvent.clear(portInput)
      userEvent.type(portInput, '0')
      await waitFor(() => {
        expect(screen.getByTestId('ListenerPortInvalidMessage')).toBeInTheDocument()
        expect(screen.getByTestId('ListenerPortInvalidMessage')).toHaveTextContent('Value less than 1 is not allowed.')
      })

      userEvent.clear(portInput)
      userEvent.type(portInput, 'a')
      await waitFor(() => {
        expect(screen.getByTestId('ListenerPortInvalidMessage')).toBeInTheDocument()
        expect(screen.getByTestId('ListenerPortInvalidMessage')).toHaveTextContent('Value less than 1 is not allowed.')
      })

      userEvent.clear(portInput)
      userEvent.type(portInput, '99898')
      await waitFor(() => {
        expect(screen.getByTestId('ListenerPortInvalidMessage')).toBeInTheDocument()
        expect(screen.getByTestId('ListenerPortInvalidMessage')).toHaveTextContent(
          'Value more than 65535 is not allowed.'
        )
      })

      userEvent.clear(portInput)
      await waitFor(() => {
        expect(screen.getByTestId('ListenerPortInvalidMessage')).toBeInTheDocument()
        expect(screen.getByTestId('ListenerPortInvalidMessage')).toHaveTextContent('Listener Port is required')
      })
    })

    it('Validate Error message for Instance Port input', async () => {
      await act(async () => {
        render(<TestComponent history={history} />)
      })
      const portInput = screen.getByTestId('InstancePortInput')

      userEvent.clear(portInput)
      userEvent.type(portInput, '0')
      await waitFor(() => {
        expect(screen.getByTestId('InstancePortInvalidMessage')).toBeInTheDocument()
        expect(screen.getByTestId('InstancePortInvalidMessage')).toHaveTextContent('Value less than 1 is not allowed.')
      })

      userEvent.clear(portInput)
      userEvent.type(portInput, 'a')
      await waitFor(() => {
        expect(screen.getByTestId('InstancePortInvalidMessage')).toBeInTheDocument()
        expect(screen.getByTestId('InstancePortInvalidMessage')).toHaveTextContent('Value less than 1 is not allowed.')
      })

      userEvent.clear(portInput)
      userEvent.type(portInput, '99898')
      await waitFor(() => {
        expect(screen.getByTestId('InstancePortInvalidMessage')).toBeInTheDocument()
        expect(screen.getByTestId('InstancePortInvalidMessage')).toHaveTextContent(
          'Value more than 65535 is not allowed.'
        )
      })

      userEvent.clear(portInput)
      await waitFor(() => {
        expect(screen.getByTestId('InstancePortInvalidMessage')).toBeInTheDocument()
        expect(screen.getByTestId('InstancePortInvalidMessage')).toHaveTextContent('Instance Port is required')
      })
    })

    it('Validate if form is valid for submission', async () => {
      await act(async () => {
        render(<TestComponent history={history} />)
      })

      expect(await screen.findByTestId('btn-LoadBalancer-navigationBottomLaunch')).toBeVisible()

      const inputName = screen.getByTestId('NameInput')
      fireEvent.change(inputName, ('name', { target: { value: 'test' } }))
      expect(inputName).toHaveValue('test')

      const ipInput = screen.getByTestId('SourceIPInput')
      fireEvent.change(ipInput, ('name', { target: { value: '10.0.0.1' } }))
      expect(ipInput).toHaveValue('10.0.0.1')

      const listenerPortInput = screen.getByTestId('ListenerPortInput')
      fireEvent.change(listenerPortInput, ('name', { target: { value: '80' } }))
      expect(listenerPortInput).toHaveValue('80')

      const instancePortInput = screen.getByTestId('InstancePortInput')
      fireEvent.change(instancePortInput, ('name', { target: { value: '80' } }))
      expect(instancePortInput).toHaveValue('80')

      const monitorSelect = screen.getByTestId('Monitortype-form-select')
      userEvent.click(monitorSelect)
      const tcpOption = await screen.findByTestId('Monitortype-form-select-option-TCP')
      userEvent.click(tcpOption)

      // const tagsRadio = screen.getByTestId('Selectortype-Radio-option-InstanceLabels')
      // userEvent.click(tagsRadio)

      // const tagKeyInput = screen.getByTestId('Instancestags-input-dictionary-Key-0')
      // fireEvent.change(tagKeyInput, ('name', { target: { value: 'tag1' } }))
      // expect(tagKeyInput).toHaveValue('tag1')

      // const tagValueInput = screen.getByTestId('Instancestags-input-dictionary-Value-0')
      // fireEvent.change(tagValueInput, ('name', { target: { value: 'value1' } }))
      // expect(tagValueInput).toHaveValue('value1')

      expect(await screen.findByTestId('btn-LoadBalancer-navigationBottomLaunch')).not.toHaveAttribute('disabled')

      const addListenerButton = screen.getByTestId('btn-loadBalancer-addListener')
      userEvent.click(addListenerButton)

      expect(await screen.findByTestId('btn-LoadBalancer-navigationBottomLaunch')).toBeVisible()

      const listenerPortInput1 = screen.getAllByTestId('ListenerPortInput')[1]
      fireEvent.change(listenerPortInput1, ('name', { target: { value: '78' } }))
      expect(listenerPortInput1).toHaveValue('78')

      const instancePortInput1 = screen.getAllByTestId('InstancePortInput')[1]
      fireEvent.change(instancePortInput1, ('name', { target: { value: '80' } }))
      expect(instancePortInput1).toHaveValue('80')

      const monitorSelect1 = screen.getAllByTestId('Monitortype-form-select')[1]
      userEvent.click(monitorSelect1)
      const tcpOptions = await screen.findAllByTestId('Monitortype-form-select-option-TCP')
      userEvent.click(tcpOptions[1])

      // const tagsRadio1 = screen.getAllByTestId('Selectortype-Radio-option-InstanceLabels')[1]
      // userEvent.click(tagsRadio1)

      // const tagKeyInput1 = screen.getAllByTestId('Instancestags-input-dictionary-Key-0')[1]
      // fireEvent.change(tagKeyInput1, ('name', { target: { value: 'tag1' } }))
      // expect(tagKeyInput1).toHaveValue('tag1')

      // const tagValueInput1 = screen.getAllByTestId('Instancestags-input-dictionary-Value-0')[1]
      // fireEvent.change(tagValueInput1, ('name', { target: { value: 'value1' } }))
      // expect(tagValueInput1).toHaveValue('value1')

      expect(await screen.findByTestId('btn-LoadBalancer-navigationBottomLaunch')).not.toHaveAttribute('disabled')
    })
  })
})
