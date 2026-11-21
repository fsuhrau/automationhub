import {render, screen} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import {describe, expect, it, vi} from 'vitest';
import PageLoader from './PageLoader';

describe('PageLoader', () => {
    it('spins while loading', () => {
        render(<PageLoader/>);
        expect(screen.getByRole('progressbar', {hidden: true})).toBeInTheDocument();
    });

    it('shows the server message of an error with a retry button', async () => {
        const retry = vi.fn();
        render(<PageLoader error={{response: {data: {message: 'no such page'}}}} onRetry={retry}/>);
        expect(screen.getByText('no such page')).toBeInTheDocument();
        await userEvent.click(screen.getByRole('button', {name: 'Retry'}));
        expect(retry).toHaveBeenCalledTimes(1);
    });

    it('shows an error without a retry button when there is nothing to retry', () => {
        render(<PageLoader error={new Error('broken')}/>);
        expect(screen.getByText('broken')).toBeInTheDocument();
        expect(screen.queryByRole('button')).toBeNull();
    });
});
