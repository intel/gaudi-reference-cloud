// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import { useLocation } from 'react-router-dom'
import LineGraph from '../../../utils/metrics/LineGraph'
import EmptyView from '../../../utils/emptyView/EmptyView'
import './ChartItem.scss'
import { Button, ButtonGroup } from 'react-bootstrap'
import { BsDownload } from 'react-icons/bs'

const ChartItem = (props: any): JSX.Element => {
  const { pathname } = useLocation()
  const isKubernetesChart =
    pathname.toLowerCase().startsWith('/cluster/') || pathname.toLowerCase().startsWith('/metrics/clusters')

  const metricData = props.metricData
  const title = props.title
  const testId = 'metrics' + String(title.replaceAll(' ', '')) + 'Title'

  const options = [
    {
      id: 'download',
      label: (
        <>
          <BsDownload />
          Download
        </>
      ),
      variant: 'link'
    }
  ]

  // Handles the click action to generate and download a JSON file based on the provided data.
  const onClickAction = (infoName: string, data: any): void => {
    const fileContentArray: any[] = []
    for (const item of data) {
      const fileContent = {
        item: '',
        content: [] as any
      }

      fileContent.item = `${item.label}-${item.item}`
      fileContent.content = [...item.data]

      fileContentArray.push(fileContent)
    }

    const jsonString = JSON.stringify(fileContentArray)

    const element = document.createElement('a')
    const file = new Blob([jsonString], { type: 'text/json' })
    element.href = URL.createObjectURL(file)
    const fileName = `${infoName}.json`
    element.download = fileName
    document.body.appendChild(element)
    element.click()
    document.body.removeChild(element)
  }

  const getChartTitle = (): JSX.Element => {
    if (isKubernetesChart) {
      return (
        <h4 intc-id={testId} className="h6">
          {title}
        </h4>
      )
    } else {
      return (
        <h3 intc-id={testId} className="h6">
          {title}
        </h3>
      )
    }
  }

  return (
    <div className="col-12 col-lg-6 col-xxl-4">
      <div className="d-flex flex-column rounded p-s6 gap-s6 h-100 chartItem">
        {getChartTitle()}
        <>
          {metricData[0]?.isError ? (
            <div className="d-flex flex-column h-100 justify-content-center">
              <EmptyView title="Data unavailable" subTitle={metricData[0]?.errorMessage} />
            </div>
          ) : (
            <>
              <div>
                <LineGraph data={metricData} />
              </div>
              <div className="d-flex justify-content-center">
                <ButtonGroup>
                  {options.map((item: any, index: number) => (
                    <Button
                      intc-id={`btn-graph-option-${item.label}`}
                      data-wap_ref={`btn-graph-option-${item.label}`}
                      aria-label={item.label}
                      key={index}
                      variant={item.variant}
                      onClick={() => {
                        onClickAction(title, metricData)
                      }}
                    >
                      {item.label}
                    </Button>
                  ))}
                </ButtonGroup>
              </div>
            </>
          )}
        </>
      </div>
    </div>
  )
}

export default ChartItem
