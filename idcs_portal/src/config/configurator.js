// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import configurator from '@intc/configurator'
import AppSettingsService from '../services/AppSettingsService'

/**
 * Update idcConfig definition and registrations everytime a property is removed or added
 */

/**
 * IDC Configuration variables
 * @typedef {Object} idcConfig
 * ################### GENERAL  ###################
 * @property {String} REACT_APP_ENV Name of environment.
 * @property {String} REACT_APP_CONSOLE_LONG_NAME The commercial long name of the console.
 * @property {String} REACT_APP_CONSOLE_SHORT_NAME The commercial short name of the console.
 * @property {String} REACT_APP_COMPANY_LONG_NAME The company long name.
 * @property {String} REACT_APP_COMPANY_SHORT_NAME The company short name.
 * @property {String} REACT_APP_GUI_DOMAIN The url for IDC Console.
 * @property {String} REACT_APP_GUI_BETA_DOMAIN The url for IDC Beta Console.
 * @property {string} REACT_APP_CLOUD_CONNECT_URL URL for Cloud Connect.
 * ################### API SERVICES  ###################
 * @property {String} REACT_APP_API_GLOBAL_SERVICE URL of proxy for global APIS.
 * @property {String} REACT_APP_API_REGIONAL_SERVICES Array of URL of proxy for regional APIS.
 * @property {String} REACT_APP_API_REGIONAL_SERVICE URL of proxy for regional APIS.
 * @property {String} REACT_APP_API_SLURM_SERVICE URL of proxy for SLURM API.
 * @property {String} REACT_APP_API_LEARNING_LABS_SERVICE URL of proxy for Learning labs APIS.
 * ################### AZURE CONFIGURATION  ###################
 * @property {string} REACT_APP_AZURE_CLIENT_ID AZURE B2C App registration ID.
 * @property {string} REACT_APP_AZURE_CLIENT_API_SCOPE AZURE B2C API Scope.
 * @property {string} REACT_APP_AZURE_B2C_UNIFIED_FLOW AZURE B2C Unified Flow Name.
 * @property {string} REACT_APP_AZURE_B2C_SIGNIN_SIGNUP_AUTHORITY AZURE B2C Authority URL.
 * @property {string} REACT_APP_AZURE_B2C_AUTHORITY_DOMAIN AZURE B2C Authority Domain ID.
 * @property {string} REACT_APP_AZURE_LANDING_PAGE_URL AZURE B2C Logout URL.
 * @property {string} REACT_APP_AZURE_STANDARD_ENROLL_URL AZURE B2C Standard registration URL.
 * @property {string} REACT_APP_AZURE_PREMIUM_ENROLL_URL AZURE B2C Premium registration URL.
 * @property {string} REACT_APP_AZURE_ENTERPRISE_ENROLL_URL AZURE B2C Enterprise registration URL.
 * ################### APP CONFIGURATION  ###################
 * @property {number} REACT_APP_AXIOS_TIMEOUT AXIOS Instance request timeout.
 * @property {number} REACT_APP_ENABLE_MSAL_LOGGING Enable MSAL logging.
 * @property {number} REACT_APP_NOTIFICATIONS_HEARTBEAT Notifications request rate
 * @property {number} REACT_APP_TOAST_DELAY Default delay for toast messages
 * @property {number} REACT_APP_TOAST_ERROR_DELAY Default delay for toast error messages
 * @property {Array<any>} REACT_APP_SITE_BANNERS Contains the banners for all regions.
 * @property {Array<any>} REACT_APP_LEARNING_DOCS Contains the doc configuration for learning panel
 * ################### APP CLOUD ACCOUNT CONFIGURATION  ###################
 * @property {string} REACT_APP_SELECTED_CLOUD_ACCOUNT Default Cloud Account.
 * ################### APP REGIONS CONFIGURATION  ###################
 * @property {string} REACT_APP_DEFAULT_REGION_NAMES Array of default region names.
 * @property {string} REACT_APP_DEFAULT_REGION_NAME Default region name.
 * @property {string} REACT_APP_DEFAULT_REGION_AVAILABILITY_ZONES Array of default region availability zones.
 * @property {string} REACT_APP_DEFAULT_REGION_AVAILABILITY_ZONE Default region availability zone.
 * @property {string} REACT_APP_DEFAULT_REGION_PREFIXES Array of default region prefixes.
 * @property {string} REACT_APP_DEFAULT_REGION_PREFIX Default region prefix.
 * @property {Array<string>} REACT_APP_DEFAULT_REGIONS Array of default regions.
 * @property {string} REACT_APP_DEFAULT_REGION Default region.
 * @property {string} REACT_APP_SELECTED_REGION Current region.
 * @property {Object} REACT_APP_REGION_BLOCKED_FEATURES Contains the Feature flags that must be turn off for specific regions.
 * ################### HELP LINKS CONFIGURATION  ###################
 * @property {string} REACT_APP_SUPPORT_PAGE IDC support page link.
 * @property {string} REACT_APP_SUBMIT_TICKET IDC submit a support ticket link for standard users.
 * @property {string} REACT_APP_SUBMIT_TICKET_ENTERPRISE IDC submit a support ticket link for enterprise users.
 * @property {string} REACT_APP_SUBMIT_TICKET_PREMIUM IDC submit a support ticket link for premium users.
 * @property {string} REACT_APP_PUBLIC_DOCUMENTATION IDC public documentation link.
 * @property {string} REACT_APP_PUBLIC_FEEDBACK_URL public Feedback Link
 * @property {string} REACT_APP_GETTING_STARTED_URL IDC public documentation link.
 * @property {string} REACT_APP_TUTORIALS_URL IDC public documentation link.
 * @property {string} REACT_APP_WHATSNEW_URL IDC public documentation link.
 * @property {string} REACT_APP_KUBERNETES_RELEASE_URL kubernetes release documentation link.
 * @property {string} REACT_APP_COMMUNITY IDC Community page link.
 * @property {string} REACT_APP_KNOWLEDGE_BASE IDC public documentation
 * @property {string} REACT_APP_SHH_KEYS IDC public documentation link.
 * @property {string} REACT_APP_INSTANCE_SPEC IDC public documentation link.
 * @property {string} REACT_APP_INSTANCE_CONNECT IDC public documentation link.
 * @property {string} REACT_APP_CLUSTER_HOW_TO_CONNECT IDC public documentation link.
 * @property {string} REACT_APP_CLUSTER_GUIDE IDC public cluster documentation link.
 * @property {string} REACT_APP_MULTIUSER_GUIDE IDC multi user documentation link.
 * @property {string} REACT_APP_LEARNING_LABS_DISCLAIMER IDC Learning labs disclaimer link.
 * @property {string} REACT_APP_SERVICE_AGREEMENT_URL IDC Service Agreement link.
 * @property {string} REACT_APP_SOFTWARE_AGREEMENT_URL IDC Software and Services Terms and Conditions link.
 * @property {string} REACT_APP_GUIDES_STORAGE_OVERVIEW_URL IDC public storage overview link.
 * @property {string} REACT_APP_GUIDES_STORAGE_FILE_URL IDC public storage file link.
 * @property {string} REACT_APP_GUIDES_OBJECT_STORAGE_URL IDC public object storage link.
 * ################### ARIA CONFIGURATION  ###################
 * @property {string} REACT_APP_ARIA_DIRECT_POST_CLIENT_NO Client No for Intel ARIA instance.
 * ################### FEATURE FLAGS  ###################
 * @property {number} REACT_APP_FEATURE_NOTIFICATIONS Feature flag for notifications use 0 to turn off and 1 to turn on.
 * @property {number} REACT_APP_FEATURE_KaaS Feature flag for kubernetes use 0 to turn off and 1 to turn on.
 * @property {number} REACT_APP_FEATURE_IKS_STORAGE Feature flag for kubernetes Storage use 0 to turn off and 1 to turn on.
 * @property {number} REACT_APP_FEATURE_IKS_SECURITY Feature flag for kubernetes security use 0 to turn off and 1 to turn on.
 * @property {number} REACT_APP_FEATURE_API_KEYS Feature flag for API Keys page use 0 to turn off and 1 to turn on.
 * @property {number} REACT_APP_FEATURE_UPGRADE_TO_PREMIUM Feature flag for upgrade standard to premium use 0 to turn off and 1 to turn on.
 * @property {number} REACT_APP_FEATURE_STORAGE Feature flag for storage use 0 to turn off and 1 to turn on.
 * @property {number} REACT_APP_FEATURE_STORAGE_EDIT Feature flag for storage use 0 to turn off and 1 to turn on.
 * @property {number} REACT_APP_FEATURE_SOFTWARE Feature flag for software use 0 to turn off and 1 to turn on.
 * @property {number} REACT_APP_FEATURE_TRAINING Feature flag for training use 0 to turn off and 1 to turn on.
 * @property {number} REACT_APP_FEATURE_UPGRADE_TO_ENTERPRISE Feature flag for upgrade premium to enterprise use 0 to turn off and 1 to turn on.
 * @property {number} REACT_APP_FEATURE_DIRECTPOST Feature flag for direct post use 0 to turn off and 1 to turn on.
 * @property {number} REACT_APP_FEATURE_UX_WHITELIST Feature flag for UX whitelist use 0 to turn off and 1 to turn on.
 * @property {number} REACT_APP_FEATURE_UPGRADE_PREMIUM_COUPON_CODE Feature flag for upgrade to premium coupon code. use 0 to turn off and 1 to turn on.
 * @property {number} REACT_APP_FEATURE_MULTIUSER Feature flag for multiuser support. use 0 to turn off and 1 to turn on.
 * @property {number} REACT_APP_FEATURE_OBJECT_STORAGE Feature flag for object storage use 0 or not include to turn off and 1 to turn on.
 * @property {number} REACT_APP_FEATURE_FEEDBACK Feature flag for Feedback menu use 0 or not include to turn off and 1 to turn on.
 * @property {number} REACT_APP_FEATURE_LOAD_BALANCER Feature flag for Load Balancer module use 0 or not include to turn off and 1 to turn on.
 * @property {number} REACT_APP_FEATURE_SUPER_COMPUTER Feature flag for Super Computer module use 0 or not include to turn off and 1 to turn on.
 * @property {number} REACT_APP_FEATURE_SC_SECURITY Feature flag for Super computer security use 0 to turn off and 1 to turn on.
 * @property {number} REACT_APP_FEATURE_LEARNING_LABS Feature flag for Learning labs module use 0 or not include to turn off and 1 to turn on.
 * @property {number} REACT_APP_FEATURE_NAVBAR_SEARCH Feature flag for Navbar Search use 0 to turn off and 1 to turn on.
 * @property {number} REACT_APP_FEATURE_METRICS Feature flag for metrics Graphs use 0 to turn off and 1 to turn on.
 * @property {number} REACT_APP_FEATURE_COMPUTE_EDIT_LABELS Feature flag for Compute Edit Labels. Use 0 to turn off and 1 to turn on.
 * @property {number} REACT_APP_FEATURE_COMPUTE_SHOW_LABELS Feature flag for Compute Show Labels tab. Use 0 to turn off and 1 to turn on.
 * @property {number} REACT_APP_FEATURE_USER_ROLES Feature flag for account roles use 0 to turn off and 1 to turn on.
 * @property {number} REACT_APP_FEATURE_USER_CREDENTIALS Feature flag for user credentials and 1 to turn on.
 * @property {number} REACT_APP_FEATURE_USER_ROLE_EDIT Feature flag for account role edit. Use 0 to turn off and 1 to turn on.
 * @property {number} REACT_APP_FEATURE_STORAGE_VAST Feature flag for storage vast use 0 to turn off and 1 to turn on.
 * @property {number} REACT_APP_FEATURE_METRICS_BARE_METAL  Feature flag for Bare Metal metrics. Use 0 to turn off and 1 to turn on.
 * @property {number} REACT_APP_FEATURE_METRICS_CLUSTER  Feature flag for cluster metrics. Use 0 to turn off and 1 to turn on.
 * @property {number} REACT_APP_FEATURE_METRICS_GROUPS  Feature flag for cluster metrics. Use 0 to turn off and 1 to turn on.
 */

