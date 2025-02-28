// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { mockAxios } from '../../setupTests'

export const clearAxiosMock = () => {
  mockAxios.get.mockReset()
  mockAxios.post.mockReset()
  mockAxios.put.mockReset()
}

export const expectValueForInputElement = async (element, inputValue, expectValue) => {
  await clearInputElement(element)
  userEvent.type(element, inputValue)
  await waitFor(() => {
    expect(element).toHaveValue(expectValue)
  })
}

export const expectClassForInputElement = async (element, inputValue, expectValue) => {
  await clearInputElement(element)
  userEvent.type(element, inputValue)
  await waitFor(() => {
    expect(element).toHaveClass(expectValue)
  })
}

export const expectTextContentForInputElement = async (element, expectValue) => {
  await waitFor(() => {
    expect(element).toHaveTextContent(expectValue)
  })
}

export const clearInputElement = async (element) => {
  userEvent.clear(element)
}
