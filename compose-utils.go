package main

import (
	"bytes"
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"mast/protobufs"
	"net/http"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/golang/protobuf/proto"
	"github.com/zeebo/blake3"
)

const farcasterEpoch int64 = 1609459200 // January 1, 2021 UTC

func SendCast(castData CastData) error {
	rawFid, privateKeyHex, err := FindFidAndPrivateKey()
	if err != nil {
		log.Fatalf("Problem retrieving credentials, run cast auth to authorize tbe CLI")
	}
	fid := uint64(rawFid) // FID of the user submitting the message
	network := protobufs.FarcasterNetwork_FARCASTER_NETWORK_MAINNET

	s := spinner.New(spinner.CharSets[35], 25*time.Millisecond)
	s.Color("bold", "fgHiWhite")
	s.Suffix = " Sending.."
	s.Start()
	// Create a slice to hold our embeds
	var embeds []*protobufs.Embed

	// Only add non-empty URLs as embeds
	if castData.URL1 != "" {
		embeds = append(embeds, &protobufs.Embed{
			Embed: &protobufs.Embed_Url{
				Url: castData.URL1,
			},
		})
	}

	if castData.URL2 != "" {
		embeds = append(embeds, &protobufs.Embed{
			Embed: &protobufs.Embed_Url{
				Url: castData.URL2,
			},
		})
	}

	// Construct the cast add message
	castAdd := &protobufs.CastAddBody{
		Text:   castData.Message,
		Embeds: embeds, // This will be an empty slice if no URLs were provided
	}

	if castData.Channel != "" {
		url := fmt.Sprintf("https://api.warpcast.com/v1/channel?channelId=%s", castData.Channel)
		resp, err := http.Get(url)
		if err != nil {
			log.Fatalf("Failed to send GET request: %v", err)
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			log.Fatalf("Channel fetch failed")
		}
		var response GetChannelResonse
		err = json.NewDecoder(resp.Body).Decode(&response)
		if err != nil {
			log.Fatalf("Faled to decode json")
		}
		castAdd.Parent = &protobufs.CastAddBody_ParentUrl{
			ParentUrl: response.Result.Channel.URL,
		}
	}

	// Construct the message data object
	msgData := &protobufs.MessageData{
		Type:      protobufs.MessageType_MESSAGE_TYPE_CAST_ADD,
		Fid:       fid,
		Timestamp: uint32(time.Now().Unix() - farcasterEpoch),
		Network:   network,
		Body:      &protobufs.MessageData_CastAddBody{castAdd},
	}

	// Serialize the message data to bytes
	msgDataBytes, err := proto.Marshal(msgData)
	if err != nil {
		log.Fatalf("Failed to encode message data: %v", err)
		return err
	}

	// Calculate the blake3 hash, truncated to 20 bytes
	hasher := blake3.New()
	hasher.Write(msgDataBytes)
	hash := hasher.Sum(nil)[:20]

	// Construct the actual message
	msg := &protobufs.Message{
		HashScheme:      protobufs.HashScheme_HASH_SCHEME_BLAKE3,
		Hash:            hash,
		SignatureScheme: protobufs.SignatureScheme_SIGNATURE_SCHEME_ED25519,
	}

	// Sign the message
	// REPLACE THE PRIVATE KEY WITH YOUR OWN
	if strings.HasPrefix(privateKeyHex, "0x") {
		privateKeyHex = privateKeyHex[2:]
	}
	privateKeyBytes, err := hex.DecodeString(privateKeyHex)
	if err != nil {
		log.Fatalf("Invalid hex string: %v", err)
		return err
	}
	privateKey := ed25519.NewKeyFromSeed(privateKeyBytes)
	signature := ed25519.Sign(privateKey, hash)

	// Continue constructing the message
	msg.Signature = signature
	msg.Signer = privateKey.Public().(ed25519.PublicKey)

	// Serialize the message
	msg.DataBytes = msgDataBytes
	msgBytes, err := proto.Marshal(msg)
	if err != nil {
		log.Fatalf("Failed to encode message: %v", err)
		return err
	}

	// Finally, submit the message to the network
	// url := "https://hub.farcaster.standardcrypto.vc:2281/v1/submitMessage"
	url := "https://hub.pinata.cloud/v1/submitMessage"
	resp, err := http.Post(url, "application/octet-stream", bytes.NewBuffer(msgBytes))
	if err != nil {
		log.Fatalf("Failed to send POST request: %v", err)
		return err
	}
	defer resp.Body.Close()

	var response CastResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		log.Fatalf("Faled to decode json")
	}

	if resp.StatusCode == http.StatusOK {
		s.Stop()
		fmt.Printf("Cast Successful! \nHash: %s\n", response.Hash)
		return nil
	} else {
		fmt.Printf("Failed to send the message. HTTP status: %d\n", resp.StatusCode)
		return nil
	}
}
