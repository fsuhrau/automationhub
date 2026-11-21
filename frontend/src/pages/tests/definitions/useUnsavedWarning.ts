import {useEffect} from 'react';

// useUnsavedWarning has the browser ask before the page is left while dirty
export const useUnsavedWarning = (dirty: boolean): void => {
    useEffect(() => {
        if (!dirty) return;
        const warn = (e: BeforeUnloadEvent): void => e.preventDefault();
        window.addEventListener('beforeunload', warn);
        return () => window.removeEventListener('beforeunload', warn);
    }, [dirty]);
};
