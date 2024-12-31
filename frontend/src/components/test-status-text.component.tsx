import React from 'react';

import {TestResultState} from '../types/test.result.state.enum';
import {Typography} from '@mui/material';
import {getTestStatusColor, getTestStatusText} from "../helper/TestStatusHelper";

export interface TestStatusTextProps {
    status: TestResultState
}

const TestStatusTextComponent: React.FC<TestStatusTextProps> = (props) => {
    const {status} = props;

    return (
        <Typography style={{color: getTestStatusColor(status)}}>{getTestStatusText(status)}</Typography>
    );
};

export default TestStatusTextComponent;
