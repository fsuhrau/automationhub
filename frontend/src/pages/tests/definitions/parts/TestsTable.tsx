import React from 'react';
import {Table, TableBody, TableCell, TableContainer, TableHead, TableRow} from '@mui/material';
import ITestData from '../../../../types/test';
import {AppRoutes} from '../../../../lib/routes';
import {TestRow} from './TestRow';
import {useAgoClock} from '../useAgoClock';

interface TestsTableProps {
    tests: ITestData[],
    historyKnown: boolean,
    appRoutes: AppRoutes,
    canRun: boolean,
    onRun: (test: ITestData) => void,
}

// TestsTable lists the tests, one row each
export const TestsTable: React.FC<TestsTableProps> = ({tests, historyKnown, appRoutes, canRun, onRun}) => {
    const now = useAgoClock();
    return <TableContainer sx={{mx: -2, width: 'auto'}}>
        <Table size="small" aria-label="Tests">
            <TableHead>
                <TableRow>
                    <TableCell sx={{pl: 2}}>Test</TableCell>
                    <TableCell>Last run</TableCell>
                    <TableCell sx={{display: {xs: 'none', md: 'table-cell'}}}>Latest runs</TableCell>
                    <TableCell/>
                </TableRow>
            </TableHead>
            <TableBody>
                {tests.map(t => <TestRow key={t.id} test={t} historyKnown={historyKnown} now={now}
                    testRoutes={appRoutes.test(t.id)} canRun={canRun} onRun={onRun}/>)}
            </TableBody>
        </Table>
    </TableContainer>;
};
