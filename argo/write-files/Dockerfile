
FROM alpine/curl:8.12.2

WORKDIR /script

COPY ./write-files.sh . 

RUN chmod +x ./write-files.sh

ENTRYPOINT [ "./write-files.sh" ]
