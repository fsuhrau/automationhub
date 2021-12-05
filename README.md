# Automation Hub
Automation Hub is a Service which makes device automation as easy as possible by reducing the pain of starting stopping and observing devices. Its current focus is on Game UITesting as most of the other available solutions leek certen features.

## Supports
- Android Devices (via USB ADB)
- Android Emulator (via ADB)
- Android Device Cloud(s) (via remote ADB)
- iOS Simulator (via xcrun)
- iOS Devices (via ios_deploy)
- Unity Editor

## Features
- device handling like starting, stopping and observing devices
- easy to use web interface
- building test scenarios for more complex test cases (app migration, backend compatibility tests)
- running tests on real devices
- running tests on unity editor
- building test reports
- measure performance on test runs
- provide native ui testing

## Installation
### Requirements
- Android ADB
- ios_deploy
- XCode

### Install via Brew on MacOS
you can install the hub via brew on macos its part of a private tap for now.
`brew install fsuhrau/tap/automationhub` (not released)

### Install via golang
you can also install it directly from sources
`go install github.com/fsuhrau/automationhub`

## Configure
to run the hub it needs a basic configuration you can create your configuration by running the configuration wizard.
for a more complex configuration you can checkout the [example_config.yaml](example_config.yaml)
`automationhub configure`

## Run
to run the hub you need a configuration check the configuration set Configure
`automationhub master --config config.yaml`
it will start observing the device state and provide an api and webinterface via ip on port 8002 [Link](http://localhost:8002)

## Usage
### Apps
### Devices
### Tests
#### Create Tests
#### Run Tests
#### Test Results