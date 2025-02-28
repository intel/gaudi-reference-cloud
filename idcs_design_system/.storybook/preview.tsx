import React from 'react'
import type { Preview } from '@storybook/react'
import { useDarkMode } from 'storybook-dark-mode'
import './preview.scss'

const preview: Preview = {
  decorators: [
    (Story) => (
      <div data-bs-theme={!useDarkMode() ? 'light' : 'dark'} className="sheet">
        <Story />
      </div>
    )
  ],
  parameters: {
    darkMode: {
      stylePreview: true
    },
    controls: {
      matchers: {
        color: /(background|color)$/i,
        date: /Date$/i
      }
    },
    options: {
      storySort: {
        order: ['Documentation', ['Introduction', 'GetStarted', 'Contributing', 'Changelog']]
      }
    }
  }
}

export default preview
