import {useMemo, useState} from 'react';
import IProtocolEntryData from '../../../types/protocol.entry';
import {filterLog, LogEntry} from './logModel';

const PAGE = 500;

export interface LogView {
    hidden: Set<string>,
    toggleHidden: (key: string) => void,
    errorsOnly: boolean,
    setErrorsOnly: (on: boolean) => void,
    search: string,
    setSearch: (text: string) => void,
    // how many entries are shown, and more of them
    limit: number,
    showMore: () => void,
    // the entry "show in log" jumped to
    focus: number | null,
    visible: LogEntry[],
    resetFilters: () => void,
    // clears the filters and makes the entry visible and focused
    showEntry: (entry: IProtocolEntryData) => void,
}

// useLogView holds the filters of the log tab and "show in log" from the overview
export const useLogView = (log: LogEntry[]): LogView => {
    const [hidden, setHidden] = useState<Set<string>>(new Set());
    const [errorsOnly, setErrorsOnly] = useState(false);
    const [search, setSearch] = useState('');
    const [limit, setLimit] = useState(PAGE);
    const [focus, setFocus] = useState<number | null>(null);
    const visible = useMemo(() => filterLog(log, {hidden, errorsOnly, text: search}), [log, hidden, errorsOnly, search]);
    const resetFilters = (): void => {
        setHidden(new Set());
        setErrorsOnly(false);
        setSearch('');
    };
    const toggleHidden = (key: string): void => setHidden(prev => {
        const next = new Set(prev);
        if (next.has(key)) next.delete(key); else next.add(key);
        return next;
    });
    const showEntry = (entry: IProtocolEntryData): void => {
        resetFilters();
        setLimit(l => Math.max(l, log.indexOf(entry) + 50));
        setFocus(entry.id as number);
    };
    return {
        hidden, toggleHidden, errorsOnly, setErrorsOnly, search, setSearch, limit, focus, visible, resetFilters, showEntry,
        showMore: () => setLimit(l => l + PAGE),
    };
};
