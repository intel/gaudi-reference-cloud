// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import NotificationsService from '../../../services/NotificationsService'

export const mockBaseAlert = {
  serviceName: 'billing',
  message: 'test',
  creation: '07/18/2023 9:54 am'
}

export const mockBaseNotificationsNullStore = {
  lastUpdated: '2023-06-13T16:56:55.720541785Z',
  notifications: [],
  alerts: []
}

export const mockBaseNotificationsStore = {
  notifications: [],
  alerts: [mockBaseAlert, mockBaseAlert, mockBaseAlert, mockBaseAlert, mockBaseAlert]
}

export const mockNotifications = () => {
  jest.spyOn(NotificationsService, 'getNotifications').mockReturnValue({
    data: mockBaseNotificationsStore
  })
}

export const mockNotificationsNull = () => {
  jest.spyOn(NotificationsService, 'getNotifications').mockReturnValue({
    data: mockBaseNotificationsNullStore
  })
}