/**
 * Collection of feature flags used by the application, use in conjunction with isFeatureFlagEnable function
 * Feature flag must be defined on idcConfig object as well
 */
export const appFeatureFlags = {
  REACT_APP_FEATURE_NOTIFICATIONS: 'REACT_APP_FEATURE_NOTIFICATIONS',
  REACT_APP_FEATURE_KaaS: 'REACT_APP_FEATURE_KaaS',
  REACT_APP_FEATURE_IKS_STORAGE: 'REACT_APP_FEATURE_IKS_STORAGE',
  REACT_APP_FEATURE_IKS_SECURITY: 'REACT_APP_FEATURE_IKS_SECURITY',
  REACT_APP_FEATURE_API_KEYS: 'REACT_APP_FEATURE_API_KEYS',
  REACT_APP_FEATURE_USER_CREDENTIALS: 'REACT_APP_FEATURE_USER_CREDENTIALS',
  REACT_APP_FEATURE_UPGRADE_TO_PREMIUM: 'REACT_APP_FEATURE_UPGRADE_TO_PREMIUM',
  REACT_APP_FEATURE_STORAGE: 'REACT_APP_FEATURE_STORAGE',
  REACT_APP_FEATURE_STORAGE_EDIT: 'REACT_APP_FEATURE_STORAGE_EDIT',
  REACT_APP_FEATURE_SOFTWARE: 'REACT_APP_FEATURE_SOFTWARE',
  REACT_APP_FEATURE_TRAINING: 'REACT_APP_FEATURE_TRAINING',
  REACT_APP_FEATURE_UPGRADE_TO_ENTERPRISE: 'REACT_APP_FEATURE_UPGRADE_TO_ENTERPRISE',
  REACT_APP_FEATURE_DIRECTPOST: 'REACT_APP_FEATURE_DIRECTPOST',
  REACT_APP_FEATURE_UX_WHITELIST: 'REACT_APP_FEATURE_UX_WHITELIST',
  REACT_APP_FEATURE_UPGRADE_PREMIUM_COUPON_CODE: 'REACT_APP_FEATURE_UPGRADE_PREMIUM_COUPON_CODE',
  REACT_APP_FEATURE_MULTIUSER: 'REACT_APP_FEATURE_MULTIUSER',
  REACT_APP_FEATURE_OBJECT_STORAGE: 'REACT_APP_FEATURE_OBJECT_STORAGE',
  REACT_APP_FEATURE_FEEDBACK: 'REACT_APP_FEATURE_FEEDBACK',
  REACT_APP_FEATURE_LOAD_BALANCER: 'REACT_APP_FEATURE_LOAD_BALANCER',
  REACT_APP_FEATURE_SUPER_COMPUTER: 'REACT_APP_FEATURE_SUPER_COMPUTER',
  REACT_APP_FEATURE_SC_SECURITY: 'REACT_APP_FEATURE_SC_SECURITY',
  REACT_APP_FEATURE_LEARNING_LABS: 'REACT_APP_FEATURE_LEARNING_LABS',
  REACT_APP_FEATURE_NAVBAR_SEARCH: 'REACT_APP_FEATURE_NAVBAR_SEARCH',
  REACT_APP_FEATURE_METRICS: 'REACT_APP_FEATURE_METRICS',
  REACT_APP_FEATURE_QUICK_CONNECT: 'REACT_APP_FEATURE_QUICK_CONNECT',
  REACT_APP_FEATURE_COMPUTE_EDIT_LABELS: 'REACT_APP_FEATURE_COMPUTE_EDIT_LABELS',
  REACT_APP_FEATURE_COMPUTE_SHOW_LABELS: 'REACT_APP_FEATURE_COMPUTE_SHOW_LABELS',
  REACT_APP_FEATURE_USER_ROLES: 'REACT_APP_FEATURE_USER_ROLES',
  REACT_APP_FEATURE_USER_ROLE_EDIT: 'REACT_APP_FEATURE_USER_ROLE_EDIT',
  REACT_APP_FEATURE_STORAGE_VAST: 'REACT_APP_FEATURE_STORAGE_VAST',
  REACT_APP_FEATURE_METRICS_BARE_METAL: 'REACT_APP_FEATURE_METRICS_BARE_METAL',
  REACT_APP_FEATURE_METRICS_CLUSTER: 'REACT_APP_FEATURE_METRICS_CLUSTER',
  REACT_APP_FEATURE_METRICS_GROUPS: 'REACT_APP_FEATURE_METRICS_GROUPS',
  REACT_APP_FEATURE_IKS_KUBE_CONFIG: 'REACT_APP_FEATURE_IKS_KUBE_CONFIG'
}
/**
 * Update all region configuration variables based on default region name
 * @param {idcConfig} idcConfig IDC configuration to update
 */
