# Cosmos Blockchain Simulator

## Overview

The Cosmos SDK offers a full fledged simulation framework to fuzz test every message defined by a module.

This functionality is provided by the the [`simulation`](https://github.com/cosmos/cosmos-sdk/tree/master/x/simulation) module, and a dummy application [`SimApp`](https://github.com/cosmos/cosmos-sdk/blob/master/simapp/app.go).
The simulation module defines all the simulation logic as well as the operations for randomized parameters like accounts, balances etc. The sim app defines a blockchain app which is run during the simulations.
Individual SDK modules define the set of messages that should be generated and delivered to a simulated app.

### Goals

The blockchain simulator tests how the blockchain application would behave under real life circumstances by generating and sending randomized messages.
The goal of this is to detect and debug failures that could halt a live chain, by providing logs and statistics about the operations run by the simulator as well as exporting the latest application state when a failure was found.

Its main difference with integration testing is that the simulator app allows you to pass parameters to customize the chain that's being simulated.
This comes in handy when trying to reproduce bugs that were generated in the provided operations (randomized or not).

## Running Simulations

Simulations are run through the go testing framework, and controlled with various flags.

For example:

```bash
 $ go test -mod=readonly github.com/cosmos/cosmos-sdk/simapp \
  -run=TestAppFullAppSimulation \
  -Enabled=true \
  -NumBlocks=100 \
  -BlockSize=200 \
  -Commit=true \
  -Period=5 \
  -v -timeout 24h
```

### Simulation commands

The simulation app has different commands (written as go tests), each of which tests a different failure type:

- `FullAppSimulation`: General simulation mode. Runs the chain and the specified operations for a given number of blocks. Tests that there're no `panics` on the simulation. It does also run invariant checks on every `Period` but they are not benchmarked.
- `AppImportExport`: The simulator exports the initial app state and then it creates a new app with the exported `genesis.json` as an input, checking for inconsistencies between the stores.
- `AppSimulationAfterImport`: Queues two simulations together. The first one provides the app state (_i.e_ genesis) to the second. Useful to test software upgrades or hard-forks from a live chain.
- `AppStateDeterminism`: Checks that all the nodes return the same values, in the same order.
- `BenchmarkInvariants`: Analyses the performance of running all the modules' invariants (_i.e_ sequentially runs a [benchmark](https://golang.org/pkg/testing/#hdr-Benchmarks) test). An invariant checks for differences between the values that are on the store and the passive tracker. Eg: total coins held by accounts vs total supply tracker.

Each simulation must receive a set of inputs (_i.e_ flags) such as the number of blocks that the simulation is run, seed, block size, etc.
Check the full list of flags [here](https://github.com/cosmos/cosmos-sdk/blob/adf6ddd4a807c8363e33083a3281f6a5e112ab89/simapp/sim_test.go#L34-L50).

> Note: sims must pass the `-Enabled=true` flag for them to run

### Simulator Modes

<!-- TODO make this explanation less confusing

module state, module parameters - aka the genesis file
simulation params -  I think small set of things like transition matrix
the "params" - overrides for op weights, module parameters, couple of other things
-->

In addition to the various inputs and commands, the simulator runs in three modes:

1. Completely random where the initial state, module parameters and simulation parameters are **pseudo-randomly generated**.
2. From a `genesis.json` file where the initial state and the module parameters are defined.
This mode is helpful for running simulations on a known state such as a live network export where a new (mostly likely breaking) version of the application needs to be tested.
3. From a `params.json` file where the initial state is pseudo-randomly generated but the module and simulation parameters can be provided manually.
This allows for a more controlled and deterministic simulation setup while allowing the state space to still be pseudo-randomly simulated. The list of available parameters is listed [here](https://github.com/cosmos/cosmos-sdk/blob/adf6ddd4a807c8363e33083a3281f6a5e112ab89/x/simulation/params.go#L170-L178).

::: tip
These modes are not mutually exclusive. So you can for example run a randomly generated genesis state (`1`) with manually generated simulation params (`3`).
:::

### Usage Examples

For more examples check the SDK [Makefile](https://github.com/cosmos/cosmos-sdk/blob/adf6ddd4a807c8363e33083a3281f6a5e112ab89/Makefile#L88-L123).

```bash
 $ go test -mod=readonly github.com/cosmos/cosmos-sdk/simapp \
  -run=TestApp<simulation_command> \
  ...<flags>
  -v -timeout 24h
```

### The `runsim` Tool

Cosmos uses a custom tool for running many simulations together and organizing the results.

### Debugging Tips

Here are some suggestions when encountering a simulation failure:

- Export the app state at the height were the failure was found. You can do this by passing the `-ExportStatePath` flag to the simulator.
- Use `-Verbose` logs. They could give you a better hint on all the operations involved.
- Reduce the simulation `-Period`. This will run the invariants checks more frequently.
- Print all the failed invariants at once with `-PrintAllInvariants`.
- Try using another `-Seed`. If it can reproduce the same error and if it fails sooner you will spend less time running the simulations.
- Reduce the `-NumBlocks` . How's the app state at the height previous to the failure?
- Run invariants on every operation with `-SimulateEveryOperation`. _Note_: this will slow down your simulation **a lot**.
- Try adding logs to operations that are not logged. You will have to define a [Logger](https://github.com/cosmos/cosmos-sdk/blob/adf6ddd4a807c8363e33083a3281f6a5e112ab89/x/staking/keeper/keeper.go#L65:17) on your `Keeper`.

## Writing Simulations

### How The Simulator Works

The simulator does the following:

 1. set up an app (using a randomly generated or provided genesis file)
 2. randomly generate message types and their contents
 3. deliver these messages in blocks to the app
 4. checks "invariants" every few blocks to make sure nothing has broken

**App** The app used is defined in /simapp. This is a fairly normal sdk app featuring all the modules.

**Operations** Messages are generated and delivered using "Operations". These are functions that generate and deliver various message types. They are defined in the modules.
For example the bank module has an operation for the `MsgSend` message type. It picks random to and from address, and a random coin amount, creates a MsgSend with these values then calls the bank handler with this msg. Some operations call handlers directly rather than the app's DeliverTx method.

**Invariants** check certain properties of the state to make sure they have not been broken - for example do all account balances add up to the total supply?
They are defined in the modules.
The simulator runs the invariants the same way a live chain would: through the crisis module. Every few blocks the crisis end blocker runs all the app's invariants. If any return `false` then the app panics.

**Mock Validators** A simulation does not run tendermint, instead the behavior of validators is mocked and randomly generated.
Every block a random set of mock validators are chosen to have signed a block correctly. Then some do not sign the block--simulating validator downtime, and some double sign - which is submitted as evidence.
Individual validator behavior is simulated through markov chains, where they hop between one of three states - offline, spotty connection, online.

**Block Size** The size of each block and the number of blocks to run are set at the beginning. Note block size is actually varied randomly through a markov chain.

<!--
### Use simulation in your SDK-based application

TODO: link to the simulation section on the tutorial for how to add your own simulation messages
-->
