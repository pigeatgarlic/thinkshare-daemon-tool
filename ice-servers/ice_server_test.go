package iceservers

import (
	"encoding/json"
	"testing"

	"github.com/OnePlay-Internet/daemon-tool/log"
	"github.com/pion/webrtc/v3"
)

func TestFilter(t *testing.T) {
	rtc := webrtc.Configuration{ICEServers: []webrtc.ICEServer{{
		URLs: []string{
			"stun:stun.l.google.com:19302",
		}}, {
		URLs:           []string{"turn:workstation.thinkmay.net:3478"},
		Username:       "oneplay",
		Credential:     "oneplay",
		CredentialType: webrtc.ICECredentialTypePassword,
	}, {
		URLs:           []string{"turn:52.66.204.210:3478"},
		Username:       "oneplay",
		Credential:     "oneplay",
		CredentialType: webrtc.ICECredentialTypePassword,
	}},
	}

	result := FilterWebRTCConfig(rtc)
	str := FilterAndEncodeWebRTCConfig(rtc)
	result2 := DecodeWebRTCConfig(str)

	str3, _ := json.MarshalIndent(rtc, " ", " ")
	str2, _ := json.MarshalIndent(result2, " ", " ")
	str1, _ := json.MarshalIndent(result, " ", " ")

	log.PushLog("%s\n%s\n%s\n%s", str3, str, str1, str2)
}
