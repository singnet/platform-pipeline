Feature: Publish example service

  Scenario: Run all services and publish example service
    Given Ethereum network is running on port 8545
    Given Contracts are deployed using Truffle
    Given IPFS is running