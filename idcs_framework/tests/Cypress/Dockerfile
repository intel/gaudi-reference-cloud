# Build the manager binary
FROM internal-placeholder.com/intelcloud/cypress/browsers:node16.17.1-chrome106-ff105-edge 

# Default Intel proxy settings can be overridden with --build-args.
ARG http_proxy='http://internal-placeholder.com:912/'
ARG https_proxy='http://internal-placeholder.com:912/'
ARG no_proxy='intel.com,.intel.com,localhost,undefined'

# Configure proxy environment variables from ARG values above.
ENV proxy ${http_proxy}
ENV https_proxy ${https_proxy}
ENV no_proxy ${no_proxy}
RUN mkdir /cydocker
WORKDIR /cydocker
COPY ./package.json .
COPY ./.npmrc .
COPY ./cypress.config.js .
COPY ./cypress ./cypress
COPY ./reporter.json .
RUN npm config set registry https://pixi2.intel.com/ && npm install && npm install cy2 
#RUN sed -i -e 's|VERIFY_TEST_RUNNER_TIMEOUT_MS = 30000|VERIFY_TEST_RUNNER_TIMEOUT_MS = 100000|1' ./node_modules/cypress/lib/tasks/verify.js
ENTRYPOINT ["npx","cypress","run"]
CMD [""]