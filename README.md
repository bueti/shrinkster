# Shrinkster

Shrinkster is a URL shortener project built with Go. The project is deployed to https://shrink.ch

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
