import { FC } from 'react';
import { TestResultState } from '../types/test.result.state.enum';
import { Typography } from '@material-ui/core';

export interface TestStatusTextProps {
    status: TestResultState
}

const TestStatusTextComponent: FC<TestStatusTextProps> = (props) => {
    const { status } = props;

    return (
        <div>
            {status == TestResultState.TestResultSuccess && <Typography>Success</Typography>}
            {status == TestResultState.TestResultFailed && <Typography color={'error'}>Failed</Typography>}
            {status == TestResultState.TestResultOpen && <Typography>Running</Typography>}
            {status == TestResultState.TestResultUnstable && <Typography>Unstable</Typography>}
        </div>
    );
};

export default TestStatusTextComponent;
