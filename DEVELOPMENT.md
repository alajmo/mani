# Development

```
# Stand in Example Directory
(cd .. && go build && cd - && ./mani sync)

# Stand in Example Directory
(cd ../../ && make build-and-link && cd - && mani run status --cwd)

# Stand in root
go build && ./mani sync -c example/mani.yaml

# Run specific test
TEST_OPTIONS="-v" TEST_PATTERN="TestInit" make test

# Tests with verbose flag
TEST_OPTIONS="-v" make test

# Update golden files
TEST_OPTIONS="-v" make update-golden
```

