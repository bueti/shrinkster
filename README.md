# Shrinkster

Shrinkster is a URL shortener project built to get more familiar with Go.

## Setup

This project uses Docker Compose to run the application and its dependencies. To get started, run the following command:

```sh
docker compose up --build
```

This will start the server and a Postgres database. You will have to configure the environment variables, or adapt the startup parameters to match your environment.

## Deployment

This project uses Github Actions to build a Docker image and push it to Docker Hub. Then it uses a Github Action to deploy the image to a server using SSH.

## License
This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details
