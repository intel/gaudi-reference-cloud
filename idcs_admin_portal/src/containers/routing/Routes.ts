// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { appFeatureFlags } from '../../config/configurator'
import { type IdcRoute } from './Routes.types'
import NotFound from '../../pages/error/NotFound'
import AccessDenied from '../../pages/error/AccessDenied'
import SomethingWentWrong from '../../pages/error/SomethingWentWrong'
import { AppRolesEnum } from '../../utility/Enums'
import ApiKeysContainer from '../profile/ApiKeysContainer'
import DashboardContainer from '../dashboard/DashboardContainer'
import RequestAGSGroupDetails from '../../pages/RequestAGSGroupDetails'
import CouponsContainer from '../coupons/CouponsContainer'
import CloudCreditsContainer from '../cloudCredits/CloudCreditsContainer'
import SkuQuotaDetailsContainer from '../skuManagement/SkuQuotaDetailsContainer'
import SkuQuotaCreateContainer from '../skuManagement/SkuQuotaCreateContainer'
import TerminateInstanceContainer from '../instanceManagement/TerminateInstanceContainer'
import UserManagementContainer from '../userManagement/UserManagementContainer'
import CloudAccountApprovalList from '../cloudApproveList/CloudAccountApprovalList'
import BannerManagementContainer from '../bannerManagement/BannerManagementContainer'
import BannerCreateContainer from '../bannerManagement/BannerCreateContainer'
import NodeListContainer from '../nodePoolManagement/NodeListContainer'
import NodeEditContainer from '../nodePoolManagement/NodeEditContainer'
import PoolListContainer from '../nodePoolManagement/PoolListContainer'
import PoolEditContainer from '../nodePoolManagement/PoolEditContainer'
import CloudAccountListContainer from '../nodePoolManagement/CloudAccountListContainer'
import AddCloudAccountContainer from '../nodePoolManagement/AddCloudAccountContainer'
import IMIContainer from '../imis/IMIContainer'
import K8SContainer from '../k8s/K8S'
import ClusterDetailsListContainer from '../cluster/ClusterDetailsListContainer'
import InstanceTypeContainer from '../instanceTypes/InstanceTypeContainer'
import StorageUsagesContainer from '../storageManagement/StorageUsagesContainer'
import AddQuotaContainer from '../storageManagement/AddQuotaContainer'
import QuotaAssigmentsContainer from '../storageManagement/QuotaAssigmentsContainer'
import QuotaManagementServiceContainer from '../storageManagement/QuotaManagementServiceContainer'
import QuotaManagementServiceCreateContainer from '../storageManagement/QuotaManagementServiceCreateContainer'
import QuotaManagementServiceEditContainer from '../storageManagement/QuotaManagementServiceEditContainer'
import QuotaManagementServiceQuotasContainer from '../storageManagement/QuotaManagementServiceQuotasContainer'
import QuotaManagementServiceAddQuotaContainer from '../storageManagement/QuotaManagementServiceAddQuotaContainer'
import QuotaManagementServiceEditQuotaContainer from '../storageManagement/QuotaManagementServiceEditQuotaContainer'
import NodeStatesContainer from '../nodePoolManagement/NodeStatesContainer'
import ProductVendorsContainer from '../productCatalog/ProductVendorsContainer'
import ProductFamiliesContainer from '../productCatalog/ProductFamiliesContainer'
import ProductCatalogContainer from '../productCatalog/ProductCatalogContainer'
import ProductCatalogEditContainer from '../productCatalog/ProductCatalogEditContainer'
import ProductCatalogCreate from '../productCatalog/ProductCatalogCreateContainer'
import ProductCatalogApprovalContainer from '../productCatalog/ProductCatalogApprovalContainer'
import ProductFamiliesCreateContainer from '../productCatalog/ProductFamiliesCreateContainer'
import ProductFamiliesEditContainer from '../productCatalog/ProductFamiliesEditContainer'
import ProductVendorsCreateContainer from '../productCatalog/ProductVendorsCreateContainer'
import ProductVendorsEditContainer from '../productCatalog/ProductVendorsEditContainer'
import KubeScoreContainer from '../cluster/KubeScoreContainer'
import ProductCatalogCreateMetaDatasetContainer from '../productCatalog/ProductCatalogCreateMetaDatasetContainer'
import UserSummaryContainer from '../userSummary/UserSummaryContainer'
import ProductCatalogEditMetaDatasetContainer from '../productCatalog/ProductCatalogEditMetaDatasetContainer'
import RegionManagementContainer from '../regionManagement/RegionManagementContainer'
import RegionCreateContainer from '../regionManagement/RegionCreateContainer'
import RegionEditContainer from '../regionManagement/RegionEditContainer'
import AccountRegionContainer from '../regionManagement/AccountRegionContainer'
import AddAccountRegionContainer from '../regionManagement/AddAccountRegionContainer'

