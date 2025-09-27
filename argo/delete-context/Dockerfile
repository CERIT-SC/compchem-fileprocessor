FROM alpine/curl:8.12.1

RUN apk add --no-cache bash

WORKDIR /script

COPY ./delete-context.sh .

RUN chmod +x ./delete-context.sh

ENTRYPOINT [ "/script/delete-context.sh" ]