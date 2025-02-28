// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React, { useMemo, useState } from 'react'
import { type IdcBreadcrum } from './Navigation.types'
import useUserStore from '../../store/userStore/UserStore'

import { getAllowedRoutes } from '../routing/Routing'
import TopBreadcrums from '../../components/header/TopBreadcrumbs'
import useAppStore from '../../store/appStore/AppStore'

const BreadcrumbContainer: React.FC = () => {
  const [showBreadcrums, setShowBreadcrums] = useState(false)
  const isOwnCloudAccount = useUserStore((state) => state.isOwnCloudAccount)
  const breadcrumCustomTitles = useAppStore((state) => state.breadcrumCustomTitles)
  const currentPath = window.location.pathname

  const items = useMemo(() => {
    const allowedRoutes = getAllowedRoutes(isOwnCloudAccount)
    if (currentPath === '/') {
      return []
    }
    const pathChunks = currentPath.split('/')
    const breadCrums: IdcBreadcrum[] = []
    let latestPath = ''
    let latestPathForSearch = ''
    // eslint-disable-next-line no-labels
    mainloop: for (let i = 1; i < pathChunks.length; i++) {
      const currentChunk = pathChunks[i]
      const searchPath = `${latestPathForSearch}/${currentChunk}`
      const path = `${latestPath}/${currentChunk}`
      let routeMatch = allowedRoutes.find((x) => x.path === searchPath)
      if (routeMatch !== undefined) {
        const breadCrum = {
          title: routeMatch.breadcrumTitle ?? 'No title',
          codePath: searchPath,
          path,
          hide: routeMatch.showBreadcrums === false
        }
        breadCrums.push(breadCrum)
        latestPath = path
        latestPathForSearch = searchPath
        // eslint-disable-next-line no-labels
        continue mainloop
      }
      for (let j = 1; j <= 10; j++) {
        const searchPathWithParam = `${latestPathForSearch}/:param${j > 1 ? j : ''}`
        routeMatch = allowedRoutes.find((x) => x.path === searchPathWithParam)
        if (routeMatch !== undefined) {
          let title: string | undefined = ''
          if (breadcrumCustomTitles.get(path)) {
            title = breadcrumCustomTitles.get(path)
          } else if (routeMatch.breadcrumTitle) {
            title = routeMatch.breadcrumTitle
          } else {
            title = currentChunk
          }
          const breadCrum = {
            title: title ?? 'No title',
            codePath: searchPathWithParam,
            path,
            hide: routeMatch.showBreadcrums === false
          }
          breadCrums.push(breadCrum)
          latestPath = path
          latestPathForSearch = searchPathWithParam
          // eslint-disable-next-line no-labels
          continue mainloop
        }
      }
      latestPath = path
      latestPathForSearch = searchPath
    }
    const lastBreadcrum = breadCrums.length > 0 ? breadCrums[breadCrums.length - 1] : null
    const shouldShowBreadcrum =
      lastBreadcrum !== null && allowedRoutes.some((x) => x.path === lastBreadcrum.codePath && x.showBreadcrums)
    setShowBreadcrums(shouldShowBreadcrum)
    return breadCrums.filter((x) => !x.hide)
  }, [isOwnCloudAccount, window.location.pathname, breadcrumCustomTitles])

  return <TopBreadcrums currentPath={currentPath} items={items} shouldShowBreadcrum={showBreadcrums} />
}

export default BreadcrumbContainer
