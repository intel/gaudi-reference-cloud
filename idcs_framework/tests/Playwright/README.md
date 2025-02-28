## SRE End-to-End tests

We are using [Playwright|https://playwright.dev/] (PW) framework to run End-to-End (E2E) tests against INTEL cloud services exposed via web. Like other test frameworks (Selenium, Cypress) Playwright allows to run scripted test scenarios using headless browsers. We chose to use TypeScript for tests development because it is widely used, has wealth of best practices available, can produce reliable and fast tests, leverages all benefits of JavaScript (fast, asynchrounous).
Also Playwright tests are easily debugged, provide great tracing capabilities and nice reports.

## How to run locally

Pre-requisites on your machine:
- NodeJS, NPM and NPX
Recommended:
- VSCode with standard MS Playwright plugin for debugging

In working directory E2E-tests:
```
$ npm install
$ npx playwright install chromium
$ npx playwright --version
```

Create .env file (see example in .env.template) in same directory, with necessary credentials.

Run test:
```
$ npx playwright test [IDC00-UI.spec.ts]
```
Or in VSCode plugin panel ("Testing") choose desired test spec (or particular test inside it) and run test or debug. You can set breakpoints and perform step-by-step debugging. Test progress you can see in terminal, or in Debug console.
After the test, you can pull the test report by:
```
$ npx playwright show-report
```
According to playwright config, Playwright can generate standard HTML report, and/or other types of reports. It is very customizeable (eg. you can add Allure reports).

## Trace viewer
In addition to standard HTML reports, PW can generate Video captures, Screenshots and Trace archive.
You can configure these to be generated always, or on failure only. Traces are particularly great because they provide very detailed timeline view of all steps, all checks and expects, standard output, API activity, console logs, etc.
```
$ npx playwright show-trace path/to/trace.zip
```
Current PW config retains traces only on failed tests, to save disk space. In pipeline we have additional logic in PW config - based on env variable "CI" we allow one retry upon failure, to avoid flaky tests.

## Code generation tool
Playwright provide a great tool to visually navigate (simulating test user) and record your activities as a script, which is pretty much ready to run as a web UI test.
```
$ npx playwright codegen
```

## Monitoring
Primary purpose of these E2E tests is:
- IDC platform continuous monitoring (see automation in repo https://github.com/intel-innersource/frameworks.sre.infrastructure.pipelines/actions)
- on-demand IDC validation

Following comprehensive tests are available at the moment:
- IDC00-UI - verifying all UI features, including specific ones for Standard and Premium user (also may be different per-region)
- IDC01-VMaas - running Compute provisioning (small VM instance) with SSH validation. May be adjusted to provision any type of Compute (eg, BMaas).
- IDC02-Preview - provisioning test and verifications specific for IDC-Preview
- IDC03-STaaS-FS - File Storage provisioning and validation by mounting it on test VM (installing all pre-requisites)
- IDC04-STaaS-OS - Object Storage provisioning test

## Links
https://internal-placeholder.com/display/SRE/E2E+tests+pipeline+-+Playwright
