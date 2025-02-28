// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import idcConfig, { appFeatureFlags } from '../../config/configurator'
import HardwareCatalogContainer from '../hardwareCatalog/HardwareCatalogContainer'
import HomePageContainer from '../homePage/HomePageContainer'
import KeyPairsContainer from '../keypairs/KeyPairsContainer'
import ImportKeysConstainer from '../keypairs/ImportKeysConstainer'
import SoftwareCatalogContainer from '../software/SoftwareCatalogContainer'
import SoftwareDetailContainer from '../software/SoftwareDetailContainer'
import TrainingAndWorkshopsContainer from '../trainingAndWorkshops/TrainingAndWorkshopsContainer'
import TrainingDetailContainer from '../trainingAndWorkshops/TrainingDetailContainer'
import { type IdcRoute } from './Routes.types'
import NotFound from '../../pages/error/NotFound'
import AccessDenied from '../../pages/error/AccessDenied'
import SomethingWentWrong from '../../pages/error/SomethingWentWrong'
import ComputeGroupsLaunchContainer from '../compute-groups/ComputeGroupsLaunchContainer'
import ComputeGroupsEditReservationContainer from '../compute-groups/ComputeGroupsEditReservationContainer'
import ComputeGroupsReservationsContainers from '../compute-groups/ComputeGroupsReservationsContainers'
import ComputeLaunchContainer from '../compute/ComputeLaunchContainer'
import ComputeReservationsContainers from '../compute/ComputeReservationsContainers'
import StorageReservationsContainer from '../storage/StorageReservationsContainer'
import StorageLaunchContainer from '../storage/StorageLaunchContainer'
import { AppRolesEnum } from '../../utils/Enums'
import ObjectStorageReservationsContainer from '../objectStorage/ObjectStorageReservationsContainer'
import ObjectStorageLaunchContainer from '../objectStorage/ObjectStorageLaunchContainer'
import ObjectStorageUsersReservationsContainer from '../objectStorage/ObjectStorageUsersReservationsContainer'
import ObjectStorageUsersLaunchContainer from '../objectStorage/ObjectStorageUsersLaunchContainer'
import InvoicesContainers from '../billing/InvoicesContainers'
import UsagesContainers from '../billing/UsagesContainer'
import NotificationsContainer from '../notifications/NotificationsContainer'
import CloudCreditsContainers from '../billing/CloudCreditsContainers'
import ManagePaymentMethodsContainer from '../billing/ManagePaymentMethodsContainer'
import ManageCreditCardContainers from '../billing/ManageCreditCardContainers'
import CreditCardResponseContainers from '../billing/paymentMethods/CreditCardResponseContainers'
import PremiumContainers from '../billing/PremiumContainers'
import UpgradeAccountContainers from '../billing/UpgradeAccountContainers'
import ClusterMyReservationsContainer from '../cluster/ClusterMyReservationsContainer'
import ClusterReserveContainer from '../cluster/ClusterReserveContainer'
import LoadbalancerReservationsContainer from '../cluster/LoadbalancerReservationsContainer'
import ClusterAddNodeGroupContainer from '../cluster/ClusterAddNodeGroupContainer'
import ClusterAddStorageContainer from '../cluster/ClusterAddStorageContainer'
import ClusterEditStorageContainer from '../cluster/ClusterEditStorageContainer'
import ApiKeysContainer from '../profile/ApiKeysContainer'
import AccountSettingsContainer from '../profile/AccountSettingsContainer'
import AccountsContainerInRoute from '../profile/AccountsContainerInRoute'
import AccessManagementContainer from '../profile/AccessManagementContainer'
import useUserStore from '../../store/userStore/UserStore'
import ManageCouponCodeContainers from '../billing/ManageCouponCodeContainers'
import ComputeEditReservationContainer from '../compute/ComputeEditReservationContainer'
import ObjectStorageUsersEditContainer from '../objectStorage/ObjectStorageUsersEditContainer'
import ObjectStorageRuleLaunchContainer from '../objectStorage/ObjectStorageRuleLaunchContainer'
import ObjectStorageRuleEditContainer from '../objectStorage/ObjectStorageRuleEditContainer'
import StorageEditContainer from '../storage/StorageEditContainer'
import SoftwareLaunchContainer from '../software/SoftwareLaunchContainer'
import SuperComputerReservationsContainer from '../superComputer/SuperComputerReservationsContainer'
import SuperComputerHomePageContainer from '../superComputer/SuperComputerHomePageContainer'
import SuperComputerLaunchContainer from '../superComputer/SuperComputerLaunchContainer'
import SuperComputerDetailContainer from '../superComputer/SuperComputerDetailContainer'
import SuperComputerAddWorkerNodeContainer from '../superComputer/SuperComputerAddWorkerNodeContainer'
import SuperComputerAddLoadBalancerContainer from '../superComputer/SuperComputerAddLoadBalancerContainer'
import SuperComputerAddStorageContainer from '../superComputer/SuperComputerAddStorageContainer'
import SuperComputerSecurityRuleEditContainer from '../superComputer/SuperComputerSecurityRuleEditContainer'
import MetricsContainer from '../metrics/MetricsContainer'
import ComputeReservationsDetailsContainer from '../compute/ComputeReservationsDetailsContainer'
import ComputeGroupsReservationsDetailsContainer from '../compute-groups/ComputeGroupsReservationsDetailsContainers'
import ObjectStorageReservationsDetailsContainer from '../objectStorage/ObjectStorageReservationsDetailsContainer'
import ObjectStorageUsersReservationsDetailsContainer from '../objectStorage/ObjectStorageUsersReservationsDetailsContainer'
import StorageReservationsDetailsContainer from '../storage/StorageReservationsDetailsContainer'
import ClusterHomePageContainer from '../cluster/ClusterHomePageContainer'
import ClusterMyReservationsDetailsContainer from '../cluster/ClusterMyReservationsDetailsContainer'
import LoadBalancerLaunchContainer from '../loadBalancer/LoadBalancerLaunchContainer'
import LoadBalancerReservationsContainer from '../loadBalancer/LoadBalancerReservationsContainer'
import LearningLabsCatalogContainer from '../trainingAndWorkshops/labs/LearningLabsCatalogContainer'
import LearningLabsLLMChatContainer from '../trainingAndWorkshops/labs/LearningLabsLLMChatContainer'
import LoadBalancerReservationsDetailsContainer from '../loadBalancer/LoadBalancerReservationsDetailsContainer'
import LoadBalancerEditContainer from '../loadBalancer/LoadBalancerEditContainer'
import LearningLabsTextToImageContainer from '../trainingAndWorkshops/labs/LearningLabsTextToImageContainer'
import GetStartedContainer from '../getStarted/GetStartedContainer'
import ClusterSecurityRulesEditContainer from '../cluster/ClusterSecurityRulesEditContainer'
import UserCredentialsContainer from '../profile/UserCredentialsContainer'
import UserCredentialsLaunchContainer from '../profile/UserCredentialsLaunchContainer'
import AccountRolesContainer from '../profile/AccountRolesContainer'
import AccountRolesDetailsContainer from '../profile/AccountRolesDetailsContainer'
import AccountRolesCreateContainer from '../profile/AccountRolesCreateContainer'
import AccountRolesEditContainer from '../profile/AccountRolesEditContainer'
import UserRolesContainer from '../profile/UserRolesContainer'
import Redirect from '../../utils/redirect/Redirect'
import RedirectWithParam from '../../utils/redirect/RedirectWithParam'
import MaaSLLMChatContainer from '../maas/MaaSLLMChatContainer'
import ClusterMetricsContainer from '../metrics/ClusterMetricsContainer'
import InstanceGroupsMetricsContainer from '../metrics/InstanceGroupsMetricsContainer'

