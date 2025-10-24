FROM alpine/curl:8.12.1

RUN apk add --no-cache bash jq

WORKDIR /script

COPY ./simulation-annotation.sh .

RUN chmod +x ./simulation-annotation.sh
RUN mkdir /files

ENTRYPOINT [ "/script/simulation-annotation.sh" ]
