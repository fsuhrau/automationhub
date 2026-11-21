import React from 'react';
import {Alert, Button, LinearProgress} from '@mui/material';
import ITestData from '../../../../types/test';
import {AppRoutes} from '../../../../lib/routes';
import {NoTests} from './NoTests';
import {TestsBrowser} from './TestsBrowser';

interface TestsListBodyProps {
    tests: ITestData[] | null,
    error: string | null,
    onRetry: () => void,
    appRoutes: AppRoutes,
    canRun: boolean,
    onRun: (test: ITestData) => void,
}

// TestsListBody is what the list shows: the error, the progress, the empty
// state or the tests
export const TestsListBody: React.FC<TestsListBodyProps> = ({tests, error, onRetry, appRoutes, ...browser}) => {
    if (error) {
        return <Alert severity="error" action={<Button color="inherit" size="small" onClick={onRetry}>Retry</Button>}>{error}</Alert>;
    }
    if (tests === null) return <LinearProgress/>;
    if (tests.length === 0) return <NoTests newTestUrl={appRoutes.newTest()}/>;
    return <TestsBrowser tests={tests} appRoutes={appRoutes} {...browser}/>;
};
