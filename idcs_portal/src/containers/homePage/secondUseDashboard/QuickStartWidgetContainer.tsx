// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import QuickStartWidget from '../../../components/homePage/secondUseDashboard/QuickStartWidget'
import { BsKey, BsCpu, BsHddRack, BsBoxes, BsDeviceSsd, BsCircle } from 'react-icons/bs'
import { appFeatureFlags, isFeatureFlagEnable } from '../../../config/configurator'
import { checkRoles } from '../../../utils/accessControlWrapper/AccessControlWrapper'
import { AppRolesEnum } from '../../../utils/Enums'

interface ActionItem {
  label: string
  icon: JSX.Element
  redirectTo: string
  disabled: boolean
}

export interface InitialState {
  actions: ActionItem[]
}

interface QuickStartWidgetContainerProps {
  navigateTo: (url: string) => void
}

const QuickStartWidgetContainer: React.FC<QuickStartWidgetContainerProps> = (props): JSX.Element => {
  const initialState: InitialState = {
    actions: [
      {
        label: 'Launch Compute Instance',
        icon: <BsCpu />,
        redirectTo: '/compute/reserve',
        disabled: false
      },
      {
        label: 'Launch Instance Group',
        icon: <BsHddRack />,
        redirectTo: '/compute-groups/reserve',
        disabled: false
      },
      {
        label: 'Create a Kubernetes Cluster',
        icon: <BsBoxes />,
        redirectTo: '/cluster/reserve',
        disabled:
          !isFeatureFlagEnable(appFeatureFlags.REACT_APP_FEATURE_KaaS) ||
          !checkRoles([AppRolesEnum.Intel, AppRolesEnum.Premium, AppRolesEnum.Enterprise])
      },
      {
        label: 'Create a Storage Volume',
        icon: <BsDeviceSsd />,
        redirectTo: '/storage/reserve',
        disabled: !isFeatureFlagEnable(appFeatureFlags.REACT_APP_FEATURE_STORAGE)
      },
      {
        label: 'Launch a JupyterLab',
        icon: <BsCircle />,
        redirectTo: '/learning/notebooks',
        disabled: !isFeatureFlagEnable(appFeatureFlags.REACT_APP_FEATURE_TRAINING)
      },
      {
        label: 'Upload a Key',
        icon: <BsKey />,
        redirectTo: '/security/publickeys/import',
        disabled: false
      }
    ]
  }

  return <QuickStartWidget state={initialState} navigateTo={props.navigateTo} />
}

export default QuickStartWidgetContainer
