Feature: Publish example service

    Background: Run all services
        Given Ethereum network is running on port 8545
        Given Contracts are deployed using Truffle
        Given IPFS is running with API port 5002 and Gateway port 8081
        Given snet is configured local rpc
            | Ethereum RPC port | user name   | IPFS port |
            | 8545              | "snet-user" | 5002      |
        Given Organization is added:
            | organization        |etcd endpoint|group name|payment address|
            | ExampleOrganization |http://127.0.0.1:2379|default_group|0x3b2b3C2e2E7C93db335E69D827F3CC4bC2A2A2cB|

    Scenario: Publish example-service-services
        When  example-service service is registered
            | name            | display name      | organization name   | daemon port |group name|
            | ExampleService  | Example Service | ExampleOrganization | 8090        |default_group|
        When  example-service service snet-daemon config file is created
            | name            | organization name   | daemon port | price |daemon group|
            | ExampleService  | ExampleOrganization | 8090        | 10    |default_group|
        When example-service service is running
        Then example-service make a call using payment channel
            | group name     | organization name   | daemon port |service name|
            | default_group  | ExampleOrganization | 8090        |ExampleService|
        Then example-service claim channel by treasurer server
            | daemon port |
            | 8090        |
        Then example-service make a call using payment channel
            | name            | organization name   | daemon port |service name|
            | ExampleService  | ExampleOrganization | 8090        |ExampleService|
        Then example-service claim channel by treasurer server
            | daemon port |
            | 8090        |
