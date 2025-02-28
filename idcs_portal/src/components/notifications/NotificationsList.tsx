// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import NotificationItem from './NotificationItem'
import { Link } from 'react-router-dom'
import EmptyView from '../../utils/emptyView/EmptyView'

const NotificationsList = (props: any): JSX.Element => {
  let notifications = props.notifications
  const emptyNotification = props.emptyNotification
  const isWidget = props.isWidget

  if (isWidget) {
    notifications = notifications.slice(0).slice(-4)
  }

  const emptyView = (): JSX.Element => (
    <EmptyView title={emptyNotification.title} subTitle={emptyNotification.subTitle} />
  )

  const notificationList = (): any[] => {
    return notifications.map((notification: any, index: number) => (
      <div
        className={`d-flex flex-row justify-content section-component gap-s6 ${isWidget ? 'shadow-sm' : 'shadow'}`}
        key={index}
      >
        <NotificationItem isWidget={isWidget} notification={notification} />
      </div>
    ))
  }

  const endOfWidget = (): JSX.Element | string => {
    let html: any = ''
    if (isWidget && notifications.length > 0) {
      html = (
        <div className="d-flex flex-column justify-content-center">
          <Link intc-id="goToNotificationsButton" to="/notifications" className="text-center">
            <span className="btn btn-link">Go to notifications</span>
          </Link>
        </div>
      )
    }
    return html
  }

  const getNotifications = (): JSX.Element | any[] => {
    return notifications.length === 0 ? emptyView() : notificationList()
  }

  return (
    <>
      {isWidget && (
        <div className="section">
          <h2>Notifications</h2>
        </div>
      )}
      <div className="section">
        {isWidget ? (
          <>
            {getNotifications()}
            {endOfWidget()}
          </>
        ) : (
          getNotifications()
        )}
      </div>
    </>
  )
}

export default NotificationsList
