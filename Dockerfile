FROM golang:1.20 as builder

# default application port
ENV APP_PORT=${APP_PORT:-"8080"}
ENV DATABASE_URL=${DATABASE_URL}

WORKDIR /app
# copy everything
COPY . ./

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -o /runner

FROM scratch 

COPY --from=builder /runner /runner

# port number as specified by the app
EXPOSE ${APP_PORT}

# runs the application
CMD ["/runner"]




