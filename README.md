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

### Using the commands

1. List all the available docker service and see their status

   ```bash
   $ composectl list
   ```

   For self host compose repository at other path

   ```bash
   $ composectl list --repo-path=../../SelfHostCompose
   ```

2. See all _files_ of a docker service and its status

   - by service sequence (as per `composectl list`)

      ```bash
      $ composectl service -s 12 -a
      ```

   - by service name (as per `composectl list`)

      ```bash
      $ composectl service -n gitea -a
      ```

3. See all _secrets_ of a docker service and its status

   - by service sequence (as per `composectl list`)

      ```bash
      $ composectl service -s 12
      ```

   - by service name (as per `composectl list`)

      ```bash
      $ composectl service -n gitea
      ``` 

4. Set the configuration/setting for the CLI application:

   - Set a `age` public key that will be used to encrypt secrets

      ```bash
      $ composectl set age-pubkey=age1....
      ```

   - Set the root of the self host compose repository

      ```bash
      $ composectl set repo-path=../SelfHostCompose
      ```

      Note that using `--repo-path` will always override this option

5. Unset the configuration/setting for the CLI application so tthat it uses default option

   - Unset a `age` public key

      ```bash
      $ composectl unset age-pubkey
      ```

   - Unset the root of the self host compose repository

      ```bash
      $ composectl unset repo-path
      ```

6. Decrypt a docker service secrets

   - by service sequence (as per `composectl list`)

      ```bash
      $ composectl decrypt -s 12 -a
      ```

   - by service name (as per `composectl list`)

      ```bash
      $ composectl decrypt -n gitea -a
      ``` 
   
   - for a particular secret only by index (as per `composectl service`)

      ```bash
      $ composectl decrypt -s 12 -i 1
      ```

   - for all secrets of the service

      ```bash
      $ composectl decrypt -n gitea -a
      ``` 

   - to overwrite existing secrets

      ```bash
      $ composectl decrypt -n gitea -i 1 -o
      ``` 

7. Encrypt a docker service secrets

   - by service sequence (as per `composectl list`) for a file relative to the service root

      ```bash
      $ composectl encrypt -s 12 -f .env
      ```

   - by service name (as per `composectl list`) for a file relative to the service root

      ```bash
      $ composectl encrypt -n gitea -f .env
      ``` 

   - to overwrite existing encrypted secret

      ```bash
      $ composectl encrypt -n gitea -f .env -o
      ``` 

   - to specify a `age` public key if not set with `composectl set`

      ```bash
      $ composectl encrypt -n gitea -f .env -p age1....
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
