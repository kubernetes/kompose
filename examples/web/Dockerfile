FROM golang:1.21.2

# Debugging within the container
RUN apt-get update && apt-get install dnsutils redis-tools -y

# Set the working directory in the container
WORKDIR /app

# Copy the entire project which includes the public directory, vendoring, etc.
COPY . .

# Build your application
RUN CGO_ENABLED=0 GOOS=linux go build -o /frontend

# Change the permissions so that all users can execute it
RUN chmod +x /frontend

# Although setting permissions on /frontend should suffice, set wider permissions if needed
RUN chown -R 1001:0 /app && \
    chmod -R g=u /app

# This directive ensures the container does not run as root
USER 1001

EXPOSE 8080
CMD ["/frontend"]
