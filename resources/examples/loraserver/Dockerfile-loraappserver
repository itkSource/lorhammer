FROM debian:9.2

# get version list : https://docs.loraserver.io/lora-app-server/overview/downloads/
ARG VERSION=0.14.1

RUN echo 'Acquire::http::Pipeline-Depth "0";' > /etc/apt/apt.conf.d/http-pipeline && \
    apt update && \
    apt install -y procps wget

RUN wget https://dl.loraserver.io/tar/lora-app-server_${VERSION}_linux_amd64.tar.gz -O /tmp/lora-app-server.tgz &&\
    tar xvf /tmp/lora-app-server.tgz -C /usr/local/bin && \
    chmod 755 /usr/local/bin/lora-app-server && \
    rm /tmp/lora-app-server.tgz

ADD ./rsa/issued/ /etc/ssl/certs/
ADD ./rsa/private/ /etc/ssl/private/

EXPOSE 8001 8080

HEALTHCHECK CMD bash -c '[[ $(ps faux | grep lora-app-server | grep -v grep | wc -l) > 0 ]]' && exit 0 || exit 1

CMD ["lora-app-server"]
