FROM alpine:3.21.3

RUN apk add --no-cache bash

WORKDIR /script

COPY ./read-files.sh . 

RUN chmod +x ./read-files.sh

ENTRYPOINT [ "/script/read-files.sh" ]
