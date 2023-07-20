FROM golang:1.20-alpine

RUN apk add --update tini 
RUN mkdir -p /app/prebid-cache/
WORKDIR /app/prebid-cache/

COPY ./ ./

RUN go mod download
RUN go mod tidy
RUN go mod vendor
RUN go build -mod=vendor -o /prebid-app

RUN addgroup -g 89874 prebid-cache
RUN adduser -D -H -u 89874 -G prebid-cache prebid-cache
RUN chown -R prebid-cache:prebid-cache /app/prebid-cache/
USER prebid-cache

EXPOSE 2424
EXPOSE 2525

ENTRYPOINT ["/sbin/tini", "--"]
CMD ["/prebid-app", "-v", "1", "-logtostderr"]

