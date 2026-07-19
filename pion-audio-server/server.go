package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/pion/webrtc/v4"
)

type SDP struct {
	SDP  string `json:"sdp"`
	Type string `json:"type"`
}

func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {
	mux := http.NewServeMux()
	config := webrtc.Configuration{}

	api := webrtc.NewAPI()

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

		peerConnection.OnConnectionStateChange(func(state webrtc.PeerConnectionState) {
			log.Println("Connection State:", state.String())
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

	log.Println("Listening on :9090")
	log.Fatal(http.ListenAndServe(":9090", cors(mux)))
}
