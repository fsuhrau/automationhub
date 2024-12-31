import React from 'react';

import {TestResultState} from '../types/test.result.state.enum';
import {Cancel, CheckCircle, DirectionsRun, Explicit} from '@mui/icons-material';
import {getTestStatusColor} from "../helper/TestStatusHelper";

export interface TestStatusIconProps {
    status: TestResultState
}

const TestStatusIconComponent: React.FC<TestStatusIconProps> = (props) => {
    const {status} = props;

    return (
        <>
            {status == TestResultState.TestResultSuccess && <CheckCircle htmlColor={getTestStatusColor(status)}/>}
            {status == TestResultState.TestResultFailed && <Cancel htmlColor={getTestStatusColor(status)}/>}
            {status == TestResultState.TestResultOpen && <DirectionsRun htmlColor={getTestStatusColor(status)}/>}
            {status == TestResultState.TestResultUnstable && <Explicit htmlColor={getTestStatusColor(status)}/>}
        </>
    );
};

export default TestStatusIconComponent;
