// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { render, screen, waitFor, within } from '@testing-library/react'
import { BrowserRouter } from 'react-router-dom'
import userEvent from '@testing-library/user-event'
import { clearAxiosMock } from '../../mocks/utils'
import { mockStandardUser } from '../../mocks/authentication/authHelper'
import Header from '../../../components/header/Header'
import idcConfig, { updateCurrentRegion } from '../../../config/configurator'
import useUserStore from '../../../store/userStore/UserStore'

const TestComponent = () => {
  const user = useUserStore((state) => state.user)
  const mockUser = { ...user }

  return (
    <BrowserRouter>
      <Header userDetails={mockUser} pathname={'/'} />
    </BrowserRouter>
  )
}

describe('Region Menu', () => {
  const originalWindowLocation = window.location

  beforeAll(() => {
    sessionStorage.clear()
    Object.defineProperty(window, 'location', {
      configurable: true,
      enumerable: true,
      value: {
        ...originalWindowLocation,
        search: '?region=us-dev-1',
        // Mock as if we reload config because of reload beign called
        reload: () => {
          updateCurrentRegion(idcConfig)
        }
      }
    })
  })

  beforeEach(() => {
    clearAxiosMock()
    Object.defineProperty(window, 'matchMedia', {
      writable: true,
      value: jest.fn().mockImplementation((query) => ({
        matches: false,
        media: query,
        onchange: null,
        addListener: jest.fn(), // deprecated
        removeListener: jest.fn(), // deprecated
        addEventListener: jest.fn(),
        removeEventListener: jest.fn(),
        dispatchEvent: jest.fn()
      }))
    })
  })

  beforeEach(() => {
    mockStandardUser()
  })

  afterAll(() => {
    sessionStorage.clear()
    Object.defineProperty(window, 'location', {
      configurable: true,
      enumerable: true,
      value: originalWindowLocation
    })
  })

  it('Selected region should have check icon and bold class', async () => {
    render(<TestComponent />)
    const regionLabel = await screen.findByTestId('regionLabel')
    const regionMenu = screen.getByTestId('regionMenu')
    await userEvent.click(regionMenu)
    const selectedRegionOption = await screen.findByTestId(`region-option-${regionLabel.textContent.trim()}`)
    const selectedIcon = within(selectedRegionOption).getByTestId(`selectedIcon-${regionLabel.textContent.trim()}`)
    expect(selectedIcon).not.toHaveClass('invisible')
    await userEvent.keyboard('{Escape}')
  })

  it('Default region should be selected by default if no region is set on session', async () => {
    render(<TestComponent />)
    const defaultRegion = 'us-dev-1'
    expect(idcConfig.REACT_APP_SELECTED_REGION).toBe(defaultRegion)
    const regionLabel = await screen.findByTestId('regionLabel')
    expect(regionLabel).toHaveTextContent(defaultRegion)

    const defaultRegionApi = 'https://internal-placeholder.com/v1'
    expect(idcConfig.REACT_APP_API_REGIONAL_SERVICE).toBe(defaultRegionApi)
    const defaultRegionName = 'us-dev-1a-default'
    expect(idcConfig.REACT_APP_DEFAULT_REGION_NAME).toBe(defaultRegionName)
    const defaultRegionAvailabilityZone = 'us-dev-1a'
    expect(idcConfig.REACT_APP_DEFAULT_REGION_AVAILABILITY_ZONE).toBe(defaultRegionAvailabilityZone)
    const defaultRegionPrefix = '22'
    expect(idcConfig.REACT_APP_DEFAULT_REGION_PREFIX).toBe(defaultRegionPrefix)
  })

  it('Change region should set region, reload page and update region on configuration', async () => {
    render(<TestComponent />)
    const testRegion = 'us-dev-2'
    Object.defineProperty(window.location, 'search', {
      writable: true,
      value: '?region=' + testRegion
    })

    const regionMenu = screen.getByTestId('regionMenu')
    await userEvent.click(regionMenu)
    const testRegionOption = await screen.findByTestId('region-option-us-dev-2')
    await userEvent.click(testRegionOption)

    await waitFor(() => {
      expect(idcConfig.REACT_APP_SELECTED_REGION).toBe(testRegion)
    })
  })

  it('After reload page, change configuration variables and update region label', async () => {
    render(<TestComponent />)
    const testRegion = 'us-dev-2'
    expect(idcConfig.REACT_APP_SELECTED_REGION).toBe(testRegion)
    const regionLabel = await screen.findByTestId('regionLabel')
    expect(regionLabel).toHaveTextContent(testRegion)

    const testRegionApi = 'https://internal-placeholder.com/v1'
    expect(idcConfig.REACT_APP_API_REGIONAL_SERVICE).toBe(testRegionApi)
    const testRegionName = 'us-dev-2a-default'
    expect(idcConfig.REACT_APP_DEFAULT_REGION_NAME).toBe(testRegionName)
    const testRegionAvailabilityZone = 'us-dev-2a'
    expect(idcConfig.REACT_APP_DEFAULT_REGION_AVAILABILITY_ZONE).toBe(testRegionAvailabilityZone)
    const testRegionPrefix = '24'
    expect(idcConfig.REACT_APP_DEFAULT_REGION_PREFIX).toBe(testRegionPrefix)
  })
})
