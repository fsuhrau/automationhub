// jest-dom adds DOM matchers to vitest's expect, e.g.
// expect(element).toHaveTextContent(/react/i)
import '@testing-library/jest-dom/vitest';

import {afterEach} from 'vitest';
import {cleanup} from '@testing-library/react';

// vitest runs without globals, so Testing Library does not unmount by itself
afterEach(() => cleanup());
