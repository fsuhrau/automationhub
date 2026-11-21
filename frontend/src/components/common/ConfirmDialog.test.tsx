import {describe, expect, it, vi} from 'vitest';
import {fireEvent, render, screen} from '@testing-library/react';
import {ConfirmDialog} from './ConfirmDialog';

describe('ConfirmDialog', () => {
    const open = (props = {}): { onConfirm: () => void, onClose: () => void } => {
        const onConfirm = vi.fn();
        const onClose = vi.fn();
        render(<ConfirmDialog open={true} title="Delete?" confirm="Delete" onConfirm={onConfirm} onClose={onClose} {...props}>It is gone</ConfirmDialog>);
        return {onConfirm, onClose};
    };

    it('confirms', () => {
        const {onConfirm, onClose} = open();
        expect(screen.getByText('It is gone')).toBeInTheDocument();
        fireEvent.click(screen.getByRole('button', {name: 'Delete'}));
        expect(onConfirm).toHaveBeenCalled();
        expect(onClose).not.toHaveBeenCalled();
    });

    it('cancels', () => {
        const {onConfirm, onClose} = open();
        fireEvent.click(screen.getByRole('button', {name: 'Cancel'}));
        expect(onClose).toHaveBeenCalled();
        expect(onConfirm).not.toHaveBeenCalled();
    });

    it('disables the confirm button while busy', () => {
        open({busy: true, color: 'warning'});
        expect(screen.getByRole('button', {name: 'Delete'})).toBeDisabled();
    });

    it('renders nothing while closed', () => {
        open({open: false});
        expect(screen.queryByText('Delete?')).toBeNull();
    });
});
