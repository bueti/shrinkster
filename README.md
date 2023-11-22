# Shrinkster
<img align="right" src="https://github.com/bueti/shrinkster/assets/383917/e65d7181-31a5-4da1-8d0c-d97e640cbd33">
Shrinkster is a URL shortener project built with Go. The project is deployed to https://shrink.ch

## Installation

Shrinkster comes with a CLI application that can be used to create and manage short URLs. To install the CLI, run the following command:

```sh
brew install bueti/tap/shrink
```

or if you prefer to install it manually download it from the [releases](https://github.com/bueti/shrinkster/releases) page.

## Setup

Docker Compose is used to run the application and its dependencies. To get everything up and running locally, run the following command:

```sh
docker compose up --build
```

This will start the server and a Postgres database. You will have to configure the environment variables (via and .env file), or adapt the startup parameters to match your environment.

## Deployment

Shrinkster uses Github Actions to build a Docker image and push it to Docker Hub. Lastly, the image is deployed to an OVH VM using Docker Compose.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details
