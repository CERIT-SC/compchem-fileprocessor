FROM alpine:3.21.3

RUN apk add --no-cache bash
RUN apk add --no-cache --upgrade wget

WORKDIR /script

COPY ./read-files.sh . 

RUN chmod +x ./read-files.sh

ENTRYPOINT [ "/script/read-files.sh" ]
