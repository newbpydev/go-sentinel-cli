# TypeScript Migration Plan for Test Files

This document outlines the step-by-step plan to migrate JavaScript test files to TypeScript in the `web/static/js/test` directory.

## Migration Checklist

### Phase 1: Preparation
- [ ] **Audit Current Test Files**
  - [ ] List all `.js` test files in the directory
  - [ ] Identify corresponding source files for each test
  - [ ] Check for any existing TypeScript configurations

- [ ] **Setup TypeScript Configuration**
  - [ ] Ensure `tsconfig.json` exists and is properly configured
  - [ ] Verify test-related TypeScript types are installed (`@types/jest`, `@types/node`, etc.)
  - [ ] Update build/test scripts in `package.json` if needed

### Phase 2: File-by-File Migration

1. **coverage.test.ts** (Partially converted)
   - [x] File already exists in TypeScript
   - [ ] Verify all tests pass
   - [ ] Remove `.js` version if it exists

2. **settings.test.ts**
   - [ ] Create `settings.test.ts`
   - [ ] Convert test cases from JS to TS
   - [ ] Add proper type annotations
   - [ ] Verify tests pass
   - [ ] Delete `settings.test.js`

3. **websocket.test.js**
   - [ ] Create `websocket.test.ts`
   - [ ] Convert test cases from JS to TS
   - [ ] Add WebSocket type definitions
   - [ ] Verify tests pass
   - [ ] Delete `websocket.test.js`

4. **setup.ts**
   - [ ] Create `setup.ts`
   - [ ] Convert setup code from JS to TS
   - [ ] Add proper type definitions for test environment
   - [ ] Verify it works with test files
   - [ ] Delete `setup.js`

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
