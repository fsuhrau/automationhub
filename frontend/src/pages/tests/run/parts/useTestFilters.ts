import {useState} from 'react';
import {ResultFilter} from '../../testRunModel';

export interface TestFilters {
    result: ResultFilter,
    setResult: (r: ResultFilter) => void,
    search: string,
    setSearch: (s: string) => void,
    // rows that have children are open while filtering, else when toggled
    isExpanded: (key: number) => boolean,
    toggle: (key: number) => void,
    clear: () => void,
}

// useTestFilters holds what the tests list is narrowed by and what is expanded
export const useTestFilters = (): TestFilters => {
    const [result, setResult] = useState<ResultFilter>('all');
    const [search, setSearch] = useState('');
    const [expanded, setExpanded] = useState<Set<number>>(new Set());
    const filtering = result !== 'all' || search.trim() !== '';
    const toggle = (key: number): void => setExpanded(prev => {
        const next = new Set(prev);
        if (next.has(key)) next.delete(key); else next.add(key);
        return next;
    });
    const clear = (): void => {
        setResult('all');
        setSearch('');
    };
    return {result, setResult, search, setSearch, isExpanded: key => filtering || expanded.has(key), toggle, clear};
};
