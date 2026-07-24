package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/pion/webrtc/v4"
)

type SDP struct {
	SDP  string `json:"sdp"`
	Type string `json:"type"`
}

func main() {
	var connectionCountMu sync.Mutex
	var connectionsDatachannels sync.Map
	connectionCount := 0

	updateConnectionCount := func() {
		connectionsDatachannels.Range(func(key, value any) bool {
			dc := value.(*webrtc.DataChannel)
			err := dc.SendText(fmt.Sprintf("%d", connectionCount))
			if err != nil {
				log.Println("failed to update data channel", err)
			}

			return true
		})
	}

	mux := http.NewServeMux()
	config := webrtc.Configuration{}

	api := webrtc.NewAPI()

	mux.Handle("/", http.FileServer(http.Dir("./static")))

	mux.HandleFunc("/offer", func(w http.ResponseWriter, r *http.Request) {
		log.Println("offer received")

		var offer SDP

		if err := json.NewDecoder(r.Body).Decode(&offer); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		peerConnection, err := api.NewPeerConnection(config)
		if err != nil {
			panic(err)
		}

		// Create an outgoing audio track.
		echoTrack, err := webrtc.NewTrackLocalStaticRTP(
			webrtc.RTPCodecCapability{
				MimeType: webrtc.MimeTypeOpus,
			},
			"audio",
			"echo",
		)
		if err != nil {
			panic(err)
		}

		_, err = peerConnection.AddTrack(echoTrack)
		if err != nil {
			panic(err)
		}

		// create the data channel and store it as the server connections
		var connectionId *uint16
		dc, err := peerConnection.CreateDataChannel("connections", &webrtc.DataChannelInit{})
		dc.OnError(func(err error) {
			log.Println("data channel error", err)
		})

		dc.OnOpen(func() {
			connectionId = dc.ID()
			connectionsDatachannels.Store(connectionId, dc)
			connectionCountMu.Lock()
			connectionCount += 1
			connectionCountMu.Unlock()
			updateConnectionCount()
		})

		peerConnection.OnConnectionStateChange(func(state webrtc.PeerConnectionState) {
			log.Println("Connection State:", state.String())

			removeDataChannel := func() {
				// Remove connection data channel from the server state
				connectionsDatachannels.Delete(connectionId)
				connectionCountMu.Lock()
				connectionCount -= 1
				connectionCountMu.Unlock()
				updateConnectionCount()
			}

			switch state.String() {
			case "closed":
				removeDataChannel()
			case "disconnected":
				removeDataChannel()
			}
		})

		peerConnection.OnICEConnectionStateChange(func(state webrtc.ICEConnectionState) {
			log.Printf("ICE Connection State: %s", state.String())
		})

		peerConnection.OnICEGatheringStateChange(func(state webrtc.ICEGatheringState) {
			log.Printf("ICE Gathering State: %s", state.String())
		})

		peerConnection.OnSignalingStateChange(func(state webrtc.SignalingState) {
			log.Printf("Signaling State: %s", state.String())
		})

		peerConnection.OnICECandidate(func(c *webrtc.ICECandidate) {
			if c == nil {
				log.Println("Finished gathering")
				return
			}

			log.Printf("Candidate: %s", c.ToJSON().Candidate)
		})

		peerConnection.OnTrack(func(track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
			log.Println("Incoming Track")
			log.Println("Kind:", track.Kind())
			log.Println("Codec:", track.Codec().MimeType)

			go func() {

				for {
					packet, _, err := track.ReadRTP()
					if err != nil {
						log.Println("Track ended:", err)
						return
					}

					if err := echoTrack.WriteRTP(packet); err != nil {
						log.Println(err)
						return
					}
				}

			}()

		})

		err = peerConnection.SetRemoteDescription(
			webrtc.SessionDescription{
				Type: webrtc.SDPTypeOffer,
				SDP:  offer.SDP,
			},
		)
		if err != nil {
			panic(err)
		}

		answer, err := peerConnection.CreateAnswer(nil)
		if err != nil {
			panic(err)
		}

		gatherComplete := webrtc.GatheringCompletePromise(peerConnection)

		err = peerConnection.SetLocalDescription(answer)
		if err != nil {
			panic(err)
		}

		<-gatherComplete

		response := SDP{
			Type: peerConnection.LocalDescription().Type.String(),
			SDP:  peerConnection.LocalDescription().SDP,
		}

		json.NewEncoder(w).Encode(response)

	})

	log.Println("Listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
