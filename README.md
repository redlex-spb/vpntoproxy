# Router, creating multiple VPN connections
The program allows you to create VPN connections in a separate container with a proxy utility, thereby proxying traffic through the VPN.  
Thus, using the program, you can create many tunnels and proxy certain traffic through the tunnel.
## Features
- Create VPN tunnels;
- Route connections;
- Proxy a separate project / program on a specific tunnel;
- Automatically download ovpn config (_TODO_);
## Requirements
- Go 1.15+ (recent changes have only been tested on 1.15);
- Docker (for create containers);
## Using
You can interact with the project using the [API](https://github.com/redlex-spb/vpntoproxy/wiki/API). UI is in development.
## TODO
- [ ] Create scheduler with automatic vpn / proxy check;
- [ ] Automatically download VPN config;
- [ ] UI;
- [ ] Create different types of proxy connections;
- [ ] Use various VPN providers;
