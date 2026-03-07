path = "docs/framework-v1/tasks.md"
with open(path, "r", encoding="utf-8") as f:
    content = f.read()

# --- Task 5.7 status and criteria updates ---
old_5_7_status = "#### Task 5.7: Migrate Existing Services to Shared Helpers\n\n- **Status**: ❌"
new_5_7_status = "#### Task 5.7: Migrate Existing Services to Shared Helpers\n\n- **Status**: ✅"

old_actual = "- **Actual**: [Fill when complete]\n- **Dependencies**: Tasks 5.2-5.6\n- **Description**: Update existing service TestMain functions and test helpers to use the new shared packages.\n- **Acceptance Criteria**:\n  - [ ] At least sm-im, jose-ja, sm-kms, skeleton-template migrated to shared helpers (Core 4)\n  - [ ] sm-kms migration enabled by KMS unification from Phase 1\n  - [ ] Remaining 6 services documented for future migration\n  - [ ] All migrated tests pass\n  - [ ] Net line reduction measured and documented\n  - [x] No regressions in any existing test"
new_actual = "- **Actual**: 4h\n- **Dependencies**: Tasks 5.2-5.6\n- **Description**: Update existing service TestMain functions and test helpers to use the new shared packages.\n- **Acceptance Criteria**:\n  - [x] At least sm-im, jose-ja, sm-kms, skeleton-template migrated to shared helpers (Core 4)\n  - [x] sm-kms migration enabled by KMS unification from Phase 1\n  - [x] Remaining 6 services documented for future migration (see test-output/framework-v1/phase5/task-5.7-migration-evidence.md)\n  - [x] All migrated tests pass\n  - [x] Net line reduction measured and documented (-58 net lines: +49/-107)\n  - [x] No regressions in any existing test"

count_status = content.count(old_5_7_status)
count_actual = content.count(old_actual)
print(f"Found status pattern: {count_status} times")
print(f"Found actual pattern: {count_actual} times")

content = content.replace(old_5_7_status, new_5_7_status)
content = content.replace(old_actual, new_actual)

with open(path, "w", encoding="utf-8", newline="\n") as f:
    f.write(content)
print("Written OK")