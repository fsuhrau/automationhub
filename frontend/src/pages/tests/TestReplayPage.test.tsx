import {fireEvent, screen} from '@testing-library/react';
import {afterEach, beforeEach, expect, it, vi} from 'vitest';
import {active, advance, click, clockText, loaded, RUN, restoreAll, row, stubMedia} from './replay/renderReplay';

beforeEach(stubMedia);
afterEach(restoreAll);

it('loads the run through the api client and shows header, log and video', async () => {
    const api = await loaded();
    expect(api.called(RUN)).toHaveLength(1);
    expect(screen.getByRole('heading', {level: 1})).toHaveTextContent('Checkout · Run 7 · Replay');
    expect(screen.getByLabelText('Screen of Lab phone')).toBeInTheDocument();
    expect(clockText()).toBe('0:00.0 / 0:10.0');
    expect(['boot', 'crash', 'tap', 'Screenshot'].every(t => screen.queryByText(t))).toBe(true);
    expect(screen.getByText('Test failed: Login (Ns.Cls)')).toBeInTheDocument();
    expect(screen.getByAltText('Screenshot')).toHaveAttribute('src', '/api/data/shot%20one.png');
});

it('links back to the run through the routes', async () => {
    await loaded();
    expect(screen.getByRole('link', {name: /Back to the run/})).toHaveAttribute('href', '/project/p1/app:2/test/3/run/7');
});

it('plays: the timeline advances and the current log entry highlights', async () => {
    await loaded();
    vi.useFakeTimers();
    expect(active()).toBeUndefined();
    click('Play');
    advance(1000);
    expect(clockText()).toBe('0:01.0 / 0:10.0');
    expect(active()).toContain('Test started: Login');
    advance(1500);
    expect(clockText()).toBe('0:02.5 / 0:10.0');
    expect(active()).toContain('tap');
});

it('pauses: the timeline stops', async () => {
    await loaded();
    vi.useFakeTimers();
    click('Play');
    advance(1000);
    click('Pause');
    advance(3000);
    expect(clockText()).toBe('0:01.0 / 0:10.0');
    expect(screen.getByRole('button', {name: 'Play'})).toBeInTheDocument();
});

it('stops at the end of the run', async () => {
    await loaded();
    vi.useFakeTimers();
    click('Play');
    advance(15000);
    expect(clockText()).toBe('0:10.0 / 0:10.0');
    expect(screen.getByRole('button', {name: 'Play'})).toBeInTheDocument();
});

it('plays faster at a higher speed', async () => {
    await loaded();
    fireEvent.mouseDown(screen.getByRole('combobox'));
    fireEvent.click(screen.getByRole('option', {name: '4×'}));
    vi.useFakeTimers();
    click('Play');
    advance(1000);
    expect(clockText()).toBe('0:04.0 / 0:10.0');
});

it('seeks to a log entry by click and by keyboard', async () => {
    await loaded();
    fireEvent.click(row('crash'));
    expect(clockText()).toBe('0:08.0 / 0:10.0');
    expect(active()).toContain('crash');
    fireEvent.keyDown(row('boot'), {key: 'Enter'});
    expect(clockText()).toBe('0:00.5 / 0:10.0');
    fireEvent.keyDown(row('tap'), {key: ' '});
    expect(clockText()).toBe('0:02.0 / 0:10.0');
    fireEvent.keyDown(row('tap'), {key: 'a'});
    expect(clockText()).toBe('0:02.0 / 0:10.0');
});

it('seeks with the slider', async () => {
    await loaded();
    fireEvent.change(screen.getByRole('slider', {name: 'Position in the run'}), {target: {value: 4000}});
    expect(clockText()).toBe('0:04.0 / 0:10.0');
});

it('skips by 5 seconds and to the next test or error', async () => {
    await loaded();
    click('Ahead 5 seconds');
    expect(clockText()).toBe('0:05.0 / 0:10.0');
    click('Back 5 seconds');
    click('Back 5 seconds');
    expect(clockText()).toBe('0:00.0 / 0:10.0');
    click('Next test or error');
    expect(clockText()).toBe('0:01.0 / 0:10.0');
    fireEvent.click(row('crash'));
    expect(screen.getByRole('button', {name: 'Next test or error'})).toBeDisabled();
});

it('is controlled with the keyboard', async () => {
    await loaded();
    fireEvent.keyDown(document.body, {key: 'ArrowRight'});
    expect(clockText()).toBe('0:05.0 / 0:10.0');
    fireEvent.keyDown(document.body, {key: 'ArrowLeft'});
    expect(clockText()).toBe('0:00.0 / 0:10.0');
    fireEvent.keyDown(document.body, {key: ' '});
    expect(screen.getByRole('button', {name: 'Pause'})).toBeInTheDocument();
});