export const updateCurrentRegion = (idcConfig) => {
  const urlParams = new URLSearchParams(window.location.search)
  const lastSelectedRegion = urlParams.get('region')
  let region = lastSelectedRegion || idcConfig.REACT_APP_SELECTED_REGION
  let regionIndex = idcConfig.REACT_APP_DEFAULT_REGIONS.findIndex((x) => x === region)
  if (regionIndex === -1) {
    region = idcConfig.REACT_APP_SELECTED_REGION
    regionIndex = idcConfig.REACT_APP_DEFAULT_REGIONS.findIndex((x) => x === region)
  }
  idcConfig.REACT_APP_SELECTED_REGION = region
  idcConfig.REACT_APP_API_REGIONAL_SERVICE = idcConfig.REACT_APP_API_REGIONAL_SERVICES[regionIndex]
  idcConfig.REACT_APP_DEFAULT_REGION_NAME = idcConfig.REACT_APP_DEFAULT_REGION_NAMES[regionIndex]
  idcConfig.REACT_APP_DEFAULT_REGION_AVAILABILITY_ZONE =
    idcConfig.REACT_APP_DEFAULT_REGION_AVAILABILITY_ZONES[regionIndex]
  idcConfig.REACT_APP_DEFAULT_REGION_PREFIX = idcConfig.REACT_APP_DEFAULT_REGION_PREFIXES[regionIndex]
  idcConfig.REACT_APP_DEFAULT_REGION = idcConfig.REACT_APP_DEFAULT_REGIONS[regionIndex]
}

