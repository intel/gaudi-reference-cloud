import userEvent from '@testing-library/user-event'
import { act } from 'react'
import { fireEvent, waitFor } from '@testing-library/react'
import { mockAxios } from '../../setupTests'

export const clearAxiosMock = () => {
  mockAxios.get.mockReset()
  mockAxios.post.mockReset()
  mockAxios.put.mockReset()
}

export const typeOnInputElement = async(element, value) => {
  await act(async () => {
    await userEvent.type(element, value)
  })
  await waitFor(() => {
    expect(element).toHaveValue(value)
  })
}

export const setValueOnDateTimeInput = async(element, dateTimeISO) => {
  await act(() => {
    fireEvent.change(element, { target: { value: dateTimeISO } })
  })
  await waitFor(() => {
    expect(element).not.toHaveValue('')
  })
}

export const clickButton = async(element) => {
  await waitFor(() => {
    expect(element).toBeEnabled()
  })
  await act(async () => {
    await userEvent.click(element)
  })
}

export const selectRadioButton = async(element) => {
  await act(async () => {
    await userEvent.click(element)
  })
  await waitFor(() => {
    expect(element).toBeChecked()
  })
}
