FROM golang:1.8.0

ADD ./ $GOPATH/src/lorhammer/

WORKDIR $GOPATH/src/lorhammer/

RUN sh install.sh
RUN sh build.sh && \
    ln -s $GOPATH/src/lorhammer/build/* /bin && \
    cp -r $GOPATH/src/lorhammer/resources /bin

