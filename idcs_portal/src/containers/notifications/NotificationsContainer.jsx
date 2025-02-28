// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import NotificationsList from '../../components/notifications/NotificationsList'
import useNotificationsStore from '../../store/notificationStore/NotificationsStore'

const NotificationsContainer = (props) => {
  const isWidget = props.isWidget
  const { loading, notifications } = useNotificationsStore()

  const emptyNotification = {
    title: 'No notifications yet',
    subTitle: (
      <span>
        Stay tuned for exciting updates!
        <br /> No new notifications at the moment.
      </span>
    )
  }

  return (
    <NotificationsList
      isWidget={isWidget}
      loading={loading}
      notifications={notifications}
      emptyNotification={emptyNotification}
    />
  )
}

export default NotificationsContainer
