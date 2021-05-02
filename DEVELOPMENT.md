# Development

```
# Stand in Example Directory
(cd .. && go build && cd - && ./mani sync)

# Stand in Example Directory
(cd ../../ && make build-and-link && cd - && mani run status --cwd)

# Stand in root
go build && ./mani sync -c example/mani.yaml
```
