# ---- build stage ----
FROM golang:1.22-alpine AS build
WORKDIR /src
# No external deps, so just copy everything (incl. embedded patch.html) and build.
COPY . .
RUN CGO_ENABLED=0 go build -o /proxy .

# ---- run stage ----
# distroless/static includes ca-certificates, which the proxy needs to make
# the outbound HTTPS request to theknot.com.
FROM gcr.io/distroless/static-debian12
COPY --from=build /proxy /proxy
EXPOSE 8080
ENTRYPOINT ["/proxy"]
