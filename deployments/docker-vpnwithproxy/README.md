# Image «vpnwithproxy»
## Description
## Manual run image
- docker build -t vpnwithproxy .
- docker run -it --cap-add=NET_ADMIN --name vpn -v /${PWD}/vpn/japan.ovpn:/vpn/config.ovpn -p 1080:1080 vpnwithproxy