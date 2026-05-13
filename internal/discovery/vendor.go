package discovery

import (
	"strings"

	"netdash/internal/device"
)

// ouiTable maps the first 3 octets of a MAC (uppercase, no separators) to a vendor name.
// Focused on vendors commonly seen in home networks.
var ouiTable = map[string]string{
	// Amazon / Amazon subsidiaries
	"B0FC0D": "Amazon", "A40801": "Amazon", "FC65DE": "Amazon",
	"747548": "Amazon", "74C246": "Amazon", "6837E9": "Amazon",
	"0C47C9": "Amazon", "34D270": "Amazon", "44650D": "Amazon",
	"8C8590": "Amazon", "B47C9C": "Amazon", "FCA183": "Amazon",
	"40B4CD": "Amazon", "F08173": "Amazon", "AC63BE": "Amazon",
	"4CEFC0": "Amazon", "3874B0": "Amazon", "CC9EAE": "Amazon",
	"1C12B0": "Amazon", "A4508F": "Amazon", "78FE3D": "Amazon",
	// Ring (Amazon subsidiary)
	"D011C3": "Ring", "A461EE": "Ring", "18B430": "Ring", "34EA34": "Ring",
	// Eero (Amazon subsidiary)
	"F81A67": "Eero", "A8A69F": "Eero", "A4D196": "Eero",
	// TP-Link / Kasa
	"14EBB6": "TP-Link", "50C7BF": "TP-Link", "A842A1": "TP-Link",
	"B04E26": "TP-Link", "002719": "TP-Link", "98DAC4": "TP-Link",
	"6032B1": "TP-Link", "18D6C7": "TP-Link", "C46E1F": "TP-Link",
	"D83066": "TP-Link", "B0A7B9": "TP-Link", "9469FB": "TP-Link",
	// Ubiquiti
	"002722": "Ubiquiti", "0418D6": "Ubiquiti", "24A43C": "Ubiquiti",
	"44D9E7": "Ubiquiti", "68D79A": "Ubiquiti", "788A20": "Ubiquiti",
	"802AA8": "Ubiquiti", "B4FBE4": "Ubiquiti", "DC9FDB": "Ubiquiti",
	"E063DA": "Ubiquiti", "F09FC2": "Ubiquiti", "F492BF": "Ubiquiti",
	"245A4C": "Ubiquiti", "FCECDA": "Ubiquiti", "18E829": "Ubiquiti",
	// Rachio
	"007F28": "Rachio",
	// Apple
	"3CA6F6": "Apple", "A4C3F0": "Apple", "F0B479": "Apple",
	"DC2B2A": "Apple", "3C2EFF": "Apple", "28CFE9": "Apple",
	"4C8D79": "Apple", "A45E60": "Apple", "BC67E6": "Apple",
	// Samsung
	"BC7E8B": "Samsung", "8CCA72": "Samsung", "A8F274": "Samsung",
	"6C2A0D": "Samsung", "D487D8": "Samsung",
	// Netgear
	"70F220": "Netgear", "28C68E": "Netgear", "A040A0": "Netgear",
	"C03F0E": "Netgear", "9C3DCF": "Netgear",
	// Sonos
	"94907E": "Sonos", "5CACF6": "Sonos", "D89561": "Sonos",
	"48A6B8": "Sonos", "78282E": "Sonos",
	// Google / Nest
	"F88FCA": "Google", "54607E": "Google", "6C5AB5": "Google",
	"A47733": "Google", "30FDA7": "Google",
	// Belkin / WeMo
	"B4750E": "Belkin", "EC1A59": "Belkin", "94103E": "Belkin",
	// LIFX
	"D073D5": "LIFX",
	// Philips Hue (Signify)
	"001788": "Philips Hue", "ECB5FA": "Philips Hue",
	// Shelly
	"C45BBE": "Shelly", "3C6105": "Shelly",
	// Tuya (many white-label smart plugs)
	"D81E86": "Tuya", "8899DC": "Tuya", "A4B1C1": "Tuya",
	// Verizon / Fios router
	"B8F853": "Verizon",
}

// hostnamePatterns maps hostname prefixes to human-readable device labels.
var hostnamePatterns = []struct {
	prefix string
	label  string
}{
	{"amazonplug", "Smart Plug"},
	{"amazon-", "Amazon Device"},
	{"kasa", "Smart Plug"},
	{"hs100", "Smart Plug"},
	{"hs110", "Smart Plug"},
	{"hs200", "Smart Switch"},
	{"hs300", "Smart Power Strip"},
	{"ring-", "Security Camera"},
	{"rachio-", "Sprinkler Controller"},
	{"eero", "WiFi Router"},
	{"iphone", "iPhone"},
	{"ipad", "iPad"},
	{"macbook", "MacBook"},
	{"xbox", "Xbox"},
	{"playstation", "PlayStation"},
	{"samsung", "Smart TV"},
	{"audioengine", "Smart Speaker"},
	{"sonos", "Smart Speaker"},
	{"nest-", "Nest Device"},
	{"chromecast", "Chromecast"},
	{"wemo-", "Smart Plug"},
	{"shelly", "Smart Plug"},
}

// vendorLabels maps vendor names to a device label when no hostname pattern matches.
var vendorLabels = map[string]string{
	"Ring":    "Security Camera",
	"Rachio":  "Sprinkler Controller",
	"Eero":    "WiFi Router",
	"Sonos":   "Smart Speaker",
	"LIFX":    "Smart Bulb",
	"Philips Hue": "Smart Bulb",
	"Shelly":  "Smart Plug",
}

// LookupVendor returns the vendor name for a MAC address, or "" if unknown.
func LookupVendor(mac string) string {
	oui := macToOUI(mac)
	if oui == "" {
		return ""
	}
	return ouiTable[oui]
}

// InferLabel returns a human-readable device label from hostname and/or vendor.
func InferLabel(hostname, vendor string) string {
	h := strings.ToLower(hostname)
	for _, p := range hostnamePatterns {
		if strings.HasPrefix(h, p.prefix) {
			return p.label
		}
	}
	if label, ok := vendorLabels[vendor]; ok {
		return label
	}
	return ""
}

// EnrichDevice sets MAC and vendor on the device, and infers a label if none is set.
// Call this inside a store.Update closure.
func EnrichDevice(d *device.Device, mac string) {
	if mac != "" {
		d.MAC = mac
		if d.Vendor == "" {
			d.Vendor = LookupVendor(mac)
		}
	}
	if d.Label == "" {
		d.Label = InferLabel(d.Hostname, d.Vendor)
	}
}

func macToOUI(mac string) string {
	mac = strings.ToUpper(mac)
	var b strings.Builder
	for _, c := range mac {
		if c != ':' && c != '-' && c != '.' {
			b.WriteRune(c)
		}
	}
	s := b.String()
	if len(s) < 6 {
		return ""
	}
	return s[:6]
}
