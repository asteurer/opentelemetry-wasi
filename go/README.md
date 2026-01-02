# OpenTelemetry WASI for Go

> [!Caution]
> This is a work-in-progress.

## Running the examples
### Requirements
- [**go**](https://go.dev/dl/) - v1.25+
- [**componentize-go**](https://github.com/asteurer/componentize-go) - Latest version

### Usage
Build a version of [Spin](https://github.com/spinframework/spin) from this [branch](https://github.com/asteurer/spin/tree/wasi-otel) and install the relevant plugins:
```sh
git clone --branch wasi-otel --depth 1 https://github.com/asteurer/spin
cd spin
cargo install --path .
spin plugin update
spin plugin install otel
```

Then, run the example of your choosing:
```sh
cd examples/spin-basic
spin build
spin otel setup
spin otel up
# In a different terminal...
curl localhost:3000
```

## Generating the WIT bindings
Whenever WIT files are changed/added to the `../wit` directory, the bindings  in `./wit_component` need to be regenerated.

### Prerequisites
- [**wit-bindgen**](https://github.com/bytecodealliance/wit-bindgen) - Latest version

### Run
```sh
wit-bindgen go -w imports --out-dir wit_component ../wit
```
