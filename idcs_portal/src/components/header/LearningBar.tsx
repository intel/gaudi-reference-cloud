// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import { type ArticleItem } from '../../containers/navigation/Navigation.types'
import ArticleList from '../../utils/articleList/ArticleList'
import { Button } from 'react-bootstrap'

interface LearningBarProps {
  items: ArticleItem[]
  showLearningBar: boolean
  setShowLearningBar: (show: boolean, savePreference: boolean) => void
}

const LearningBar: React.FC<LearningBarProps> = ({ items, showLearningBar, setShowLearningBar }) => {
  return (
    <div
      intc-id="LearningBarNavigationMain"
      className={`learningBar ${showLearningBar ? 'showLearningBar' : ''} offcanvas-sm-width`}
    >
      <div intc-id="LearningBarNavigationHeader" className="w-100 learningBar-header px-s6 py-s6">
        <h2 className="mb-0 h6">Documentation</h2>
        <Button
          variant="close"
          size="sm"
          onClick={() => {
            setShowLearningBar(false, true)
          }}
        ></Button>
      </div>
      <div intc-id="LearningBarNavigationBody" className="offcanvas-body">
        <ArticleList items={items} />
      </div>
    </div>
  )
}

export default LearningBar
