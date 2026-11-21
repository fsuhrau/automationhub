import {useEffect, useState} from 'react';

// useDebouncedValue follows `value` once it stayed unchanged for `delayMs`
export const useDebouncedValue = <T,>(value: T, delayMs: number): T => {
    const [settled, setSettled] = useState(value);
    useEffect(() => {
        const timer = setTimeout(() => setSettled(value), delayMs);
        return () => clearTimeout(timer);
    }, [value, delayMs]);
    return settled;
};
