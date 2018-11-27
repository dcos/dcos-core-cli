.PHONY: plugin
plugin: python
	@python3 scripts/plugin/package_plugin.py

.PHONY: python
python:
	@cd python/lib/dcoscli; \
		make binary
