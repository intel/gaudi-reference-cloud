// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import Button from 'react-bootstrap/Button'
import idcConfig, { appFeatureFlags, isFeatureFlagEnable } from '../../../config/configurator'
import { EnrollAccountType } from '../../../utils/Enums'
import UserBadge from '../../../utils/userBadge/UserBadge'
import LabelValuePair from '../../../utils/labelValuePair/LabelValuePair'
import { BsCopy } from 'react-icons/bs'

const AccountSettingsView = ({
  displayName,
  email,
  cloudAccountId,
  cloudAccountType,
  upgrade,
  isOwnCloudAccount,
  copyToClipboard
}) => {
  const accountTypeInformation = [
    {
      accountType: EnrollAccountType.standard,
      includesTitle: 'Standard tier includes:',
      included: [
        'Explore and evaluate the latest Intel® AI products.',
        'Develop AI skills.',
        'Access cutting edge learning resources.',
        'Limited access to compute instances.'
      ],
      notIncluded: [
        'Access to latest compute generations.',
        'Billed subscription for teams.',
        'Intel premium support.'
      ],
      allowUpgrade: isFeatureFlagEnable(appFeatureFlags.REACT_APP_FEATURE_UPGRADE_TO_PREMIUM),
      upgradeLabel: 'Upgrade to premium'
    },
    {
      accountType: EnrollAccountType.premium,
      includesTitle: 'Premium tier includes:',
      included: [
        'Access to latest compute generations.',
        'AI and machine learning software toolkits.',
        'Intel premium support.'
      ],
      notIncluded: [
        'Subscription based usage.',
        'Private cloud offering.',
        '24 x 7 Intel premium plus support.',
        'Enterprise services like supercomputing clusters.'
      ],
      allowUpgrade: isFeatureFlagEnable(appFeatureFlags.REACT_APP_FEATURE_UPGRADE_TO_ENTERPRISE),
      upgradeLabel: 'Upgrade to enterprise'
    },
    {
      accountType: EnrollAccountType.intel,
      included: [],
      notIncluded: [],
      allowUpgrade: false
    },
    {
      accountType: EnrollAccountType.enterprise,
      includesTitle: 'Enterprise tier includes:',
      included: [
        'Access to latest compute generations.',
        'Subscription based usage.',
        'Private cloud offering.',
        '24 x 7 Intel premium plus support.',
        'Enterprise services like supercomputing clusters.'
      ],
      notIncluded: [],
      allowUpgrade: false
    }
  ]

  const accountInformation = accountTypeInformation.find(
    (x) =>
      x.accountType === cloudAccountType ||
      (cloudAccountType === EnrollAccountType.enterprise_pending && x.accountType === EnrollAccountType.enterprise)
  )

  return (
    <>
      <div className="section">
        <h2>
          {isOwnCloudAccount ? 'Your' : 'Selected'} {`${idcConfig.REACT_APP_CONSOLE_SHORT_NAME} Account`}
        </h2>
        <div className="d-flex flex-row flex-wrap gap-s8">
          <LabelValuePair label="Cloud Account ID">
            {cloudAccountId}{' '}
            <Button
              onClick={() => {
                copyToClipboard(cloudAccountId)
              }}
              variant="icon-simple"
              intc-id="btn-copy-cloud-account-id-account-settings"
              data-wap_ref="btn-copy-cloud-account-id-account-settings"
              aria-label="Copy cloud account id"
            >
              <BsCopy />
            </Button>
          </LabelValuePair>
          <div className="d-inline">
            <label className="fw-semibold" htmlFor="planTypeInfo">
              Tier:
            </label>
            <span className="fw-normal ms-s4" id="planTypeInfo">
              <UserBadge />
            </span>
            {accountInformation.allowUpgrade ? (
              <Button
                variant="primary"
                className="ms-s4"
                aria-label={accountInformation.upgradeLabel}
                intc-id="btn-account-upgrade"
                data-wap_ref="btn-account-upgrade"
                onClick={upgrade}
              >
                Upgrade
              </Button>
            ) : null}
          </div>
        </div>
        <div className="d-flex flex-row gap-s8 flex-wrap">
          {accountInformation.included.length > 0 ? (
            <div className="d-flex flex-column gap-s6">
              <label className="fw-semibold">{accountInformation.includesTitle}</label>
              <ul>
                {accountInformation.included.map((offering, index) => (
                  <li key={index}>{offering}</li>
                ))}
              </ul>
            </div>
          ) : null}
          {accountInformation.notIncluded.length > 0 ? (
            <div className="d-flex flex-column gap-s6">
              <label className="fw-semibold">Not included:</label>
              <ul>
                {accountInformation.notIncluded.map((offering, index) => (
                  <li key={index}>{offering}</li>
                ))}
              </ul>
            </div>
          ) : null}
        </div>
      </div>
      <div className="section">
        <h2>Your intel.com Account</h2>
        <LabelValuePair label="Name">{displayName}</LabelValuePair>
        <LabelValuePair label="Email">{email}</LabelValuePair>
      </div>
    </>
  )
}

export default AccountSettingsView
