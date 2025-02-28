// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { create } from 'zustand'
import PublicService from '../../services/PublicService'
import type Training from '../models/Training/Training'
import moment from 'moment'
import HappyFace from '../../assets/images/HappyFace.svg'
import Mountains from '../../assets/images/Mountains.svg'
import Robot from '../../assets/images/Robot.svg'
import JupiterLogo from '../../assets/images/JupiterLog.svg'

export interface TrainingStore {
  trainings: Training[] | []
  trainingDetail: Training | null
  loading: boolean
  enrolling: boolean
  refreshRate: boolean
  getTraining: (trainingId: string) => Promise<void>
  setTrainings: (background?: boolean) => Promise<void>
  reset: () => void
}

const getImage = (trainingPicture: string): any => {
  let imageSrc = null

  switch (trainingPicture) {
    case 'txt2img.svg':
      imageSrc = HappyFace
      break
    case 'img2img.svg':
      imageSrc = Mountains
      break
    case 'llm.svg':
      imageSrc = Robot
      break
    default:
      imageSrc = JupiterLogo
      break
  }

  return imageSrc
}

const initialState = {
  trainings: [],
  trainingDetail: null,
  loading: false,
  enrolling: false,
  refreshRate: false
}

const useTrainingStore = create<TrainingStore>()((set, get) => ({
  ...initialState,
  reset: () => {
    set(initialState)
  },
  setTrainings: async (background) => {
    if (!background) {
      set({ loading: true })
    }
    try {
      const response = await PublicService.getTrainingCatalog()

      const trainingDetail = { ...response.data.products }
      const trainings: Training[] = []
      for (const index in trainingDetail) {
        const detail = { ...trainingDetail[index] }
        const training = buildData(detail)
        trainings.push(training)
      }
      set({ loading: false })
      set({ trainings })
    } catch (error) {
      set({ loading: false })
      throw error
    }
  },
  getTraining: async (trainingId: string) => {
    set({ loading: true })
    try {
      const response = await PublicService.getTrainingDetail(trainingId)

      const responseDetail = [...response.data.products]
      if (responseDetail.length > 0) {
        const training = buildData(responseDetail[0])
        set({ trainingDetail: training })
      }
      set({ loading: false })
    } catch (error) {
      set({ loading: false })
      throw error
    }
  }
}))

const buildData = (trainingResponse: any): Training => {
  const metadata = {
    ...trainingResponse.metadata
  }
  const rates = trainingResponse.rates
  let accountType = ''
  let unit = ''
  let rateValue = ''
  let usageExpr = ''

  if (rates.length > 0) {
    const rate = rates[0]

    if (rate) {
      accountType = rate.accountType
      unit = rate.unit
      rateValue = rate.rate
      usageExpr = rate.usageExpr
    }
  }
  const training: Training = {
    name: trainingResponse.name,
    id: trainingResponse.id,
    created: moment(trainingResponse.created).format('MM/DD/YYYY h:mm a'),
    vendorId: trainingResponse.vendorId,
    familyId: trainingResponse.familyId,
    description: trainingResponse.description,
    category: metadata.category,
    gettingStarted: metadata['detail.gettingStarted'],
    audience: metadata['detail.objectives.audience'],
    expectations: metadata['detail.objectives.expectations'],
    overview: metadata['detail.overview'],
    prerrequisites: metadata['detail.prerrequisites'],
    shortDesc: metadata.shortDesc,
    displayCatalogDesc: metadata.displayCatalogDesc,
    displayName: metadata.displayName,
    familyDisplayDescription: metadata['family.displayDescription'],
    familyDisplayName: metadata['family.displayName'],
    launch: metadata.launch,
    region: metadata.region,
    service: metadata.service,
    eccn: trainingResponse.eccn,
    pcq: trainingResponse.pcq,
    matchExpr: trainingResponse.matchExpr,
    accountType,
    unit,
    rate: rateValue,
    usageExpr,
    imageSource: getImage(metadata.displayPicture),
    displayInHomepage: metadata.displayInHomepage === 'true',
    homepageDisplayGroup: metadata.homepageDisplayGroup,
    featuredSoftware: metadata['detail.featuredSoftware'] || null
  }

  return training
}

export default useTrainingStore
