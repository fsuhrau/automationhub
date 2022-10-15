import React from 'react';

import { TestResultState } from '../types/test.result.state.enum';
import { Cancel, CheckCircle, DirectionsRun, Explicit } from '@mui/icons-material';

export interface TestStatusIconProps {
    status: TestResultState
}

const TestStatusIconComponent: React.FC<TestStatusIconProps> = (props) => {
    const { status } = props;

    return (
        <>
            {status == TestResultState.TestResultSuccess && <CheckCircle htmlColor={'green'}/>}
            {status == TestResultState.TestResultFailed && <Cancel htmlColor={'red'}/>}
            {status == TestResultState.TestResultOpen && <DirectionsRun htmlColor={'gray'}/>}
            {status == TestResultState.TestResultUnstable && <Explicit htmlColor={'yellow'}/>}
        </>
    );
};

export default TestStatusIconComponent;
