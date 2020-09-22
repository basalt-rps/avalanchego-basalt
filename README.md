# Bᴀsᴀʟᴛ mod for AvalancheGo

This repository is a fork from the [official AvalancheGo repository](https://github.com/ava-labs/gecko) by AVA Labs,
that implements the novel Bᴀsᴀʟᴛ peer sampling algorithm.
The code of the Bᴀsᴀʟᴛ algorithm provided in this repository is neither
produced nor endorsed by the original AvalancheGo authors at AVA Labs.

Bᴀsᴀʟᴛ is a new peer sampling algorithm which is built to provide resilience
to Sybil attacks on the Internet. Bᴀsᴀʟᴛ is an alternative to Proof-of-Stake
in this regard: it enables for the sampling of validator nodes not with probability
weighted by their stake, but by giving similar weights to all nodes.
To protect against Sybil attacks, a uniform probability distribution
cannot be used. Instead, Bᴀsᴀʟᴛ spreads the sampled nodes over IP address
prefixes in a hierarchical manner, preventing any single large entity
from owning all our samples.

Bᴀsᴀʟᴛ will be used as soon as the `--staking-tls-enabled=false`
option is used.

# AvalancheGo

## Installation

Avalanche is an incredibly lightweight protocol, so the minimum computer requirements are quite modest.

- Hardware: 2 GHz or faster CPU, 4 GB RAM, 2 GB hard disk.
- OS: Ubuntu >= 18.04 or Mac OS X >= Catalina.
- Software: [Go](https://golang.org/doc/install) version >= 1.13.X and set up [`$GOPATH`](https://github.com/golang/go/wiki/SettingGOPATH).
- Network: IPv4 or IPv6 network connection, with an open public port.

### Native Install

Clone the AvalancheGo repository:

```sh
go get -v -d github.com/ava-labs/avalanchego/...
cd $GOPATH/src/github.com/ava-labs/avalanchego
```

#### Building the Avalanche Executable

Build Avalanche using the build script:

```sh
./scripts/build.sh
```

The Avalanche binary, named `avalanchego`, is in the `build` directory.

### Docker Install

- Make sure you have docker installed on your machine (so commands like `docker run` etc. are available).
- Build the docker image of latest avalanchego branch by `scripts/build_image.sh`.
- Check the built image by `docker image ls`, you should see some image tagged
  `avalanchego-xxxxxxxx`, where `xxxxxxxx` is the commit id of the Avalanche source it was built from.
- Test Avalanche by `docker run -ti -p 9650:9650 -p 9651:9651 avalanchego-xxxxxxxx /avalanchego/build/avalanchego
   --network-id=local --staking-enabled=false --snow-sample-size=1 --snow-quorum-size=1`. (For a production deployment,
  you may want to extend the docker image with required credentials for
  staking and TLS.)

## Running Avalanche

### Connecting to Mainnet

To connect to the Avalanche Mainnet, run:

```sh
./build/avalanchego
```

You should see some pretty ASCII art and log messages.

You can use `Ctrl+C` to kill the node.

### Connecting to Fuji

To connect to the Fuji Testnet, run:

```sh
./build/avalanchego --network-id=fuji
```

### Creating a Local Testnet

To create a single node testnet, run:

```sh
./build/avalanchego --network-id=local --staking-enabled=false --snow-sample-size=1 --snow-quorum-size=1
```

This launches an Avalanche network with one node.
