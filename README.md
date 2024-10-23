# Mast

![img](https://cdn.stevedylan.dev/files/bafkreic4ha4atqbzrfjbrnzqw6uhzramsbh2gvhjsh3lausgtnfnraw7yi)
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

![mast-auth](https://cdn.stevedylan.dev/files/bafybeid3zxtmhrw3zdxfm5y7r6jm7w3c4ir3esnejz66pofn4znudpliki)

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

![mast-new](https://cdn.stevedylan.dev/files/bafybeihsj7qpqy4sbgfrpzvh5nciepwlwl646hnkfcyrtk2hdl2hvdmd6y)

> [!NOTE]
> To cast in a channel make sure you are already a member

## Questions

If you have an quesitons or issues feel free to [contact me](https://stevedylan.dev/links)!
