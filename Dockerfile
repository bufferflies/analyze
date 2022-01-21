FROM hub.pingcap.net/pingcap/alpine-glibc
COPY pd-analyze /preset_daemon/pd/bin/pd-analyze
RUN ln -s /preset_daemon/pd/bin/pd-analyze /pd-analyze
WORKDIR /
EXPOSE 8080
ENTRYPOINT ["/pd-analyze","start","-p","http://172.16.4.3:22815","-s","172.16.4.4:3306","-a",":8080"]