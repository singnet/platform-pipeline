Feature: Publish example service

    Background: Run all services
        Given Ethereum network is running on port 8545
        Given Contracts are deployed using Truffle
        Given IPFS is running with API port 5002 and Gateway port 8081
        Given snet is configured local rpc
            | Ethereum RPC port | user name   | IPFS port |
            | 8545              | "snet-user" | 5002      |
        Given Organization is added:
            | organization        |
            | ExampleOrganization |

    Scenario: Publish example-service-services
        When  example-service service is registered
            | name            | display name      | organization name   | daemon port |
            | ExampleService  | Example Service | ExampleOrganization | 8090        |
        When  example-service service snet-daemon config file is created
            | name            | organization name   | daemon port | price |
            | ExampleService  | ExampleOrganization | 8090        | 10    |
        When example-service service is running
        Then example-service make a call using payment channel
            | name            | organization name   | daemon port |
            | ExampleService  | ExampleOrganization | 8090        |
        Then example-service claim channel by treasurer server
            | daemon port |
            | 8090        |
        Then example-service make a call using payment channel
            | name            | organization name   | daemon port |
            | ExampleService  | ExampleOrganization | 8090        |
        Then example-service claim channel by treasurer server
            | daemon port |
            | 8090        |
