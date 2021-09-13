FROM hub.pingcap.net/pingcap/alpine-glibc
COPY pd-analyze /preset_daemon/pd/bin/pd-analyze
RUN ln -s /preset_daemon/pd/bin/pd-analyze /pd-analyze
WORKDIR /
EXPOSE 8080
ENTRYPOINT ["/pd-analyze","-p","http://pd-regression-prometheus:9090","-s","localhost:3306"]