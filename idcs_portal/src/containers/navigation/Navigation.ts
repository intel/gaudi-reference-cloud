// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import {
  BsCpu,
  BsHddStack,
  BsMortarboard,
  BsBook,
  BsWallet2,
  BsHouse,
  BsActivity,
  BsBell,
  BsPeople
} from 'react-icons/bs'
import { ReactComponent as Catalog } from '../../assets/images/custom-catalog.svg'
import { ReactComponent as K8s } from '../../assets/images/custom-k8.svg'
import { ReactComponent as SuperComputing } from '../../assets/images/custom-supercomp.svg'
import { type IdcNavigation } from './Navigation.types'
import idcConfig from '../../config/configurator'
import { getAllowedRoutes } from '../routing/Routing'

export const getAllowedNavigations = (isOwnCloudAccount: boolean): IdcNavigation[] => {
  const routesArr = getAllowedRoutes(isOwnCloudAccount)
  const navigationArr = idcNavigations()
  const navigationsArr = navigationArr.filter((x) => !x.path || routesArr.some((y) => y.path === x.path))
  navigationsArr.forEach((nav) => {
    if (nav.children) {
      nav.children = nav.children.filter((x) => !x.path || routesArr.some((y) => y.path === x.path))
    }
  })
  return navigationsArr.filter((x) => x.path || (!x.path && x.children !== undefined && x.children.length > 0))
}

const idcNavigations: () => IdcNavigation[] = () => [
  {
    path: '/',
    name: 'Home',
    icon: BsHouse,
    showInMenu: false,
    showBadge: false
  },
  {
    path: '/home',
    name: 'Home',
    icon: BsHouse,
    showInMenu: true,
    showBadge: false,
    children: [
      {
        path: '/home/getstarted',
        name: 'Get Started',
        icon: BsHouse,
        showInMenu: false,
        showBadge: false,
        showInToolbar: false
      }
    ]
  },
  // *********************************************
  // Catalog
  // *********************************************
  {
    path: '',
    name: 'Catalog',
    icon: Catalog,
    showInMenu: true,
    showBadge: false,
    children: [
      {
        path: '/hardware',
        name: 'Hardware',
        showBadge: false
      },
      {
        path: '/software',
        name: 'Software',
        showBadge: false
      }
    ]
  },
  // *********************************************
  // Compute
  // *********************************************
  {
    name: 'Compute',
    path: '',
    showInMenu: true,
    icon: BsCpu,
    showBadge: false,
    children: [
      {
        path: '/compute/overview',
        name: 'Overview',
        showBadge: false
      },
      {
        path: '/compute',
        name: 'Instances',
        showBadge: false
      },
      {
        path: '/compute-groups',
        name: 'Instance Groups',
        showBadge: false
      },
      {
        path: '/load-balancer',
        name: 'Load Balancers',
        showBadge: false
      },
      {
        path: '/security/publickeys',
        name: 'Keys',
        showBadge: false
      }
    ]
  },
  // *********************************************
  // K8s Services
  // *********************************************
  {
    name: 'Kubernetes',
    path: '',
    showInMenu: true,
    icon: K8s,
    showBadge: false,
    children: [
      {
        path: '/cluster/overview',
        name: 'Overview',
        showBadge: false
      },
      {
        path: '/cluster',
        name: 'Clusters',
        showBadge: false
      }
    ]
  },
  // *********************************************
  // Super Computing
  // *********************************************
  {
    name: 'Supercomputing',
    path: '',
    showInMenu: true,
    icon: SuperComputing,
    showBadge: false,
    children: [
      {
        path: '/supercomputer/overview',
        name: 'Overview',
        showBadge: false
      },
      {
        path: '/supercomputer',
        name: 'Clusters',
        showInToolbar: true
      }
    ]
  },
  // *********************************************
  // Storage
  // *********************************************
  {
    name: 'Storage',
    path: '',
    showInMenu: true,
    icon: BsHddStack,
    showBadge: false,
    children: [
      {
        path: '/storage/overview',
        name: 'Overview',
        showInMenu: false,
        showBadge: false
      },
      {
        path: '/storage',
        name: 'File Storage',
        showBadge: false
      },
      {
        path: '/buckets',
        name: 'Object Storage',
        showBadge: false
      }
    ]
  },

  // *********************************************
  // Billing
  // *********************************************
  {
    name: 'Billing',
    path: '',
    icon: BsWallet2,
    showInMenu: false,
    dividerAtTop: true,
    showBadge: false,
    children: [
      {
        path: '/billing/invoices',
        name: 'Invoices',
        showBadge: false
      },
      {
        path: '/billing/usages',
        name: 'Usage',
        showBadge: false
      },
      {
        path: '/billing/credits',
        name: 'Cloud Credits',
        showBadge: false
      },
      {
        path: '/billing/managePaymentMethods',
        name: 'Payment Methods',
        showBadge: false
      }
    ]
  },
  {
    path: '/premium',
    name: 'Account setup',
    icon: BsWallet2,
    showInMenu: false,
    showBadge: false
  },
  {
    path: '',
    name: 'Account Settings',
    icon: BsPeople,
    showInMenu: false,
    showBadge: false,
    dividerAtTop: true,
    children: [
      {
        path: '/profile/accountsettings',
        name: 'My Information',
        showBadge: false
      },
      {
        path: '/profile/accountAccessManagement',
        name: 'Members',
        showBadge: false
      },
      {
        path: '/profile/roles',
        name: 'Roles',
        showBadge: false
      },
      {
        path: '/profile/credentials',
        name: 'Credentials',
        showBadge: false
      }
    ]
  },
  // *********************************************
  // Metrics
  // *********************************************
  {
    name: 'Cloud Monitor',
    path: '',
    icon: BsActivity,
    showInMenu: true,
    dividerAtTop: true,
    showBadge: false,
    children: [
      {
        path: '/metrics/instances',
        name: 'Instances',
        showBadge: false,
        showInToolbar: true
      },
      {
        path: '/metrics/clusters',
        name: 'Kubernetes',
        showBadge: false,
        showInToolbar: true
      },
      {
        path: '/metrics/instance-groups',
        name: 'Instance Groups',
        showBadge: false,
        showInToolbar: true
      }
    ]
  },
  {
    path: '',
    name: 'Learning',
    icon: BsMortarboard,
    showInMenu: true,
    dividerAtTop: true,
    showBadge: false,
    children: [
      {
        path: '/learning/notebooks',
        name: 'Notebooks',
        showInMenu: true,
        showBadge: false
      },
      {
        path: '/learning/labs',
        name: 'Labs',
        showInMenu: true,
        showBadge: false
      }
    ]
  },
  {
    path: '/notifications',
    name: 'Notification',
    icon: BsBell,
    showInMenu: false,
    showBadge: false
  },
  {
    path: '/docs',
    externalPath: `${window.location.origin}${idcConfig.REACT_APP_PUBLIC_DOCUMENTATION}`,
    name: 'Documentation',
    icon: BsBook,
    showInMenu: true,
    showBadge: false
  }
]
