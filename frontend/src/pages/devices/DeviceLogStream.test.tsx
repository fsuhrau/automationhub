import {describe, expect, it} from 'vitest';
import {act, fireEvent, render, screen} from '@testing-library/react';
import DeviceLogStream from '../../components/DeviceLogStream';
import {fakeSources} from '../../testing/fakeSse';

const INFO = '11-06 13:51:05.123 1234 5678 I Unity: all fine';
const WARN = '11-06 13:51:06.123 1234 5678 W Unity: careful';
const FAIL = '11-06 13:51:07.123 1234 5678 E Unity: broken';

const open = (maxHeight?: number): ReturnType<typeof fakeSources> => {
    const log = fakeSources();
    render(<DeviceLogStream projectId="p1" deviceId={3} createSource={log.create} maxHeight={maxHeight}/>);
    return log;
};

const emit = (log: ReturnType<typeof fakeSources>, type: string, data?: unknown): void => {
    act(() => log.sources[0].emit(type, data));
};

describe('DeviceLogStream', () => {
    it('opens the log stream of the device and waits for lines', () => {
        const log = open();
        expect(log.sources[0].url).toBe('/api/p1/device/3/logs/stream?lines=300');
        expect(screen.getByText('Connecting…')).toBeInTheDocument();
        expect(screen.getByText('Waiting for log lines…')).toBeInTheDocument();
    });

    it('shows the lines the stream sends and goes live', () => {
        const log = open(200);
        emit(log, 'lines', [INFO, FAIL]);
        expect(screen.getByText(INFO)).toBeInTheDocument();
        expect(screen.getByText(FAIL)).toBeInTheDocument();
        expect(screen.getByText('Live')).toBeInTheDocument();
    });

    it('shows the failure the device reports', () => {
        const log = open();
        emit(log, 'failure', {message: 'adb is gone'});
        expect(screen.getByText('adb is gone')).toBeInTheDocument();
        expect(screen.getByText('No log')).toBeInTheDocument();
        expect(screen.getByText('The device log could not be read.')).toBeInTheDocument();
    });

});

describe('DeviceLogStream controls', () => {
    it('filters by text and to problems only', () => {
        const log = open();
        emit(log, 'lines', [INFO, WARN, FAIL]);
        fireEvent.change(screen.getByLabelText('Filter the log'), {target: {value: 'careful'}});
        expect(screen.queryByText(INFO)).toBeNull();
        expect(screen.getByText(WARN)).toBeInTheDocument();
        fireEvent.change(screen.getByLabelText('Filter the log'), {target: {value: 'nothing like it'}});
        expect(screen.getByText('No line matches.')).toBeInTheDocument();
        fireEvent.change(screen.getByLabelText('Filter the log'), {target: {value: ''}});
        fireEvent.click(screen.getByRole('button', {name: 'Problems only'}));
        expect(screen.queryByText(INFO)).toBeNull();
        expect(screen.getByText(FAIL)).toBeInTheDocument();
    });

});

describe('DeviceLogStream pausing', () => {
    it('holds the lines while paused and shows them again on resume', () => {
        const log = open();
        emit(log, 'lines', [INFO]);
        fireEvent.click(screen.getByRole('button', {name: 'Pause the log'}));
        emit(log, 'lines', [FAIL]);
        expect(screen.queryByText(FAIL)).toBeNull();
        expect(screen.getByText('Paused · 1 new')).toBeInTheDocument();
        fireEvent.click(screen.getByRole('button', {name: 'Resume the log'}));
        expect(screen.getByText(FAIL)).toBeInTheDocument();
    });

    it('clears the shown lines, also while paused', () => {
        const log = open();
        emit(log, 'lines', [INFO]);
        fireEvent.click(screen.getByRole('button', {name: 'Clear the log'}));
        expect(screen.queryByText(INFO)).toBeNull();
        emit(log, 'lines', [WARN]);
        fireEvent.click(screen.getByRole('button', {name: 'Pause the log'}));
        fireEvent.click(screen.getByRole('button', {name: 'Clear the log'}));
        expect(screen.getByText('Paused · 0 new')).toBeInTheDocument();
    });

    it('stays at the end unless scrolled up', () => {
        const log = open();
        const box = screen.getByRole('log');
        Object.defineProperty(box, 'scrollHeight', {value: 500, configurable: true});
        emit(log, 'lines', [INFO]);
        expect(box.scrollTop).toBe(500);
        Object.defineProperty(box, 'clientHeight', {value: 100, configurable: true});
        box.scrollTop = 0;
        fireEvent.scroll(box);
        emit(log, 'lines', [WARN]);
        expect(box.scrollTop).toBe(0);
    });
});
