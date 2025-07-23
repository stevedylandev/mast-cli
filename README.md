# Mast

![img](https://cdn.stevedylan.dev/files/bafkreicb5ravot4fg6cvjmasp7l4n3c5x26lpejefx5mx6byubcq4oib4y)
A simple TUI for sending casts on Farcaster

## Install

There are several ways you can install Mast

### Install Script

You can copy this command to download and run the install script

```bash
curl -fsSL https://stevedylan.dev/mast.sh | bash
```

[Download and view script](https://stevedylan.dev/mast.sh)

### Homebrew

Mast can be installed with [Brew](https://brew.sh) by using the command below.

```
brew install stevedylandev/mast-cli/mast-cli
```

### Prebuilt Binary

Releases are prebuilt binaries that can be downloaded from the [releases page](https://github.com/stevedylandev/mast-cli/releases)

### Build From Source

Building the CLI from source is pretty easy, just clone the repo, build, and install.

```
git clone https://github.com/stevedylandev/mast-cli && cd mast-cli && go build . && go install .
```

## Setup

Before you start hoisting some bangers, run the auth command to authorize the CLI. You will need your FID and a Signer Private Key. You can easily generate one within the CLI by running the `login` command.

```
mast login
```

![mast-login](https://cdn.stevedylan.dev/files/bafybeicpkgpef2dn5dcxf3a34pu2mop4x4udjpjazva2taugpxeompdej4)

This will provide a QR code for you to scan and will open an approval screen within Warpcast. If you prefer to provide your own signer you can do so with the `auth` command.

```
mast auth
```

![mast-auth](https://cdn.stevedylan.dev/files/bafybeib5fji7gxx54wpk2oy3f3medklkclwwz6tl73si6ejugsgzqlcvya)

> [!TIP]
> If you're not sure how to make a signer or prefer to make one locally, check out [CastKeys](https://github.com/stevedylandev/cast-keys) or [Farcaster Keys Server](https://github.com/stevedylandev/farcaster-keys-server)

### Automatic Hub Setup

When you run `mast auth` or `mast login`, the CLI will automatically set up your hub configuration if it's not already configured. This includes:

1. **Hub Selection**: You'll be prompted to select your preferred hub (Neynar is recommended)
2. **API Key Setup**: If you choose Neynar, you'll be prompted to enter your API key
3. **Connection Verification**: The CLI will test the connection to ensure everything is working

This ensures you have everything configured in one go, so you can start casting immediately after authentication.

### Manual Hub Configuration

If you need to change your hub configuration later, you can use the `hub` command:

```
mast hub set
```

### Hub Options

Mast supports multiple Farcaster hubs for sending casts. Currently, **Neynar** is the recommended and default choice as other hubs have been shut down.

Available options:

- **Neynar** (Recommended) - Requires an API key
- **Pinata** (DEPRECATED) - No longer available
- **Standard Crypto** (DEPRECATED) - No longer available
- **Custom** - Enter your own hub URL

### Neynar API Key

Neynar is currently the primary hub provider for Farcaster. To use it:

1. **Get an API Key**: Visit [Neynar](https://neynar.com) to create an account and obtain your API key
2. **Enter API Key**: You'll be prompted to enter your Neynar API key (it will be stored securely)
3. **Verify Connection**: The CLI will test the connection to ensure everything is working

> [!TIP]
> Your API key is stored securely in your home directory and will be automatically used for all future casts.

### Troubleshooting

If you encounter issues with your hub connection:

- **402 Payment Required**: Check your Neynar account status and billing
- **401 Authentication Failed**: Verify your API key is correct
- **403 Forbidden**: Ensure you have the necessary permissions for the hub API

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

You can also use optional flags to bypass the interactive TUI for a quick cast

```
NAME:
   mast new - Send a new Cast

USAGE:
   mast new [command options]

OPTIONS:
   --message value, -m value  Cast message text
   --url value, -u value      URL to embed in the cast
   --url2 value, --u2 value   Second URL to embed in the cast
   --channel value, -c value  Channel ID for the cast
   --help, -h                 show help
```

![mast-new](https://cdn.stevedylan.dev/files/bafybeievnzmfviuwq7v57nyd4bprtk3khvtelegrqqiabswfwvblmksewy)

> [!NOTE]
> To cast in a channel make sure you are already a member

## Questions

If you have an quesitons or issues feel free to [contact me](https://stevedylan.dev/links)!
