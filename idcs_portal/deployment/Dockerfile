
##########################  PRODUCTION IMAGE ##############################

FROM  internal-placeholder.com/cache/library/node:20.13 AS build-ui-stage

ENV http_proxy="http://internal-placeholder.com:911/" \
    https_proxy="http://internal-placeholder.com:912/" \
    no_proxy="127.0.0.1,localhost,intel.com,internal-placeholder.com"

WORKDIR /app

COPY ./idcs_portal ./idcs_portal
COPY ./idcs_design_system ./idcs_design_system

RUN cd ./idcs_design_system && npm install

RUN cd ./idcs_portal && npm install

RUN cd ./idcs_portal && npm run build

FROM  internal-placeholder.com/cache/library/python:3.10.12 AS build-docs-stage

ENV http_proxy="http://internal-placeholder.com:911/" \
    https_proxy="http://internal-placeholder.com:912/" \
    no_proxy="127.0.0.1,localhost,intel.com,internal-placeholder.com"

WORKDIR /app

COPY ./idcs_portal/NOTICE.txt ../idcs_portal/NOTICE.txt
COPY ./docs ./

SHELL ["/bin/bash", "-c"]

RUN curl https://raw.githubusercontent.com/creationix/nvm/master/install.sh | bash
RUN source ~/.profile && nvm install v20.13.0 && node -v && npm -v
RUN python3 --version
RUN python3 -m pip install -U sphinx==5.3.0
RUN sphinx-build --version
RUN python3 -m pip install -r ./requirements.txt --no-cache-dir

RUN source ~/.profile && make clean && PROJECT=public make html
RUN cp -r ./source/_build/html ./source/public_docs
RUN ls ./source/public_docs

FROM internal-placeholder.com/cache/library/nginx:1.15

COPY /idcs_portal/nginx.conf /etc/nginx/nginx.conf
COPY --from=build-ui-stage /app/idcs_portal/build /frontend/build
COPY --from=build-docs-stage /app/source/public_docs /frontend/build/docs
