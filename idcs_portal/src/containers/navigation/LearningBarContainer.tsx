// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React, { useEffect, useState } from 'react'
import useAppStore from '../../store/appStore/AppStore'
import { type ArticleItem } from './Navigation.types'
import LearningBar from '../../components/header/LearningBar'
import idcConfig from '../../config/configurator'
import { useLocation } from 'react-router-dom'

type CategoryRoutes = Record<string, string[]>

const LearningBarContainer: React.FC = () => {
  const trainingItems: ArticleItem[] = idcConfig.REACT_APP_LEARNING_DOCS
    ? idcConfig.REACT_APP_LEARNING_DOCS.map((item) => ({
        ...item,
        openInNewPage: true
      }))
    : []

  const showLearningBar = useAppStore((state) => state.showLearningBar)
  const setShowLearningBar = useAppStore((state) => state.setShowLearningBar)
  const learningArticlesAvailable = useAppStore((state) => state.learningArticlesAvailable)
  const setLearningArticlesAvailable = useAppStore((state) => state.setLearningArticlesAvailable)
  const [activeTrainingItems, setActiveTrainingItems] = useState<ArticleItem[]>([...trainingItems])
  const location = useLocation()

  const routes: CategoryRoutes = {
    accounts: ['/profile/accountsettings', '/profile/accountAccessManagement', '/profile/roles'],
    credentials: ['/profile/credentials'],
    billing: ['/billing'],
    catalog: ['/hardware', '/software'],
    software: ['/software'],
    compute: ['/compute', '/compute-groups', '/security'],
    'load balancer': ['/load-balancer'],
    documentation: [],
    'cloud monitor': ['/metrics'],
    'intel K8S service': ['/cluster'],
    storage: ['/storage', '/buckets'],
    'super computing': ['/supercomputer'],
    learning: ['/learning/notebooks']
  }

  const mapCategoryToPaths = (categories: string[]): string[] => {
    const result: string[] = []

    categories.forEach((category) => {
      if (routes[category]) {
        result.push(...routes[category])
      }
    })

    return result
  }

  const filterTrainingItems = (): void => {
    const filteredTrainingItems = trainingItems.filter((item) =>
      item.category.includes('dashboard')
        ? location.pathname.toLowerCase() === '/home'
        : mapCategoryToPaths(item.category).some((path) =>
            location.pathname.toLowerCase().startsWith(path?.toLowerCase())
          )
    )
    filteredTrainingItems.sort((a, b) => b.rating - a.rating) // Sort descending by rating, higher rate shows first
    const learningArticlesAvailable = filteredTrainingItems.length > 0
    setLearningArticlesAvailable(learningArticlesAvailable)
    setActiveTrainingItems(filteredTrainingItems)
  }

  useEffect(() => {
    filterTrainingItems()
  }, [location])

  return (
    <LearningBar
      items={activeTrainingItems}
      setShowLearningBar={setShowLearningBar}
      showLearningBar={learningArticlesAvailable && showLearningBar}
    />
  )
}

export default LearningBarContainer
