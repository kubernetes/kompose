FROM golang:1.21.2
WORKDIR /app

# Copy the entire project which includes the public directory, vendoring, etc.
COPY . .

# Build your application
RUN CGO_ENABLED=0 GOOS=linux go build -o /frontend

EXPOSE 8080
CMD ["/frontend"]
