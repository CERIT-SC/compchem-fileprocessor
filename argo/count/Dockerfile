FROM alpine:3.21.3

RUN apk add --no-cache bash coreutils

WORKDIR /script

COPY ./count-words.sh . 

RUN chmod +x ./count-words.sh

ENTRYPOINT [ "/script/count-words.sh" ]
