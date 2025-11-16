container-registry-prefix := 'ghcr.io/asteurer'
otel-collector-version := "0.139.0"
go-version := "1.25.4"

build-test-base tag="latest":
    docker build -f test/Dockerfile.base \
        --build-arg OTEL_VERSION={{otel-collector-version}} \
        --build-arg GO_VERSION={{go-version}} \
        -t "{{container-registry-prefix}}/opentelemetry-wasi-integration-tests-base:{{tag}}" .

run-integration-tests-locally base_image_tag="latest":
    docker build -f test/Dockerfile.local_runner \
        --build-arg BASE_IMAGE="{{container-registry-prefix}}/opentelemetry-wasi-integration-tests-base:{{base_image_tag}}" \
        -t opentelemetry-wasi-integration-tests .

    docker run opentelemetry-wasi-integration-tests
