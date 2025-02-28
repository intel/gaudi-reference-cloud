// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import Button from 'react-bootstrap/Button'
import { Dropdown, DropdownButton, ButtonGroup } from 'react-bootstrap'
import EmptyView from '../../../utils/emptyView/EmptyView'
import ActionConfirmation from '../../../utils/modals/actionConfirmation/ActionConfirmation'
import { NavLink } from 'react-router-dom'
import TabsNavigation from '../../../utils/tabsNavigation/TabsNavagation'

const SuperComputerDetail = ({
  noFoundItem,
  navigationTop,
  clusterDetail,
  setAction,
  tabs,
  setActiveTab,
  activeTab,
  actionModal,
  onActionModal
}) => {
  // *****
  // functions
  // *****
  function getNavigationItem(item, key) {
    switch (item.type) {
      case 'link':
        return (
          <Button
            key={key}
            intc-id={`btn-super-computer-navigationTop ${item.label}`}
            variant="outline-primary"
            onClick={item.buttonFunction}
          >
            {item.label}
          </Button>
        )
      case 'title':
        return <h2 key={key}>{item.label}</h2>
      case 'dropdown':
        return (
          <DropdownButton
            key={key}
            as={ButtonGroup}
            variant="outline-primary"
            title="Actions"
            intc-id="SuperComputerActionsDropdownButton"
          >
            {item.actions.map((action, index) => {
              return action?.isTextItem ? (
                <div key={index}>
                  <Dropdown.Divider />
                  <Dropdown.ItemText className="fw-semibold" intc-id={`myReservationActionsDropdownItemText${index}`}>
                    {action.name}
                  </Dropdown.ItemText>
                </div>
              ) : (
                <Dropdown.Item
                  key={index}
                  onClick={() => setAction(action, clusterDetail)}
                  intc-id={`myReservationActionsDropdownItemButton${index}`}
                >
                  {action.name}
                </Dropdown.Item>
              )
            })}
          </DropdownButton>
        )
      case 'documentation':
        return (
          <NavLink key={key} to={item.redirecTo} className="link" target="_blank" rel="noopener noreferrer">
            {item.label}
          </NavLink>
        )
      default:
        return <h2 key={key}>{item.label}</h2>
    }
  }

  let content = null

  if (clusterDetail) {
    content = (
      <>
        <div className="section flex-row">
          {navigationTop.map((element, index) => {
            return getNavigationItem(element, index)
          })}
        </div>
        <div className="section">
          <TabsNavigation tabs={tabs} activeTab={activeTab} setTabActive={setActiveTab} />
          {tabs[activeTab].content}
        </div>
      </>
    )
  } else {
    content = (
      <div className="section">
        <EmptyView title={noFoundItem.title} subTitle={noFoundItem.subTitle} action={noFoundItem.action} />
      </div>
    )
  }
  // }
  return (
    <>
      <ActionConfirmation
        actionModalContent={actionModal}
        onClickModalConfirmation={onActionModal}
        showModalActionConfirmation={actionModal.show}
      />
      {content}
    </>
  )
}

export default SuperComputerDetail
