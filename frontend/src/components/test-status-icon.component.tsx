import React from 'react';

import { TestResultState } from '../types/test.result.state.enum';
import { ArrowRightRounded, Cancel, CheckCircle, Explicit } from '@mui/icons-material';

export interface TestStatusIconProps {
    status: TestResultState
}

const TestStatusIconComponent: React.FC<TestStatusIconProps> = (props) => {
    const { status } = props;

    return (
        <div>
            {status == TestResultState.TestResultSuccess && <CheckCircle htmlColor={'green'}/>}
            {status == TestResultState.TestResultFailed && <Cancel htmlColor={'red'}/>}
            {status == TestResultState.TestResultOpen && <ArrowRightRounded htmlColor={'gray'}/>}
            {status == TestResultState.TestResultUnstable && <Explicit htmlColor={'yellow'}/>}
        </div>
    );
};

export default TestStatusIconComponent;
