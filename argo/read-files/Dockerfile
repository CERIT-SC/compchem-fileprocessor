FROM alpine:3.21.3

WORKDIR /script

COPY ./read-files.sh . 

RUN chmod +x ./read-files.sh

ENTRYPOINT [ "./read-files.sh" ]
