## Brief
A simple exporter from [gravitee-tcp-reporter](https://github.com/gravitee-io/gravitee-reporter-tcp) to prometheus metrics.

Tested with:
* gravitee-apim-gateway 3.5
* gravitee-tcp-reporter 1.0.0-1.1.3  

## Getting start
1. Check logs of your gateway  
   ```
   [graviteeio-node] [] INFO  i.g.p.c.internal.PluginRegistryImpl -   > reporter-tcp [1.1.3] has been loaded
   ```
   Tcp-reporter-plugin must be loaded after start. Otherwise you need to install it.
2. Enable it in gravitee.yml:
   ```yml
   reporters:
      tcp:
        enabled: true
        host: localhost
        port: 8123
   ```
3. Start exporter and check metrics:  
   ```
   docker pull androndo/gravitee-gateway-tcp2prom
   docker run --rm --network=host --name tcp2prom androndo/gravitee-gateway-tcp2prom
   curl localhost:8080/metrics
   ```

## Options
Just ispect Dockerfile for ENV.
