import {KeyboardEvent, useEffect, useRef} from 'react';
import {keyForEvent} from '../deviceControlModel';
import {DeviceInput} from './deviceInput';

// key presses are sent as text once typing pauses this long
export const TYPE_BATCH_MS = 60;

export interface Typing {
    onScreenKey: (e: KeyboardEvent) => void,
    // send what was typed so far
    flush: () => void,
}

// useTyping sends what is typed while the screen has focus to the device:
// characters as text (batched, so a fast typist sends words), special keys
// by name.
export const useTyping = (input: DeviceInput, hasLease: boolean): Typing => {
    const typed = useRef('');
    const timer = useRef(0);
    useEffect(() => () => window.clearTimeout(timer.current), []);

    const flush = (): void => {
        window.clearTimeout(timer.current);
        const value = typed.current;
        typed.current = '';
        if (value) input.typeText(value);
    };
    const onScreenKey = (e: KeyboardEvent): void => {
        const key = hasLease ? keyForEvent(e.nativeEvent) : null;
        if (!key) return;
        e.preventDefault();
        if ('text' in key) {
            typed.current += key.text;
            window.clearTimeout(timer.current);
            timer.current = window.setTimeout(flush, TYPE_BATCH_MS);
            return;
        }
        flush();
        input.sendKey(key.key);
    };
    return {onScreenKey, flush};
};
