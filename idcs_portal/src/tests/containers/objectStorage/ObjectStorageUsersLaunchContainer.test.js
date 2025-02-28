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
import { mockBaseObjectStorageStore, mockObjectStorage } from '../../mocks/objectStorage/ObjectStorage'
import ObjectStorageUsersLaunchContainer from '../../../containers/objectStorage/ObjectStorageUsersLaunchContainer'

const TestComponent = ({ history }) => {
  return (
    <AuthWrapper>
      <Router location={history.location} navigator={history}>
        <ObjectStorageUsersLaunchContainer />
      </Router>
    </AuthWrapper>
  )
}

const btnTestId = 'btn-objectStorageUsersLaunch-navigationBottom Create'

describe('Object Storage Users Launch container unit test cases', () => {
  let history = null

  beforeEach(() => {
    clearAxiosMock()
    history = createMemoryHistory()
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
      render(<TestComponent history={history} />)
    })

    expect(await screen.findByText('Create principal')).toBeInTheDocument()
  })

  it('Show navigation bottom buttons', async () => {
    await act(async () => {
      render(<TestComponent history={history} />)
    })

    expect(await screen.findByTestId(btnTestId)).toBeInTheDocument()

    expect(await screen.findByTestId('btn-objectStorageUsersLaunch-navigationBottom Cancel')).toBeInTheDocument()
  })

  it('Show form elements', async () => {
    await act(async () => {
      render(<TestComponent history={history} />)
    })

    expect(await screen.findByTestId('NameInput')).toBeInTheDocument()
    expect(await screen.findByTestId('ObjectStorageUsersLaunchNameValidMessage')).toBeInTheDocument()
    expect(await screen.findByTestId('ObjectStorageUsersLaunchNameValidMessage')).toHaveTextContent(
      'Max length 12 characters.'
    )

    expect(await screen.findByTestId('Applypermissions-Radio-option-Forallbuckets')).toBeInTheDocument()
    expect(await screen.findByTestId('Applypermissions-Radio-option-Forallbuckets')).toBeChecked()
    expect(await screen.findByTestId('Applypermissions-Radio-option-Perbucket')).toBeInTheDocument()
    expect(await screen.findByTestId('Allowedactions-Input-option-selectAll')).toBeInTheDocument()
    expect(await screen.findByTestId('Allowedpolicies-Input-option-selectAll')).toBeInTheDocument()
    expect(await screen.findByTestId('AllowedpoliciespathInput')).toBeInTheDocument()
  })

  describe('Object Storage Users Launch container form validation', () => {
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

      const inputLetters13 = '1234567890123' // 13 letters
      const inputLetters12 = '123456789012' // 12 letters

      userEvent.clear(nameInput)
      userEvent.type(nameInput, inputLetters13)
      await waitFor(() => {
        expect(screen.getByTestId('NameInput')).toHaveValue(inputLetters12)
      })

      userEvent.clear(nameInput)
      await waitFor(() => {
        expect(screen.getByTestId('NameInvalidMessage')).toBeInTheDocument()
        expect(screen.getByTestId('NameInvalidMessage')).toHaveTextContent('Name is required')
      })
    })

    it('Create button disabled', async () => {
      await act(async () => {
        render(<TestComponent history={history} />)
      })

      expect(await screen.findByTestId(btnTestId)).toBeVisible()
    })

    it('Validate if form is valid for submission - All buckets allowed actions', async () => {
      await act(async () => {
        render(<TestComponent history={history} />)
      })

      expect(await screen.findByTestId(btnTestId)).toBeVisible()

      const inputName = screen.getByTestId('NameInput')
      fireEvent.change(inputName, ('name', { target: { value: 'test' } }))
      expect(inputName).toHaveValue('test')

      const actionsSelectAll = await screen.findByTestId('Allowedactions-Input-option-selectAll')
      userEvent.click(actionsSelectAll)

      expect(await screen.findByTestId(btnTestId)).not.toHaveAttribute('disabled')
    })

    it('Validate if form is valid for submission - All buckets allowed policies', async () => {
      await act(async () => {
        render(<TestComponent history={history} />)
      })

      expect(await screen.findByTestId(btnTestId)).toBeVisible()

      const inputName = screen.getByTestId('NameInput')
      fireEvent.change(inputName, ('name', { target: { value: 'test' } }))
      expect(inputName).toHaveValue('test')

      const actionsSelectAll = await screen.findByTestId('Allowedpolicies-Input-option-selectAll')
      userEvent.click(actionsSelectAll)

      expect(await screen.findByTestId(btnTestId)).not.toHaveAttribute('disabled')

      userEvent.click(actionsSelectAll)
      expect(await screen.findByTestId(btnTestId)).toBeVisible()

      const pathName = screen.getByTestId('AllowedpoliciespathInput')
      fireEvent.change(pathName, ('name', { target: { value: 'test' } }))
      expect(pathName).toHaveValue('test')
      expect(await screen.findByTestId(btnTestId)).not.toHaveAttribute('disabled')
    })

    it('Validate if form is valid for submission - Per buckets allowed actions', async () => {
      await act(async () => {
        render(<TestComponent history={history} />)
      })

      expect(await screen.findByTestId(btnTestId)).toBeVisible()

      const inputName = screen.getByTestId('NameInput')
      fireEvent.change(inputName, ('name', { target: { value: 'test' } }))
      expect(inputName).toHaveValue('test')

      const radioPerBucket = await screen.findByTestId('Applypermissions-Radio-option-Perbucket')
      userEvent.click(radioPerBucket)
      expect(await screen.findByText('Buckets:')).toBeInTheDocument()
      expect(screen.queryByTestId('Allowedactions-Input-option-selectAll')).not.toBeInTheDocument()
      expect(screen.queryByTestId('Allowedpolicies-Input-option-selectAll')).not.toBeInTheDocument()
      expect(screen.queryByTestId('AllowedpoliciespathInput')).not.toBeInTheDocument()

      const allBuckets = screen.getAllByTestId(/^Buckets-Radio-option-/i)
      expect(allBuckets.length).toEqual(mockBaseObjectStorageStore().items.length)

      const selectBucket = await screen.findByTestId(
        `Buckets-Radio-option-${mockBaseObjectStorageStore().items[0].metadata.name}`
      )
      userEvent.click(selectBucket)

      expect(await screen.findByTestId('Allowedactions-Input-option-selectAll')).toBeInTheDocument()
      expect(await screen.findByTestId('Allowedpolicies-Input-option-selectAll')).toBeInTheDocument()
      expect(await screen.findByTestId('AllowedpoliciespathInput')).toBeInTheDocument()

      expect(await screen.findByTestId(btnTestId)).toBeVisible()

      const actionsSelectAll = await screen.findByTestId('Allowedactions-Input-option-selectAll')
      userEvent.click(actionsSelectAll)

      expect(await screen.findByTestId(btnTestId)).not.toHaveAttribute('disabled')
    })

    it('Validate if form is valid for submission - Per buckets allowed policies', async () => {
      await act(async () => {
        render(<TestComponent history={history} />)
      })

      expect(await screen.findByTestId(btnTestId)).toBeVisible()

      const inputName = screen.getByTestId('NameInput')
      fireEvent.change(inputName, ('name', { target: { value: 'test' } }))
      expect(inputName).toHaveValue('test')

      const radioPerBucket = await screen.findByTestId('Applypermissions-Radio-option-Perbucket')
      userEvent.click(radioPerBucket)
      expect(await screen.findByText('Buckets:')).toBeInTheDocument()
      expect(screen.queryByTestId('Allowedactions-Input-option-selectAll')).not.toBeInTheDocument()
      expect(screen.queryByTestId('Allowedpolicies-Input-option-selectAll')).not.toBeInTheDocument()
      expect(screen.queryByTestId('AllowedpoliciespathInput')).not.toBeInTheDocument()

      const allBuckets = screen.getAllByTestId(/^Buckets-Radio-option-/i)
      expect(allBuckets.length).toEqual(mockBaseObjectStorageStore().items.length)

      const selectBucket = await screen.findByTestId(
        `Buckets-Radio-option-${mockBaseObjectStorageStore().items[0].metadata.name}`
      )
      userEvent.click(selectBucket)

      expect(await screen.findByTestId('Allowedactions-Input-option-selectAll')).toBeInTheDocument()
      expect(await screen.findByTestId('Allowedpolicies-Input-option-selectAll')).toBeInTheDocument()
      expect(await screen.findByTestId('AllowedpoliciespathInput')).toBeInTheDocument()

      expect(await screen.findByTestId(btnTestId)).toBeVisible()

      const policiesSelectAll = await screen.findByTestId('Allowedpolicies-Input-option-selectAll')
      userEvent.click(policiesSelectAll)

      expect(await screen.findByTestId(btnTestId)).not.toHaveAttribute('disabled')

      userEvent.click(policiesSelectAll)
      expect(await screen.findByTestId(btnTestId)).toBeVisible()

      const pathName = screen.getByTestId('AllowedpoliciespathInput')
      fireEvent.change(pathName, ('name', { target: { value: 'test' } }))
      expect(pathName).toHaveValue('test')
      expect(await screen.findByTestId(btnTestId)).not.toHaveAttribute('disabled')
    })

    it('Validate if form is valid for submission - Change permissions type', async () => {
      await act(async () => {
        render(<TestComponent history={history} />)
      })

      expect(await screen.findByTestId(btnTestId)).toBeVisible()

      const inputName = screen.getByTestId('NameInput')
      fireEvent.change(inputName, ('name', { target: { value: 'test' } }))
      expect(inputName).toHaveValue('test')

      const actionCheckboxAll = await screen.findByTestId('Allowedactions-Input-option-selectAll')
      userEvent.click(actionCheckboxAll)

      expect(await screen.findByTestId(btnTestId)).not.toHaveAttribute('disabled')

      const radioPerBucket = await screen.findByTestId('Applypermissions-Radio-option-Perbucket')
      userEvent.click(radioPerBucket)
      expect(await screen.findByTestId(btnTestId)).toBeVisible()

      const selectBucket = await screen.findByTestId(
        `Buckets-Radio-option-${mockBaseObjectStorageStore().items[0].metadata.name}`
      )
      userEvent.click(selectBucket)

      const policiesSelectAll = await screen.findByTestId('Allowedpolicies-Input-option-selectAll')
      userEvent.click(policiesSelectAll)

      expect(await screen.findByTestId(btnTestId)).not.toHaveAttribute('disabled')

      const radioAll = await screen.findByTestId('Applypermissions-Radio-option-Forallbuckets')
      userEvent.click(radioAll)
      expect(await screen.findByTestId(btnTestId)).toBeVisible()
    })
  })
})
