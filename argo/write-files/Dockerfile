FROM alpine:3.21.3

WORKDIR /script

COPY ./write-files.sh . 

RUN chmod +x ./write-files.sh

ENTRYPOINT [ "./write-files.sh" ]
