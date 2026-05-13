I now have a comprehensive picture. Here's what I found about your smart home outlets:

Smart Home Outlets on Your Network
Currently Online: Amazon Smart Outlet — 192.168.1.219
MAC: b0:fc:0d:ac:1f:93 (Amazon Technologies Inc.)
mDNS hostname: AE586E9C9758.local
Protocol: Matter (_matter._tcp) — the new cross-ecosystem smart home standard from Apple/Google/Amazon
Operational port: 5541 (Matter's standard TCP port, secured with DTLS)
Service ID: BE141FEDFD7A67F5-05E6CD268D9BB541 (compressed Fabric ID + Node ID — unique per Matter commissioning)
What this means: This is almost certainly an Amazon Smart Plug (2nd gen or later) with Matter support. It communicates locally over your LAN using the Matter protocol instead of going through Amazon's cloud for local control.
Currently Offline / Sleeping (from ARP cache):
Hostname	IP	Type
amazonplug15jn	192.168.1.151	Amazon Smart Plug
amazonplug13rr	192.168.1.246	Amazon Smart Plug
amazonplug12r9	192.168.1.252	Amazon Smart Plug
hs100	192.168.1.206	TP-Link HS100 Smart Plug
The hs100 at .206 was actively toggling on/off on April 22 (uptime records show it flipping every few minutes) — it's currently offline but was very active then.

How They Communicate
Amazon Smart Plugs (Matter):
Matter uses encrypted TCP connections on port 5541 with DTLS for session security. The device announces itself via mDNS (_matter._tcp.local), and controllers (your phone, Alexa app, HomeKit, etc.) discover it locally. No internet required for local control — commands go directly LAN-to-device.

TP-Link HS100 (legacy):
The HS100 uses TP-Link's older proprietary protocol on TCP port 9999 — unencrypted JSON-over-TCP. It can be controlled completely locally (tools like tplink-smarthome-api speak to it directly). It does NOT use Matter.

Other Smart Home Devices Spotted
rachio-5888ec (192.168.1.152) — Rachio smart sprinkler controller
ring-f34e3a (192.168.1.245) — Ring doorbell/security camera
amazon-4da125b99 (192.168.1.153) — Amazon Echo or similar hub device
The big story here is that your network has both generations of smart plug tech: the old TP-Link proprietary protocol and newer Matter-based Amazon plugs. Would you like to add MAC vendor lookup or device-type labeling to netdash so these show up with proper labels in the dashboard?

Read bbmx2s6p6.output
One extra detail from the background port scan: port 55443 is open on 192.168.1.219 (the Amazon Matter plug). That's Amazon's proprietary Frustration-Free Setup / device registration port — used during initial pairing and by the Alexa app for device management, separate from Matter's port 5541. It's HTTPS-like but with Amazon's own certificate pinning.

So that device exposes two ports: 55443 (Amazon proprietary setup/management) and 5541 (Matter operational, only accepts connections from commissioned controllers). No plain HTTP, no MQTT, no legacy TP-Link protocol — it's a fairly locked-down modern device.