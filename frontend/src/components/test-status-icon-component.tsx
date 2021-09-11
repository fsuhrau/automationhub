import { FC } from 'react';
import { TestResultState } from '../types/test.result.state.enum';
import { ArrowRightRounded, Cancel, CheckCircle, Explicit } from '@material-ui/icons';

export interface TestStatusIconProps {
    status: TestResultState
}

const TestStatusIconComponent: FC<TestStatusIconProps> = (props) => {
    const { status } = props;

    return (
        <div>
            {status == TestResultState.TestResultSuccess && <CheckCircle htmlColor="green"/>}
            {status == TestResultState.TestResultFailed && <Cancel color="error"/>}
            {status == TestResultState.TestResultOpen && <ArrowRightRounded />}
            {status == TestResultState.TestResultUnstable && <Explicit />}
        </div>
    );
};

export default TestStatusIconComponent;
