// Copyright 2020 VMware, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package entities

import (
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var uniqueTemplateID uint16 = 256

func TestPrepareRecord(t *testing.T) {
	prepareRecordTests := []struct {
		record    Record
		expectLen uint16
		expectErr error
	}{
		{NewDataRecord(uniqueTemplateID, 1, 0, false), 0, nil},
		{NewTemplateRecord(uniqueTemplateID, 1, false), 4, nil},
	}

	for _, test := range prepareRecordTests {
		actualErr := test.record.PrepareRecord()
		if actualErr != test.expectErr {
			t.Errorf("Prepare record expects no error, but got %v", actualErr)
		}
	}
}

func TestAddInfoElements(t *testing.T) {
	testIEs := []*InfoElement{
		// Test element of each type
		NewInfoElement("protocolIdentifier", 4, 1, 0, 1),  // unsigned8
		NewInfoElement("sourceTransportPort", 7, 2, 0, 2), // unsigned16
		NewInfoElement("ingressInterface", 10, 3, 0, 4),   // unsigned32
		NewInfoElement("packetDeltaCount", 2, 4, 0, 8),    // unsigned64
		// No elements of signed8, signed16 and signed64 in IANA registry
		NewInfoElement("mibObjectValueInteger", 434, 7, 0, 4), // signed32
		// No elements of float32 in IANA registry
		NewInfoElement("samplingProbability", 311, 10, 0, 8),     // float64
		NewInfoElement("dataRecordsReliability", 276, 11, 0, 1),  // boolean
		NewInfoElement("sourceMacAddress", 56, 12, 0, 6),         // mac address
		NewInfoElement("sourceIPv4Address", 8, 18, 0, 4),         // IP Address
		NewInfoElement("interfaceDescription", 83, 13, 0, 65535), // String
		NewInfoElement("flowStartSeconds", 150, 14, 0, 4),        // dateTimeSeconds
		NewInfoElement("flowStartMilliseconds", 152, 15, 0, 8),   // dateTimeMilliseconds
	}
	macAddress, _ := net.ParseMAC("aa:bb:cc:dd:ee:ff")
	valData := []interface{}{
		uint8(0x1),                  // ICMP proto
		uint16(443),                 // https port
		uint32(1000),                // ingress interface ID
		uint64(100000),              // packet count
		int32(-12345),               // mibObjectValueInteger
		0.856,                       // samplingProbability
		true,                        // dataRecordsReliability
		macAddress,                  // mac address
		net.ParseIP("1.2.3.4"),      // IP Address
		"My Interface in IPFIX lib", // String
		uint32(time.Now().Unix()),   // dateTimeSeconds
		uint64(time.Now().Unix()),   // dateTimeMilliseconds
	}
	addIETests := []struct {
		record  Record
		ieList  []*InfoElement
		valList []interface{}
	}{
		{NewTemplateRecord(uniqueTemplateID, 12, false), testIEs, nil},
		{NewDataRecord(uniqueTemplateID, len(testIEs), 0, false), testIEs, valData},
	}

	for i, test := range addIETests {
		for j, testIE := range test.ieList {
			var actualErr error
			if i == 0 {
				// For template record
				ie := NewInfoElementWithValue(testIE, nil)
				actualErr = test.record.AddInfoElement(&ie)
			} else {
				// For data record
				ie := NewInfoElementWithValue(testIE, test.valList[j])
				actualErr = test.record.AddInfoElement(&ie)
				if testIE.Len == VariableLength {
					_, ok := test.valList[j].(string)
					if !ok {
						t.Errorf("val argument is not of valid type string")
					}
				}
			}
			assert.Equal(t, nil, actualErr, "Error returned is not nil")
		}
		// go test -v to see data buffer in hex format
		if i == 0 {
			t.Logf("template record %x", test.record.GetBuffer())
		} else {
			t.Logf("data record %x", test.record.GetBuffer())
		}
	}
}

func TestGetInfoElementWithValue(t *testing.T) {
	templateRec := NewTemplateRecord(256, 1, true)
	templateRec.orderedElementList = make([]InfoElementWithValue, 0)
	ie := NewInfoElementWithValue(NewInfoElement("sourceIPv4Address", 8, 18, 0, 4), nil)
	templateRec.orderedElementList = append(templateRec.orderedElementList, ie)
	_, _, exist := templateRec.GetInfoElementWithValue("sourceIPv4Address")
	assert.Equal(t, true, exist)
	_, _, exist = templateRec.GetInfoElementWithValue("destinationIPv4Address")
	assert.Equal(t, false, exist)
	dataRec := NewDataRecord(256, 1, 0, true)
	dataRec.orderedElementList = make([]InfoElementWithValue, 0)
	ie = NewInfoElementWithValue(NewInfoElement("sourceIPv4Address", 8, 18, 0, 4), net.ParseIP("10.0.0.1"))
	dataRec.orderedElementList = append(dataRec.orderedElementList, ie)
	infoElementWithValue, _, _ := dataRec.GetInfoElementWithValue("sourceIPv4Address")
	assert.Equal(t, net.ParseIP("10.0.0.1"), infoElementWithValue.Value)
	infoElementWithValue, _, _ = dataRec.GetInfoElementWithValue("destinationIPv4Address")
	assert.Empty(t, infoElementWithValue)
}
