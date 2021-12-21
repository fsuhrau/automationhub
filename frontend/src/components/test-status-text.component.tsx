import { FC } from 'react';
import { TestResultState } from '../types/test.result.state.enum';
import { Typography } from '@mui/material';

export interface TestStatusTextProps {
    status: TestResultState
}

const TestStatusTextComponent: FC<TestStatusTextProps> = (props) => {
    const { status } = props;

    return (
        <div>
            {status == TestResultState.TestResultSuccess && <Typography style={{ color: 'green' }} >Success</Typography>}
            {status == TestResultState.TestResultFailed && <Typography  style={{ color: 'red' }} >Failed</Typography>}
            {status == TestResultState.TestResultOpen && <Typography  style={{ color: 'gray' }} >Running</Typography>}
            {status == TestResultState.TestResultUnstable && <Typography  style={{ color: 'yellow' }} >Unstable</Typography>}
        </div>
    );
};

export default TestStatusTextComponent;
