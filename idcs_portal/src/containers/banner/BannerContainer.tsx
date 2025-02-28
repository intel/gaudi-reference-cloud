// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React, { useEffect, useState } from 'react'
import { useLocation } from 'react-router-dom'
import useBannerStore, { type BannerDefinition } from '../../store/bannerStore/BannerStore'
import useUserStore from '../../store/userStore/UserStore'
import idcConfig from '../../config/configurator'
import CustomAlerts from '../../utils/customAlerts/CustomAlerts'
import moment from 'moment'
import useAppStore from '../../store/appStore/AppStore'
import { Carousel } from 'react-bootstrap'

const BannerContainer: React.FC = (): JSX.Element => {
  const location = useLocation()

  const bannerList = useBannerStore((state) => state.bannerList)
  const hiddenBannerList = useBannerStore((state) => state.hiddenBannerList)
  const addBanner = useBannerStore((state) => state.addBanner)
  const hideBanner = useBannerStore((state) => state.hideBanner)
  const showLearningBar = useAppStore((state) => state.showLearningBar)
  const learningArticlesAvailable = useAppStore((state) => state.learningArticlesAvailable)
  const showSideNavBar = useAppStore((state) => state.showSideNavBar)

  const isIntelUser = useUserStore((state) => state.isIntelUser)
  const isEnterpriseUser = useUserStore((state) => state.isEnterpriseUser)
  const isEnterprisePendingUser = useUserStore((state) => state.isEnterprisePendingUser)
  const isPremiumUser = useUserStore((state) => state.isPremiumUser)
  const isStandardUser = useUserStore((state) => state.isStandardUser)

  const [activeBanners, setActiveBanners] = useState<BannerDefinition[]>([])
  const [hasScrollBar, setHasScrollBar] = useState(false)

  const filterBannersByUserType = (banners: BannerDefinition[]): BannerDefinition[] => {
    if (isIntelUser()) {
      return banners.filter((x) => x.userTypes.some((x) => x === 'intel' || x === 'all'))
    }
    if (isEnterpriseUser() || isEnterprisePendingUser()) {
      return banners.filter((x) => x.userTypes.some((x) => x === 'enterprise' || x === 'all'))
    }
    if (isPremiumUser()) {
      return banners.filter((x) => x.userTypes.some((x) => x === 'premium' || x === 'all'))
    }

    if (isStandardUser()) {
      return banners.filter((x) => x.userTypes.some((x) => x === 'standard' || x === 'all'))
    }
    return banners
  }

  const filterBannersByRoute = (banners: BannerDefinition[]): BannerDefinition[] => {
    return banners.filter((x) =>
      x.routes.some((x) => window.location.pathname.toLowerCase().startsWith(x?.toLowerCase()) || x === 'all')
    )
  }

  const filterBannersByRegion = (banners: BannerDefinition[]): BannerDefinition[] => {
    return banners.filter((banner) =>
      banner?.regions.some((region) => region === idcConfig.REACT_APP_SELECTED_REGION || region === 'all')
    )
  }

  const filterBannersByExpiration = (banners: BannerDefinition[]): BannerDefinition[] => {
    const filteredBanners: BannerDefinition[] = []

    banners.forEach((x) => {
      if (x.expirationDatetime) {
        const currentDate = moment().valueOf()
        const expirationDate = moment(x.expirationDatetime, 'YYYY-MM-DDTHH:mm:ssZ').valueOf()
        if (currentDate < expirationDate) filteredBanners.push(x)
      } else {
        filteredBanners.push(x)
      }
    })

    return filteredBanners
  }

  const sortBanners = (a: any, b: any): number => {
    // Check for maintenance status first
    if (a.isMaintenance === 'True') return -1
    if (b.isMaintenance === 'True') return 1

    // Handle empty updatedTimestamp by setting them to 0 for comparison
    const timestampA = a.updatedTimestamp || 0
    const timestampB = b.updatedTimestamp || 0

    // Compare timestamps for sorting
    return timestampB - timestampA
  }

  const getActiveBanners = (): BannerDefinition[] => {
    if (!bannerList || !hiddenBannerList || bannerList.length === 0) {
      return []
    }

    let banners = bannerList.filter((x) => !hiddenBannerList.includes(x.id))
    banners = filterBannersByUserType(banners)
    banners = filterBannersByRoute(banners)
    banners = filterBannersByExpiration(banners)
    banners = filterBannersByRegion(banners)

    banners.sort(sortBanners)

    return banners
  }

  const loadBannersFromConfig = (): void => {
    if (!idcConfig.REACT_APP_SITE_BANNERS || idcConfig.REACT_APP_SITE_BANNERS.length === 0) {
      return
    }

    const banners = idcConfig.REACT_APP_SITE_BANNERS.filter((x) => x.status === 'active')

    banners.forEach((banner) => {
      const expirationDatetime = banner.expirationDatetime ? banner.expirationDatetime : null
      addBanner(
        banner.id,
        banner.type,
        banner.title,
        banner.message,
        banner.userTypes,
        banner.routes,
        banner.regions,
        expirationDatetime,
        banner.link,
        banner?.isMaintenance ?? 'False',
        banner?.updatedTimestamp ?? ''
      )
    })
  }

  const loadBanners = (): void => {
    if (!bannerList || !hiddenBannerList || bannerList.length === 0) {
      loadBannersFromConfig()
    }
    setActiveBanners(getActiveBanners())
  }

  const getBannersList = (): JSX.Element => {
    return (
      <Carousel
        controls={false}
        indicators={activeBanners?.length > 1}
        className={`siteBanner ${showLearningBar && learningArticlesAvailable ? 'learningBarMargingEnd' : ''} ${showSideNavBar ? 'sideNavBarMargingStart' : ''} ${hasScrollBar ? 'withScrollbar' : ''}`}
        fade={true}
      >
        {activeBanners.map((banner) => (
          <Carousel.Item key={banner.id}>
            <CustomAlerts
              showAlert={true}
              alertType={banner.type}
              title={banner.title}
              message={banner.message}
              link={banner.link}
              onCloseAlert={
                banner.isMaintenance === 'True'
                  ? undefined
                  : () => {
                      hideBanner(banner.id)
                    }
              }
              showIcon={false}
            />
          </Carousel.Item>
        ))}
      </Carousel>
    )
  }

  useEffect(() => {
    loadBanners()
  }, [location, bannerList, hiddenBannerList])

  useEffect(() => {
    try {
      const rootElement = document.getElementById('root')
      const resizeObserver = new ResizeObserver(() => {
        const scrollHeight = rootElement?.scrollHeight ?? 0
        const clientHeight = rootElement?.clientHeight ?? 0
        const hasScrollBar = scrollHeight > clientHeight
        setHasScrollBar(hasScrollBar)
      })
      resizeObserver.observe(rootElement as Element)
      return () => {
        resizeObserver?.disconnect()
      }
    } catch (error) {
      // No Support for resizeObserver
    }
  }, [])

  return <>{activeBanners && activeBanners.length > 0 ? getBannersList() : null}</>
}

export default BannerContainer
