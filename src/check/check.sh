MAX_CYCLO=10
gocyclo -over "$MAX_CYCLO" -ignore '^(.*\/controller\/.*|.*_mock.go)$' .
if [ $? -ne 0 ]; then \
  echo "Cyclomatic complexity checker returned error, most likely due to exceeding max complexity = $MAX_CYCLO.\n"; \
  exit 1; \
fi
echo "Cyclomatic complexity: ok\n"

go run check/check.go
echo "\nHalstead test: ok"

# staticcheck ./... || exit 1
# echo "\staticcheck: ok"
# echo "\Code complexity checks done."
