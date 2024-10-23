# Mast

![img](https://cdn.stevedylan.dev/files/bafkreigqnynyjfax3loj5maiwnvv3qqxotpoajiq4p6r6glmt6pjmowjke)
A simple TUI for sending casts on Farcaster

## Install

Copy this command to download and run the install script

```bash
curl -fsSL https://stevedylan.dev/mast.sh | bash
```

[Download and view script](https://stevedylan.dev/mast.sh)

Alternatively download a prebuilt binary from the [releases page](https://github.com/stevedylandev/mast-cli/releases)

It is also available as a Homebrew tab

```
brew install stevedylandev/mast-cli/mast-cli
```

## Setup

Before you start hoisting some bangers, run the auth command to authorize the CLI. You will need your FID and a Signer Private Key. If you don't have a signer you can make one at [castkeys.xyz](https://castkeys.xyz)

```
mast auth
```

![mast-auth](https://cdn.stevedylan.dev/files/bafybeib5fji7gxx54wpk2oy3f3medklkclwwz6tl73si6ejugsgzqlcvya)

## Usage

To send a cast, simply run the command below.

```
mast new
```

You will be given the option to fill out different fields for the cast

```
 Message
 Main text body of your cast

 URL
 https://github.com/stevedylandev/mast-cli

 URL
 https://docs.farcaster.xyz

 Channel ID
 dev
```

![mast-new](https://cdn.stevedylan.dev/files/bafybeievnzmfviuwq7v57nyd4bprtk3khvtelegrqqiabswfwvblmksewy)

> [!NOTE]
> To cast in a channel make sure you are already a member

## Questions

If you have an quesitons or issues feel free to [contact me](https://stevedylan.dev/links)!
