import React from 'react';

// the form props that save: the submit button and Cmd/Ctrl+S, as in an editor
export const saveHandlers = (save: () => void): {
    onSubmit: (e: React.FormEvent) => void,
    onKeyDown: (e: React.KeyboardEvent) => void,
} => ({
    onSubmit: e => {
        e.preventDefault();
        save();
    },
    onKeyDown: e => {
        if (!(e.metaKey || e.ctrlKey) || e.key !== 's') return;
        e.preventDefault();
        save();
    },
});
