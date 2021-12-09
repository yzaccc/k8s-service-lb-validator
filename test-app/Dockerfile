#FROM golang:alpine AS build
FROM golang:1.14.4-stretch AS build

WORKDIR /build
COPY . .
RUN go build -o main .

FROM scratch
COPY --from=build /build/main /
EXPOSE 80 81

ENTRYPOINT ["sh","/main"]