package compose

import (
	"bytes"
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mast/hub"
	"mast/protobufs"
	"net/http"
	"strings"
	"time"

	auth "mast/auth"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/golang/protobuf/proto"
	"github.com/zeebo/blake3"
)

type spinnerModel struct {
	spinner  spinner.Model
	done     bool
	err      error
	castHash string
}

func initialSpinnerModel() spinnerModel {
	s := spinner.New()
	s.Spinner = spinner.Points
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#7C65C1"))
	return spinnerModel{spinner: s}
}

func (m spinnerModel) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m spinnerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		default:
			return m, nil
		}
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case doneMsg:
		m.done = true
		m.castHash = string(msg)
		return m, tea.Quit
	case errMsg:
		m.err = msg
		return m, tea.Quit
	default:
		return m, nil
	}
}

func (m spinnerModel) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v\n", m.err)
	}
	if m.done {
		return fmt.Sprintf("Cast Successful!\nHash: %s\n", m.castHash)
	}
	return fmt.Sprintf("%s Sending cast...\n", m.spinner.View())
}

type doneMsg string

const farcasterEpoch int64 = 1609459200 // January 1, 2021 UTC

func SendCast(castData CastData) error {
	rawFid, privateKeyHex, err := auth.FindFidAndPrivateKey()
	if err != nil {
		log.Fatalf("Problem retrieving credentials, run cast auth to authorize tbe CLI")
	}
	fid := uint64(rawFid)
	network := protobufs.FarcasterNetwork_FARCASTER_NETWORK_MAINNET

	resultChan := make(chan string)
	errorChan := make(chan error)
	go func() {
		var embeds []*protobufs.Embed

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

		castAdd := &protobufs.CastAddBody{
			Text:   castData.Message,
			Embeds: embeds,
		}

		if castData.Channel != "" {
			url := fmt.Sprintf("https://api.warpcast.com/v1/channel?channelId=%s", castData.Channel)
			resp, err := http.Get(url)
			if err != nil {
				log.Fatalf("Failed to send GET request: %v", err)
				return
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

		msgData := &protobufs.MessageData{
			Type:      protobufs.MessageType_MESSAGE_TYPE_CAST_ADD,
			Fid:       fid,
			Timestamp: uint32(time.Now().Unix() - farcasterEpoch),
			Network:   network,
			Body:      &protobufs.MessageData_CastAddBody{castAdd},
		}

		msgDataBytes, err := proto.Marshal(msgData)
		if err != nil {
			log.Fatalf("Failed to encode message data: %v", err)
			return
		}

		hasher := blake3.New()
		hasher.Write(msgDataBytes)
		hash := hasher.Sum(nil)[:20]

		msg := &protobufs.Message{
			HashScheme:      protobufs.HashScheme_HASH_SCHEME_BLAKE3,
			Hash:            hash,
			SignatureScheme: protobufs.SignatureScheme_SIGNATURE_SCHEME_ED25519,
		}

		if strings.HasPrefix(privateKeyHex, "0x") {
			privateKeyHex = privateKeyHex[2:]
		}
		privateKeyBytes, err := hex.DecodeString(privateKeyHex)
		if err != nil {
			log.Fatalf("Invalid hex string: %v", err)
			return
		}
		privateKey := ed25519.NewKeyFromSeed(privateKeyBytes)
		signature := ed25519.Sign(privateKey, hash)

		msg.Signature = signature
		msg.Signer = privateKey.Public().(ed25519.PublicKey)

		msg.DataBytes = msgDataBytes
		msgBytes, err := proto.Marshal(msg)
		if err != nil {
			log.Fatalf("Failed to encode message: %v", err)
			return
		}

		hub, apiKey, err := hub.RetrieveHubPreference()
		if err != nil {
			log.Fatal(err)
		}

		url := hub + "/v1/submitMessage"
		req, err := http.NewRequest("POST", url, bytes.NewBuffer(msgBytes))
		if err != nil {
			log.Fatalf("Failed to create request: %v", err)
		}
		req.Header.Set("Content-Type", "application/octet-stream")
		
		// Add API key header if available (for Neynar)
		if apiKey != "" {
			req.Header.Set("x-api-key", apiKey)
		}

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			log.Fatalf("Failed to send POST request: %v", err)
			return
		}
		defer resp.Body.Close()

		var response CastResponse
		err = json.NewDecoder(resp.Body).Decode(&response)
		if err != nil {
			log.Fatalf("Faled to decode json")
		}

		if resp.StatusCode == http.StatusOK {
			resultChan <- response.Hash
		} else {
			// Read the response body for more detailed error information
			bodyBytes, _ := io.ReadAll(resp.Body)
			bodyStr := string(bodyBytes)
			
			var errorMsg string
			switch resp.StatusCode {
			case 401:
				errorMsg = fmt.Sprintf("Authentication failed (401). Please check your API key.")
			case 402:
				errorMsg = fmt.Sprintf("Payment required (402). Please check your Neynar account status and billing.")
			case 403:
				errorMsg = fmt.Sprintf("Forbidden (403). You may not have permission to use this endpoint.")
			case 429:
				errorMsg = fmt.Sprintf("Rate limited (429). Please try again later.")
			default:
				errorMsg = fmt.Sprintf("Failed to send the message. HTTP status: %d. Response: %s", resp.StatusCode, bodyStr)
			}
			errorChan <- fmt.Errorf(errorMsg)
		}
	}()
	p := tea.NewProgram(initialSpinnerModel())

	go func() {
		select {
		case hash := <-resultChan:
			p.Send(doneMsg(hash))
		case err := <-errorChan:
			p.Send(errMsg(err))
		}
	}()

	if _, err := p.Run(); err != nil {
		return err
	}

	return nil
}
