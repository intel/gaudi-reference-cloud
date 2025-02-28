import configurator from '@intc/configurator'
import RegionService from '../services/RegionService'

/**
 * Update idcConfig definition and registrations everytime a property is removed or added
 */

/**
 * IDC Configuration variables
 * @typedef {Object} idcConfig
 * ################### GENERAL  ###################
 * @property {String} REACT_APP_ENV Name of environment.
 * @property {String} REACT_APP_COMPANY_LONG_NAME The company long name.
 * ################### API SERVICES  ###################
 * @property {String} REACT_APP_API_GLOBAL_SERVICE URL of proxy for global APIS.
 * @property {String} REACT_APP_API_REGIONAL_SERVICES Array of URL of proxy for regional APIS
 * @property {String} REACT_APP_API_REGIONAL_SERVICE URL of proxy for regional APIS.
 * @property {String} REACT_APP_GUI_DOMAIN The url for IDC Console.
 * @property {String} REACT_APP_API_PREVIEW_SERVICE URL of proxy for Preview APIS.
 * ################### AZURE CONFIGURATION  ###################
 * @property {string} REACT_APP_AZURE_CLIENT_ID AZURE B2C App registration ID.
 * @property {string} REACT_APP_AZURE_CLIENT_API_SCOPE AZURE B2C API Scope.
 * @property {string} REACT_APP_AZURE_CLIENT_AUTHORITY AZURE B2C Authority Domain ID.
 * @property {string} REACT_APP_AZURE_LANDING_PAGE_URL AZURE B2C Logout URL.
 * ################### APP CONFIGURATION  ###################
 * @property {number} REACT_APP_AXIOS_TIMEOUT AXIOS Instance request timeout.
 * @property {number} REACT_APP_ENABLE_MSAL_LOGGING Enable MSAL logging.
 * @property {string} REACT_APP_AGS_GLOBAL_ADMIN AGS ROLE to access application
 * @property {string} REACT_APP_AGS_SRE_ADMIN AGS ROLE to access application.
 * @property {string} REACT_APP_AGS_IKS_ADMIN AGS ROLE to access application.
 * @property {string} REACT_APP_AGS_COMPUTE_ADMIN AGS ROLE to access application.
 * @property {string} REACT_APP_AGS_PRODUCT_ADMIN AGS ROLE to access application.
 * @property {string} REACT_APP_AGS_SLURM_ADMIN AGS ROLE to access application.
 * @property {string} REACT_APP_DEFAULT_REGIONS Array of default regions.
 * @property {string} REACT_APP_SELECTED_REGION Current region.
 * @property {string} REACT_APP_BANNER_AWS_ROLE_ARN aws role arn to update site banner messages file.
 * @property {string} REACT_APP_AWS_REGION S3 bucket region.
 * @property {string} REACT_APP_AWS_S3_BUCKET_NAME S3 bucket name.
 * @property {number} REACT_APP_TOAST_DELAY Default delay for toast messages
 * @property {number} REACT_APP_TOAST_ERROR_DELAY Default delay for toast error messages
 * ################### HELP LINKS CONFIGURATION  ###################
 * @property {string} REACT_APP_SUPPORT_PAGE IDC support page link.
 * @property {string} REACT_APP_KNOWLEDGE_BASE IDC Knowledge base page link.
 * @property {string} REACT_APP_SUBMIT_TICKET IDC submit a support ticket link.
 * ##################### FEATURE FLAGS  ######################
 * @property {number} REACT_APP_FEATURE_BANNER_MANAGEMENT Feature flag for banner management use 0 to turn off and 1 to turn on.
 * @property {number} REACT_APP_FEATURE_NODE_POOL_MANAGEMENT Feature flag for node pool management use 0 to turn off and 1 to turn on.
 * @property {number} REACT_APP_FEATURE_STORAGE_MANAGEMENT Feature flag for storage quotas management use 0 to turn off and 1 to turn on.
 * @property {number} REACT_APP_FEATURE_PRODUCT_CATALOG Feature flag for product catalog management use 0 to turn off and 1 to turn on.
 * @property {number} REACT_APP_FEATURE_REGION_MANAGEMENT Feature flag for regions management use 0 to turn off and 1 to turn on.
 * @property {number} REACT_APP_FEATURE_KUBE_SCORE Feature flag for Iks kube score use 0 to turn off and 1 to turn on.
 * @property {Object} REACT_APP_REGION_BLOCKED_FEATURES Contains the Feature flags that must be turn off for specific regions.
 */

/**
 * Collection of feature flags used by the application, use in conjunction with isFeatureFlagEnable function
 * Feature flag must be defided on idcConfig object as well
 */
export const appFeatureFlags = {
  REACT_APP_FEATURE_BANNER_MANAGEMENT: 'REACT_APP_FEATURE_BANNER_MANAGEMENT',
  REACT_APP_FEATURE_NODE_POOL_MANAGEMENT: 'REACT_APP_FEATURE_NODE_POOL_MANAGEMENT',
  REACT_APP_FEATURE_STORAGE_MANAGEMENT: 'REACT_APP_FEATURE_STORAGE_MANAGEMENT',
  REACT_APP_FEATURE_PRODUCT_CATALOG: 'REACT_APP_FEATURE_PRODUCT_CATALOG',
  REACT_APP_FEATURE_REGION_MANAGEMENT: 'REACT_APP_FEATURE_REGION_MANAGEMENT',
  REACT_APP_FEATURE_KUBE_SCORE: 'REACT_APP_FEATURE_KUBE_SCORE'
}

/**
 * Update all region configuration variables based on default region name
 * @param {idcConfig} currentConfig IDC configuration to update
 */
export const updateCurrentRegion = (idcConfig) => {
  const lastSelectedRegion = RegionService.getLastChangedRegion()
  let region = lastSelectedRegion || idcConfig.REACT_APP_SELECTED_REGION
  let regionIndex = idcConfig.REACT_APP_DEFAULT_REGIONS.findIndex((x) => x === region)
  if (regionIndex === -1) {
    region = idcConfig.REACT_APP_SELECTED_REGION
    regionIndex = idcConfig.REACT_APP_DEFAULT_REGIONS.findIndex((x) => x === region)
  }
  idcConfig.REACT_APP_SELECTED_REGION = region
  idcConfig.REACT_APP_API_REGIONAL_SERVICE = idcConfig.REACT_APP_API_REGIONAL_SERVICES[regionIndex]
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
  REACT_APP_ENV: 'production'
})

registerIdcConfiguration([/^127.0.0.1/, /^localhost/], {
  REACT_APP_ENV: 'Local'
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

export default idcConfig
