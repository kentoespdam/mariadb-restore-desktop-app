// Vitest setup: make sure React's development build is loaded
// (React 19 strips `act` from the production build, which breaks
// @testing-library/react@16). Vitest already sets NODE_ENV=test,
// but we set it here as well so the import-graph picks the dev
// build deterministically.
// https://github.com/testing-library/react-testing-library/issues/1392
process.env.NODE_ENV ??= 'test';
(globalThis as { IS_REACT_ACT_ENVIRONMENT?: boolean }).IS_REACT_ACT_ENVIRONMENT = true;
