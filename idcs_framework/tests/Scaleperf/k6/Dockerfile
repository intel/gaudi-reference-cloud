FROM internal-placeholder.com/cache/loadimpact/k6:latest

ADD . /app
WORKDIR /app/scaleperf-suite

RUN ["chmod", "+x", "productTests.sh"]
ENTRYPOINT ["productTests.sh"]