// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { BsWallet2, BsHouse, BsBell, BsHddStack } from 'react-icons/bs'
import { TbCpuOff } from 'react-icons/tb'
import { MdManageAccounts, MdKey } from 'react-icons/md'
import { ReactComponent as K8s } from '../../assets/images/k8.svg'
import { ReactComponent as Catalog } from '../../assets/images/custom-catalog.svg'
import { type IdcNavigation } from './Navigation.types'
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
    name: 'Dashboard',
    icon: BsHouse,
    showInMenu: true,
    showBadge: false
  },
  // *********************************************
  // Cloud Credits
  // *********************************************
  {
    name: 'Cloud Credits',
    path: '',
    showInMenu: true,
    icon: BsWallet2,
    showBadge: false,
    children: [
      {
        path: '/billing/coupons',
        name: 'Details',
        showBadge: false,
        showInMenu: false
      },
      {
        path: '/cloudcredits/create',
        name: 'Create Cloud Credits',
        showBadge: false,
        showInMenu: false
      }
    ]
  },
  // *********************************************
  // Cloud Account Instance Whitelist
  // *********************************************
  {
    name: 'Cloud Account Instance Whitelist',
    path: '',
    showInMenu: true,
    icon: MdManageAccounts,
    showBadge: false,
    children: [
      {
        path: '/camanagement/details',
        name: 'Details',
        showBadge: false
      },
      {
        path: '/camanagement/create',
        name: 'Account Assignment',
        showBadge: false
      }
    ]
  },
  // *********************************************
  // Terminate Instances
  // *********************************************
  {
    name: 'Terminate Instance',
    path: '/instances/terminate',
    showInMenu: true,
    icon: TbCpuOff,
    showBadge: false
  },
  // *********************************************
  // Developer Tools
  // *********************************************
  {
    name: 'Developer Tools',
    path: '/profile/apikeys',
    showInMenu: true,
    icon: MdKey,
    showBadge: false
  },
  // *********************************************
  // Cloud Account Management
  // *********************************************
  {
    name: 'Cloud Account Management',
    path: '/usermanagement/details',
    showInMenu: true,
    icon: MdManageAccounts,
    showBadge: false
  },
  // *********************************************
  // IKS Management
  // *********************************************
  {
    name: 'IKS Management',
    path: '',
    showInMenu: true,
    icon: K8s,
    showBadge: false,
    children: [
      {
        path: '/iks/supercomputecluster/details',
        name: 'Super Compute Cluster',
        showBadge: false
      },
      {
        path: '/iks/cluster/details',
        name: 'Cluster',
        showBadge: false
      },
      {
        path: '/iks/imis',
        name: 'IMI',
        showBadge: false
      },
      {
        path: '/iks/kubescore',
        name: 'Kube Score',
        showBadge: false
      },
      {
        path: '/iks/k8s',
        name: 'K8S Version',
        showBadge: false
      },
      {
        path: '/iks/instancetypes',
        name: 'Instance Types',
        showBadge: false
      }
    ]
  },
  // *********************************************
  // Cloud Account Kubernetes Approval List
  // *********************************************
  {
    name: 'K8S Accounts Approval List',
    path: '/cloudaccounts/approvelist',
    showInMenu: true,
    icon: Catalog,
    showBadge: false
  },
  // *********************************************
  // Banners
  // *********************************************
  {
    name: 'Banners',
    path: '',
    showInMenu: true,
    icon: BsBell,
    showBadge: false,
    children: [
      {
        path: '/bannermanagement',
        name: 'Details',
        showBadge: false
      }
    ]
  },
  // *********************************************
  // Node Pool Management
  // *********************************************
  {
    name: 'Node Pool Management',
    path: '',
    showInMenu: true,
    icon: Catalog,
    showBadge: false,
    children: [
      {
        path: '/npm/pools',
        name: 'Pools',
        showBadge: false
      },
      {
        path: '/npm/nodes',
        name: 'Nodes',
        showBadge: false
      }
    ]
  },
  // *********************************************
  // Storage Management
  // *********************************************
  {
    name: 'Storage Management',
    path: '',
    showInMenu: true,
    icon: BsHddStack,
    showBadge: false,
    children: [
      {
        path: '/storagemanagement/services',
        name: 'Quota Services',
        showBadge: false
      },
      {
        path: '/storagemanagement/usages',
        name: 'Storage Usages',
        showBadge: false
      },
      {
        path: '/storagemanagement/quota',
        name: 'Quota Assigments',
        showBadge: false
      },
      {
        path: '/storagemanagement/managequota',
        name: 'Quota Management',
        showBadge: false
      }
    ]
  },
  // *********************************************
  // Quota Management
  // *********************************************
  {
    name: 'Quota Management',
    path: '',
    showInMenu: true,
    icon: BsHddStack,
    showBadge: false,
    children: [
      {
        path: '/quotamanagement/services',
        name: 'Quota Services',
        showBadge: false
      },
      {
        path: '/quotamanagement/services/create',
        name: 'Create Service',
        showBadge: false
      }
    ]
  },
  // *********************************************
  // Product Catalog Management
  // *********************************************
  {
    name: 'Product Catalog Self Service',
    path: '',
    showInMenu: true,
    icon: Catalog,
    showBadge: false,
    children: [
      {
        path: '/products/vendors',
        name: 'Vendors',
        showBadge: false
      },
      {
        path: '/products/families',
        name: 'Families',
        showBadge: false
      },
      {
        path: '/products',
        name: 'Products',
        showBadge: false
      }
    ]
  },
  // *********************************************
  // User Account Summary
  // *********************************************
  {
    name: 'User Account Summary',
    path: '',
    showInMenu: true,
    icon: MdManageAccounts,
    showBadge: false,
    children: [
      {
        path: '/usersummary',
        name: 'User Details',
        showBadge: false
      }
    ]
  },
  // *********************************************
  // Region Management
  // *********************************************
  {
    name: 'Region Management',
    path: '',
    showInMenu: true,
    icon: BsHddStack,
    showBadge: false,
    children: [
      {
        path: '/regionmanagement/regions',
        name: 'Regions',
        showBadge: false
      },
      {
        path: '/regionmanagement/whitelist',
        name: 'Whitelist Accounts',
        showBadge: false
      }
    ]
  }
]
