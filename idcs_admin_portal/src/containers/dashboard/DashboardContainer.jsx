import Dashboard from '../../components/dashboard/Dashboard'
import { appFeatureFlags } from '../../config/configurator'
import { AppRolesEnum } from '../../utility/Enums'
import { checkRoles } from '../../utility/wrapper/AccessControlWrapper'

const DashboardContainer = () => {
  // A container handle the function, api call and redux for each module
  const options = [
    {
      title: 'Request AGS Groups',
      description: 'View the AGS Groups that you would like to request.',
      buttons: [
        {
          href: '/request/agsgroup',
          text: 'View'
        }
      ],
      isCurrentRoleMatching: true
    },
    {
      title: 'Cloud Credits',
      description: 'Perform actions related to manage cloud credit codes.',
      buttons: [
        {
          href: '/billing/coupons',
          text: 'View Cloud Credits'
        },
        {
          href: '/cloudcredits/create',
          text: 'Create Cloud Credits'
        }
      ],
      isCurrentRoleMatching: checkRoles([
        AppRolesEnum.SREAdmin,
        AppRolesEnum.ComputeAdmin,
        AppRolesEnum.ProductAdmin,
        AppRolesEnum.SuperAdmin
      ])
    },
    {
      title: 'Cloud Account Instance Whitelist',
      description: 'Whitelist the Cloud Accounts for products.',
      buttons: [
        {
          href: '/camanagement/details',
          text: 'Cloud Account Details'
        },
        {
          href: '/camanagement/create',
          text: 'Cloud Account Assignment'
        }
      ],
      isCurrentRoleMatching: checkRoles([AppRolesEnum.ProductAdmin, AppRolesEnum.SuperAdmin, AppRolesEnum.SREAdmin])
    },
    {
      title: 'Terminate Instances',
      description: 'Terminate the Instances for the specific accounts.',
      buttons: [
        {
          href: '/instances/terminate',
          text: 'View Instances'
        }
      ],
      isCurrentRoleMatching: checkRoles([AppRolesEnum.SREAdmin, AppRolesEnum.ComputeAdmin, AppRolesEnum.SuperAdmin])
    },
    {
      title: 'Developer Tools',
      description: 'Get Admin Token to Execute API calls',
      buttons: [
        {
          href: '/profile/apikeys',
          text: 'Admin Token'
        }
      ],
      isCurrentRoleMatching: checkRoles([
        AppRolesEnum.IKSAdmin,
        AppRolesEnum.ComputeAdmin,
        AppRolesEnum.SuperAdmin,
        AppRolesEnum.SREAdmin
      ])
    },
    {
      title: 'Cloud Account Management',
      description: 'Block/ Unblock the Cloud Accounts.',
      buttons: [
        {
          href: '/usermanagement/details',
          text: 'Cloud Account Details'
        }
      ],
      isCurrentRoleMatching: checkRoles([AppRolesEnum.SREAdmin, AppRolesEnum.ComputeAdmin, AppRolesEnum.SuperAdmin])
    },
    {
      title: 'Intel Kubernetes Service Mangement',
      description: 'Manage Intel Kubernetes Service Components',
      buttons: [
        {
          href: '/iks/cluster/details',
          text: 'Manage Intel Kubernetes Service'
        }
      ],
      isCurrentRoleMatching: checkRoles([AppRolesEnum.IKSAdmin, AppRolesEnum.SuperAdmin, AppRolesEnum.SREAdmin])
    },
    {
      title: 'Cloud Account Kubernetes Approval List',
      description: 'Approve Cloud Accounts to use Kubernetes Clusters.',
      buttons: [
        {
          href: '/cloudaccounts/approvelist',
          text: 'Add / Remove Cloud Accounts'
        }
      ],
      isCurrentRoleMatching: checkRoles([AppRolesEnum.IKSAdmin, AppRolesEnum.SuperAdmin, AppRolesEnum.ProductAdmin])
    },
    {
      title: 'Storage',
      description: 'Update/ Create storage quota for IKS',
      buttons: [
        {
          href: '/storagemanagement/quota',
          text: 'Manage Storage Quota'
        },
        {
          href: '/storagemanagement/usages',
          text: 'Usages Storage'
        }
      ],
      featureFlag: appFeatureFlags.REACT_APP_FEATURE_STORAGE_MANAGEMENT,
      isCurrentRoleMatching: checkRoles([AppRolesEnum.StorageAdmin, AppRolesEnum.SuperAdmin])
    },
    {
      title: 'Quota Management',
      description: 'Update/ Create service quotas',
      buttons: [
        {
          href: '/quotamanagement/services',
          text: 'Manage Services'
        }
      ],
      isCurrentRoleMatching: checkRoles([AppRolesEnum.QuotaAdmin, AppRolesEnum.SuperAdmin])
    },
    {
      title: 'UI Console Banners',
      description: 'Update/ Create new UI console banners',
      buttons: [
        {
          href: '/bannermanagement',
          text: 'Manage Banners'
        }
      ],
      isCurrentRoleMatching: checkRoles([AppRolesEnum.SuperAdmin, AppRolesEnum.BannerAdmin]),
      featureFlag: appFeatureFlags.REACT_APP_FEATURE_BANNER_MANAGEMENT
    },
    {
      title: 'Node Pool Management',
      description: 'Manage Node Pools, Pool List, and its associated Cloud Accounts',
      buttons: [
        {
          href: '/npm/pools',
          text: 'Pool Management'
        },
        {
          href: '/nodepoolmanagement/nodes',
          text: 'Manage Node Pool List'
        },
        {
          href: '/npm/nodes',
          text: 'Node Management'
        }
      ],
      isCurrentRoleMatching: checkRoles([AppRolesEnum.SuperAdmin, AppRolesEnum.NodePoolAdmin]),
      featureFlag: appFeatureFlags.REACT_APP_FEATURE_NODE_POOL_MANAGEMENT
    },
    {
      title: 'Product Catalog Self Service',
      description: 'Create/Edit Products',
      buttons: [
        {
          href: '/products/vendors',
          text: 'Manage Catalog'
        }
      ],
      isCurrentRoleMatching: checkRoles([AppRolesEnum.SuperAdmin, AppRolesEnum.CatalogAdmin]),
      featureFlag: appFeatureFlags.REACT_APP_FEATURE_PRODUCT_CATALOG
    },
    {
      title: 'User Account Summary',
      description: 'User Status, Details, and related Deployed Services',
      buttons: [
        {
          href: '/usersummary',
          text: 'User Details'
        }
      ],
      isCurrentRoleMatching: checkRoles([AppRolesEnum.SuperAdmin, AppRolesEnum.ProductAdmin])
    },
    {
      title: 'Region Management',
      description: 'Create/ Update/ Delete Regions',
      buttons: [
        {
          href: '/regionmanagement/regions',
          text: 'Manage Regions'
        },
        {
          href: '/regionmanagement/whitelist',
          text: 'Whitelist Accounts'
        }
      ],
      isCurrentRoleMatching: checkRoles([AppRolesEnum.SuperAdmin, AppRolesEnum.RegionAdmin]),
      featureFlag: appFeatureFlags.REACT_APP_FEATURE_REGION_MANAGEMENT
    }
  ]
  return <Dashboard options={options} />
}

export default DashboardContainer
