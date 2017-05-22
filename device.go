package onvif

import (
	"encoding/json"
	"fmt"
	"strings"
)

var deviceXMLNs = []string{
	`xmlns:tds="http://www.onvif.org/ver10/device/wsdl"`,
	`xmlns:tt="http://www.onvif.org/ver10/schema"`,
}

// GetDeviceInformation fetch information of ONVIF camera
func (device Device) GetDeviceInformation() (DeviceInformation, error) {
	// Create SOAP
	soap := SOAP{
		Body:  "<tds:GetDeviceInformation/>",
		XMLNs: deviceXMLNs,
	}

	// Send SOAP request
	response, err := soap.SendRequest(device.XAddr)
	if err != nil {
		return DeviceInformation{}, err
	}

	// Parse response to interface
	deviceInfo, err := response.ValueForPath("Envelope.Body.GetDeviceInformationResponse")
	if err != nil {
		return DeviceInformation{}, err
	}

	// Parse interface to struct
	result := DeviceInformation{}
	err = interfaceToStruct(&deviceInfo, &result)
	if err != nil {
		return result, err
	}

	return result, nil
}

// GetSystemDateAndTime fetch date and time from ONVIF camera
func (device Device) GetSystemDateAndTime() (string, error) {
	// Create SOAP
	soap := SOAP{
		Body:  "</tds:GetSystemDateAndTime>",
		XMLNs: deviceXMLNs,
	}

	// Send SOAP request
	response, err := soap.SendRequest(device.XAddr)
	if err != nil {
		return "", err
	}

	// Parse response
	dateTime, _ := response.ValueForPathString("Envelope.Body.GetSystemDateAndTimeResponse.SystemDateAndTime")
	return dateTime, nil
}

// GetCapabilities fetch info of ONVIF camera's capabilities
func (device Device) GetCapabilities() (DeviceCapabilities, error) {
	// Create SOAP
	soap := SOAP{
		XMLNs: deviceXMLNs,
		Body:  `<tds:GetCapabilities></tds:Category></tds:GetCapabilities>`,
	}

	// Send SOAP request
	response, err := soap.SendRequest(device.XAddr)
	if err != nil {
		return DeviceCapabilities{}, err
	}

	// Get network capabilities
	envelopeBodyPath := "Envelope.Body.GetCapabilitiesResponse.Capabilities"
	ifaceNetCap, err := response.ValueForPath(envelopeBodyPath + ".Device.Network")
	if err != nil {
		return DeviceCapabilities{}, err
	}

	netCap := NetworkCapabilities{}
	mapNetCap, ok := ifaceNetCap.(map[string]interface{})
	if ok {
		netCap.DynDNS = interfaceToBool(mapNetCap["DynDNS"])
		netCap.IPFilter = interfaceToBool(mapNetCap["IPFilter"])
		netCap.IPVersion6 = interfaceToBool(mapNetCap["IPVersion6"])
		netCap.ZeroConfig = interfaceToBool(mapNetCap["ZeroConfiguration"])
		netCap.Extension = make(map[string]bool)

		if mapNetExtension, ok := mapNetCap["Extension"].(map[string]interface{}); ok {
			for key, value := range mapNetExtension {
				netCap.Extension[key] = interfaceToBool(value)
			}
		}
	}

	// Get events capabilities
	ifaceEventsCap, err := response.ValueForPath(envelopeBodyPath + ".Events")
	if err != nil {
		return DeviceCapabilities{}, err
	}

	eventsCap := make(map[string]bool)
	if mapEventsCap, ok := ifaceEventsCap.(map[string]interface{}); ok {
		for key, value := range mapEventsCap {
			if strings.ToLower(key) == "xaddr" {
				continue
			}

			key = strings.Replace(key, "WS", "", 1)
			eventsCap[key] = interfaceToBool(value)
		}
	}

	// Get events capabilities
	ifaceStreamingCap, err := response.ValueForPath(envelopeBodyPath + ".Media.StreamingCapabilities")
	if err != nil {
		return DeviceCapabilities{}, err
	}

	streamingCap := make(map[string]bool)
	if mapStreamingCap, ok := ifaceStreamingCap.(map[string]interface{}); ok {
		for key, value := range mapStreamingCap {
			key = strings.Replace(key, "_", " ", -1)
			streamingCap[key] = interfaceToBool(value)
		}
	}

	// Get PTZ capabilities
	ptzCap := true
	if _, err = response.ValueForPath(envelopeBodyPath + ".PTZ"); err != nil {
		ptzCap = false
	}

	// Create final result
	deviceCapabilities := DeviceCapabilities{
		Network:   netCap,
		Events:    eventsCap,
		Streaming: streamingCap,
		PTZ:       ptzCap,
	}

	return deviceCapabilities, nil
}

