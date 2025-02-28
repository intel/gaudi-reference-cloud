
##########################  PRODUCTION IMAGE ##############################
FROM  internal-placeholder.com/cache/library/node:20.13 AS build-stage
# Set working directory

ENV http_proxy="http://internal-placeholder.com:911/" \
    https_proxy="http://internal-placeholder.com:912/" \
    no_proxy="127.0.0.1,localhost,intel.com,internal-placeholder.com"

WORKDIR /app

#RUN npm config set registry https://pixi.intel.com/
COPY ./idcs_admin_portal ./idcs_admin_portal
COPY ./idcs_design_system ./idcs_design_system

RUN cd ./idcs_design_system && npm install

RUN cd ./idcs_admin_portal && npm install

RUN cd ./idcs_admin_portal && npm run build


RUN ls

FROM internal-placeholder.com/cache/library/nginx:1.15

COPY /idcs_admin_portal/nginx.conf /etc/nginx/nginx.conf
COPY --from=build-stage /app/idcs_admin_portal/build /frontend/build




