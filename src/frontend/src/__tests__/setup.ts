// Vitest setup: make sure React's development build is loaded
// (React 19 strips `act` from the production build, which breaks
// @testing-library/react@16). Vitest already sets NODE_ENV=test,
// but we set it here as well so the import-graph picks the dev
// build deterministically.
// https://github.com/testing-library/react-testing-library/issues/1392
process.env.NODE_ENV ??= 'test';
(globalThis as { IS_REACT_ACT_ENVIRONMENT?: boolean }).IS_REACT_ACT_ENVIRONMENT = true;

// jest-dom matchers (toBeInTheDocument, toHaveTextContent, ...).
// ponytail: register once here so every test file gets them.
import '@testing-library/jest-dom/vitest';

// ponytail: explicit RTL cleanup between tests so dialogs from a
// previous render don't leak into the next (vitest's auto-cleanup
// fires after the file, not after each test).
import { cleanup } from '@testing-library/react';
import { afterEach } from 'vitest';

afterEach(() => {
  cleanup();
});
