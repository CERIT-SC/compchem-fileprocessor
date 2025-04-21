FROM alpine/curl:8.12.1

RUN apk add --no-cache bash

WORKDIR /script

COPY ./write-files.sh . 

RUN chmod +x ./write-files.sh

ENTRYPOINT [ "/script/write-files.sh" ]
