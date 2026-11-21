import js from '@eslint/js';
import globals from 'globals';
import react from 'eslint-plugin-react';
import reactHooks from 'eslint-plugin-react-hooks';
import tsPlugin from '@typescript-eslint/eslint-plugin';
import tsParser from '@typescript-eslint/parser';
import importPlugin from 'eslint-plugin-import';

// Globs of the folders and files that already meet the size limits.
// Inside them the size rules are errors; everywhere else they only warn.
// Each package adds its folders here: append a glob such as 'src/api/**'.
const refactored = [
    'src/setupTests.ts',
    'src/services/device.video.test.ts',
    // F2 shared kit
    'src/lib/**',
    'src/components/common/**',
    'src/hooks/useAsyncAction*',
    'src/hooks/useNow*',
    'src/hooks/useSseLog*',
    // F3 state and API core
    'src/api/**',
    'src/application/HubState*',
    'src/hooks/HubStateProvider*',
    'src/types/**',
];

// Generated-looking theme files are data tables and stay exempt.
const sizeExempt = ['src/shared-theme/**', 'src/theme/**'];

const sizeRules = (level) => ({
    'max-lines': [level, { max: 200, skipBlankLines: true, skipComments: true }],
    'max-lines-per-function': [level, { max: 40, skipBlankLines: true, skipComments: true }],
    complexity: [level, 10],
    'max-params': [level, 3],
});

// Rules the legacy code still violates: warnings there, errors in `refactored`.
const legacyWarnRules = [
    'react/jsx-indent',
    'react/jsx-indent-props',
    'react-hooks/exhaustive-deps',
    'react-hooks/set-state-in-effect',
    'react-hooks/purity',
    'react-hooks/refs',
    'react-hooks/use-memo',
    '@typescript-eslint/explicit-function-return-type',
    '@typescript-eslint/no-unused-vars',
    '@typescript-eslint/no-explicit-any',
    '@typescript-eslint/no-empty-object-type',
    '@typescript-eslint/no-wrapper-object-types',
    '@typescript-eslint/ban-ts-comment',
    'prefer-const',
];
const legacyRules = Object.fromEntries(legacyWarnRules.map((rule) => [rule, 'warn']));

export default [
    { ignores: ['node_modules/**', 'build/**', 'dist/**', 'coverage/**', 'public/**', '**/*.js'] },
    js.configs.recommended,
    ...tsPlugin.configs['flat/recommended'],
    react.configs.flat['jsx-runtime'],
    reactHooks.configs.flat.recommended,
    {
        files: ['**/*.{ts,tsx}'],
        languageOptions: {
            parser: tsParser,
            parserOptions: { project: ['./tsconfig.json'], ecmaFeatures: { jsx: true } },
            globals: { ...globals.browser },
        },
        plugins: { react, import: importPlugin },
        settings: { react: { version: 'detect' } },
        rules: {
            'no-prototype-builtins': 'off',
            'import/prefer-default-export': 'off',
            '@typescript-eslint/no-shadow': 'warn',
            'max-len': ['warn', { code: 160 }],
            '@typescript-eslint/no-unused-expressions': ['error', { allowShortCircuit: true, allowTernary: true }],
            '@typescript-eslint/explicit-function-return-type': ['error', { allowExpressions: true, allowTypedFunctionExpressions: true }],
            'react/no-direct-mutation-state': 'error',
            'react/no-access-state-in-setstate': 'warn',
            'react/prefer-stateless-function': 'warn',
            'react/no-array-index-key': 'warn',
            'react/no-did-update-set-state': 'warn',
            'react/jsx-boolean-value': ['warn', 'always'],
            'react/jsx-indent': ['error', 4],
            'react/jsx-indent-props': ['error', 4],
            'react-hooks/rules-of-hooks': 'error',
            'react-hooks/exhaustive-deps': 'error',
        },
    },
    { files: ['**/*.{ts,tsx}'], ignores: refactored, rules: legacyRules },
    { files: ['**/*.{ts,tsx}'], ignores: sizeExempt, rules: sizeRules('warn') },
    { files: refactored, ignores: sizeExempt, rules: sizeRules('error') },
];
