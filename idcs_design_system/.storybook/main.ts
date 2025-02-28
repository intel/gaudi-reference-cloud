import type { StorybookConfig } from '@storybook/react-vite'

const config: StorybookConfig = {
  stories: ['../src/**/*.mdx', '../src/components/**/*.stories.@(js|jsx|mjs|ts|tsx)'],
  staticDirs: [
    { from: '../src/assets', to: '/assets' },
    { from: '../src/scss/dlux/fonts', to: '/fonts' }
  ],
  addons: [
    '@storybook/addon-links',
    '@storybook/addon-onboarding',
    '@storybook/addon-essentials',
    '@chromatic-com/storybook',
    '@storybook/addon-interactions',
    '@storybook/addon-storysource',
    'storybook-dark-mode',
    '@storybook/addon-styling-webpack',
    '@storybook/addon-docs'
  ],
  framework: {
    name: '@storybook/react-vite',
    options: {}
  }
}
export default config
