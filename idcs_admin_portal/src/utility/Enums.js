export const AzureRolesEnum = {
  GlobalAdmin: 'IDC.Admin',
  SREAdmin: 'IDC.SRE',
  IKSAdmin: 'IDC.IKS',
  ComputeAdmin: 'IDC.Compute',
  ProductAdmin: 'IDC.Product',
  SlurmAdmin: 'IDC.Slurm',
  SuperAdmin: 'IDC.Super.Admin',
  BannerAdmin: 'IDC.Banners',
  StorageAdmin: 'IDC.Storage',
  QuotaAdmin: 'IDC.Quota',
  CatalogAdmin: 'IDC.Catalog',
  RegionAdmin: 'IDC.Region',
  NodePoolAdmin: 'IDC.NodePools'
}

export const AppRolesEnum = {
  GlobalAdmin: 'GlobalAdmin',
  SREAdmin: 'SREAdmin',
  ProductAdmin: 'ProductAdmin',
  SlurmAdmin: 'SlurmAdmin',
  ComputeAdmin: 'ComputeAdmin',
  IKSAdmin: 'IKSAdmin',
  SuperAdmin: 'SuperAdmin',
  BannerAdmin: 'BannerAdmin',
  StorageAdmin: 'StorageAdmin',
  QuotaAdmin: 'QuotaAdmin',
  CatalogAdmin: 'CatalogAdmin',
  RegionAdmin: 'RegionAdmin',
  NodePoolAdmin: 'NodePoolAdmin'
}

export const ErrorBoundaryLevel = {
  AppLevel: 'AppLevel',
  RouteLevel: 'RouteLevel'
}

export const EnrollAccountType = {
  Intel: 'intel',
  Enterprise: 'enterprise',
  Premium: 'premium',
  Standard: 'standard'
}

export const CloudAccountType = {
  standard: 'ACCOUNT_TYPE_STANDARD',
  premium: 'ACCOUNT_TYPE_PREMIUM',
  enterprise_pending: 'ACCOUNT_TYPE_ENTERPRISE_PENDING',
  enterprise: 'ACCOUNT_TYPE_ENTERPRISE',
  intel: 'ACCOUNT_TYPE_INTEL',
  member: 'ACCOUNT_TYPE_MEMBER'
}

export const BannerType = {
  Info: 'info',
  Warning: 'warning'
}

export const AppRoutesEnum = {
  All: 'all',
  Home: '/home',
  'Hardware Catalog': '/hardware',
  'Software Catalog': '/software',
  'Compute Instances': '/compute',
  'Instance Groups': '/compute-groups',
  'Load Balancers': '/load-balancer',
  Keys: '/security/publickeys',
  'Intel K8s Service': '/cluster',
  Supercomputing: '/supercomputer',
  Metrics: '/metrics',
  Billing: '/billing',
  'Upgrade Account': '/upgradeaccount',
  'File Storage': '/storage',
  'Object Storage': '/buckets',
  'Learning Labs': '/learning/labs',
  'Learning Notebooks': '/learning/notebooks',
  Documentation: '/docs'
}

export const IDCVendorFamilies = {
  Compute: 'compute',
  Network: 'network',
  Storage: 'storage',
  Training: 'training',
  Software: 'software',
  SuperComputer: 'supercomputer',
  UserInterface: 'userinterface',
  Kubernetes: 'kubernetes'
}

export const toastMessageEnum = {
  formValidationError: 'Please complete all required and invalid fields.'
}
