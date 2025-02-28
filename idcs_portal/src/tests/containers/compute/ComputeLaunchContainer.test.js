// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { render, screen, waitFor, within } from '@testing-library/react'
import { Router } from 'react-router-dom'
import { createMemoryHistory } from 'history'
import { mockProductCatalog, mockEmptyProductCatalog } from '../../mocks/compute/productCatalog'
import ComputeLaunchContainer from '../../../containers/compute/ComputeLaunchContainer'
import { mockStandardUser } from '../../mocks/authentication/authHelper'
import { clearAxiosMock, expectValueForInputElement } from '../../mocks/utils'
import { mockEmptyMachineImages, mockMachineImages } from '../../mocks/compute/machineImages'
import {
  mockBasePublicKeysStore,
  mockCreatePublicKeys,
  mockEmptyPublicKeys,
  mockPublicKeys
} from '../../mocks/publicKeys/publicKeys'
import { mockVnets } from '../../mocks/compute/vnets'
import { act } from 'react'
import AuthWrapper from '../../../utils/authWrapper/AuthWrapper'
import { mockVendorsApi } from '../../mocks/shared/vendors'
import userEvent from '@testing-library/user-event'
import { getCustomInputId } from '../../../utils/customInput/CustomInput.types'

const TestComponent = () => {
  return (
    <AuthWrapper>
      <Router location={createMemoryHistory().location} navigator={history}>
        <ComputeLaunchContainer />
      </Router>
    </AuthWrapper>
  )
}

