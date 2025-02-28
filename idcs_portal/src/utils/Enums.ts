// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

export const AppRolesEnum = {
  Standard: 'Standard',
  Premium: 'Premium',
  Enterprise: 'Enterprise',
  EnterprisePending: 'EnterprisePending',
  Intel: 'Intel'
}

export const ErrorBoundaryLevel = {
  AppLevel: 'AppLevel',
  RouteLevel: 'RouteLevel'
}

export const EnrollAccountType = {
  standard: 'ACCOUNT_TYPE_STANDARD',
  premium: 'ACCOUNT_TYPE_PREMIUM',
  enterprise_pending: 'ACCOUNT_TYPE_ENTERPRISE_PENDING',
  enterprise: 'ACCOUNT_TYPE_ENTERPRISE',
  intel: 'ACCOUNT_TYPE_INTEL',
  member: 'ACCOUNT_TYPE_MEMBER'
}

export const EnrollActionResponse = {
  ENROLL_ACTION_NONE: 'ENROLL_ACTION_NONE',
  ENROLL_ACTION_REGISTER: 'ENROLL_ACTION_REGISTER',
  ENROLL_ACTION_RETRY: 'ENROLL_ACTION_RETRY',
  ENROLL_ACTION_COUPON_OR_CREDIT_CARD: 'ENROLL_ACTION_COUPON_OR_CREDIT_CARD',
  ENROLL_ACTION_TC: 'ENROLL_ACTION_TC'
}

export const CreditCardsTypes = {
  VisaCreditCard: 'Visa',
  AmexCreditCard: 'American Express',
  DiscoverCreditCard: 'Discover',
  MasterCardCreditCard: 'MasterCard'
}

export const IDCVendorFamilies = {
  Compute: 'compute',
  Network: 'network',
  Storage: 'storage',
  Training: 'training',
  Software: 'software',
  SuperComputer: 'supercomputer',
  UserInterface: 'userinterface',
  Kubernetes: 'kubernetes',
  Dpai: 'dpai',
  MaaS: 'convergedinference',
  payment: 'paymentservices',
  labs: 'labs'
}

export const iksNodeGroupActionsEnum = {
  deleteNodeGroup: 'deleteNodeGroup',
  addNode: 'addNode',
  removeNode: 'removeNode',
  upgradeImage: 'upgradeImage'
}

export const computeCategoriesEnum = {
  singleNode: 'singlenode',
  cluster: 'cluster'
}

export const InvitationState = {
  INVITE_STATE_UNSPECIFIED: '',
  INVITE_STATE_PENDING_ACCEPT: 'Invited',
  INVITE_STATE_ACCEPTED: 'Joined',
  INVITE_STATE_REVOKED: 'Revoked',
  INVITE_STATE_EXPIRED: 'Expired',
  INVITE_STATE_REMOVED: 'Removed',
  INVITE_STATE_REJECTED: 'Rejected'
}

export const InvitationStateSelection = {
  INVITE_STATE_UNSPECIFIED: 'INVITE_STATE_UNSPECIFIED',
  INVITE_STATE_PENDING_ACCEPT: 'INVITE_STATE_PENDING_ACCEPT',
  INVITE_STATE_ACCEPTED: 'INVITE_STATE_ACCEPTED',
  INVITE_STATE_REVOKED: 'INVITE_STATE_REVOKED',
  INVITE_STATE_EXPIRED: 'INVITE_STATE_EXPIRED',
  INVITE_STATE_REMOVED: 'INVITE_STATE_REMOVED',
  INVITE_STATE_REJECTED: 'INVITE_STATE_REJECTED'
}

export const StorageServicesEnum = {
  fileStorage: 'Storage Service - File',
  objectStorage: 'Storage Service - Object'
}

export const objectStorageLifecycleRule = {
  deleteMarker: 'Delete Marker',
  expireDays: 'Expire Days',
  noncurrentExpireDays: 'Non Current Expire Days',
  prefix: 'Prefix'
}

export const objectStorageUsersPermissionPolicies = {
  ReadBucket: 'Read',
  WriteBucket: 'Write',
  DeleteBucket: 'Delete'
}

export const objectStorageUsersPermissionActions = {
  GetBucketLocation: 'Get Bucket Location',
  GetBucketPolicy: 'Get Bucket Policy',
  ListBucket: 'List Bucket',
  ListBucketMultipartUploads: 'List Bucket Multipart Uploads',
  ListMultipartUploadParts: 'List Multipart Upload Parts',
  GetBucketTagging: 'Get Bucket Tagging'
}

export const superComputerProductCatalogTypes = {
  coreCompute: 'CPU',
  aiCompute: 'AI',
  fileStorage: 'FileStorage',
  controlPlane: 'sc-cluster'
}

export const superComputerNodeGroupTypes = {
  aiCompute: 'supercompute-ai',
  gpCompute: 'supercompute-gp'
}

export const iksClusterTypes = {
  iksCluster: 'generalpurpose',
  superCluster: 'supercompute'
}

export const toastMessageEnum = {
  formValidationError: 'Please complete all required and invalid fields.'
}

export const costTimeFactorEnum = {
  minute: 1,
  hour: 60,
  day: 1440,
  week: 10080
}

export const specificErrorMessageEnum = {
  authErrorMessage: ['permission to access the resource is denied', 'permission denied for this resource'],
  noCreditsErrorMessage: 'paid service not allowed',
  capacityErrorMessage: 'insufficient capacity',
  memberExisted: 'memberEmail already exists',
  maxMemberLimitReached: 'member max add limit reached'
}

export const countryCodesForAcceptedCountries = [
  'AT',
  'BE',
  'BG',
  'CY',
  'CZ',
  'DE',
  'DK',
  'EE',
  'ES',
  'FI',
  'FR',
  'GR',
  'HR',
  'HU',
  'IE',
  'IT',
  'LT',
  'LU',
  'LV',
  'MT',
  'NL',
  'PL',
  'PT',
  'RO',
  'SE',
  'SI',
  'SK',
  'US'
]
