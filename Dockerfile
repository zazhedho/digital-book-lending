FROM alpine:latest
RUN apk add --no-cache tzdata
ENV TZ=Asia/Jakarta

WORKDIR /app

ARG SERVICE_NAME
ENV SERVICE_NAME=${SERVICE_NAME}

COPY ${SERVICE_NAME} .
COPY entrypoint.sh .
#COPY .env . #parsing in docker run
#COPY config config

RUN chmod +x entrypoint.sh
ENTRYPOINT ["./entrypoint.sh"]