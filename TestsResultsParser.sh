#!/bin/bash

GITHUB_STEP_SUMMARY="test-results/unit-tests.md"

echo "-----------" >> $GITHUB_STEP_SUMMARY
echo "-----------" >> $GITHUB_STEP_SUMMARY
echo "" >> $GITHUB_STEP_SUMMARY
echo "## 🧪 Test Results Summary: $(date)<br/>" >> $GITHUB_STEP_SUMMARY
echo "" >> $GITHUB_STEP_SUMMARY

echo "| Result | Test | Time (ms) | Errors |" >> $GITHUB_STEP_SUMMARY
echo "| :--- | :--- | :--- | :--- |" >> $GITHUB_STEP_SUMMARY
jq -rs '
(
  [.[] | select(.Test != null and .Action == "output")] |
  group_by(.Test) |
  map({key: .[0].Test, value: ([.[].Output] | join(""))}) |
  from_entries
) as $outputs |
.[] |
select(.Test != null and (.Action == "pass" or .Action == "fail" or .Action == "skip")) |
"| " +
(if .Action == "fail" then "❌ " elif .Action == "pass" then "✅ " elif .Action == "skip" then "⚠️ " else "? " end) +
" | " + .Test + " | " +
((.Elapsed // 0) * 1000 | round | tostring) + " | " +
(if .Action == "fail" then (($outputs[.Test] // "N/A") | gsub("\n"; " ")) else "" end) +
" | "
' test-results/unit-tests.json | sed 's/_/\\_/g' >> $GITHUB_STEP_SUMMARY

PASSED=$(jq -s '[.[] | select(.Test != null and .Action == "pass")] | length' test-results/unit-tests.json)
FAILED=$(jq -s '[.[] | select(.Test != null and .Action == "fail")] | length' test-results/unit-tests.json)
SKIPPED=$(jq -s '[.[] | select(.Test != null and .Action == "skip")] | length' test-results/unit-tests.json)

echo "" >> $GITHUB_STEP_SUMMARY
echo "-----------" >> $GITHUB_STEP_SUMMARY
echo "" >> $GITHUB_STEP_SUMMARY
echo "## 📊 Test Results Summary:" >> $GITHUB_STEP_SUMMARY
echo "" >> $GITHUB_STEP_SUMMARY
echo " * **✅ Passed:** $PASSED" >> $GITHUB_STEP_SUMMARY
echo " * **❌ Failed:** $FAILED" >> $GITHUB_STEP_SUMMARY
echo " * **⚠️ Skipped:** $SKIPPED" >> $GITHUB_STEP_SUMMARY
echo "" >> $GITHUB_STEP_SUMMARY
echo "-----------" >> $GITHUB_STEP_SUMMARY

if [[ "$FAILED" -gt 0 ]]; then
echo "::error title=Test Failures::❌ $FAILED test(s) failed. See the test report for details."
exit 1
fi
