import React from 'react'
import { type Meta, type StoryObj } from '@storybook/react'
import { BsExclamationTriangle, BsCheckCircle, BsInfoCircle, BsExclamationOctagon } from 'react-icons/bs'
import Alert, { type AlertProps } from 'react-bootstrap/Alert'
import './Alert.stories.scss'

const meta: Meta<typeof Alert> = {
  component: Alert,
  parameters: {
    layout: 'centered',
    options: {
      showPanel: true
    }
  },
  decorators: [(Story) => <Story />],
  title: 'Components/Alert',
  tags: ['autodocs'],
  argTypes: {}
}
export default meta

type Story = StoryObj<typeof Alert>

const defaultArgs: AlertProps = {
  children: null,
  'aria-label': 'Click on this Button for...',
  onClose: () => {
    console.log('closed')
  },
  dismissible: false
}

const variants = [
  {
    variant: 'info',
    icon: <BsInfoCircle />
  },
  {
    variant: 'secondary',
    icon: <BsInfoCircle />
  },
  {
    variant: 'success',
    icon: <BsCheckCircle />
  },
  {
    variant: 'warning',
    icon: <BsExclamationTriangle />
  },
  {
    variant: 'danger',
    icon: <BsExclamationOctagon />
  }
]

export const Variants: Story = (args: React.JSX.IntrinsicAttributes & AlertProps) => (
  <div className="d-flex flex-column gap-s8 w-100">
    {variants.map((x, index) => (
      <Alert key={index} {...args} variant={x.variant} className="w-100" role="alert">
        <div className="d-flex flex-column justify-content-center align-self-center">{x.icon}</div>
        <div className="alert-body gap-s3">
          <Alert.Heading className="h6">Alert {x.variant}</Alert.Heading>
          <span className="alert-body">
            {' '}
            <span className="d-flex-inline gap-s4">
              This is the alert Message content. Also prefer links at the end like this one
              <>
                &nbsp;
                <a
                  className="link"
                  aria-label="Open Intel Tiber AI Console"
                  target={'_blank'}
                  rel={'noreferrer'}
                  href="https://console.cloud.intel.com"
                >
                  Intel Tiber AI Console
                </a>
                .
              </>
            </span>
          </span>
        </div>
      </Alert>
    ))}
  </div>
)
Variants.args = {
  ...defaultArgs
}
