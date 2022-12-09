package iceservers

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/pion/webrtc/v3"
)

func TestFilter(t *testing.T) {
	rtc := webrtc.Configuration{ ICEServers: []webrtc.ICEServer{{
			URLs: []string{
				"stun:stun.l.google.com:19302",
			}, }, {
				URLs:           []string{"turn:workstation.thinkmay.net:3478"},
				Username:       "oneplay",
				Credential:     "oneplay",
				CredentialType: webrtc.ICECredentialTypePassword,
			}, {
				URLs:           []string{"turn:stun.l.google.com:19302"},
				Username:       "oneplay",
				Credential:     "oneplay",
				CredentialType: webrtc.ICECredentialTypePassword,
		}},
	}

	result  := FilterWebRTCConfig(rtc)
	str     := FilterAndEncodeWebRTCConfig(rtc);
	result2 := DecodeWebRTCConfig(str)

	str3,_ := json.MarshalIndent(rtc," ", " ");
	str2,_ := json.MarshalIndent(result2," ", " ");
	str1,_ := json.MarshalIndent(result," ", " ");

	fmt.Printf("%s\n%s\n%s\n%s",str3,str,str1,str2);
}



