// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { type IconType } from 'react-icons'
import { ReactComponent as Catalog } from '../../assets/images/custom-catalog.svg'
import { ReactComponent as K8s } from '../../assets/images/custom-k8.svg'
import { BsBoxArrowUpRight, BsBook, BsMortarboard } from 'react-icons/bs'
import idcConfig from '../../config/configurator'

interface GetStartedCardResourceAction {
  label: string
  leftIcon?: IconType | React.FunctionComponent<React.SVGProps<SVGSVGElement>>
  rigthIcon?: IconType | React.FunctionComponent<React.SVGProps<SVGSVGElement>>
  variant: 'primary' | 'outline-primary'
  openInNewTab?: boolean
  redirectTo: string
  badge?: string
  type: 'button' | 'dropdownButton'
  actions?: GetStartedCardResourceAction[]
}

export interface GetStartedCardResource {
  label: string
  text: string
  resources?: GetStartedCardResource[]
  actions?: GetStartedCardResourceAction
}

export interface GetStartedCard {
  imgSrc: string
  imgSrcSet: string
  title: string
  subTitle: string
  homePageText: string
  getStartedPageText: string
  redirectTo: string
  badge?: string
  resources: GetStartedCardResource[]
}

export interface initialCardSate {
  learn: GetStartedCard
  deploy: GetStartedCard
}

// Local state
export const initialCardState: initialCardSate = {
  learn: {
    imgSrc: 'images/dashboard/learning-cloud.png',
    imgSrcSet: 'images/dashboard/learning-cloud.png 1x, images/dashboard/learning-cloud@2x.png 2x',
    title: 'Learn',
    subTitle: 'For developers, students, and AI/ML researchers',
    homePageText: 'Easy access to Intel software learning and evaluation resources for accelerated computing.',
    getStartedPageText:
      'Gain mastery in AI and accelerated computing with Jupyter notebooks running on Intel GPUs and AI accelerators.',
    redirectTo: '/home/getstarted?tab=learn',
    badge: 'Free',
    resources: [
      {
        label: 'Start now',
        text: 'Use our shared environments to discover what Intel enables you to do. Get access to Intel GPUs or Intel AI accelerators.',
        actions: {
          variant: 'primary',
          label: 'Connect now',
          redirectTo: 'jupyterLab',
          badge: 'New',
          type: 'dropdownButton',
          actions: [
            {
              variant: 'primary',
              label: 'AI Accelerator',
              redirectTo: 'tr-gi-start', // Training Name for Intel Gaudi in Product Catalog
              type: 'button'
            },
            {
              variant: 'primary',
              label: 'GPU',
              redirectTo: '',
              type: 'button'
            }
          ]
        }
      },
      {
        label: 'Learning paths',
        text: 'Gain mastery in AI and cutting-edge topics with  our Jupyter notebooks running on GPUs and AI accelerators',
        actions: {
          variant: 'outline-primary',
          leftIcon: BsMortarboard,
          label: 'Learning',
          redirectTo: '/learning/notebooks',
          type: 'button'
        }
      },
      {
        label: 'Additional resources',
        text: 'Check out our tutorials, guides. code samples and videos for step-by-step guidance.',
        actions: {
          variant: 'outline-primary',
          type: 'button',
          leftIcon: BsBook,
          rigthIcon: BsBoxArrowUpRight,
          label: 'Documentation',
          openInNewTab: true,
          redirectTo: idcConfig.REACT_APP_PUBLIC_DOCUMENTATION
        }
      }
    ]
  },
  deploy: {
    imgSrc: 'images/dashboard/production-cloud.png',
    imgSrcSet: 'images/dashboard/production-cloud.png 1x, images/dashboard/production-cloud@2x.png 2x',
    title: 'Deploy',
    subTitle: 'For AI startups and AI-focused enterprise customers',
    homePageText:
      'Intel-optimized computing and infrastructure services for deploying AI workloads and services at scale.',
    getStartedPageText: 'Optimize, launch and scale your AI solutions on Intel. Higher performance at a lower cost.',
    redirectTo: '/home/getstarted?tab=deploy',
    resources: [
      {
        label: 'GPU and AI Accelerators',
        text: 'Physical and virtual compute instances for small to medium-scale AI inferencing and other use cases.',
        actions: {
          variant: 'outline-primary',
          type: 'button',
          leftIcon: Catalog,
          label: 'Select Instance',
          redirectTo: '/hardware?fctg=GPU%2CAI'
        }
      },
      {
        label: 'AI cluster',
        text: 'Multi-node compute clusters for medium- to large-scale AI model training and fine tuning.',
        actions: {
          variant: 'outline-primary',
          type: 'button',
          leftIcon: K8s,
          label: 'Launch Cluster',
          redirectTo: '/cluster/reserve'
        }
      }
    ]
  }
}
