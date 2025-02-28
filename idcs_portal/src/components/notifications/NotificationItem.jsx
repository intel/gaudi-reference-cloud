// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { BsCreditCard, BsExclamationCircle, BsCpu, BsBell } from 'react-icons/bs'
import moment from 'moment'

const NotificationItem = (props) => {
  const notification = props.notification
  const popup = props.popup
  const header = props.header
  const dateFormat = 'MM/DD/YYYY h:mm a'

  function getIcon(notificationType) {
    if (notificationType === 'billing') {
      return (
        <BsCreditCard
          intc-id="billingNotificationIcon"
          className="m-auto text-center p-2"
          style={{ fontSize: '50px', background: '#653171', color: 'white' }}
        />
      )
    }
    if (notificationType === 'compute') {
      return (
        <BsCpu
          intc-id="computeNotificationIcon"
          className="m-auto text-center p-2"
          style={{ fontSize: '50px', background: '#00377C', color: 'white' }}
        />
      )
    }
    if (notificationType === 'warning-alert') {
      return (
        <BsExclamationCircle
          intc-id="warningAlertNotificationIcon"
          className="m-auto text-center p-2"
          style={{ fontSize: '50px', background: '#B13D4A', color: 'white' }}
        />
      )
    }
    return (
      <BsBell
        intc-id="defaultNotificationIcon"
        className="m-auto text-center p-2"
        style={{ fontSize: '50px', background: '#0089B9', color: 'white' }}
      />
    )
  }

  return (
    <div intc-id="notificationItem" className="flex-row d-flex bd-highlight p-2">
      <div className="mx-3 d-flex align-items-center">{getIcon(notification?.serviceName)}</div>
      <div>
        <div>
          <strong className="text-capitalize"> {notification?.serviceName} </strong>
          {popup ? <br /> : null}
          {header ? <br /> : <span> {notification?.message} </span>}
        </div>
        <span> {moment(new Date(notification?.creation)).format(dateFormat)} </span>
      </div>
    </div>
  )
}

export default NotificationItem
