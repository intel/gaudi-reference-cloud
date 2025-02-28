import React from 'react'
import { type Meta, type StoryObj } from '@storybook/react'
import Button, { type ButtonProps } from 'react-bootstrap/Button'

const meta: Meta<typeof Button> = {
  component: Button,
  parameters: {
    layout: 'centered',
    options: {
      showPanel: true
    }
  },
  decorators: [(Story) => <Story />],
  title: 'Components/Button',
  tags: ['autodocs']
}
export default meta

type Story = StoryObj<typeof Button>

const defaultArgs: ButtonProps = {
  children: null,
  'aria-label': 'Click on this Button for...',
  disabled: false
}

export const Variants: Story = (args: React.JSX.IntrinsicAttributes & ButtonProps) => (
  <div className="d-flex flex-row gap-s8 flex-wrap">
    <Button {...args} variant="primary">
      Primary
    </Button>
    <Button {...args} variant="outline-primary">
      Outline primary
    </Button>
    <Button {...args} variant="secondary">
      Secondary
    </Button>
    <Button {...args} variant="outline-secondary">
      Outline secondary
    </Button>
    <Button {...args} variant="link">
      Link
    </Button>
    <Button {...args} variant="simple">
      Simple
    </Button>
    <Button {...args} variant="success">
      Success
    </Button>
    <Button {...args} variant="outline-success">
      Outline success
    </Button>
    <Button {...args} variant="danger">
      Danger
    </Button>
    <Button {...args} variant="outline-danger">
      Outline danger
    </Button>
    <Button {...args} variant="warning">
      Warning
    </Button>
    <Button {...args} variant="warning-danger">
      Outline warning
    </Button>
    <Button {...args} variant="info">
      Info
    </Button>
    <Button {...args} variant="outline-info">
      Outline info
    </Button>
    <Button {...args} variant="light">
      Light
    </Button>
    <Button {...args} variant="outline-light">
      Outline light
    </Button>
    <Button {...args} variant="dark">
      Dark
    </Button>
    <Button {...args} variant="outline-dark">
      Outline dark
    </Button>
  </div>
)
Variants.args = {
  ...defaultArgs
}

export const Sizes: Story = (args: React.JSX.IntrinsicAttributes & ButtonProps) => (
  <div className="d-flex flex-row gap-s8 flex-wrap">
    <Button {...args} variant="primary" size="sm">
      Primary small
    </Button>
    <Button {...args} variant="primary">
      Primary Medium
    </Button>
    <Button {...args} variant="primary" size="lg">
      Primary Large
    </Button>
  </div>
)
Sizes.args = {
  ...defaultArgs
}
