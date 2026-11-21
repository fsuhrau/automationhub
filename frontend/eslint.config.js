import js from '@eslint/js';
import globals from 'globals';
import react from 'eslint-plugin-react';
import reactHooks from 'eslint-plugin-react-hooks';
import tsPlugin from '@typescript-eslint/eslint-plugin';
import tsParser from '@typescript-eslint/parser';
import importPlugin from 'eslint-plugin-import';

// The size rules are errors for all of src; the generated-looking theme files are data tables and stay exempt.
const sizeExempt = ['src/shared-theme/**', 'src/theme/**'];
const testFiles = ['src/**/*.test.{ts,tsx}', 'src/testing/**'];

// The theme files come from a template and still break these rules: warnings there, errors everywhere else.
const themeWarnRules = [
    'react/jsx-indent',
    'react/jsx-indent-props',
    '@typescript-eslint/explicit-function-return-type',
    '@typescript-eslint/no-unused-vars',
    '@typescript-eslint/no-empty-object-type',
    '@typescript-eslint/ban-ts-comment',
];
const themeRules = Object.fromEntries(themeWarnRules.map((rule) => [rule, 'warn']));

const sizeRules = (level) => ({
    'max-lines': [level, { max: 200, skipBlankLines: true, skipComments: true }],
    'max-lines-per-function': [level, { max: 40, skipBlankLines: true, skipComments: true }],
    complexity: [level, 10],
    'max-params': [level, 3],
});

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
    { files: sizeExempt.map((glob) => `${glob}/*.{ts,tsx}`), rules: themeRules },
    { files: ['**/*.{ts,tsx}'], ignores: sizeExempt, rules: sizeRules('error') },
    // Tests group many cases in describe blocks and the kit builds fixtures: only the length limits are off.
    { files: testFiles, rules: { 'max-lines': 'off', 'max-lines-per-function': 'off' } },
];
