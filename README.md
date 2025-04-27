# Graviola

Graviola is a "Prometheus-like" API. It should be used for cases that you have several Prometheus instances and want a single point of entry for queries through API.

Docs are intentionally small because I don't consider it a "ready to use" ptoject. I mean, it would work, but it doesn't have any feature that is different than the main existing projects that do the same thing. As such you should not use it.

If you really want to use it, read the [config_example.yml](https://github.com/jademcosta/graviola/blob/main/config_example.yml) file. It has comments to explain, and you might be able to understand how it works.

## Concepts

Pending. I'll write this when Graviola has some new and important features.

## Roadmap

This is the "mostly" prioritized roadmap. I may change it as I will, though:

* More meaningful metrics. Right now, it doesn't have many metrics and this mean a less than ideal monitoring experience, which is one of the main shortcomings of other similar tools.
* "Warnings" returned by all remotes are not being returned on Graviola. This might hide some bug in a remote.
* Configurable timeouts for each remote (in a cascading style).
* Compressed responses.
* Allow to define time-windows to different remotes.
* Allow to define default labels in a remote, both to add on every response and to allow to not even hit the remote with a query if it is possible to determine it doesn't have the data.
* Allow to define API-KEYs to access it.
* Allow to configure SSO access.
* Add tracing!
* Allow a label to be "denylisted" from a remote, meaning it will be removed from the response.




## Bugs and unimplemented features

These are shortcomings that will be implemented in next releases, but right now can't be used.

- Exemplars - No support at all
- Native histograms - no support at all
- Using query filters when querying for label values on API (this affects only the `/labels/values` and `labels/names` endpoints)
- It doesn't have an UI. API access is the only possible way to access it.
