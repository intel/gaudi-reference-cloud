module.exports = {
  roots: ['<rootDir>'],
  transform: {
    '\\.(js|jsx)$': ['babel-jest', { configFile: './babel.config.js' }]
  },
  testMatch: [
    '**/*.test.js'
  ],
  moduleFileExtensions: ['ts', 'js', 'jsx', 'json', 'node'],
  collectCoverage: true,
  clearMocks: true,
  coverageDirectory: 'coverage'
  // transform: {
  //     '^.+\\.jsx?$': 'babel-jest'
  // },
  // verbose: true,
  // testEnvironment: "node",
  // testEnvironment: "jest-environment-jsdom",
  // testEnvironmentOptions: {
  //     browsers: [
  //         "chrome",
  //         "firefox",
  //         "safari"
  //     ]
  // },
  // globals: {
  //     crypto: require("crypto")
  // },
  // coverageReporters: [
  //     [
  //         "lcov",
  //         {
  //             "projectRoot": "../../"
  //         }
  //     ]
  // ],
  // testRegex: '(/__tests__/.*|(\\.|/)(test|spec))\\.ts?$',
  // moduleFileExtensions: ['ts', 'jsx', 'js', 'json', 'node'],
  // collectCoverage: true,
  // clearMocks: true,
  // coverageDirectory: "coverage",
}
