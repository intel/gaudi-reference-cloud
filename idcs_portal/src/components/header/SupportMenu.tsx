// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import { Nav, Dropdown, Button } from 'react-bootstrap'
import { AppRolesEnum } from '../../utils/Enums'
import { checkRoles } from '../../utils/accessControlWrapper/AccessControlWrapper'
import { BsQuestionCircle } from 'react-icons/bs'
import idcConfig, { appFeatureFlags, isFeatureFlagEnable } from '../../config/configurator'

const SupportMenuItems = [
  {
    label: 'Browse documentation',
    href: idcConfig.REACT_APP_PUBLIC_DOCUMENTATION,
    allowedRoles: [
      AppRolesEnum.Standard,
      AppRolesEnum.Premium,
      AppRolesEnum.Enterprise,
      AppRolesEnum.EnterprisePending,
      AppRolesEnum.Intel
    ]
  },
  {
    label: 'Community',
    href: idcConfig.REACT_APP_COMMUNITY,
    allowedRoles: [
      AppRolesEnum.Standard,
      AppRolesEnum.Premium,
      AppRolesEnum.Enterprise,
      AppRolesEnum.EnterprisePending,
      AppRolesEnum.Intel
    ]
  },
  {
    label: 'Knowledge base',
    href: idcConfig.REACT_APP_KNOWLEDGE_BASE,
    allowedRoles: [
      AppRolesEnum.Standard,
      AppRolesEnum.Premium,
      AppRolesEnum.Enterprise,
      AppRolesEnum.EnterprisePending,
      AppRolesEnum.Intel
    ]
  },
  {
    label: 'Submit a ticket',
    href: idcConfig.REACT_APP_SUBMIT_TICKET,
    allowedRoles: [AppRolesEnum.Premium, AppRolesEnum.Enterprise, AppRolesEnum.Intel, AppRolesEnum.Standard]
  },
  {
    label: 'Contact support',
    href: idcConfig.REACT_APP_SUPPORT_PAGE,
    allowedRoles: [AppRolesEnum.Premium, AppRolesEnum.Enterprise, AppRolesEnum.Intel]
  },
  {
    label: 'Send feedback',
    href: idcConfig.REACT_APP_PUBLIC_FEEDBACK_URL,
    allowedRoles: [
      AppRolesEnum.Standard,
      AppRolesEnum.Premium,
      AppRolesEnum.Enterprise,
      AppRolesEnum.EnterprisePending,
      AppRolesEnum.Intel
    ]
  }
]

const SupportMenu: React.FC = (): JSX.Element => {
  const { isPremiumUser, isEnterpriseUser } = useUserStore((state) => state)
  const supportLink = isEnterpriseUser()
    ? idcConfig.REACT_APP_SUBMIT_TICKET_ENTERPRISE
    : isPremiumUser()
      ? idcConfig.REACT_APP_SUBMIT_TICKET_PREMIUM
      : idcConfig.REACT_APP_SUBMIT_TICKET

  const getSupportMenu = (): JSX.Element[] => {
    const allowedOptions = SupportMenuItems.filter((x) => checkRoles(x.allowedRoles))
    return allowedOptions.map(function (item) {
      return item.href !== null ? (
        item.label === 'Send Feedback' ? (
          isFeatureFlagEnable(appFeatureFlags.REACT_APP_FEATURE_FEEDBACK) ? (
            <div className="m-0 p-0" key={`div-${item.label}`}>
              <Dropdown.Divider className="mt-0" />
              <Dropdown.Item
                intc-id={`help-menu-${item.label.replaceAll(' ', '-')}`}
                href={item.href}
                key={item.label}
                target="_blank"
                rel="noreferrer"
                role="button"
                aria-label={`Go to ${item.label}`}
              >
                {item.label}
              </Dropdown.Item>
            </div>
          ) : (
            <></>
          )
        ) : (
          <Dropdown.Item
            intc-id={`help-menu-${item.label.replaceAll(' ', '-')}`}
            href={item.label === 'Submit a ticket' ? supportLink : item.href}
            key={item.label}
            target="_blank"
            rel="noreferrer"
            role="button"
            aria-label={`Go to ${item.label}`}
          >
            {item.label}
          </Dropdown.Item>
        )
      ) : (
        <Dropdown.ItemText key={item.label} role="button" aria-label={`Go to ${item.label}`}>
          {item.label}
        </Dropdown.ItemText>
      )
    })
  }

  return (
    <Dropdown intc-id="help-menu" className="align-self-center" as={Nav.Item}>
      <Dropdown.Toggle
        id="dropdown-header-support-toggle"
        role="combobox"
        variant="icon-simple"
        className="d-flex align-items-center"
        aria-label="Support menu"
        data-bs-toggle="dropdown"
        aria-expanded="false"
        aria-controls="dropdown-header-menu-support"
        as={Button}
      >
        <BsQuestionCircle intc-id="helpIcon" title="Support Icon" />
      </Dropdown.Toggle>
      <Dropdown.Menu
        renderOnMount
        id="dropdown-header-menu-support"
        align="end"
        aria-labelledby="dropdown-header-support-toggle"
      >
        {getSupportMenu()}
      </Dropdown.Menu>
    </Dropdown>
  )
}

export default SupportMenu
