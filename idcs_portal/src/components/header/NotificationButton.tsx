// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import { Nav } from 'react-bootstrap'
import { NavLink } from 'react-router-dom'
import { BsBell } from 'react-icons/bs'
import useNotificationsStore from '../../store/notificationStore/NotificationsStore'

const NotificationButton: React.FC = (): JSX.Element => {
  const { notifications } = useNotificationsStore()
  return (
    <Nav.Item
      intc-id="notificationsHeaderButton"
      className="btn btn-icon-simple"
      as={NavLink}
      aria-label="Notification Button"
      to="/notifications"
    >
      <BsBell intc-id="notificationIcon" />
      {notifications.length > 0 ? (
        <span className="border rounded-circle notificationsCount"> {notifications.length} </span>
      ) : null}
    </Nav.Item>
  )
}

export default NotificationButton
