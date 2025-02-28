// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { render, screen } from '@testing-library/react'
import AuthWrapper from '../../../utils/authWrapper/AuthWrapper'
import { mockIntelUser } from '../../mocks/authentication/authHelper'
import { act } from 'react'
import idcConfig from '../../../config/configurator'
import { clearAxiosMock } from '../../mocks/utils'
import { createMemoryHistory } from 'history'
import { Router, Routes, Route } from 'react-router-dom'
import { mockVendorsApi } from '../../mocks/shared/vendors'
import userEvent from '@testing-library/user-event'
import {
  mockBaseObjectStorageStore,
  mockObjectStorage,
  mockObjectStorageData
} from '../../mocks/objectStorage/ObjectStorage'
import ObjectStorageRuleEditContainer from '../../../containers/objectStorage/ObjectStorageRuleEditContainer'

const TestComponent = ({ history, name }) => {
  history.push(`/buckets/d/${name}/lifecyclerule/e/rule1`)
  return (
    <AuthWrapper>
      <Router location={history.location} navigator={history}>
        <Routes>
          <Route
            path="/buckets/d/:param/lifecyclerule/e/:param2"
            element={<ObjectStorageRuleEditContainer history={history} name={name} />}
          />
        </Routes>
      </Router>
    </AuthWrapper>
  )
}

describe('Object Storage Lifecycle Rule Launch container unit test cases', () => {
  let history = null
  const selectedRule = mockObjectStorageData()[1].lifecycleRulePolicies[0]
  const name = mockBaseObjectStorageStore().items[1].metadata.name

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

    expect(await screen.findByText('Edit Lifecycle Rule - ' + selectedRule.ruleName)).toBeInTheDocument()
  })

  it('Show navigation bottom buttons', async () => {
    await act(async () => {
      render(<TestComponent history={history} name={name} />)
    })

    expect(await screen.findByTestId('btn-ObjectStorageRuleEdit-navigationBottom Save')).toBeInTheDocument()

    expect(await screen.findByTestId('btn-ObjectStorageRuleEdit-navigationBottom Cancel')).toBeInTheDocument()
  })

  it('Show form elements with values', async () => {
    await act(async () => {
      render(<TestComponent history={history} name={name} />)
    })

    expect(await screen.findByTestId('PrefixInput')).toBeInTheDocument()
    expect(await screen.findByTestId('PrefixInput')).toHaveValue(selectedRule.prefix)
    expect(await screen.findByTestId('Input')).toBeInTheDocument()
    expect(await screen.findByTestId('Input')).toHaveValue(selectedRule.expireDays.toString())
    expect(await screen.findByTestId('Input')).not.toHaveAttribute('disabled')
    expect(await screen.findByTestId('NoncurrentexpirydaysInput')).toBeInTheDocument()
    expect(await screen.findByTestId('NoncurrentexpirydaysInput')).toHaveValue(
      selectedRule.noncurrentExpireDays.toString()
    )
    expect(await screen.findByTestId('-Radio-option-DeleteMarker')).toBeInTheDocument()
    expect(await screen.findByTestId('-Radio-option-DeleteMarker')).not.toBeChecked()
    expect(await screen.findByTestId('-Radio-option-ExpiryDays')).toBeInTheDocument()
    expect(await screen.findByTestId('-Radio-option-ExpiryDays')).toBeChecked()
  })

  describe('Object Storage Rule Launch container form validation', () => {
    it('Disable/ Enable expiry days based on delete marker', async () => {
      await act(async () => {
        render(<TestComponent history={history} name={name} />)
      })
      const expiryDays = screen.getByTestId('Input')
      const deleteMarkerRadio = screen.getByTestId('-Radio-option-DeleteMarker')
      const expiryDaysRadio = screen.getByTestId('-Radio-option-ExpiryDays')

      expect(expiryDays).not.toHaveAttribute('disabled')
      userEvent.click(deleteMarkerRadio)
      expect(deleteMarkerRadio).toBeChecked()
      expect(expiryDays).toHaveAttribute('disabled')

      userEvent.click(expiryDaysRadio)
      expect(expiryDaysRadio).toBeChecked()
      expect(expiryDays).not.toHaveAttribute('disabled')
    })

    it('Edit button enabled', async () => {
      await act(async () => {
        render(<TestComponent history={history} name={name} />)
      })

      expect(await screen.findByTestId('btn-ObjectStorageRuleEdit-navigationBottom Save')).not.toHaveAttribute(
        'disabled'
      )
    })
  })
})
