FROM docker.io/library/golang:1.23.10-alpine3.21 AS compiler
WORKDIR /app/
COPY . .
RUN go build -o main .

FROM scratch AS runner
COPY --from=compiler /app/main /main
USER 1001:1001
ENTRYPOINT ["/main"]
EXPOSE 8080/tcp