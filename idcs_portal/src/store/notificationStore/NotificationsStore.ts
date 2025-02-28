// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { useQuery } from 'react-query'
import NotificationsService from '../../services/NotificationsService'
import idcConfig, { appFeatureFlags, isFeatureFlagEnable } from '../../config/configurator'

export interface Notification {
  alertType: string
  clientRecordId: string
  cloudAccountId: string
  creation: string
  expiration: string
  id: string
  message: string
  properties: any
  serviceName: string
  severity: string
  status: string
  userId: string
}

interface NotificationsStore {
  loading: boolean
  notifications: Notification[] | []
}

const useNotificationsStore = (): NotificationsStore => {
  const { REACT_APP_NOTIFICATIONS_HEARTBEAT } = idcConfig as any
  const { isLoading, data } = isFeatureFlagEnable(appFeatureFlags.REACT_APP_FEATURE_NOTIFICATIONS)
    ? useQuery(['notifications'],
      NotificationsService.getNotifications,
      { refetchInterval: REACT_APP_NOTIFICATIONS_HEARTBEAT }
    )
    : {
        isLoading: false,
        data: {
          data: []
        }
      }
  let notifications: Notification[] = data?.data.alerts || []
  if (data?.data?.notifications != null) {
    notifications = data?.data.alerts.concat(data?.data?.notifications) || []
  }
  const loading = isLoading

  return {
    loading,
    notifications
  }
}

export default useNotificationsStore
