import {MutableRefObject, useEffect, useRef} from 'react';

// useLatest keeps the value of the last render in a ref, for callbacks of
// long-lived listeners (timers, animation frames, sockets) that must not
// capture an old one.
export const useLatest = <T, >(value: T): MutableRefObject<T> => {
    const ref = useRef(value);
    useEffect(() => {
        ref.current = value;
    });
    return ref;
};
