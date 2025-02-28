// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

/* eslint-disable react/display-name */
import React from 'react'
import { Card } from 'react-bootstrap'
import DoughnutChart from './DoughnutChart'
import { BsArrowRightShort } from 'react-icons/bs'
import { Link } from 'react-router-dom'
import { type ChartSection } from '../../../containers/homePage/secondUseDashboard/ChartsWidgetContainer'
import { useNavigate } from 'react-router'
import { appFeatureFlags, isFeatureFlagEnable } from '../../../config/configurator'
import useAppStore from '../../../store/appStore/AppStore'

interface ChartsWidgetProps {
  isDarkMode: boolean
  isOwnCloudAccount: boolean
  computeSection: ChartSection
  storageSection: ChartSection
  kubernetesSection: ChartSection
  superComputeSection: ChartSection
}

interface ChartsRowProps {
  title: string
  total: number
  running: number
  runningDefinition: string
  url: string
  launchUrl: string
  hasError: boolean
  isDarkMode: boolean
}

const ChartRow: React.FC<ChartsRowProps> = ({
  title,
  total = 0,
  running = 0,
  runningDefinition = 'running',
  url = '',
  launchUrl = '',
  hasError = false,
  isDarkMode = false
}) => {
  // navigate
  const navigate = useNavigate()

  const redirectTo = (route: string): void => {
    navigate(route)
  }

  return (
    <div
      className="dashboard-item d-flex flex-row justify-content-between w-100 p-s3"
      onClick={() => {
        redirectTo(total > 0 ? url : launchUrl)
      }}
    >
      <div className="d-flex flex-column flex-grow-0 justify-content-center pe-s4">
        <div className="d-flex flex-row row-doughnut-chart">
          <DoughnutChart isDarkMode={isDarkMode} total={hasError ? 0 : total} running={hasError ? 0 : running} />
        </div>
      </div>
      <div className="d-flex flex-column flex-shrink-0 justify-content-center">
        <span className="fw-semibold">{title}</span>
        {hasError ? (
          <small>Data unavailable</small>
        ) : (
          <small>{`${total} total, ${running} ${runningDefinition}`}</small>
        )}
      </div>
      <div className="d-flex flex-column flex-fill justify-content-center">
        <div className="d-flex justify-content-end">
          <Link
            to={url}
            intc-id={`link-dashboard-charts-${title}`}
            data-wap_ref={`link-dashboard-charts-${title}`}
            className="btn btn-icon-simple btn-sm"
            aria-label={`${title} link`}
          >
            <BsArrowRightShort />
          </Link>
        </div>
      </div>
    </div>
  )
}

interface ChartColumnProps {
  chartSection: ChartSection
  darkMode: boolean
}

const ComputeCharts: React.FC<ChartColumnProps> = ({ chartSection, darkMode }): JSX.Element => {
  const { instances, instanceGroups, loadBalancers } = chartSection.charts
  return (
    <>
      <h2 className="h6">{chartSection.title}</h2>
      <ChartRow
        title={instances.title}
        total={instances.total}
        running={instances.running}
        runningDefinition={instances.runningDefinition}
        url={instances.url}
        launchUrl={instances.launchUrl}
        hasError={instances.hasError}
        isDarkMode={darkMode}
      />
      <ChartRow
        title={instanceGroups.title}
        total={instanceGroups.total}
        running={instanceGroups.running}
        runningDefinition={instanceGroups.runningDefinition}
        url={instanceGroups.url}
        launchUrl={instanceGroups.launchUrl}
        hasError={instanceGroups.hasError}
        isDarkMode={darkMode}
      />
      {isFeatureFlagEnable(appFeatureFlags.REACT_APP_FEATURE_LOAD_BALANCER) ? (
        <ChartRow
          title={loadBalancers.title}
          total={loadBalancers.total}
          running={loadBalancers.running}
          runningDefinition={loadBalancers.runningDefinition}
          url={loadBalancers.url}
          launchUrl={loadBalancers.launchUrl}
          hasError={loadBalancers.hasError}
          isDarkMode={darkMode}
        />
      ) : null}
    </>
  )
}

const StorageCharts: React.FC<ChartColumnProps> = ({ chartSection, darkMode }): JSX.Element => {
  const { buckets, volumes } = chartSection.charts
  return (
    <>
      <h2 className="h6">{chartSection.title}</h2>
      <ChartRow
        title={buckets.title}
        total={buckets.total}
        running={buckets.running}
        runningDefinition={buckets.runningDefinition}
        url={buckets.url}
        launchUrl={buckets.launchUrl}
        hasError={buckets.hasError}
        isDarkMode={darkMode}
      />
      <ChartRow
        title={volumes.title}
        total={volumes.total}
        running={volumes.running}
        runningDefinition={volumes.runningDefinition}
        url={volumes.url}
        launchUrl={volumes.launchUrl}
        hasError={volumes.hasError}
        isDarkMode={darkMode}
      />
    </>
  )
}