export const routes: IdcRoute[] = [
  // *********************************************
  // Global
  // *********************************************
  {
    path: '/',
    component: HomePageContainer,
    breadcrumTitle: 'Home',
    showBreadcrums: true
  },
  {
    path: '/home',
    breadcrumTitle: 'Home',
    showBreadcrums: true,
    component: HomePageContainer
  },
  {
    path: '/home/getstarted',
    component: GetStartedContainer,
    breadcrumTitle: 'Get Started',
    showBreadcrums: true
  },
  {
    path: '/error/notfound',
    component: NotFound
  },
  {
    path: '/error/accessdenied',
    component: AccessDenied
  },
  {
    path: '/error/somethingwentwrong',
    component: SomethingWentWrong
  },
  {
    path: '/notifications',
    component: NotificationsContainer,
    breadcrumTitle: 'Notifications',
    recentlyVisitedTitle: 'Notifications',
    showBreadcrums: true,
    featureFlag: appFeatureFlags.REACT_APP_FEATURE_NOTIFICATIONS
  },
  {
    path: '/profile/apikeys',
    component: ApiKeysContainer,
    recentlyVisitedTitle: 'API Keys',
    featureFlag: appFeatureFlags.REACT_APP_FEATURE_API_KEYS
  },
  {
    path: '/profile/accountsettings',
    component: AccountSettingsContainer,
    breadcrumTitle: 'Account settings',
    recentlyVisitedTitle: 'My information',
    showBreadcrums: true
  },
  {
    path: '/profile/accountAccessManagement',
    component: AccessManagementContainer,
    breadcrumTitle: 'Members',
    recentlyVisitedTitle: 'Members',
    showBreadcrums: true,
    allowedFn: () => {
      return useUserStore.getState().shouldShowAccessManagement() ?? false
    }
  },
  {
    path: '/profile/accountAccessManagement/user-role/:param',
    component: UserRolesContainer,
    breadcrumTitle: 'User roles',
    showBreadcrums: true,
    featureFlag: appFeatureFlags.REACT_APP_FEATURE_USER_ROLES,
    allowedFn: () => {
      return useUserStore.getState().shouldShowAccessManagement() ?? false
    }
  },
  {
    path: '/profile/roles',
    breadcrumTitle: 'Roles',
    recentlyVisitedTitle: 'Roles',
    component: AccountRolesContainer,
    featureFlag: appFeatureFlags.REACT_APP_FEATURE_USER_ROLES && appFeatureFlags.REACT_APP_FEATURE_MULTIUSER,
    showBreadcrums: true,
    allowedFn: () => {
      return useUserStore.getState().shouldShowAccessManagement() || !useUserStore.getState().isOwnCloudAccount
    }
  },
  {
    path: '/profile/roles/d/:param',
    breadcrumTitle: '',
    component: AccountRolesDetailsContainer,
    featureFlag: appFeatureFlags.REACT_APP_FEATURE_USER_ROLES,
    showBreadcrums: true
  },
  {
    path: '/profile/roles/reserve',
    breadcrumTitle: 'Create role',
    recentlyVisitedTitle: 'Create role',
    component: AccountRolesCreateContainer,
    featureFlag: appFeatureFlags.REACT_APP_FEATURE_USER_ROLES,
    showBreadcrums: true,
    allowedFn: () => {
      return useUserStore.getState().shouldShowAccessManagement() ?? false
    }
  },
  {
    path: '/profile/roles/d/:param/edit',
    breadcrumTitle: 'Update role',
    component: AccountRolesEditContainer,
    featureFlag: appFeatureFlags.REACT_APP_FEATURE_USER_ROLES && appFeatureFlags.REACT_APP_FEATURE_USER_ROLE_EDIT,
    showBreadcrums: true,
    allowedFn: () => {
      return useUserStore.getState().shouldShowAccessManagement() ?? false
    }
  },
  {
    path: '/profile/credentials',
    breadcrumTitle: 'Credentials',
    recentlyVisitedTitle: 'Credentials',
    component: UserCredentialsContainer,
    featureFlag: appFeatureFlags.REACT_APP_FEATURE_USER_CREDENTIALS,
    showBreadcrums: true,
    allowedFn: () => {
      return useUserStore.getState().isOwnCloudAccount
    }
  },
  {
    path: '/profile/credentials/launch',
    component: UserCredentialsLaunchContainer,
    breadcrumTitle: 'Generate API key',
    recentlyVisitedTitle: 'Generate API key',
    featureFlag: appFeatureFlags.REACT_APP_FEATURE_USER_CREDENTIALS,
    showBreadcrums: true,
    allowedFn: () => {
      return useUserStore.getState().isOwnCloudAccount
    }
  },
  {
    path: '/accounts',
    recentlyVisitedTitle: 'Accounts',
    component: AccountsContainerInRoute,
    allowedFn: () => {
      return useUserStore.getState().user?.hasInvitations ?? false
    }
  },
  {
    path: '/resources',
    href: 'https://www.intel.com/content/www/us/en/resources-documentation/developer.html'
  },
  {
    path: '/docs',
    href: idcConfig.REACT_APP_PUBLIC_DOCUMENTATION
  },
  // *********************************************
  // Billing
  // *********************************************
  {
    path: '/billing',
    component: () => <Redirect path="/billing/usages" />,
    showBreadcrums: false
  },
  {
    path: '/billing/invoices',
    component: InvoicesContainers,
    roles: [AppRolesEnum.Premium],
    memberNotAllowed: true,
    breadcrumTitle: 'Invoices',
    recentlyVisitedTitle: 'Invoices',
    showBreadcrums: true
  },
  {
    path: '/billing/usages',
    component: UsagesContainers,
    memberNotAllowed: true,
    breadcrumTitle: 'Usages',
    recentlyVisitedTitle: 'Usages',
    showBreadcrums: true
  },
  {
    path: '/billing/managePaymentMethods',
    component: ManagePaymentMethodsContainer,
    roles: [AppRolesEnum.Premium],
    memberNotAllowed: true,
    breadcrumTitle: 'Payment methods',
    recentlyVisitedTitle: 'Payment methods',
    showBreadcrums: true
  },
  {
    path: '/billing/managePaymentMethods/managecreditcard',
    component: ManageCreditCardContainers,
    roles: [AppRolesEnum.Premium],
    featureFlag: appFeatureFlags.REACT_APP_FEATURE_DIRECTPOST,
    memberNotAllowed: true,
    breadcrumTitle: 'Add a credit card',
    recentlyVisitedTitle: 'Add credit card',
    showBreadcrums: true
  },
  {
    path: '/billing/creditResponse',
    component: CreditCardResponseContainers,
    roles: [AppRolesEnum.Premium],
    featureFlag: appFeatureFlags.REACT_APP_FEATURE_DIRECTPOST,
    memberNotAllowed: true
  },
  {
    path: '/billing/managecouponcode',
    component: () => <Redirect path="/billing/credits/managecouponcode" />,
    showBreadcrums: false,
    roles: [AppRolesEnum.Enterprise, AppRolesEnum.Premium, AppRolesEnum.Standard, AppRolesEnum.Intel],
    memberNotAllowed: true
  },
  {
    path: '/billing/credits/managecouponcode',
    component: ManageCouponCodeContainers,
    roles: [AppRolesEnum.Enterprise, AppRolesEnum.Premium, AppRolesEnum.Standard, AppRolesEnum.Intel],
    memberNotAllowed: true,
    breadcrumTitle: 'Reedem coupon',
    recentlyVisitedTitle: 'Reedem coupon',
    showBreadcrums: true
  },
  {
    path: '/billing/credits',
    component: CloudCreditsContainers,
    roles: [AppRolesEnum.Enterprise, AppRolesEnum.Premium, AppRolesEnum.Standard, AppRolesEnum.Intel],
    memberNotAllowed: true,
    breadcrumTitle: 'Cloud credits',
    recentlyVisitedTitle: 'Cloud credits',
    showBreadcrums: true
  },
  {
    path: '/premium',
    component: PremiumContainers,
    roles: [AppRolesEnum.Premium, AppRolesEnum.Standard]
  },
  {
    path: '/upgradeaccount',
    component: UpgradeAccountContainers,
    roles: [AppRolesEnum.Premium, AppRolesEnum.Standard],
    featureFlag: appFeatureFlags.REACT_APP_FEATURE_UPGRADE_TO_PREMIUM,
    memberNotAllowed: true
  },
  // *********************************************
  // Compute
  // *********************************************
  {
    path: '/hardware',
    recentlyVisitedTitle: 'Hardware catalog',
    component: HardwareCatalogContainer,
    showBreadcrums: true
  },
  {
    path: '/compute',
    component: ComputeReservationsContainers,
    breadcrumTitle: 'Instances',
    recentlyVisitedTitle: 'Instances',
    showBreadcrums: true
  },
  {
    path: '/compute/d/:param',
    component: ComputeReservationsDetailsContainer,
    breadcrumTitle: '',
    showBreadcrums: true
  },
  {
    path: '/compute/myreservations',
    component: () => <Redirect path="/compute" />,
    showBreadcrums: false
  },
  {
    path: '/compute/reserve',
    component: ComputeLaunchContainer,
    breadcrumTitle: 'Launch a compute instance',
    recentlyVisitedTitle: 'Launch instance',
    showBreadcrums: true
  },
  {
    path: '/compute/d/:param/edit',
    component: ComputeEditReservationContainer,
    breadcrumTitle: 'Edit',
    showBreadcrums: true
  },
  {
    path: '/compute-groups',
    component: ComputeGroupsReservationsContainers,
    breadcrumTitle: 'Instance groups',
    recentlyVisitedTitle: 'Instance groups',
    showBreadcrums: true
  },
  {
    path: '/compute-groups/d/:param',
    component: ComputeGroupsReservationsDetailsContainer,
    breadcrumTitle: '',
    showBreadcrums: true
  },
  {
    path: '/compute-groups/reserve',
    component: ComputeGroupsLaunchContainer,
    breadcrumTitle: 'Launch instance group',
    recentlyVisitedTitle: 'Launch instance group',
    showBreadcrums: true
  },
  {
    path: '/compute-groups/d/:param/edit',
    component: ComputeGroupsEditReservationContainer,
    breadcrumTitle: 'Edit',
    showBreadcrums: true
  },
  {
    path: '/security/publickeys',
    component: KeyPairsContainer,
    breadcrumTitle: 'Keys',
    recentlyVisitedTitle: 'Keys',
    showBreadcrums: true
  },
  {
    path: '/security/publickeys/import',
    component: ImportKeysConstainer,
    breadcrumTitle: 'Upload key',
    recentlyVisitedTitle: 'Upload key',
    showBreadcrums: true
  },
  // *********************************************
  // Intel Kubernetes Service
  // *********************************************
  {
    path: '/cluster',
    component: ClusterMyReservationsContainer,
    featureFlag: appFeatureFlags.REACT_APP_FEATURE_KaaS,
    breadcrumTitle: 'Intel kubernetes cluster',
    recentlyVisitedTitle: 'Clusters',
    showBreadcrums: true
  },
  {
    path: '/cluster/myreservations',
    component: () => <Redirect path="/cluster" />,
    showBreadcrums: false,
    featureFlag: appFeatureFlags.REACT_APP_FEATURE_KaaS
  },
  {
    path: '/cluster/overview',
    component: ClusterHomePageContainer,
    featureFlag: appFeatureFlags.REACT_APP_FEATURE_KaaS,
    breadcrumTitle: 'Intel kubernetes cluster overview',
    recentlyVisitedTitle: 'Overview',
    showBreadcrums: false
  },
  {
    path: '/cluster/d/:param',
    component: ClusterMyReservationsDetailsContainer,
    featureFlag: appFeatureFlags.REACT_APP_FEATURE_KaaS,
    breadcrumTitle: '',
    showBreadcrums: true
  },
  {
    path: '/cluster/reserve',
    component: ClusterReserveContainer,
    featureFlag: appFeatureFlags.REACT_APP_FEATURE_KaaS,
    breadcrumTitle: 'Launch kubernetes cluster',
    recentlyVisitedTitle: 'Launch cluster',
    showBreadcrums: true
  },
  {
    path: '/cluster/d/:param/reserveLoadbalancer',
    component: LoadbalancerReservationsContainer,
    featureFlag: appFeatureFlags.REACT_APP_FEATURE_KaaS,
    breadcrumTitle: 'Add load balancer',
    showBreadcrums: true
  },
  {
    path: '/cluster/d/:param/editSecurityRule',
    component: ClusterSecurityRulesEditContainer,
    featureFlag: appFeatureFlags.REACT_APP_FEATURE_IKS_SECURITY,
    breadcrumTitle: 'Edit cluster endpoint access',
    showBreadcrums: true
  },
  {
    path: '/cluster/d/:param/addnodegroup',
    component: ClusterAddNodeGroupContainer,
    featureFlag: appFeatureFlags.REACT_APP_FEATURE_KaaS,
    breadcrumTitle: 'Add node group',
    showBreadcrums: true
  },
  {
    path: '/cluster/d/:param/addstorage',
    component: ClusterAddStorageContainer,
    featureFlag: appFeatureFlags.REACT_APP_FEATURE_IKS_STORAGE,
    breadcrumTitle: 'Add node storage',
    showBreadcrums: true
  },
  {
    path: '/cluster/d/:param/editstorage',
    component: ClusterEditStorageContainer,
    featureFlag: appFeatureFlags.REACT_APP_FEATURE_IKS_STORAGE,
    breadcrumTitle: 'Edit node storage',
    showBreadcrums: true
  },
  // *********************************************
  // Storage
  // *********************************************
  {
    path: '/storage',
    component: StorageReservationsContainer,
    featureFlag: appFeatureFlags.REACT_APP_FEATURE_STORAGE,
    breadcrumTitle: 'Storage',
    recentlyVisitedTitle: 'File storage',
    showBreadcrums: true
  },
  {
    path: '/storage/d/:param',
    component: StorageReservationsDetailsContainer,
    featureFlag: appFeatureFlags.REACT_APP_FEATURE_STORAGE,
    breadcrumTitle: '',
    showBreadcrums: true
  },
  {
    path: '/storage/myreservations',
    component: () => <Redirect path="/storage" />,
    featureFlag: appFeatureFlags.REACT_APP_FEATURE_STORAGE,
    showBreadcrums: false
  },
  {
    path: '/storage/reserve',
    component: StorageLaunchContainer,
    featureFlag: appFeatureFlags.REACT_APP_FEATURE_STORAGE,
    breadcrumTitle: 'Create volume',
    recentlyVisitedTitle: 'Create volume',
    showBreadcrums: true
  },
  {
    path: '/storage/d/:param/edit',
    component: StorageEditContainer,
    featureFlag: appFeatureFlags.REACT_APP_FEATURE_STORAGE_EDIT,
    breadcrumTitle: 'Edit',
    showBreadcrums: true
  },
  // *********************************************
  // Buckets
  // *********************************************
  {
    path: '/buckets',
    component: ObjectStorageReservationsContainer,
    featureFlag: appFeatureFlags.REACT_APP_FEATURE_OBJECT_STORAGE,
    breadcrumTitle: 'Object storage',
    recentlyVisitedTitle: 'Object storage',
    showBreadcrums: true
  },
  {
    path: '/buckets/d/:param',
    component: ObjectStorageReservationsDetailsContainer,
    featureFlag: appFeatureFlags.REACT_APP_FEATURE_OBJECT_STORAGE,
    breadcrumTitle: '',
    showBreadcrums: true
  },
  {
    path: '/buckets/reserve',
    component: ObjectStorageLaunchContainer,
    featureFlag: appFeatureFlags.REACT_APP_FEATURE_OBJECT_STORAGE,
    breadcrumTitle: 'Create storage bucket',
    recentlyVisitedTitle: 'Create bucket',
    showBreadcrums: true
  },
  {
    path: '/buckets/users',
    component: ObjectStorageUsersReservationsContainer,
    featureFlag: appFeatureFlags.REACT_APP_FEATURE_OBJECT_STORAGE,
    breadcrumTitle: 'Manage principals and permissions',
    recentlyVisitedTitle: 'Principals and permissions',
    showBreadcrums: true
  },
  {
    path: '/buckets/users/d/:param',
    component: ObjectStorageUsersReservationsDetailsContainer,
    featureFlag: appFeatureFlags.REACT_APP_FEATURE_OBJECT_STORAGE,
    breadcrumTitle: '',
    showBreadcrums: true
  },
  {
    path: '/buckets/users/reserve',
    component: ObjectStorageUsersLaunchContainer,
    featureFlag: appFeatureFlags.REACT_APP_FEATURE_OBJECT_STORAGE,
    breadcrumTitle: 'Create principal',
    recentlyVisitedTitle: 'Create principal',
    showBreadcrums: true
  },
  {
    path: '/buckets/users/d/:param/edit',
    component: ObjectStorageUsersEditContainer,
    featureFlag: appFeatureFlags.REACT_APP_FEATURE_OBJECT_STORAGE,
    breadcrumTitle: 'Edit',
    showBreadcrums: true
  },
  {
    path: '/buckets/d/:param/lifecyclerule/reserve',
    component: ObjectStorageRuleLaunchContainer,
    featureFlag: appFeatureFlags.REACT_APP_FEATURE_OBJECT_STORAGE,
    breadcrumTitle: 'Add lifecycle rule',
    showBreadcrums: true
  },
  {
    path: '/buckets/d/:param/lifecyclerule/e/:param2',
    component: ObjectStorageRuleEditContainer,
    featureFlag: appFeatureFlags.REACT_APP_FEATURE_OBJECT_STORAGE,
    breadcrumTitle: 'Edit lifecycle rule',
    showBreadcrums: true
  },
  // *********************************************
  // Software
  // *********************************************
  {
    path: '/software',
    component: SoftwareCatalogContainer,
    breadcrumTitle: 'Software',
    recentlyVisitedTitle: 'Software catalog',
    showBreadcrums: true
  },
  {
    path: '/software/d/:param',
    component: SoftwareDetailContainer,
    breadcrumTitle: ' ',
    showBreadcrums: true
  },
  {
    path: '/software/d/:param/launch',
    component: SoftwareLaunchContainer,
    breadcrumTitle: 'Launch',
    showBreadcrums: true
  },
  // *********************************************
  // Training
  // *********************************************
  {
    path: '/learning/notebooks',
    component: TrainingAndWorkshopsContainer,
    featureFlag: appFeatureFlags.REACT_APP_FEATURE_TRAINING,
    breadcrumTitle: 'Notebooks',
    recentlyVisitedTitle: 'Available notebooks',
    showBreadcrums: true
  },
  {
    path: '/learning/notebooks/detail/:param',
    component: TrainingDetailContainer,
    featureFlag: appFeatureFlags.REACT_APP_FEATURE_TRAINING,
    breadcrumTitle: ' ',
    showBreadcrums: true
  },
  {
    path: '/learning/labs',
    component: LearningLabsCatalogContainer,
    featureFlag: appFeatureFlags.REACT_APP_FEATURE_LEARNING_LABS,
    breadcrumTitle: 'Labs',
    recentlyVisitedTitle: 'Available labs',
    showBreadcrums: true
  },
  {
    path: '/learning/labs/textimage',
    component: LearningLabsTextToImageContainer,
    breadcrumTitle: 'Text-to-Image with Stable Diffusion',
    recentlyVisitedTitle: 'Lab - Text-to-Image with Stable Diffusion',
    showBreadcrums: true,
    featureFlag: appFeatureFlags.REACT_APP_FEATURE_LEARNING_LABS
  },
  {
    path: '/learning/labs/llmchat',
    component: LearningLabsLLMChatContainer,
    breadcrumTitle: 'LLM chat',
    recentlyVisitedTitle: 'Lab - LLM chat',
    showBreadcrums: true,
    featureFlag: appFeatureFlags.REACT_APP_FEATURE_LEARNING_LABS
  },
  {
    path: '/training',
    component: () => <Redirect path="/learning/notebooks" />,
    featureFlag: appFeatureFlags.REACT_APP_FEATURE_TRAINING,
    showBreadcrums: false
  },
  {
    path: '/training/detail/:param',
    component: () => <RedirectWithParam path="/learning/notebooks/detail" />,
    featureFlag: appFeatureFlags.REACT_APP_FEATURE_TRAINING,
    showBreadcrums: false
  },
  {
    path: '/learning',
    component: () => <Redirect path="/learning/notebooks" />,
    featureFlag: appFeatureFlags.REACT_APP_FEATURE_TRAINING,
    showBreadcrums: false
  },
  {
    path: '/learning/detail/:param',
    component: () => <RedirectWithParam path="/learning/notebooks/detail" />,
    featureFlag: appFeatureFlags.REACT_APP_FEATURE_TRAINING,
    showBreadcrums: false
  },
  // *********************************************
  // Super Computer
  // ********************************************
  {
    path: '/supercomputer/overview',
    component: SuperComputerHomePageContainer,
    featureFlag: appFeatureFlags.REACT_APP_FEATURE_SUPER_COMPUTER,
    breadcrumTitle: 'Supercomputing cluster overview',
    recentlyVisitedTitle: 'Overview',
    showBreadcrums: false
  },
  {
    path: '/supercomputer',
    component: SuperComputerReservationsContainer,
    featureFlag: appFeatureFlags.REACT_APP_FEATURE_SUPER_COMPUTER,
    breadcrumTitle: 'Supercomputing clusters',
    recentlyVisitedTitle: 'Clusters',
    showBreadcrums: true
  },
  {
    path: '/supercomputer/launch',
    component: SuperComputerLaunchContainer,
    featureFlag: appFeatureFlags.REACT_APP_FEATURE_SUPER_COMPUTER,
    breadcrumTitle: 'Launch',
    recentlyVisitedTitle: 'Launch cluster',
    showBreadcrums: true
  },
  {
    path: '/supercomputer/d/:param',
    component: SuperComputerDetailContainer,
    featureFlag: appFeatureFlags.REACT_APP_FEATURE_SUPER_COMPUTER,
    breadcrumTitle: '',
    showBreadcrums: true
  },
  {
    path: '/supercomputer/d/:param/addnodegroup',
    component: SuperComputerAddWorkerNodeContainer,
    featureFlag: appFeatureFlags.REACT_APP_FEATURE_SUPER_COMPUTER,
    breadcrumTitle: 'Add nodegroup',
    showBreadcrums: true
  },
  {
    path: '/supercomputer/d/:param/addloadbalancer',
    component: SuperComputerAddLoadBalancerContainer,
    featureFlag: appFeatureFlags.REACT_APP_FEATURE_SUPER_COMPUTER,
    breadcrumTitle: 'Add loadbalancer',
    showBreadcrums: true
  },
  {
    path: '/supercomputer/d/:param/addstorage',
    component: SuperComputerAddStorageContainer,
    featureFlag: appFeatureFlags.REACT_APP_FEATURE_SUPER_COMPUTER,
    breadcrumTitle: 'Add storage',
    showBreadcrums: true
  },
  {
    path: '/supercomputer/d/:param/editSecurityRule',
    component: SuperComputerSecurityRuleEditContainer,
    roles: [AppRolesEnum.Enterprise, AppRolesEnum.Premium, AppRolesEnum.Intel],
    featureFlag: appFeatureFlags.REACT_APP_FEATURE_SC_SECURITY,
    breadcrumTitle: 'Edit cluster endpoint access',
    showBreadcrums: true
  },
  // *********************************************
  // Metrics
  // *********************************************
  {
    path: '/metrics',
    component: () => <Redirect path="/metrics/instances" />,
    featureFlag: appFeatureFlags.REACT_APP_FEATURE_METRICS,
    showBreadcrums: false
  },
  {
    path: '/metrics/instances',
    component: MetricsContainer,
    breadcrumTitle: 'Instances Cloud Monitor',
    recentlyVisitedTitle: 'Compute',
    featureFlag: appFeatureFlags.REACT_APP_FEATURE_METRICS
  },
  {
    path: '/metrics/clusters',
    component: ClusterMetricsContainer,
    breadcrumTitle: 'Kubernetes Cloud Monitor',
    recentlyVisitedTitle: 'Kubernetes',
    featureFlag: appFeatureFlags.REACT_APP_FEATURE_METRICS_CLUSTER
  },
  {
    path: '/metrics/instance-groups',
    component: InstanceGroupsMetricsContainer,
    breadcrumTitle: 'Instance Groups Cloud Monitor',
    recentlyVisitedTitle: 'Compute Groups',
    featureFlag: appFeatureFlags.REACT_APP_FEATURE_METRICS_GROUPS
  },
  // *********************************************
  // Load Balancer
  // *********************************************
  {
    path: '/load-balancer',
    component: LoadBalancerReservationsContainer,
    featureFlag: appFeatureFlags.REACT_APP_FEATURE_LOAD_BALANCER,
    breadcrumTitle: 'Load balancers',
    recentlyVisitedTitle: 'Load balancers',
    showBreadcrums: true
  },
  {
    path: '/load-balancer/d/:param',
    component: LoadBalancerReservationsDetailsContainer,
    featureFlag: appFeatureFlags.REACT_APP_FEATURE_LOAD_BALANCER,
    breadcrumTitle: '',
    showBreadcrums: true
  },
  {
    path: '/load-balancer/reserve',
    component: LoadBalancerLaunchContainer,
    featureFlag: appFeatureFlags.REACT_APP_FEATURE_LOAD_BALANCER,
    breadcrumTitle: 'Launch a load balancer',
    recentlyVisitedTitle: 'Launch load balancer',
    showBreadcrums: true
  },
  {
    path: '/load-balancer/d/:param/edit',
    component: LoadBalancerEditContainer,
    featureFlag: appFeatureFlags.REACT_APP_FEATURE_LOAD_BALANCER,
    breadcrumTitle: 'Edit',
    showBreadcrums: true
  },
  // *********************************************
  // Model as a Service
  // *********************************************
  {
    path: '/software/d/:param/llm',
    component: MaaSLLMChatContainer,
    breadcrumTitle: 'LLM',
    showBreadcrums: true
  },
  // *********************************************
  // NOT FOUND
  // *********************************************
  {
    path: '*',
    component: NotFound
  }
]
