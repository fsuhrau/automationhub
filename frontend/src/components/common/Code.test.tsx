import {describe, expect, it} from 'vitest';
import {render, screen} from '@testing-library/react';
import {Code} from './Code';

describe('Code', () => {
    it('shows the text with a copy button', () => {
        render(<Code label="Copy command">npm start</Code>);
        expect(screen.getByText('npm start')).toBeInTheDocument();
        expect(screen.getByRole('button', {name: 'Copy command'})).toBeInTheDocument();
    });

    it('labels the button "Copy" by default', () => {
        render(<Code>x</Code>);
        expect(screen.getByRole('button', {name: 'Copy'})).toBeInTheDocument();
    });
});
