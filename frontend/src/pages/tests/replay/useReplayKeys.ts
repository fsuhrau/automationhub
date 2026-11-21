import {useEffect} from 'react';
import {ReplayClock} from './useReplayClock';

const SKIP_MS = 5000;
const OWN_KEYS = 'input, textarea, [role="button"], button, [role="slider"]';

// useReplayKeys: space plays and pauses, the arrows skip 5 s; controls that
// handle keys themselves keep them
export const useReplayKeys = ({position, seek, toggle}: Pick<ReplayClock, 'position' | 'seek' | 'toggle'>): void => {
    useEffect(() => {
        const onKey = (e: KeyboardEvent): void => {
            if ((e.target as HTMLElement).closest(OWN_KEYS)) return;
            if (e.key === ' ') {
                e.preventDefault();
                toggle();
            } else if (e.key === 'ArrowLeft') {
                seek(position - SKIP_MS);
            } else if (e.key === 'ArrowRight') {
                seek(position + SKIP_MS);
            }
        };
        window.addEventListener('keydown', onKey);
        return () => window.removeEventListener('keydown', onKey);
    }, [toggle, seek, position]);
};
