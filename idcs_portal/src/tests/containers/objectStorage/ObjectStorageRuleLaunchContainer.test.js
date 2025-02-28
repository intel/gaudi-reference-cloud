// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import AuthWrapper from '../../../utils/authWrapper/AuthWrapper'
import { mockIntelUser } from '../../mocks/authentication/authHelper'
import { act } from 'react'
import idcConfig from '../../../config/configurator'
import { clearAxiosMock } from '../../mocks/utils'
import { createMemoryHistory } from 'history'
import { Router, Routes, Route } from 'react-router-dom'
import { mockVendorsApi } from '../../mocks/shared/vendors'
import userEvent from '@testing-library/user-event'
import ObjectStorageRuleLaunchContainer from '../../../containers/objectStorage/ObjectStorageRuleLaunchContainer'
import { mockBaseObjectStorageStore, mockObjectStorage } from '../../mocks/objectStorage/ObjectStorage'

const TestComponent = ({ history, name }) => {
  history.push(`/buckets/d/${name}/lifecyclerule/reserve`)
  return (
    <AuthWrapper>
      <Router location={history.location} navigator={history}>
        <Routes>
          <Route
            path="/buckets/d/:param/lifecyclerule/reserve"
            element={<ObjectStorageRuleLaunchContainer history={history} name={name} />}
          />
        </Routes>
      </Router>
    </AuthWrapper>
  )
}

describe('Object Storage Lifecycle Rule Launch container unit test cases', () => {
  let history = null
  const name = mockBaseObjectStorageStore().items[2].metadata.name

  beforeEach(() => {
    clearAxiosMock()
    history = createMemoryHistory()
    mockVendorsApi()
  })

  beforeEach(() => {
    mockIntelUser()
  })

  beforeEach(() => {
    mockObjectStorage()
  })

  beforeAll(() => {
    idcConfig.REACT_APP_FEATURE_OBJECT_STORAGE = 1
  })

  it('Render component if loads correctly', async () => {
    await act(async () => {
      render(<TestComponent history={history} name={name} />)
    })

    expect(await screen.findByText('Add Lifecycle Rule')).toBeInTheDocument()
  })

  it('Show navigation bottom buttons', async () => {
    await act(async () => {
      render(<TestComponent history={history} name={name} />)
    })

    expect(await screen.findByTestId('btn-ObjectStorageRuleLaunch-navigationBottom Add')).toBeInTheDocument()

    expect(await screen.findByTestId('btn-ObjectStorageRuleLaunch-navigationBottom Cancel')).toBeInTheDocument()
  })

  it('Show form elements', async () => {
    await act(async () => {
      render(<TestComponent history={history} name={name} />)
    })

    expect(await screen.findByTestId('NameInput')).toBeInTheDocument()
    expect(await screen.findByTestId('PrefixInput')).toBeInTheDocument()
    expect(await screen.findByTestId('Input')).toBeInTheDocument()
    expect(await screen.findByTestId('NoncurrentexpirydaysInput')).toBeInTheDocument()
    expect(await screen.findByTestId('-Radio-option-DeleteMarker')).toBeInTheDocument()
    expect(await screen.findByTestId('-Radio-option-ExpiryDays')).toBeInTheDocument()
  })

  describe('Object Storage Rule Launch container form validation', () => {
    it('Validate Error message for name input', async () => {
      await act(async () => {
        render(<TestComponent history={history} name={name} />)
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
      const instanceName63 = 'instancename-instancename-instancename-instancename-instancenam' // 63 letters

      userEvent.clear(nameInput)
      userEvent.type(nameInput, instanceName64)
      await waitFor(() => {
        expect(screen.getByTestId('NameInput')).toHaveValue(instanceName63)
      })

      userEvent.clear(nameInput)
      await waitFor(() => {
        expect(screen.getByTestId('NameInvalidMessage')).toBeInTheDocument()
        expect(screen.getByTestId('NameInvalidMessage')).toHaveTextContent('Name is required')
      })
    })

    it('Create button disabled', async () => {
      await act(async () => {
        render(<TestComponent history={history} name={name} />)
      })

      expect(await screen.findByTestId('btn-ObjectStorageRuleLaunch-navigationBottom Add')).toBeVisible()
    })

    it('Disable/ Enable expiry days based on delete marker', async () => {
      await act(async () => {
        render(<TestComponent history={history} name={name} />)
      })
      const expiryDays = screen.getByTestId('Input')
      const deleteMarkerRadio = screen.getByTestId('-Radio-option-DeleteMarker')
      const expiryDaysRadio = screen.getByTestId('-Radio-option-ExpiryDays')

      expect(expiryDays).toHaveAttribute('disabled')
      userEvent.click(expiryDaysRadio)
      expect(expiryDaysRadio).toBeChecked()
      expect(expiryDays).not.toHaveAttribute('disabled')

      userEvent.click(deleteMarkerRadio)
      expect(deleteMarkerRadio).toBeChecked()
      expect(expiryDays).toHaveAttribute('disabled')
    })

    it('Validate if form is valid for submission', async () => {
      await act(async () => {
        render(<TestComponent history={history} name={name} />)
      })

      expect(await screen.findByTestId('btn-ObjectStorageRuleLaunch-navigationBottom Add')).toBeVisible()

      const inputName = screen.getByTestId('NameInput')
      fireEvent.change(inputName, ('name', { target: { value: 'test' } }))
      expect(inputName).toHaveValue('test')

      expect(await screen.findByTestId('btn-ObjectStorageRuleLaunch-navigationBottom Add')).not.toHaveAttribute(
        'disabled'
      )
    })
  })
})
