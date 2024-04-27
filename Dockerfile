FROM golang:1.22-alpine as builder
WORKDIR /
COPY . ./
RUN go mod download


RUN go build -o /dinghy-client


FROM alpine
COPY --from=builder /dinghy-client .


EXPOSE 80
CMD [ "/dinghy-client" ]