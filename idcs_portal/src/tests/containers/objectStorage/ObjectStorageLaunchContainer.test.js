// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import AuthWrapper from '../../../utils/authWrapper/AuthWrapper'
import { mockIntelUser } from '../../mocks/authentication/authHelper'
import { act } from 'react'
import idcConfig from '../../../config/configurator'
import { mockObjectStorageProduct } from '../../mocks/storage/Storage'
import { clearAxiosMock } from '../../mocks/utils'
import { createMemoryHistory } from 'history'
import { Router } from 'react-router-dom'
import { mockVendorsApi } from '../../mocks/shared/vendors'
import ObjectStorageLaunchContainer from '../../../containers/objectStorage/ObjectStorageLaunchContainer'
import userEvent from '@testing-library/user-event'

const TestComponent = ({ history }) => {
  return (
    <AuthWrapper>
      <Router location={history.location} navigator={history}>
        <ObjectStorageLaunchContainer />
      </Router>
    </AuthWrapper>
  )
}

describe('Object Storage Launch container unit test cases', () => {
  let history = null

  beforeEach(() => {
    clearAxiosMock()
    history = createMemoryHistory()
    mockVendorsApi()
  })

  beforeEach(() => {
    mockIntelUser()
  })

  beforeEach(() => {
    mockObjectStorageProduct()
  })

  beforeAll(() => {
    idcConfig.REACT_APP_FEATURE_OBJECT_STORAGE = 1
  })

  it('Render component if loads correctly', async () => {
    await act(async () => {
      render(<TestComponent history={history} />)
    })

    expect(await screen.findByText('Create storage bucket')).toBeInTheDocument()
  })

  it('Show navigation bottom buttons', async () => {
    await act(async () => {
      render(<TestComponent history={history} />)
    })

    expect(await screen.findByTestId('btn-ObjectStorage-navigationBottom Create')).toBeInTheDocument()

    expect(await screen.findByTestId('btn-ObjectStorage-navigationBottom Cancel')).toBeInTheDocument()
  })

  it('Show form elements', async () => {
    await act(async () => {
      render(<TestComponent history={history} />)
    })

    expect(await screen.findByTestId('NameInput')).toBeInTheDocument()
    expect(await screen.findByTestId('StorageBucketNameValidMessage')).toBeInTheDocument()
    expect(await screen.findByTestId('StorageBucketNameValidMessage')).toHaveTextContent(
      'Max length 50 characters. Letters, numbers and ‘- ‘ accepted.Name should start and end with an alphanumeric character.'
    )

    expect(await screen.findByTestId('DescriptionTextArea')).toBeInTheDocument()
    expect(await screen.findByTestId('StorageBucketDescriptionValidMessage')).toBeInTheDocument()
    expect(await screen.findByTestId('StorageBucketDescriptionValidMessage')).toHaveTextContent(
      'Provide a description for this bucket.'
    )
  })

  describe('Object Storage Launch container form validation', () => {
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

    it('Validate max length for description input', async () => {
      await act(async () => {
        render(<TestComponent history={history} />)
      })
      const nameInput = screen.getByTestId('DescriptionTextArea')

      const inputLetters103 =
        'instancename-instancename-instancename-instancename-instancename-instancename-instancename-instancename' // 103 letters
      const inputLetters100 =
        'instancename-instancename-instancename-instancename-instancename-instancename-instancename-instancen' // 100 letters

      userEvent.clear(nameInput)
      userEvent.type(nameInput, inputLetters103)
      await waitFor(() => {
        expect(screen.getByTestId('DescriptionTextArea')).toHaveValue(inputLetters100)
      })
    })

    it('Create button', async () => {
      await act(async () => {
        render(<TestComponent history={history} />)
      })

      expect(await screen.findByTestId('btn-ObjectStorage-navigationBottom Create')).toBeVisible()
    })

    it('Validate if form is valid for submission', async () => {
      await act(async () => {
        render(<TestComponent history={history} />)
      })

      expect(await screen.findByTestId('btn-ObjectStorage-navigationBottom Create')).toBeVisible()

      const inputName = screen.getByTestId('NameInput')
      fireEvent.change(inputName, ('name', { target: { value: 'test' } }))
      expect(inputName).toHaveValue('test')

      expect(await screen.findByTestId('btn-ObjectStorage-navigationBottom Create')).not.toHaveAttribute('disabled')
    })
  })
})
