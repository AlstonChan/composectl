# Composectl

A simple command line tool for a docker compose repository to track running/decrypted services.

## How to Use

### Build the application

1. Clone the repository to your machine

   ```bash
   git clone https://github.com/AlstonChan/composectl.git
   ```

2. Install dependencies

   ```bash
   go install
   ```

3. Build it

   ```bash
   go build
   ```

4. Run `./composectl` to see the output

### See the documentation site locally

The documentation site uses [cobra doc gen](https://pkg.go.dev/github.com/spf13/cobra@v1.10.1/doc) to generate documentation of the cli app, then uses [Hugo](https://gohugo.io/) for static site generation of the markdown file.

So you need to make sure that you have `hugo` binary installed in your local machine to run

1. After cloning the repository into your machine, change directory to `docs`

   ```bash
   cd docs
   ```

2. Run the hugo development server

   ```bash
   hugo server
   ```

3. The docs site should be available at <http://localhost:1313>

#### To regenerate the cobra docs for hugo server, run

```bash
go run main.go gen-docs -p ./docs/content/cli
```

or if you have a built binary

```bash
./composectl gen-docs -p ./docs/content/cli
```

## Contributing

To develop this application locally:

1. Clone the repository to your machine

   ```bash
   git clone https://github.com/AlstonChan/composectl.git
   ```

2. Install dependencies

   ```bash
   go install
   ```

3. Run `go run main.go` to see the output

## License

This project is license under [Apache-2.0 License](./LICENSE)
