import {useMemo, useState} from 'react';
import {DeviceControlApi} from '../../../types/device.control.api';
import {filterLogLines} from '../deviceControlModel';
import {Runner} from './useGuardedAction';

export const LOG_LINE_CHOICES = [300, 1000, 3000];

export interface LogsState {
    // null until loaded
    logs: string[] | null,
    visible: string[],
    lines: number,
    setLines: (lines: number) => void,
    filter: string,
    setFilter: (filter: string) => void,
    problemsOnly: boolean,
    toggleProblems: () => void,
    load: () => void,
}

// useLogs reads the system log of the device on demand and filters the
// loaded lines.
export const useLogs = (service: DeviceControlApi, run: Runner): LogsState => {
    const [logs, setLogs] = useState<string[] | null>(null);
    const [lines, setLines] = useState(LOG_LINE_CHOICES[0]);
    const [filter, setFilter] = useState('');
    const [problemsOnly, setProblemsOnly] = useState(false);
    const visible = useMemo(() => logs ? filterLogLines(logs, filter, problemsOnly) : [], [logs, filter, problemsOnly]);
    const load = (): void => {
        run(async () => setLogs(await service.logs(lines)));
    };
    return {logs, visible, lines, setLines, filter, setFilter, problemsOnly, toggleProblems: () => setProblemsOnly(v => !v), load};
};
