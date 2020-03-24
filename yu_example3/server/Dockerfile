FROM debian:jessie

LABEL version=1.0.0
#COPY commands.sh /scripts/commands.sh
#RUN ["chmod","+x","/scripts/commands.sh"]
#ENTRYPOINT ["/scripts/commands.sh"]
#ADD ./* /usr/local/bin/mpServer

WORKDIR /usr/local/bin/mpServer
COPY  .  /usr/local/bin/mpServer

ENTRYPOINT ["/usr/local/bin/mpServer/server2"]

