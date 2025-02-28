// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

export interface IdcRoute {
  /**
   * Route path
   */
  path: string
  /**
   * Component to render when enter route
   */
  component?: () => JSX.Element | null
  /**
   * Props to the rendered component
   */
  componentProps?: any
  /**
   * Absolute URL to enter when enter route, use to go to external sites
   */
  href?: string
  /**
   * Allowed roles, if undefined all roles can enter the route
   */
  roles?: string[]
  /**
   * Feature flag to enable or disable the route, if undefined route is always enable
   */
  featureFlag?: string
  /**
   * If true members of an account cannot enter the route
   */
  memberNotAllowed?: boolean
  /**
   * If defined the route will be enable if function returns true
   */
  allowedFn?: () => boolean
  /**
   * Name to display on Breadcrum
   */
  breadcrumTitle?: string
  /**
   * If true the breadcrums will be visible at the page in the top
   */
  showBreadcrums?: boolean
}
