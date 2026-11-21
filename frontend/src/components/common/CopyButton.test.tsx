import {describe, expect, it, vi} from 'vitest';
import {fireEvent, render, screen, waitFor} from '@testing-library/react';
import {CopyButton} from './CopyButton';

const writeText = vi.fn().mockResolvedValue(undefined);
Object.defineProperty(navigator, 'clipboard', {value: {writeText}, configurable: true});

describe('CopyButton', () => {
    it('copies the text and shows the check for a moment', async () => {
        vi.useFakeTimers({shouldAdvanceTime: true});
        render(<CopyButton text="secret" label="Copy token"/>);
        fireEvent.click(screen.getByRole('button', {name: 'Copy token'}));
        expect(writeText).toHaveBeenCalledWith('secret');
        await waitFor(() => expect(screen.getByTestId('CheckRoundedIcon')).toBeInTheDocument());
        await vi.advanceTimersByTimeAsync(1600);
        await waitFor(() => expect(screen.getByTestId('ContentCopyRoundedIcon')).toBeInTheDocument());
        vi.useRealTimers();
    });

    it('shows no check when the copy fails', async () => {
        writeText.mockRejectedValueOnce(new Error('denied'));
        render(<CopyButton text="x" label="Copy"/>);
        fireEvent.click(screen.getByRole('button', {name: 'Copy'}));
        await waitFor(() => expect(writeText).toHaveBeenCalled());
        expect(screen.queryByTestId('CheckRoundedIcon')).toBeNull();
    });
});