/**
 * Update default cloud account
 * @param {idcConfig} idcConfig IDC configuration to update
 */
export const setDefaultCloudAccount = (idcConfig) => {
  const defaultCloudAccount = AppSettingsService.getDefaultCloudAccount()
  idcConfig.REACT_APP_SELECTED_CLOUD_ACCOUNT = defaultCloudAccount
}

/**
 * Includes a new configuration
 * @param {Array} hosts An array of strings indicating the applicable hosts for the configuration
 * @param {idcConfig} newConfig New IDC configuration to register
 */
const registerIdcConfiguration = (hosts, newConfig) => {
  if (hosts !== null) {
    configurator.register({ ...newConfig, HOST: hosts })
  } else {
    configurator.register(newConfig)
  }
}

registerIdcConfiguration(null, {
  ...window._env_,
  REACT_APP_ENV: 'production',
  REACT_APP_FEATURE_DIRECTPOST: 0
})

registerIdcConfiguration([/^127.0.0.1/, /^localhost/], {
  REACT_APP_ENV: 'Local',
  REACT_APP_FEATURE_DIRECTPOST: 0
})

/**
 * Verify if feature flag is enabled
 * @param {string} featureFlagKey Use object appFeatureFlags to pass desired feature flag
 */
export const isFeatureFlagEnable = (featureFlagKey) => {
  try {
    const featureFlagValue = idcConfig[featureFlagKey]
    const isRestrictedByRegion =
      idcConfig.REACT_APP_REGION_BLOCKED_FEATURES &&
      idcConfig.REACT_APP_REGION_BLOCKED_FEATURES[featureFlagKey] &&
      idcConfig.REACT_APP_REGION_BLOCKED_FEATURES[featureFlagKey].some((x) => x === idcConfig.REACT_APP_SELECTED_REGION)
    return !isRestrictedByRegion && Boolean(Number(featureFlagValue))
  } catch {
    console.error('An error ocurred when validating if feature is enabled')
    return false
  }
}

