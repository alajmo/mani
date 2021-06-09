# Development

```
# Stand in Example Directory
(cd .. && go build && cd - && ./mani sync)

# Stand in Example Directory
(cd ../../ && make build-and-link && cd - && mani run status --cwd)

# Stand in root
go build && ./mani sync -c example/mani.yaml

# Run specific test with verbose flag
TEST_PATTERN="TestInit" TEST_OPTIONS="-v" make test

# Tests with verbose flag
TEST_OPTIONS="-v" make test

# Update all golden files
TEST_OPTIONS="-v" make update-golden

# Update specific golden file
TEST_PATTERN="TestInit" TEST_OPTIONS="-v" make update-golden

# Debug single test
TEST_PATTERN="TestInfoCmd" TEST_OPTIONS="-v" TEST_FILES="./test/integration/info_test.go" make debug-test
TEST_PATTERN="TestInfoCmd" TEST_OPTIONS="-v" TEST_FILES="./test/integration/info_test.go" make debug-test-update

# Debug completion
mani __complete list tags --projects ""
```
