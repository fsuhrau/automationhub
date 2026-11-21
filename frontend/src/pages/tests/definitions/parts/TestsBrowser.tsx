import React, {useMemo, useState} from 'react';
import {Typography} from '@mui/material';
import ITestData from '../../../../types/test';
import {AppRoutes} from '../../../../lib/routes';
import {countHealth, filterTests, HealthFilter} from '../../testsListModel';
import {TestsFilterBar} from './TestsFilterBar';
import {TestsTable} from './TestsTable';

interface TestsBrowserProps {
    tests: ITestData[],
    appRoutes: AppRoutes,
    canRun: boolean,
    onRun: (test: ITestData) => void,
}

const SEARCH_FROM = 5;

// TestsBrowser filters and searches the tests and shows the matches
export const TestsBrowser: React.FC<TestsBrowserProps> = ({tests, ...table}) => {
    const [filter, setFilter] = useState<HealthFilter>('all');
    const [search, setSearch] = useState('');
    const counts = useMemo(() => countHealth(tests), [tests]);
    // older masters don't send the latest runs; then nothing is known about them
    const historyKnown = tests.some(t => t.recentRuns !== undefined);
    const visible = useMemo(() => filterTests(tests, filter, search), [tests, filter, search]);
    const showSearch = tests.length > SEARCH_FROM;
    return <>
        {(historyKnown || showSearch) && <TestsFilterBar showChips={historyKnown} showSearch={showSearch} counts={counts}
            filter={filter} onFilter={setFilter} search={search} onSearch={setSearch}/>}
        {visible.length === 0
            ? <Typography variant="body2" color="text.secondary" sx={{py: 2}}>No test matches.</Typography>
            : <TestsTable tests={visible} historyKnown={historyKnown} {...table}/>}
    </>;
};