/**
 * Verify if feature flag is enabled but blocked in current region
 * @param {string} featureFlagKey Use object appFeatureFlags to pass desired feature flag
 */
export const isFeatureRegionBlocked = (featureFlagKey) => {
  try {
    const featureFlagValue = idcConfig[featureFlagKey]
    const isRestrictedByRegion =
      idcConfig.REACT_APP_REGION_BLOCKED_FEATURES &&
      idcConfig.REACT_APP_REGION_BLOCKED_FEATURES[featureFlagKey] &&
      idcConfig.REACT_APP_REGION_BLOCKED_FEATURES[featureFlagKey].some((x) => x === idcConfig.REACT_APP_SELECTED_REGION)
    return isRestrictedByRegion && Boolean(Number(featureFlagValue))
  } catch {
    console.error('An error ocurred when validating if feature is enabled')
    return false
  }
}

/**
 * Return list of regions in which featureflag is available
 * @param {string} featureFlagKey Use object appFeatureFlags to pass desired feature flag
 */
export const getFeatureAvailableRegions = (featureFlagKey) => {
  const regionsBlocked = idcConfig.REACT_APP_REGION_BLOCKED_FEATURES[featureFlagKey]
  if (!regionsBlocked) {
    return []
  }
  return idcConfig.REACT_APP_DEFAULT_REGIONS?.filter((x) => !regionsBlocked.some((y) => y === x))
}

/**
 * @type {idcConfig}
 */
const idcConfig = configurator.config

updateCurrentRegion(idcConfig)
setDefaultCloudAccount(idcConfig)

export default idcConfig