const KubernetesCharts: React.FC<ChartColumnProps> = ({ chartSection, darkMode }): JSX.Element => {
  const { clusters } = chartSection.charts
  return (
    <>
      <h2 className="h6">{chartSection.title}</h2>
      <ChartRow
        title={clusters.title}
        total={clusters.total}
        running={clusters.running}
        runningDefinition={clusters.runningDefinition}
        url={clusters.url}
        launchUrl={clusters.launchUrl}
        hasError={clusters.hasError}
        isDarkMode={darkMode}
      />
    </>
  )
}

const SuperComputerCharts: React.FC<ChartColumnProps> = ({ chartSection, darkMode }): JSX.Element => {
  const { clusters } = chartSection.charts
  return (
    <>
      <h2 className="h6">{chartSection.title}</h2>
      <ChartRow
        title={clusters.title}
        total={clusters.total}
        running={clusters.running}
        runningDefinition={clusters.runningDefinition}
        url={clusters.url}
        launchUrl={clusters.launchUrl}
        hasError={clusters.hasError}
        isDarkMode={darkMode}
      />
    </>
  )
}

const ChartsWidget: React.FC<ChartsWidgetProps> = ({
  isDarkMode,
  isOwnCloudAccount,
  computeSection,
  storageSection,
  kubernetesSection,
  superComputeSection
}): JSX.Element => {
  const computeCharts = <ComputeCharts darkMode={isDarkMode} chartSection={computeSection} />
  const storageCharts = <StorageCharts darkMode={isDarkMode} chartSection={storageSection} />
  const kubernetesCharts = <KubernetesCharts darkMode={isDarkMode} chartSection={kubernetesSection} />
  const superComputerCharts = <SuperComputerCharts darkMode={isDarkMode} chartSection={superComputeSection} />
  const showLearningBar = useAppStore((state) => state.showLearningBar)
  const learningArticlesAvailable = useAppStore((state) => state.learningArticlesAvailable)

  return (
    <Card className="h-100 w-100">
      <Card.Body>
        {/* XLarge View */}
        <div
          className={`d-none ${showLearningBar && learningArticlesAvailable ? '' : 'd-xxl-flex'} flex-row justify-content-between w-100 h-100 p-s2 gap-s5`}
        >
          <div className="d-flex flex-column h-100 gap-s4 flex-fill">{computeCharts}</div>
          <div className="vr" />
          <div className="d-flex flex-column h-100 gap-s4 flex-fill">{storageCharts}</div>
          <div className="vr" />
          <div className="d-flex flex-column h-100 gap-s4 flex-fill">{kubernetesCharts}</div>
          <div className="vr" />
          <div className="d-flex flex-column h-100 gap-s4 flex-fill">{superComputerCharts}</div>
        </div>

        {/* Large View */}
        <div
          className={`d-none d-xl-flex ${showLearningBar && learningArticlesAvailable ? '' : 'd-xxl-none'} flex-column justify-content-between w-100 h-100 p-s4 gap-s6`}
        >
          <div className="d-flex flex-row gap-s5 w-100">
            <div className="d-flex flex-column flex-even h-100 gap-s4">{computeCharts}</div>
            <div className="vr" />
            <div className="d-flex flex-column flex-even h-100 gap-s4">{storageCharts}</div>
            <div className="vr" />
            <div className="d-flex flex-column flex-even h-100 gap-s4">{kubernetesCharts}</div>
          </div>
          <div className="d-flex flex-row gap-s5">
            <div className="d-flex flex-column h-100 gap-s4 flex-even">{superComputerCharts}</div>
            <div className="vr" />
            <div className="d-flex flex-column h-100 gap-s4 flex-even"></div>
          </div>
        </div>

        {/* Small View */}
        <div className="d-none d-sm-flex d-xl-none w-100">
          <div className="d-flex flex-column gap-s6 w-100">
            <div className="d-flex flex-row gap-s5">
              <div className="flex-even">{computeCharts}</div>
              <div className="vr" />
              <div className="flex-even">{storageCharts}</div>
            </div>
            <div className="d-flex flex-row gap-s5">
              <div className="flex-even">{kubernetesCharts}</div>
              <div className="vr" />
              <div className="flex-even">{superComputerCharts}</div>
            </div>
          </div>
        </div>

        {/* XSmall View */}
        <div className="d-none d-xs-flex d-sm-none w-100">
          <div className="d-flex flex-column justify-content-around flex-fill gap-s6">
            <div className="d-flex flex-column gap-s4">{computeCharts}</div>
            <div className="d-flex flex-column gap-s4">{storageCharts}</div>
            <div className="d-flex flex-column gap-s4">{kubernetesCharts}</div>
            <div className="d-flex flex-column gap-s4">{superComputerCharts}</div>
          </div>
        </div>
      </Card.Body>
    </Card>
  )
}

export default ChartsWidget
