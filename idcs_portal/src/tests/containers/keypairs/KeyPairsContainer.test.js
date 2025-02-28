// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { render, screen, waitFor } from '@testing-library/react'
import { act } from 'react'
import { Router } from 'react-router-dom'
import { createMemoryHistory } from 'history'
import userEvent from '@testing-library/user-event'

import { mockStandardUser } from '../../mocks/authentication/authHelper'
import { clearAxiosMock } from '../../mocks/utils'
import {
  mockPublicKeys,
  mockEmptyPublicKeys,
  mockBasePublicKeysStore,
  mockDeletePublicKeys
} from '../../mocks/publicKeys/publicKeys'

import KeyPairsContainer from '../../../containers/keypairs/KeyPairsContainer'
import AuthWrapper from '../../../utils/authWrapper/AuthWrapper'
import ToastContainer from '../../../utils/toast/ToastContainer'

const TestComponent = ({ history }) => {
  return (
    <AuthWrapper>
      <Router location={history.location} navigator={history}>
        <ToastContainer />
        <KeyPairsContainer />
      </Router>
    </AuthWrapper>
  )
}

describe('SSH Public Keys container - My Keys', () => {
  let history = null

  beforeEach(() => {
    clearAxiosMock()
    history = createMemoryHistory()
  })

  beforeEach(() => {
    mockStandardUser()
    mockPublicKeys()
  })

  it('Render component if loads correctly', async () => {
    await act(async () => {
      render(<TestComponent history={history} />)
    })
    const uploadKeyButton = await screen.findByTestId('btn-sshview-UploadKey')
    expect(uploadKeyButton).toBeInTheDocument()
    expect(uploadKeyButton).toBeEnabled()
    expect(uploadKeyButton).toHaveAttribute('href', '/security/publickeys/import')
  })

  it('Display no key message when no data availible', async () => {
    mockEmptyPublicKeys()
    await act(async () => {
      render(<TestComponent history={history} />)
    })

    expect(await screen.findByText('No keys found')).toBeInTheDocument()
    expect(await screen.findByText('Your account currently has no keys.')).toBeInTheDocument()
    const uploadKeyOnNoKeysAvailableButton = screen.getByTestId('UploadkeyEmptyViewButton')
    expect(uploadKeyOnNoKeysAvailableButton).toBeInTheDocument()
    expect(uploadKeyOnNoKeysAvailableButton).toBeEnabled()
    expect(uploadKeyOnNoKeysAvailableButton).toHaveAttribute('href', '/security/publickeys/import')
  })

  it('Render table when data availible', async () => {
    await act(async () => {
      render(<TestComponent history={history} />)
    })

    // Get table
    const sshPublicKeysTable = screen.getByTestId('sshPublicKeysTable')
    const sshPublicKeyTable = sshPublicKeysTable.querySelector('.table')

    // Get table elements
    const tableTheadTR = sshPublicKeyTable.getElementsByTagName('thead')[0].getElementsByTagName('tr')
    const tableTbodyTR = sshPublicKeyTable.getElementsByTagName('tbody')[0].getElementsByTagName('tr')

    // Check if data loads properly
    expect(tableTheadTR.length).toEqual(1)
    expect(tableTbodyTR.length).toEqual(mockBasePublicKeysStore().items.length)

    // Check columns count
    const tableTheadTRTh = tableTheadTR[0].getElementsByTagName('th')
    expect(tableTheadTRTh.length).toEqual(3)

    // Check first two columns have sorting options and last column doesn't have
    expect(tableTheadTRTh[0].getElementsByTagName('button').length).toEqual(1)
    expect(tableTheadTRTh[1].getElementsByTagName('button').length).toEqual(1)

    // Check tbody shold have a button delete
    const tableTbodyTRTd = tableTbodyTR[0].getElementsByTagName('td')
    const deleteBtnLink = tableTbodyTRTd[1].getElementsByTagName('a')
    expect(deleteBtnLink.length).toEqual(0)
  })

  it('Name column sort should work as expected', async () => {
    await act(async () => {
      render(<TestComponent history={history} />)
    })

    // Get table
    const sshPublicKeysTable = screen.getByTestId('sshPublicKeysTable')
    const sshPublicKeyTable = sshPublicKeysTable.querySelector('.table')

    // Get table elements
    const tableTheadTR = sshPublicKeyTable.getElementsByTagName('thead')[0].getElementsByTagName('tr')
    const tableTbodyTR = sshPublicKeyTable.getElementsByTagName('tbody')[0].getElementsByTagName('tr')
    const tableTbodyTRLength = tableTbodyTR.length

    // Get name from frist row
    const firstNameBeforeSort = tableTbodyTR[0].getElementsByTagName('td')[0].getElementsByTagName('span')[0].innerHTML
    const lastNameBeforeSort = tableTbodyTR[tableTbodyTRLength - 1]
      .getElementsByTagName('td')[0]
      .getElementsByTagName('span')[0].innerHTML

    // Get name Column sort element
    const nameColumnSort = tableTheadTR[0].getElementsByTagName('th')[0].getElementsByTagName('button')[0]
    await userEvent.click(nameColumnSort)

    await waitFor(() => {
      const firstNameAfterSort = tableTbodyTR[0].getElementsByTagName('td')[0].getElementsByTagName('span')[0].innerHTML
      const lastNameAfterSort = tableTbodyTR[tableTbodyTRLength - 1]
        .getElementsByTagName('td')[0]
        .getElementsByTagName('span')[0].innerHTML
      expect(firstNameBeforeSort).toEqual(lastNameAfterSort)
      expect(lastNameBeforeSort).toEqual(firstNameAfterSort)
    })
  })

  it('Clicking on a delete button should opens the modal', async () => {
    await act(async () => {
      render(<TestComponent history={history} />)
    })

    const mockBasePublicKeysStoreItems = mockBasePublicKeysStore().items
    const mockBasePublicKeysStoreItemsLength = mockBasePublicKeysStoreItems.length

    const deleteButtons = screen.getAllByTestId('ButtonTable Delete key')
    expect(deleteButtons.length).toEqual(mockBasePublicKeysStoreItemsLength)

    await userEvent.click(deleteButtons[0])
    const deleteModal = await screen.findByTestId('deleteConfirmModal')
    expect(deleteModal).toBeInTheDocument()

    const modalMessage = 'To confirm deletion enter the name of the key below.'
    expect(await screen.findByText(modalMessage)).toBeInTheDocument()

    // Delete modal should have two buttons to close the modal
    const closeButton1 = screen.getByTestId('btn-confirm-Deletekey-cancel')
    expect(closeButton1).toBeInTheDocument()

    const closeButton2 = deleteModal.getElementsByClassName('btn-close')
    expect(closeButton2.length).toEqual(1)
  })

  it('Clicking on first close button should close the modal', async () => {
    await act(async () => {
      render(<TestComponent history={history} />)
    })

    const deleteButtons = screen.getAllByTestId('ButtonTable Delete key')
    await userEvent.click(deleteButtons[0])
    const closeButton = screen.getByTestId('btn-confirm-Deletekey-cancel')
    await userEvent.click(closeButton)
    await waitFor(() => {
      const deleteModal = screen.findByTestId('deleteConfirmModal')
      expect(deleteModal).toMatchObject({})
    })
  })

  it('Clicking on second close button should close the modal', async () => {
    await act(async () => {
      render(<TestComponent history={history} />)
    })

    const deleteButtons = screen.getAllByTestId('ButtonTable Delete key')
    await userEvent.click(deleteButtons[0])
    const deleteModal = await screen.findByTestId('deleteConfirmModal')
    const closeButton = deleteModal.getElementsByClassName('btn-close')
    await userEvent.click(closeButton[0])
    await waitFor(() => {
      const deleteModal1 = screen.findByTestId('deleteConfirmModal')
      expect(deleteModal1).toMatchObject({})
    })
  })

  it('Clicking on delete button should delete the data', async () => {
    mockDeletePublicKeys()
    await act(async () => {
      render(<TestComponent history={history} />)
    })

    const deleteButtons = screen.getAllByTestId('ButtonTable Delete key')
    await userEvent.click(deleteButtons[0])
    const deleteButton = screen.getByTestId('btn-confirm-Deletekey-delete')

    await waitFor(() => {
      const deleteModal = screen.findByTestId('deleteConfirmModal')
      expect(deleteModal).toMatchObject({})
    })

    const mockBasePublicKeysStoreItems = mockBasePublicKeysStore().items
    const mockBasePublicKeysStoreItemsLength = mockBasePublicKeysStore().items.length
    const recordName = mockBasePublicKeysStoreItems[mockBasePublicKeysStoreItemsLength - 1].metadata.name
    const nameInput = screen.getByTestId('NameInput')
    await userEvent.clear(nameInput)
    await userEvent.type(nameInput, recordName)

    await userEvent.click(deleteButton)
    const alertMessage = `The key(s) '${recordName}' was deleted successfully.`
    expect(await screen.findByText(alertMessage)).toBeInTheDocument()

    await waitFor(() => {
      expect(screen.queryByText(alertMessage)).toBeNull()
      expect(screen.queryByTestId('deleteConfirmModal')).toBeNull()
    })
  })
})