export const routes: IdcRoute[] = [
  // *********************************************
  // Global
  // *********************************************
  {
    path: '/',
    component: DashboardContainer
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
    path: '/request/agsgroup',
    component: RequestAGSGroupDetails
  },
  {
    path: '/resources',
    href: 'https://www.intel.com/content/www/us/en/resources-documentation/developer.html'
  },
  // *********************************************
  // Cloud Credit
  // *********************************************
  {
    path: '/billing/coupons',
    component: CouponsContainer,
    roles: [
      AppRolesEnum.SREAdmin,
      AppRolesEnum.ProductAdmin,
      AppRolesEnum.IKSAdmin,
      AppRolesEnum.ComputeAdmin,
      AppRolesEnum.SuperAdmin
    ]
  },
  {
    path: '/cloudcredits/create',
    component: CloudCreditsContainer,
    roles: [
      AppRolesEnum.SREAdmin,
      AppRolesEnum.ProductAdmin,
      AppRolesEnum.IKSAdmin,
      AppRolesEnum.ComputeAdmin,
      AppRolesEnum.SuperAdmin
    ]
  },
  // *********************************************
  // Cloud Account Instance Whitelist
  // *********************************************
  {
    path: '/camanagement/details',
    component: SkuQuotaDetailsContainer,
    roles: [AppRolesEnum.ProductAdmin, AppRolesEnum.SuperAdmin, AppRolesEnum.SREAdmin]
  },
  {
    path: '/camanagement/create',
    component: SkuQuotaCreateContainer,
    roles: [AppRolesEnum.ProductAdmin, AppRolesEnum.SuperAdmin, AppRolesEnum.SREAdmin]
  },
  // *********************************************
  // Terminate Instance
  // *********************************************
  {
    path: '/instances/terminate',
    component: TerminateInstanceContainer,
    roles: [AppRolesEnum.SREAdmin, AppRolesEnum.ComputeAdmin, AppRolesEnum.SuperAdmin, AppRolesEnum.SlurmAdmin]
  },
  // *********************************************
  // Developer Tools
  // *********************************************
  {
    path: '/profile/apikeys',
    component: ApiKeysContainer,
    roles: [
      AppRolesEnum.IKSAdmin,
      AppRolesEnum.ComputeAdmin,
      AppRolesEnum.SuperAdmin,
      AppRolesEnum.SlurmAdmin,
      AppRolesEnum.SREAdmin
    ]
  },
  // *********************************************
  // Cloud Account Management
  // *********************************************
  {
    path: '/usermanagement/details',
    component: UserManagementContainer,
    roles: [AppRolesEnum.SREAdmin, AppRolesEnum.ComputeAdmin, AppRolesEnum.SuperAdmin, AppRolesEnum.SlurmAdmin]
  },
  // *********************************************
  // IKS Management
  // *********************************************
  {
    path: '/iks/imis',
    component: IMIContainer,
    roles: [AppRolesEnum.IKSAdmin, AppRolesEnum.SuperAdmin]
  },
  {
    path: '/iks/k8s',
    component: K8SContainer,
    roles: [AppRolesEnum.IKSAdmin, AppRolesEnum.SuperAdmin]
  },
  {
    path: '/iks/supercomputecluster/details',
    component: ClusterDetailsListContainer,
    componentProps: {
      page: 'Super Compute Cluster'
    },
    roles: [AppRolesEnum.IKSAdmin, AppRolesEnum.SuperAdmin, AppRolesEnum.SREAdmin]
  },
  {
    path: '/iks/cluster/details',
    component: ClusterDetailsListContainer,
    componentProps: {
      page: 'Cluster'
    },
    roles: [AppRolesEnum.IKSAdmin, AppRolesEnum.SuperAdmin, AppRolesEnum.SREAdmin]
  },
  {
    path: '/iks/instancetypes',
    component: InstanceTypeContainer,
    roles: [AppRolesEnum.IKSAdmin, AppRolesEnum.SuperAdmin]
  },
  {
    path: '/iks/kubescore',
    component: KubeScoreContainer,
    featureFlag: appFeatureFlags.REACT_APP_FEATURE_KUBE_SCORE,
    roles: [AppRolesEnum.IKSAdmin, AppRolesEnum.SuperAdmin]
  },
  // *********************************************
  // Cloud Account Kubernetes Approval List
  // *********************************************
  {
    path: '/cloudaccounts/approvelist',
    component: CloudAccountApprovalList,
    roles: [AppRolesEnum.IKSAdmin, AppRolesEnum.SuperAdmin, AppRolesEnum.ProductAdmin]
  },
  // *********************************************
  // UI Console Banners
  // *********************************************
  {
    path: '/bannermanagement',
    component: BannerManagementContainer,
    featureFlag: appFeatureFlags.REACT_APP_FEATURE_BANNER_MANAGEMENT,
    roles: [AppRolesEnum.SuperAdmin, AppRolesEnum.BannerAdmin]
  },
  {
    path: '/bannermanagement/update',
    component: BannerCreateContainer,
    featureFlag: appFeatureFlags.REACT_APP_FEATURE_BANNER_MANAGEMENT,
    roles: [AppRolesEnum.SuperAdmin, AppRolesEnum.BannerAdmin]
  },
  {
    path: '/bannermanagement/create',
    component: BannerCreateContainer,
    featureFlag: appFeatureFlags.REACT_APP_FEATURE_BANNER_MANAGEMENT,
    roles: [AppRolesEnum.SuperAdmin, AppRolesEnum.BannerAdmin]
  },
  // *********************************************
  // Node Pool Management
  // *********************************************
  {
    path: '/npm/pools',
    component: PoolListContainer,
    featureFlag: appFeatureFlags.REACT_APP_FEATURE_NODE_POOL_MANAGEMENT,
    roles: [AppRolesEnum.SuperAdmin, AppRolesEnum.NodePoolAdmin]
  },
  {
    path: '/npm/pools/edit/:poolId',
    component: PoolEditContainer,
    featureFlag: appFeatureFlags.REACT_APP_FEATURE_NODE_POOL_MANAGEMENT,
    roles: [AppRolesEnum.SuperAdmin, AppRolesEnum.NodePoolAdmin]
  },
  {
    path: '/npm/pools/create',
    component: PoolEditContainer,
    featureFlag: appFeatureFlags.REACT_APP_FEATURE_NODE_POOL_MANAGEMENT,
    roles: [AppRolesEnum.SuperAdmin, AppRolesEnum.NodePoolAdmin]
  },
  {
    path: '/npm/pools/accounts/:poolId',
    component: CloudAccountListContainer,
    featureFlag: appFeatureFlags.REACT_APP_FEATURE_NODE_POOL_MANAGEMENT,
    roles: [AppRolesEnum.SuperAdmin, AppRolesEnum.NodePoolAdmin]
  },
  {
    path: '/npm/pools/accounts/add/:poolId',
    component: AddCloudAccountContainer,
    featureFlag: appFeatureFlags.REACT_APP_FEATURE_NODE_POOL_MANAGEMENT,
    roles: [AppRolesEnum.SuperAdmin, AppRolesEnum.NodePoolAdmin]
  },
  {
    path: '/npm/nodes',
    component: NodeListContainer,
    featureFlag: appFeatureFlags.REACT_APP_FEATURE_NODE_POOL_MANAGEMENT,
    roles: [AppRolesEnum.SuperAdmin, AppRolesEnum.NodePoolAdmin]
  },
  {
    path: '/npm/nodes/:poolId',
    component: NodeListContainer,
    featureFlag: appFeatureFlags.REACT_APP_FEATURE_NODE_POOL_MANAGEMENT,
    roles: [AppRolesEnum.SuperAdmin, AppRolesEnum.NodePoolAdmin]
  },
  {
    path: '/npm/nodes/edit/:nodeName',
    component: NodeEditContainer,
    featureFlag: appFeatureFlags.REACT_APP_FEATURE_NODE_POOL_MANAGEMENT,
    roles: [AppRolesEnum.SuperAdmin, AppRolesEnum.NodePoolAdmin]
  },
  {
    path: '/npm/nodes/statistic/:nodeName',
    component: NodeStatesContainer,
    featureFlag: appFeatureFlags.REACT_APP_FEATURE_NODE_POOL_MANAGEMENT,
    roles: [AppRolesEnum.SuperAdmin, AppRolesEnum.NodePoolAdmin]
  },
  // *********************************************
  // Quota Management
  // *********************************************
  {
    path: '/quotamanagement/services',
    component: QuotaManagementServiceContainer,
    roles: [AppRolesEnum.QuotaAdmin, AppRolesEnum.SuperAdmin]
  },
  {
    path: '/quotamanagement/services/create',
    component: QuotaManagementServiceCreateContainer,
    roles: [AppRolesEnum.QuotaAdmin, AppRolesEnum.SuperAdmin]
  },
  {
    path: '/quotamanagement/services/d/:param/edit',
    component: QuotaManagementServiceEditContainer,
    roles: [AppRolesEnum.QuotaAdmin, AppRolesEnum.SuperAdmin]
  },
  {
    path: '/quotamanagement/services/d/:param/quotas',
    component: QuotaManagementServiceQuotasContainer,
    roles: [AppRolesEnum.QuotaAdmin, AppRolesEnum.SuperAdmin]
  },
  {
    path: '/quotamanagement/services/d/:param/quotas/add',
    component: QuotaManagementServiceAddQuotaContainer,
    roles: [AppRolesEnum.QuotaAdmin, AppRolesEnum.SuperAdmin]
  },
  {
    path: '/quotamanagement/services/d/:param/quotas/edit/:resourceName/:ruleId',
    component: QuotaManagementServiceEditQuotaContainer,
    roles: [AppRolesEnum.QuotaAdmin, AppRolesEnum.SuperAdmin]
  },
  // *********************************************
  // Storage Quota Management
  // *********************************************
  {
    path: '/storagemanagement/usages',
    component: StorageUsagesContainer,
    featureFlag: appFeatureFlags.REACT_APP_FEATURE_STORAGE_MANAGEMENT,
    roles: [AppRolesEnum.StorageAdmin, AppRolesEnum.SuperAdmin]
  },
  {
    path: '/storagemanagement/managequota',
    component: AddQuotaContainer,
    featureFlag: appFeatureFlags.REACT_APP_FEATURE_STORAGE_MANAGEMENT,
    roles: [AppRolesEnum.StorageAdmin, AppRolesEnum.SuperAdmin]
  },
  {
    path: '/storagemanagement/quota',
    component: QuotaAssigmentsContainer,
    featureFlag: appFeatureFlags.REACT_APP_FEATURE_STORAGE_MANAGEMENT,
    roles: [AppRolesEnum.StorageAdmin, AppRolesEnum.SuperAdmin]
  },
  // *********************************************
  // Product Catalog Management
  // *********************************************
  {
    path: '/products/vendors',
    component: ProductVendorsContainer,
    featureFlag: appFeatureFlags.REACT_APP_FEATURE_PRODUCT_CATALOG,
    roles: [AppRolesEnum.SuperAdmin, AppRolesEnum.CatalogAdmin]
  },
  {
    path: '/products/vendors/create',
    component: ProductVendorsCreateContainer,
    featureFlag: appFeatureFlags.REACT_APP_FEATURE_PRODUCT_CATALOG,
    roles: [AppRolesEnum.SuperAdmin, AppRolesEnum.CatalogAdmin]
  },
  {
    path: '/products/vendors/d/:param/edit',
    component: ProductVendorsEditContainer,
    featureFlag: appFeatureFlags.REACT_APP_FEATURE_PRODUCT_CATALOG,
    roles: [AppRolesEnum.SuperAdmin, AppRolesEnum.CatalogAdmin]
  },
  {
    path: '/products/families',
    component: ProductFamiliesContainer,
    featureFlag: appFeatureFlags.REACT_APP_FEATURE_PRODUCT_CATALOG,
    roles: [AppRolesEnum.SuperAdmin, AppRolesEnum.CatalogAdmin]
  },
  {
    path: '/products/families/create',
    component: ProductFamiliesCreateContainer,
    featureFlag: appFeatureFlags.REACT_APP_FEATURE_PRODUCT_CATALOG,
    roles: [AppRolesEnum.SuperAdmin, AppRolesEnum.CatalogAdmin]
  },
  {
    path: '/products/families/d/:param/edit',
    component: ProductFamiliesEditContainer,
    featureFlag: appFeatureFlags.REACT_APP_FEATURE_PRODUCT_CATALOG,
    roles: [AppRolesEnum.SuperAdmin, AppRolesEnum.CatalogAdmin]
  },
  {
    path: '/products',
    component: ProductCatalogContainer,
    featureFlag: appFeatureFlags.REACT_APP_FEATURE_PRODUCT_CATALOG,
    roles: [AppRolesEnum.SuperAdmin, AppRolesEnum.CatalogAdmin]
  },
  {
    path: '/products/d/:param',
    component: ProductCatalogEditContainer,
    featureFlag: appFeatureFlags.REACT_APP_FEATURE_PRODUCT_CATALOG,
    roles: [AppRolesEnum.SuperAdmin, AppRolesEnum.CatalogAdmin]
  },
  {
    path: '/products/d/:param/metadata',
    component: ProductCatalogEditMetaDatasetContainer,
    featureFlag: appFeatureFlags.REACT_APP_FEATURE_PRODUCT_CATALOG,
    roles: [AppRolesEnum.SuperAdmin, AppRolesEnum.CatalogAdmin]
  },
  {
    path: '/products/create',
    component: ProductCatalogCreate,
    featureFlag: appFeatureFlags.REACT_APP_FEATURE_PRODUCT_CATALOG,
    roles: [AppRolesEnum.SuperAdmin, AppRolesEnum.CatalogAdmin]
  },
  {
    path: '/products/create/metadata',
    component: ProductCatalogCreateMetaDatasetContainer,
    featureFlag: appFeatureFlags.REACT_APP_FEATURE_PRODUCT_CATALOG,
    roles: [AppRolesEnum.SuperAdmin, AppRolesEnum.CatalogAdmin]
  },
  {
    path: '/products/approvals',
    component: ProductCatalogApprovalContainer,
    featureFlag: appFeatureFlags.REACT_APP_FEATURE_PRODUCT_CATALOG,
    roles: [AppRolesEnum.SuperAdmin, AppRolesEnum.CatalogAdmin]
  },
  // *********************************************
  // User Account Summary
  // *********************************************
  {
    path: '/usersummary',
    component: UserSummaryContainer,
    roles: [AppRolesEnum.SuperAdmin, AppRolesEnum.ProductAdmin]
  },
  // *********************************************
  // Region Management
  // *********************************************
  {
    path: '/regionmanagement/regions',
    component: RegionManagementContainer,
    featureFlag: appFeatureFlags.REACT_APP_FEATURE_REGION_MANAGEMENT,
    roles: [AppRolesEnum.SuperAdmin, AppRolesEnum.RegionAdmin]
  },
  {
    path: '/regionmanagement/regions/create',
    component: RegionCreateContainer,
    roles: [AppRolesEnum.SuperAdmin],
    featureFlag: appFeatureFlags.REACT_APP_FEATURE_REGION_MANAGEMENT
  },
  {
    path: '/regionmanagement/regions/d/:param/edit',
    component: RegionEditContainer,
    roles: [AppRolesEnum.SuperAdmin],
    featureFlag: appFeatureFlags.REACT_APP_FEATURE_REGION_MANAGEMENT
  },
  {
    path: '/regionmanagement/whitelist',
    component: AccountRegionContainer,
    roles: [AppRolesEnum.SuperAdmin],
    featureFlag: appFeatureFlags.REACT_APP_FEATURE_REGION_MANAGEMENT
  },
  {
    path: '/regionmanagement/whitelist/add',
    component: AddAccountRegionContainer,
    roles: [AppRolesEnum.SuperAdmin],
    featureFlag: appFeatureFlags.REACT_APP_FEATURE_REGION_MANAGEMENT
  },

  // *********************************************
  // NOT FOUND
  // *********************************************
  {
    path: '*',
    component: NotFound
  }
]
