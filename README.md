# Streamr Operator API

Streamr Operator API is a management API developed in Go, leveraging the Gin framework to facilitate interaction with the Streamr network for operators. It enables operators to manage their stakes, withdraw earnings, and interact with sponsorships through a set of RESTful APIs.

## Features

- **Operator Management**: View and manage operator details, including staked balances and sponsorships.
- **Earnings Withdrawal**: Support for withdrawing earnings and compounding stakes.
- **API Documentation**: Integrated Swagger UI for API documentation and interaction.

## Prerequisites

Before you begin, ensure you have met the following requirements:

- Go 1.16 or higher
- Access to an Polygon blockchain node (e.g., via Infura or a local node)

## Getting Started

Clone the repository to your local machine:

```bash
git clone https://github.com/ftkuhnsman/streamr_api.git
cd streamr_api
```
### Configuration

To configure the Streamr Operator Service, set the following environment variables according to your setup:

- `CONTRACT_ADDR`: Specifies the address of the Streamr Operator contract on the blockchain. This is required for the API to interact with the contract.
- `OWNER_ADDR`: The Polygon address of the operator's owner. This address is used to authenticate and perform operations that require ownership privileges.
- `PRIVATE_KEY`: The private key corresponding to `OWNER_ADDR`. It is used for signing transactions. **Ensure this is kept secure and not exposed in your code or version control**.
- `RPC_ADDR`: The RPC address of your Polygon node. This allows the API to communicate with the Polygon blockchain. Example: `https://polygon-mainnet.infura.io/v3/YOUR_PROJECT_ID` for Polygon mainnet or a similar URL for other providers.
- `PORT`: (Optional) The port number on which the Streamr Operator API will listen for incoming requests. The default is `8080` if not specified.

These variables can be set in your operating system's environment, or you can use a `.env` file at the root of your project with the following content:

```env
CONTRACT_ADDR=0xYourContractAddress
OWNER_ADDR=0xYourOwnerAddress
PRIVATE_KEY=YourPrivateKey
RPC_ADDR=https://yourRpcUrl 
PORT=8080
```

Note: Replace the placeholder values with your actual configuration details.

## Running the Service

After configuring the environment variables, you can start the Streamr API Service by following these steps:

1. **Build the application** (optional): While this step is optional since you can run Go applications directly through the `go run` command, building the application can help ensure all dependencies are correctly compiled. To build the application, run:

    ```bash
    go build -o streamr-api
    ```

    This command compiles the application and creates an executable named `streamr-api`.

2. **Start the service**: To run the service, you can either execute the built binary (if you followed the optional build step) or use the `go run` command with the main file as an argument. Here are the commands for both methods:

    - Using the built binary (if you have built it in the previous step):

    ```bash
    ./streamr-api
    ```

    - Using `go run`:

    ```bash
    go run main.go
    ```

    This command starts the Streamr API Service, making the API available for requests. By default, the service runs on port `8080`, unless a different port was specified in the `PORT` environment variable.

3. **Verify the service is running**: You can verify that the service is running by accessing the Swagger UI for the API documentation and interactive exploration at `http://localhost:8080/` (adjust the port number if you used a custom port).

    Opening this URL in a web browser should display the Swagger UI, where you can explore the API endpoints and try out requests directly from the interface.

Congratulations! Your Streamr Operator Service is now running and ready to interact with the Streamr network.

## Running with Docker

For ease of deployment, the Streamr Operator Service can also be run as a Docker container. This method abstracts away the need for manually managing dependencies and environment setups. Follow the steps below to get your service running in a Docker container.

### Prerequisites

Ensure Docker and Docker Compose are installed on your system. Docker Compose will use the `docker-compose.yml` file in the repository to configure and run your service.

### Configuration

Before running the service with Docker, make sure to configure the environment variables in a `.env` file located at the root of your project directory. This file will be automatically used by Docker Compose to set up your container environment. Refer to the [Configuration](#configuration) section for details on the required environment variables.

### Running the Service

1. **Navigate to the project directory**: Open a terminal and change to the directory where your project is located.

    ```bash
    cd path/to/your/streamr-operator-service
    ```

2. **Start the service using Docker Compose**: Run the following command to start the Streamr Operator Service in a Docker container. The service will start in the background.

    ```bash
    docker-compose up -d
    ```

    This command tells Docker Compose to read the `docker-compose.yml` file, pull the necessary Docker image, and start the service as configured. The `-d` flag runs the container in detached mode, allowing you to continue using the terminal.

3. **Verify the service is running**: Ensure that the Docker container is running successfully by executing:

    ```bash
    docker-compose ps
    ```

    This command lists all the containers managed by Docker Compose for this project. Look for the `api` service to confirm it's up and running.

4. **Accessing the service**: With the service running in Docker, it is accessible on the configured port. By default, the Docker Compose file maps port `8080` of the container to port `8080` on your host machine. You can access the API or Swagger UI documentation by navigating to `http://localhost:8080/` in your web browser (replace `8080` with the port you configured if different).

### Stopping the Service

To stop and remove the containers, networks, and volumes associated with your service, run:

```bash
docker-compose down
```

This will stop the Streamr Operator Service and clean up the resources used by the Docker container.

The latest docker image is also available on dockerhub:
```
docker pull ftkuhnsman/streamr_api:latest
```

## Usage

This section provides examples of how to use the Streamr Operator Service API to perform common tasks such as viewing operator details, managing stakes, and withdrawing earnings. These examples use `curl`, a command-line tool for making HTTP requests. You can also use any HTTP client, including Postman, or the integrated Swagger UI at `http://localhost:8080/docs`.

### Viewing Operator Details

To retrieve details about the operator, including the staked balance and sponsorships:

```bash
curl -X GET "http://localhost:8080/api/v1/operator" -H "accept: application/json"
```

### Withdrawing Earnings
To withdraw earnings for the operator:

```bash
Copy code
curl -X GET "http://localhost:8080/api/v1/operator/withdrawearnings" -H "accept: application/json"
```

### Staking on a Sponsor
To stake a certain amount on a given sponsor, replace <sponsorship_address> and <amount> with the sponsorship's address and the amount to stake in wei:

```bash
Copy code
curl -X GET "http://localhost:8080/api/v1/operator/stake/<sponsorship_address>/<amount>" -H "accept: application/json"
```
### Reducing Stake
To reduce the stake to a certain amount on a given sponsor, replace <sponsorship_address> and <new_amount> with the sponsorship's address and the new amount to stake in wei:

```bash
Copy code
curl -X GET "http://localhost:8080/api/v1/operator/reducestaketo/<sponsorship_address>/<new_amount>" -H "accept: application/json"
```

### Listing Sponsorships and Earnings
To list all sponsorships along with uncollected earnings:

```bash
Copy code
curl -X GET "http://localhost:8080/api/v1/operator/sponsorshipsandearnings" -H "accept: application/json"
```

### Compounding Earnings
To withdraw earnings from all sponsorships and automatically restake them:

```bash
Copy code
curl -X GET "http://localhost:8080/api/v1/operator/withdrawearningsandcompound" -H "accept: application/json"
```

Note: These examples use the default port 8080 specified in the Docker Compose file. If you are running the service without Docker or have configured a different port, adjust the URLs accordingly.

For a comprehensive list of all available API endpoints and their parameters, refer to the Swagger UI documentation at http://localhost:8080/docs after starting the service.

By following these steps, you can quickly get the Streamr Operator Service running in a Docker container, making it easy to deploy and manage.