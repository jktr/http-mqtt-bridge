# http-mqtt-bridge

This implements a trivial service that bridges received
http request bodies into an mqtt broker at a topic
derived from the requested URL's path.

While it may possibly work, this is not
production-quality code. The webserver also does not
use authentication or TLS, so you will need to deploy a
reverse proxy.
