// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import { Nav, Dropdown, Button } from 'react-bootstrap'
import { BsBoxArrowRight, BsFillArrowUpSquareFill, BsPerson, BsCopy } from 'react-icons/bs'
import { Link } from 'react-router-dom'
import Wrapper from '../../utils/Wrapper'
import { appFeatureFlags, isFeatureFlagEnable } from '../../config/configurator'
import UserBadge from '../../utils/userBadge/UserBadge'
import DarkModeContainerSwitch from '../../utils/darkMode/DarkModeContainerSwitch'
import { checkRoles } from '../../utils/accessControlWrapper/AccessControlWrapper'
import { AppRolesEnum } from '../../utils/Enums'
import { type User } from '../../store/userStore/UserStore'

interface UserMenuProps {
  userDetails: User | null
  logoutHandler: () => void
  isOwnCloudAccount: boolean
  copyToClipboard: (text: any) => void
}

const UserMenu: React.FC<UserMenuProps> = ({
  userDetails,
  logoutHandler,
  isOwnCloudAccount,
  copyToClipboard
}): JSX.Element => {
  const displayName = userDetails?.displayName
  const email = userDetails?.email

  return (
    <Dropdown intc-id="userMenu" className="align-self-center" as={Nav.Item}>
      <Dropdown.Toggle
        id="dropdown-header-user-toggle"
        role="combobox"
        className="d-flex align-items-center"
        variant="icon-simple"
        as={Button}
        data-bs-toggle="dropdown"
        aria-controls="dropdown-header-user-menu"
        aria-expanded="false"
        aria-label="User Menu"
      >
        <BsPerson intc-id="userIcon" />
      </Dropdown.Toggle>
      <Dropdown.Menu
        id="dropdown-header-user-menu"
        renderOnMount
        align="end"
        aria-labelledby="dropdown-header-user-toggle"
      >
        <div className="px-3 py-2 defaultBackground">
          <div>
            <span className="h6 mb-0">{displayName}</span>
          </div>
          <div>
            <small>{email}</small>
          </div>
        </div>
        <Dropdown.Divider className="mt-0" />
        <span className="list-unstyled mb-0">
          {checkRoles([
            AppRolesEnum.Standard,
            AppRolesEnum.Premium,
            AppRolesEnum.Enterprise,
            AppRolesEnum.Intel,
            AppRolesEnum.EnterprisePending
          ]) ? (
            <Dropdown.Item as={Link} intc-id="accountSettingsHeaderButton" to="/profile/accountsettings">
              <div className="d-flex flex-row justify-content-between align-items-baseline">
                <span>Account Settings</span>
                <div className="d-flex flex-column ms-s6">
                  <div className="d-flex flex-row align-items-center gap-s4">
                    <small>ID: {userDetails?.cloudAccountNumber} </small>
                    <Button
                      onClick={(e) => {
                        e.preventDefault()
                        e.stopPropagation()
                        copyToClipboard(userDetails?.cloudAccountNumber)
                      }}
                      variant="icon-simple"
                      aria-label="Copy cloud account id user menu"
                      intc-id="btnCopyIdUserMenu"
                      data-wap_ref="btnCopyIdUserMenu"
                    >
                      <BsCopy />
                    </Button>
                  </div>
                  <UserBadge />
                </div>
              </div>
            </Dropdown.Item>
          ) : null}

          {userDetails?.hasInvitations ? (
            <Dropdown.Item as={Link} intc-id="switchAccountsHeaderButton" to="/Accounts">
              Switch Accounts
            </Dropdown.Item>
          ) : null}

          {isOwnCloudAccount && checkRoles([AppRolesEnum.Premium]) ? (
            <Dropdown.Item as={Link} intc-id="invoicesHeaderButton" to="/billing/invoices">
              Invoices
            </Dropdown.Item>
          ) : null}

          {isOwnCloudAccount &&
          checkRoles([
            AppRolesEnum.Standard,
            AppRolesEnum.Premium,
            AppRolesEnum.Enterprise,
            AppRolesEnum.Intel,
            AppRolesEnum.EnterprisePending
          ]) ? (
            <Dropdown.Item as={Link} intc-id="currentMonthUsageHeaderButton" to="/billing/usages">
              Current month usage
            </Dropdown.Item>
          ) : null}

          {isOwnCloudAccount &&
          isFeatureFlagEnable(appFeatureFlags.REACT_APP_FEATURE_UPGRADE_TO_PREMIUM) &&
          checkRoles([AppRolesEnum.Standard]) ? (
            <Dropdown.Item as={Link} intc-id="upgradeAccountHeaderButton" to="/upgradeaccount">
              <Wrapper>
                <BsFillArrowUpSquareFill className="me-1 text-cobalt"></BsFillArrowUpSquareFill> Upgrade Account
              </Wrapper>
            </Dropdown.Item>
          ) : null}

          {isOwnCloudAccount && checkRoles([AppRolesEnum.Premium]) ? (
            <Dropdown.Item as={Link} intc-id="paymentMethodsHeaderButton" to="/billing/managePaymentMethods">
              Payment Methods
            </Dropdown.Item>
          ) : null}

          {isOwnCloudAccount &&
          checkRoles([AppRolesEnum.Standard, AppRolesEnum.Premium, AppRolesEnum.Enterprise, AppRolesEnum.Intel]) ? (
            <Dropdown.Item as={Link} intc-id="cloudCreditsHeaderButton" to="/billing/credits">
              Cloud Credits
            </Dropdown.Item>
          ) : null}

          {isFeatureFlagEnable(appFeatureFlags.REACT_APP_FEATURE_API_KEYS) &&
          checkRoles([
            AppRolesEnum.Standard,
            AppRolesEnum.Premium,
            AppRolesEnum.Enterprise,
            AppRolesEnum.Intel,
            AppRolesEnum.EnterprisePending
          ]) ? (
            <Dropdown.Item as={Link} intc-id="apiKeysHeaderButton" to="/profile/apikeys">
              Dev Token
            </Dropdown.Item>
          ) : null}
        </span>
        <Dropdown.Divider />
        <DarkModeContainerSwitch />
        <Dropdown.Item onClick={logoutHandler} intc-id="signOutHeaderButton">
          <BsBoxArrowRight className="me-1"></BsBoxArrowRight>Sign-out
        </Dropdown.Item>
      </Dropdown.Menu>
    </Dropdown>
  )
}

export default UserMenu
