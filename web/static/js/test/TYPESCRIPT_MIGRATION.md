# TypeScript Migration Plan for Test Files

This document outlines the step-by-step plan to migrate JavaScript test files to TypeScript in the `web/static/js/test` directory.

## Migration Checklist

### Phase 1: Preparation
- [x] **Audit Current Test Files**
  - [x] List all `.js` test files in the directory
  - [x] Identify corresponding source files for each test
  - [x] Check for any existing TypeScript configurations

- [x] **Setup TypeScript Configuration**
  - [x] Ensure `tsconfig.json` exists and is properly configured
  - [x] Verify test-related TypeScript types are installed (`@types/jest`, `@types/node`, etc.)
  - [x] Update build/test scripts in `package.json` if needed

### Phase 2: File-by-File Migration

1. **coverage.test.ts**
   - [x] File exists in TypeScript
   - [x] All tests pass
   - [x] No `.js` version exists

2. **settings.test.ts**
   - [x] File exists in TypeScript
   - [x] All tests pass
   - [x] No `.js` version exists

3. **websocket.test.ts**
   - [x] File exists in TypeScript
   - [x] All tests pass
   - [x] No `.js` version exists

4. **setup.ts**
   - [x] File exists in TypeScript
   - [x] All tests pass
   - [x] Removed duplicate `setup.js`

5. **example.test.ts**
   - [x] Converted from `example.test.js` to TypeScript
   - [x] All tests pass
   - [x] Removed original `example.test.js`

6. **utils/example.test.ts**
   - [x] File exists in TypeScript
   - [x] All tests pass
   - [x] No `.js` version exists

7. **utils/websocket.test.ts**
   - [x] File exists in TypeScript
   - [x] All tests pass
   - [x] No `.js` version exists

8. **main.test.ts**
   - [x] File exists in TypeScript
   - [x] All tests pass
   - [x] No `.js` version exists

### Phase 3: Verification & Cleanup
- [x] **Run All Tests**
  - [x] All 78 tests pass
  - [x] No TypeScript errors
  - [x] No linting errors

- [x] **Cleanup**
  - [x] Removed all `.js` test files that have been converted to TypeScript
  - [x] Verified no duplicate test files exist
  - [x] Updated documentation

### Phase 3: Configuration & Cleanup
- [ ] **Update Test Scripts**
  - [ ] Modify `package.json` test scripts to use `.ts` files
  - [ ] Update any test-related configurations

- [ ] **Type Definitions**
  - [ ] Create/update `global.d.ts` for any missing type definitions
  - [ ] Ensure all test utilities are properly typed

- [ ] **Final Verification**
  - [ ] Run entire test suite
  - [ ] Verify all tests pass
  - [ ] Check for any TypeScript errors

- [ ] **Cleanup**
  - [ ] Remove any remaining `.js` test files
  - [ ] Update documentation if needed

## Implementation Strategy

For each file migration, follow these steps:

1. **Preparation**
   ```bash
   # Create TypeScript version
   cp path/to/test.js path/to/test.ts
   ```

2. **Conversion Steps**
   - Add TypeScript types to all variables and functions
   - Convert `require()` to `import`
   - Add proper type imports
   - Update test assertions to use TypeScript types
   - Fix any type errors

3. **Testing**
   ```bash
   # Run specific test file
   npm test test/filename.test.ts
   ```

4. **Cleanup**
   ```bash
   # After confirming tests pass
   rm path/to/test.js
   ```

## Risk Mitigation

1. **Version Control**
   - Each file conversion should be a separate commit
   - Commit message format: `test: convert [filename].js to TypeScript`

2. **Rollback Plan**
   - Keep both `.js` and `.ts` files until all tests pass
   - Use `git checkout -- path/to/file.js` to revert if needed

3. **Testing Strategy**
   - Run tests after each file conversion
   - Verify both individual tests and full test suite