// GetDiscoveryMode fetch network discovery mode of an ONVIF camera
func (device Device) GetDiscoveryMode() (string, error) {
	// Create SOAP
	soap := SOAP{
		Body:  "<tds:GetDiscoveryMode/>",
		XMLNs: deviceXMLNs,
	}

	// Send SOAP request
	response, err := soap.SendRequest(device.XAddr)
	if err != nil {
		return "", err
	}

	// Parse response
	discoveryMode, _ := response.ValueForPathString("Envelope.Body.GetDiscoveryModeResponse.DiscoveryMode")
	return discoveryMode, nil
}

// GetScopes fetch scopes of an ONVIF camera
func (device Device) GetScopes() ([]string, error) {
	// Create SOAP
	soap := SOAP{
		Body:  "<tds:GetScopes/>",
		XMLNs: deviceXMLNs,
	}

	// Send SOAP request
	response, err := soap.SendRequest(device.XAddr)
	if err != nil {
		return nil, err
	}

	// Parse response to interface
	ifaceScopes, err := response.ValuesForPath("Envelope.Body.GetScopesResponse.Scopes")
	if err != nil {
		return nil, err
	}

	// Convert interface to array of scope
	scopes := []string{}
	for _, ifaceScope := range ifaceScopes {
		if mapScope, ok := ifaceScope.(map[string]interface{}); ok {
			scope := interfaceToString(mapScope["ScopeItem"])
			scopes = append(scopes, scope)
		}
	}

	return scopes, nil
}

// GetHostname fetch hostname of an ONVIF camera
func (device Device) GetHostname() (HostnameInformation, error) {
	// Create SOAP
	soap := SOAP{
		Body:  "<tds:GetHostname/>",
		XMLNs: deviceXMLNs,
	}

	// Send SOAP request
	response, err := soap.SendRequest(device.XAddr)
	if err != nil {
		return HostnameInformation{}, err
	}

	// Parse response to interface
	ifaceHostInfo, err := response.ValueForPath("Envelope.Body.GetHostnameResponse.HostnameInformation")
	if err != nil {
		return HostnameInformation{}, err
	}

	// Parse interface to struct
	hostnameInfo := HostnameInformation{}
	if mapHostInfo, ok := ifaceHostInfo.(map[string]interface{}); ok {
		hostnameInfo.Name = interfaceToString(mapHostInfo["Name"])
		hostnameInfo.FromDHCP = interfaceToBool(mapHostInfo["FromDHCP"])
		hostnameInfo.Extension = interfaceToString(mapHostInfo["Extension"])
	}

	return hostnameInfo, nil
}

// GetDNS fetch DNS of an ONVIF camera
func (device Device) GetDNS() (string, error) {
	// Create SOAP
	soap := SOAP{
		Body:  "<tds:GetDNS/>",
		XMLNs: deviceXMLNs,
	}

	// Send SOAP request
	response, err := soap.SendRequest(device.XAddr)
	if err != nil {
		return "", err
	}

	bt, _ := response.JsonIndent("", "    ")
	fmt.Println(string(bt))

	// Parse response
	DNS, _ := response.ValueForPathString("Envelope.Body.GetDNSResponse.DNSInformation")
	return DNS, nil
}

// GetDNS fetch DNS of an ONVIF camera
func (device Device) GetNetworkInterfaces() (string, error) {
	// Create SOAP
	soap := SOAP{
		Body:  "<tds:GetNetworkInterfaces/>",
		XMLNs: deviceXMLNs,
	}

	// Send SOAP request
	response, err := soap.SendRequest(device.XAddr)
	if err != nil {
		return "", err
	}

	bt, _ := response.JsonIndent("", "    ")
	fmt.Println(string(bt))

	// Parse response
	DNS, _ := response.ValueForPathString("Envelope.Body.GetDNSResponse.DNSInformation")
	return DNS, nil
}

func interfaceToStruct(src, dst interface{}) error {
	bt, err := json.Marshal(&src)
	if err != nil {
		return err
	}

	err = json.Unmarshal(bt, &dst)
	if err != nil {
		return err
	}

	return nil
}

func interfaceToString(src interface{}) string {
	str, _ := src.(string)
	return str
}

func interfaceToBool(src interface{}) bool {
	strBool := interfaceToString(src)
	return strings.ToLower(strBool) == "true"
}
