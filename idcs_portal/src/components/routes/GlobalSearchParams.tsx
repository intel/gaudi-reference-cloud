// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React, { useEffect, useState } from 'react'
import { useLocation, useSearchParams } from 'react-router-dom'
import idcConfig from '../../config/configurator'
import useUserStore from '../../store/userStore/UserStore'
import { getAllowedRoutes } from '../../containers/routing/Routing'
import useAppStore, { type VisitedSite } from '../../store/appStore/AppStore'
import { getAllowedNavigations } from '../../containers/navigation/Navigation'
import { type IdcNavigation } from '../../containers/navigation/Navigation.types'

const GlobalSearchParams = (): JSX.Element => {
  const { pathname, search } = useLocation()
  const [searchParams, setSearchParams] = useSearchParams()
  const [currentRegion, setCurrentRegion] = useState(searchParams.get('region'))

  const isOwnCloudAccount = useUserStore((state) => state.isOwnCloudAccount)
  const setVisitedSites = useAppStore((state) => state.setVisitedSites)
  const visitedSites = useAppStore((state) => state.visitedSites)

  const allowedRoutes = getAllowedRoutes(isOwnCloudAccount)

  useEffect(() => {
    const localStorageData: any = localStorage.getItem('recentlyVisited') ?? null
    let localStorageSites = JSON.parse(localStorageData) ?? []

    // Security filter
    if (localStorageSites.length > 0) {
      localStorageSites = localStorageSites.filter((site: any) =>
        allowedRoutes.some((route) => route.path === site.path)
      )
    }
    setVisitedSites(localStorageSites)
  }, [])

  useEffect(() => {
    const visitedSite = collectVisitedSite()
    if (visitedSite) {
      const sites = visitedSites.filter((item) => item.title !== visitedSite.title)
      if (sites.length > 5) sites.pop()
      sites.unshift(visitedSite)
      setVisitedSites(sites)
    }
  }, [pathname])

  useEffect(() => {
    if (currentRegion !== idcConfig.REACT_APP_SELECTED_REGION || !searchParams.get('region')) {
      setCurrentRegion(idcConfig.REACT_APP_SELECTED_REGION)
      setSearchParams(
        (params) => {
          params.set('region', idcConfig.REACT_APP_SELECTED_REGION)
          return params
        },
        { replace: true }
      )
    }
  }, [search, idcConfig.REACT_APP_SELECTED_REGION])

  function isChildrenOf(currentPath: string, pathname: string, children?: IdcNavigation[]): boolean {
    const paramWildcard = '/:param'
    return (
      currentPath.toLowerCase() === pathname.toLowerCase() ||
      (!children?.some((x) => currentPath.toLowerCase() === x.path) &&
        currentPath.toLowerCase().startsWith(`${pathname.toLowerCase()}/`)) ||
      (!children?.some((x) => currentPath.toLowerCase() === x.path) &&
        pathname.includes(paramWildcard) &&
        currentPath.toLowerCase().startsWith(`${pathname.split(paramWildcard)[0].toLowerCase()}/`))
    )
  }

  function collectVisitedSite(): VisitedSite | null {
    if (pathname === '/' || pathname.startsWith('/home')) return null
    const navigationArr = getAllowedNavigations(isOwnCloudAccount)
    const pathChunks = pathname.split('/')
    let latestPathForSearch = ''
    let visitedSite: VisitedSite | null = null
    for (let i = 1; i < pathChunks.length; i++) {
      const currentChunk = pathChunks[i]
      const searchPath = `${latestPathForSearch}/${currentChunk}`
      const searchPathWithParam = `${latestPathForSearch}/:param`
      let routeMatch = allowedRoutes.find((x) => x.path === searchPath)
      if (routeMatch !== undefined) {
        if (routeMatch.recentlyVisitedTitle) {
          const parentNavigation = navigationArr.find(
            (x) =>
              x.path?.toLowerCase() === routeMatch?.path.toLowerCase() ||
              x.children?.some((y) => isChildrenOf(searchPath, y.path, x.children))
          )
          const service = parentNavigation?.toolbarTitle ?? parentNavigation?.name ?? undefined
          const title = service ? `${service} - ${routeMatch.recentlyVisitedTitle}` : routeMatch.recentlyVisitedTitle
          visitedSite = {
            title,
            path: routeMatch.path
          }
        }
      }
      routeMatch = allowedRoutes.find((x) => x.path === searchPathWithParam)
      // Only apply for general service levels for current usecase
      if (routeMatch !== undefined) break

      latestPathForSearch = searchPath
    }
    return visitedSite
  }

  return <></>
}
export default GlobalSearchParams
