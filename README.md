# 🌳 AutoPDF

**AutoPDF is a tool that helps you create PDF documents from templates.**

Simple. Neat. Fast. Powerful.

Perfect for creating professional, customizable PDF documents as a service.

[![License](https://img.shields.io/badge/license-Apache2-brightgreen.svg)](LICENSE)

## Install

``` bash
go install github.com/BuddhiLW/AutoPDF/cmd/AutoPDF@latest
```

## Testing

``` bash
go test ./...
```

## Usage

``` bash
cd ./internal/autopdf/test-data

autopdf build template.tex config.yaml
cat out/output.pdf
```

Should output:

```
file out/output.pdf 
out/output.pdf: PDF document, version 1.5
```

## License

This project is licensed under the [Apache License 2.0](LICENSE).


