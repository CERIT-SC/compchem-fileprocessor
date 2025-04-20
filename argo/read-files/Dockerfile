FROM alpine/curl:8.12.2

WORKDIR /script

COPY ./read-files.sh . 

RUN chmod +x ./read-files.sh

ENTRYPOINT [ "./read-files.sh" ]
