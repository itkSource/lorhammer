FROM debian:9.2

# get version list : https://docs.loraserver.io/lora-gateway-bridge/overview/downloads/
ARG VERSION=2.1.5

RUN echo 'Acquire::http::Pipeline-Depth "0";' > /etc/apt/apt.conf.d/http-pipeline && \
    apt update && \
    apt install -y procps wget

RUN wget https://dl.loraserver.io/tar/lora-gateway-bridge_${VERSION}_linux_amd64.tar.gz -O /tmp/lora-gateway-bridge.tgz &&\
    tar xvf /tmp/lora-gateway-bridge.tgz -C /usr/local/bin && \
    chmod 755 /usr/local/bin/lora-gateway-bridge && \
    rm /tmp/lora-gateway-bridge.tgz

EXPOSE 1700/udp

HEALTHCHECK CMD bash -c '[[ $(ps faux | grep lora-gateway-bridge | grep -v grep | wc -l) > 0 ]]' && exit 0 || exit 1

CMD ["lora-gateway-bridge"]
