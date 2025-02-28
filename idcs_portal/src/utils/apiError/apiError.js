// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import useUserStore from '../../store/userStore/UserStore'
import { specificErrorMessageEnum } from '../Enums'
export const isErrorInsufficientCapacity = (errorMessage) => {
  return errorMessage.indexOf(specificErrorMessageEnum.capacityErrorMessage) !== -1
}

export const isErrorInAuthorization = (error) => {
  return (
    error?.response?.status === 403 &&
    error?.response?.data?.code === 7 &&
    specificErrorMessageEnum.authErrorMessage.some((errMessage) => error?.response?.data?.message.includes(errMessage))
  )
}

export const isErrorInsufficientCredits = (error) => {
  return (
    error?.response?.data?.code === 7 &&
    error?.response?.data?.message?.toLowerCase().indexOf(specificErrorMessageEnum.noCreditsErrorMessage) !== -1
  )
}

export const friendlyErrorMessages = {
  insufficientCapacity: 'We are currently experiencing high demand for this instance type. Please try again later.',
  unathorizedAction: 'It seems your role has not been authorized to access this item.',
  memeberExisted: 'Member already exists.',
  maxMemberLimitReached: "You've reached the maximum number of invitations."
}

export const getErrorMessageFromCodeAndMessage = (errorCode, errorMessage) => {
  if (errorCode === 7) {
    // NoCredits
    if (useUserStore.getState().isPremiumUser) {
      return "We're sorry, but we couldn't find a valid payment method on your account. Please add a valid payment method to continue with your transaction."
    }
    return "We're sorry, but you don't have an outstanding credit balance to complete this transaction."
  } else if (isErrorInsufficientCapacity(errorMessage)) {
    return friendlyErrorMessages.insufficientCapacity
  }
  return errorMessage
}
