# Metrics

This module enables prometheus metrics for Caddy.

## Use

In your `Caddyfile`:

~~~
prometheus
~~~

For each virtual host that you want to see metrics for.

It optionally takes an address where the metrics are exported, the default
is `localhost:9180`. The metrics path is fixed to `/metrics`.

With `caddyext` you'll need to put this module early in the chain, so that
the duration histogram actually makes sense. I've put at number 0.