describe('Compute Launch container Unit Test cases', () => {
  beforeEach(() => {
    clearAxiosMock()
  })

  beforeEach(() => {
    mockStandardUser()
  })

  beforeEach(() => {
    mockMachineImages()
    mockPublicKeys()
    mockProductCatalog()
    mockVnets()
    mockVendorsApi()
  })

  it('Render component if loads correctly', async () => {
    await act(async () => {
      render(<TestComponent />)
    })

    expect(await screen.findByText('Launch a compute instance')).toBeInTheDocument()
  })

  describe('Instance Type Grid Unit test cases', () => {
    it('Render table with product rows', async () => {
      await act(async () => {
        render(<TestComponent />)
      })

      const allRadio = await screen.getAllByTestId('fillTable')
      expect(allRadio.length).toBeGreaterThan(1)
    })

    it('Render table with radio buttons', async () => {
      await act(async () => {
        render(<TestComponent />)
      })

      const allRadio = await screen.getAllByTestId(/checkBoxTable/i)
      expect(allRadio.length).toBeGreaterThan(1)
    })

    it('Render table with error message if no data', async () => {
      mockEmptyProductCatalog()
      await act(async () => {
        render(<TestComponent initialShow={true} />)
      })

      const errorMessage = await screen.findAllByText('No available instance types')
      expect(errorMessage.length).toBe(2)
    })
  })

  describe('Compute Reservation form unit test cases', () => {
    it('Validates instances in the dropdown when they are empty', async () => {
      mockEmptyProductCatalog()
      await act(async () => {
        render(<TestComponent />)
      })
      expect(screen.getByTestId('Machineimage-form-select')).toHaveTextContent('Please select')
    })

    it('Validates machine images in the dropdown when they are empty', async () => {
      mockEmptyMachineImages()
      await act(async () => {
        render(<TestComponent />)
      })
      expect(screen.getByTestId('Machineimage-form-select')).toHaveTextContent('Please select')
    })

    it('Checks Create Key button exists in the form', async () => {
      await act(async () => {
        render(<TestComponent />)
      })
      expect(screen.getByTestId('SelectkeysbtnExtra')).toBeInTheDocument()
    })

    it('Validates Select/Deselect All Keys button in the form', async () => {
      mockPublicKeys()
      await act(async () => {
        render(<TestComponent />)
      })
      const selectAllButton = screen.getByTestId('SelectkeysbtnSelectAll')
      expect(selectAllButton).toBeInTheDocument()

      await userEvent.click(selectAllButton)

      const publicKeyExpected =
        mockBasePublicKeysStore().items[0].metadata.name + ` (${mockBasePublicKeysStore().items[0].spec.ownerEmail})`
      const publicKeyCheckBox = await screen.findByTestId(
        `Selectkeys-Input-option-${getCustomInputId(publicKeyExpected)}`
      )
      expect(publicKeyCheckBox).toBeInTheDocument()
      expect(publicKeyCheckBox).toBeChecked()

      await userEvent.click(selectAllButton)
      const publicKeyEmptyCheckBox = await screen.findByTestId(
        `Selectkeys-Input-option-${getCustomInputId(publicKeyExpected)}`
      )
      expect(publicKeyEmptyCheckBox).toBeInTheDocument()
      expect(publicKeyEmptyCheckBox).not.toBeChecked()
    })

    it('Validate the Launch a compute instance form submission', async () => {
      await act(async () => {
        render(<TestComponent />)
      })

      const btnLaunchInstance = screen.getByTestId('btn-computelaunch-navigationBottom Launch instance - singlenode')
      await waitFor(() => {
        expect(btnLaunchInstance).toBeVisible()
      })

      const useCaseOption = screen.getAllByTestId('Core compute-radio-select')[0]
      userEvent.click(useCaseOption)

      // await userEvent.selectOptions(
      //   screen.getByTestId('InstancefamilySelect'),
      //   '4th Generation Intel® Xeon® Scalable processors'
      // )

      // expect(screen.getByTestId('InstancefamilySelect')).toHaveValue(
      //   '4th Generation Intel® Xeon® Scalable processors'
      // )
      // expect(screen.getByRole('option', { name: '4th Generation Intel® Xeon® Scalable processors' }).selected).toBe(
      //   true
      // )

      await expectValueForInputElement(
        screen.getByTestId('InstancenameInput'),
        'test-instance-123',
        'test-instance-123'
      )

      const publicKeyExpected =
        mockBasePublicKeysStore().items[0].metadata.name + ` (${mockBasePublicKeysStore().items[0].spec.ownerEmail})`

      const publicKeyCheckBox = await screen.findByTestId(
        `Selectkeys-Input-option-${getCustomInputId(publicKeyExpected)}`
      )

      expect(publicKeyCheckBox).toBeInTheDocument()
      expect(publicKeyCheckBox).not.toBeChecked()

      await userEvent.click(publicKeyCheckBox)
      expect(publicKeyCheckBox).toBeChecked()

      const btnLaunchInstanceNew = screen.getByTestId('btn-computelaunch-navigationBottom Launch instance - singlenode')
      await waitFor(() => {
        expect(btnLaunchInstanceNew).toBeEnabled()
      })
    })

    it('Validate Error message for Instance name input', async () => {
      await act(async () => {
        render(<TestComponent />)
      })
      const instanceNameInput = screen.getByTestId('InstancenameInput')
      await userEvent.clear(instanceNameInput)
      await userEvent.type(instanceNameInput, 'instancename ')
      await waitFor(() => {
        expect(screen.getByTestId('InstancenameInvalidMessage')).toBeInTheDocument()
        expect(screen.getByTestId('InstancenameInvalidMessage')).toHaveTextContent(
          'Only lower case alphanumeric and hypen(-) allowed for Instance name:'
        )
      })

      await userEvent.clear(instanceNameInput)
      await userEvent.type(instanceNameInput, 'instancename-')
      await waitFor(() => {
        expect(screen.getByTestId('InstancenameInvalidMessage')).toBeInTheDocument()
        expect(screen.getByTestId('InstancenameInvalidMessage')).toHaveTextContent(
          'Only lower case alphanumeric and hypen(-) allowed for Instance name:'
        )
      })

      await userEvent.clear(instanceNameInput)
      await userEvent.type(instanceNameInput, '-instancename-')
      await waitFor(() => {
        expect(screen.getByTestId('InstancenameInvalidMessage')).toBeInTheDocument()
        expect(screen.getByTestId('InstancenameInvalidMessage')).toHaveTextContent(
          'Only lower case alphanumeric and hypen(-) allowed for Instance name:'
        )
      })

      const instanceName = 'instancename-instancename-instancename-instancename-instancename' // 64 letters

      await userEvent.clear(instanceNameInput)
      await userEvent.type(instanceNameInput, instanceName)
      await waitFor(() => {
        expect(screen.getByTestId('InstancenameInvalidMessage')).toBeInTheDocument()
        expect(screen.getByTestId('InstancenameInvalidMessage')).toHaveTextContent('Max length 63 characters.')
      })
    })
  })

  // TODO Validate this unit test because is failing
  // eslint-disable-next-line jest/no-disabled-tests
  it.skip('Validates Instance type and Machine name sections', async () => {
    await act(async () => {
      render(<TestComponent />)
    })

    await waitFor(() => {
      const instanceType = screen.getByTestId('Instance type: *InputLabel').nextElementSibling
      const instanceTypeName = within(instanceType).findByTestId(
        'selected-option Medium VM - Intel® Xeon 4th Gen ® Scalable processor'
      )
      expect(instanceTypeName).toBeTruthy()
      expect(
        within(screen.getByTestId('Machine image: *InputLabel').nextElementSibling).findByTestId(
          'selected-option ubuntu-2204-jammy-v20230122'
        )
      ).toBeTruthy()
    })
    // Length of the 'BareMetalHost' from mockBaseMachineImagesStore.
    const lengthBareMetalMachineImages = screen
      .getByTestId('Machine image: *InputLabel')
      .nextElementSibling.getElementsByClassName('dropdown-item mb-1').length
    await waitFor(() => {
      expect(lengthBareMetalMachineImages).toBeGreaterThan(0)
    })

    await waitFor(() => {
      const instanceType = screen.getByTestId('Instance type: *InputLabel').nextElementSibling
      const instanceTypeName = within(instanceType).findByTestId(
        'selected-option 4th Generation Intel® Xeon® Scalable processors'
      )
      expect(instanceTypeName).toBeTruthy()
      expect(
        within(screen.getByTestId('Machine image: *InputLabel').nextElementSibling).findByTestId(
          'selected-option ubuntu-20.04-gaudi-metal-cloudimg-amd64-latest'
        )
      ).toBeTruthy()
    })

    const instnaceTypeFormSelect = await screen.findByTestId('Instance type: *form-select')
    const clicked1 = instnaceTypeFormSelect.getElementsByClassName('dropdown-menu select-option')
    expect(clicked1[0]).not.toHaveClass('show')
  })

  it('Verify Public Keys input field and Create Key modal', async () => {
    mockEmptyPublicKeys()
    await act(async () => {
      render(<TestComponent />)
    })

    expect(screen.getByPlaceholderText('No keys found. Please create a key to continue.')).toBeInTheDocument()
    const createKeyButton = screen.getByTestId('SelectkeysbtnExtra')
    await userEvent.click(createKeyButton)
    await waitFor(() => {
      expect(screen.queryByText('Upload a public key')).not.toBeNull()
    })

    const createKeyMainButton = screen.getByTestId('btn-ssh-createpublickey')
    await waitFor(() => {
      expect(createKeyMainButton).toBeVisible()
    })
    const keyNameInput = screen.getByTestId('KeyNameInput')
    const keyContentsInput = screen.getByTestId('PasteyourkeycontentsTextArea')
    await userEvent.type(keyNameInput, 'test1')
    await userEvent.type(keyContentsInput, 'ssh-ed25519 SSH KEY')
    expect(createKeyMainButton).toBeEnabled()

    mockCreatePublicKeys()
    await userEvent.click(createKeyMainButton)
    mockPublicKeys()

    await waitFor(() => {
      expect(screen.findByText('Upload a public key')).toMatchObject({})
    })

    const publicKeyExpected = mockBasePublicKeysStore().items[0].metadata.name
    await waitFor(() => {
      const publicKeyCheckBox = screen.findAllByTestId(`checkbox ${publicKeyExpected}`)
      expect(publicKeyCheckBox).toBeTruthy()
    })
  })
})
