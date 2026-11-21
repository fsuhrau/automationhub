import {describe, expect, it} from 'vitest';
import {render, screen} from '@testing-library/react';
import {Section} from './Section';

describe('Section', () => {
    it('shows title, description, action and content', () => {
        render(<Section title="Tokens" description="About" action={<button>Add</button>}>body</Section>);
        expect(screen.getByRole('heading', {name: 'Tokens'})).toBeInTheDocument();
        expect(screen.getByText('About')).toBeInTheDocument();
        expect(screen.getByRole('button', {name: 'Add'})).toBeInTheDocument();
        expect(screen.getByText('body')).toBeInTheDocument();
    });

    it('works without description and action, and can fill the cell', () => {
        render(<Section title="Plain" fullHeight={true}>x</Section>);
        expect(screen.queryByRole('button')).toBeNull();
        expect(screen.getByRole('heading', {name: 'Plain'})).toBeInTheDocument();
    });
});
