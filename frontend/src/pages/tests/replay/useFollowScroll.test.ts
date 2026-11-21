import {expect, it, vi} from 'vitest';
import {renderHook} from '@testing-library/react';
import {useFollowScroll} from './useFollowScroll';

const elements = (rowTop: number): { box: HTMLElement, row: HTMLElement } => {
    const box = {offsetTop: 10, scrollTop: 0, clientHeight: 300, scrollTo: vi.fn()} as unknown as HTMLElement;
    const row = {offsetTop: 10 + rowTop, clientHeight: 20} as unknown as HTMLElement;
    return {box, row};
};

type Scroll = ReturnType<typeof vi.fn>;

const run = (rowTop: number, state: { follow: boolean, playing: boolean }, hasRow = true): Scroll => {
    const {box, row} = elements(rowTop);
    renderHook(() => useFollowScroll({current: box}, {current: hasRow ? row : null}, {...state, activeKey: 'k'}));
    return box.scrollTo as unknown as Scroll;
};

it('useFollowScroll: scrolls a hidden row into view, smoothly while playing', () => {
    expect(run(900, {follow: true, playing: true})).toHaveBeenCalledWith({top: 800, behavior: 'smooth'});
    expect(run(900, {follow: true, playing: false})).toHaveBeenCalledWith({top: 800, behavior: 'auto'});
});

it('useFollowScroll: leaves a visible row alone', () => {
    expect(run(100, {follow: true, playing: false})).not.toHaveBeenCalled();
});

it('useFollowScroll: does nothing when not following or without a row', () => {
    expect(run(900, {follow: false, playing: false})).not.toHaveBeenCalled();
    expect(run(900, {follow: true, playing: false}, false)).not.toHaveBeenCalled();
});

it('useFollowScroll: does nothing without the list', () => {
    const row = {offsetTop: 0, clientHeight: 1} as unknown as HTMLElement;
    expect(() => renderHook(() => useFollowScroll({current: null}, {current: row}, {follow: true, playing: false, activeKey: null}))).not.toThrow();
});
